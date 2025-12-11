package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"mppj"
	"mppj/api"
	"mppj/api/pb"
	"mppj/cmd/common"
	"mppj/cmd/config"
	"os"
	"runtime"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var sources mppj.SourceList

var (
	nodeID     = flag.String("id", "", "the id of the source")
	helperAddr = flag.String("helper_address", fmt.Sprintf(":%d", config.DEFAULT_PORT), "the address of the helper node")
)

func init() {
	flag.Var((*mppj.SourceList)(&sources), "sources", "the sources' ids as a comma-separated list")

	log.SetFlags(log.Flags() &^ log.Ldate)
	log.SetPrefix("> ")
}

func main() {

	flag.Parse()

	if len(sources) < 2 {
		log.Fatal("at least two sources ids must be provided")
	}

	log.Printf("MPPJ Receiver %s", *nodeID)

	// opens a helper stream
	statsHandler := api.NewStatsHandler()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials())) // no TLS for now
	opts = append(opts, grpc.WithStatsHandler(statsHandler))
	helperConn, err := grpc.NewClient(*helperAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to helper: %v", err)
	}
	defer helperConn.Close()
	helperClient := pb.NewMPPJHelperClient(helperConn)

	rsk, rpk := mppj.GetTestKeys(config.SessionID) // in real usage, keys would be randomly generated and

	r := mppj.NewReceiverWithKeys(config.SessionID, sources, rsk, rpk)

	var start, startActive time.Time
	start = time.Now() // measured time from helper connect

	stream, err := helperClient.PullRows(context.Background(), &pb.Void{})
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}

	md, err := stream.Header()
	if err != nil {
		log.Fatalf("Failed to get stream header: %v", err)
	}
	numRowsStrs := md.Get("num_rows")
	if len(numRowsStrs) == 0 {
		log.Fatalf("No num_rows header in stream")
	}
	var numRows int
	_, err = fmt.Sscanf(numRowsStrs[0], "%d", &numRows)
	if err != nil {
		log.Fatalf("Failed to parse num_rows header: %v", err)
	}
	log.Printf("expecting %d rows from helper", numRows)

	inRowApi := make(chan *pb.EncRowWithHint, numRows)
	inRows := make(chan mppj.EncRowWithHint, numRows)

	go func() {
		rc := 0
		for {
			rowMsg, err := stream.Recv()
			if rc == 0 {
				log.Println("started receiving rows from the helper")
				startActive = time.Now()
			}
			if err == io.EOF {
				log.Printf("dowloaded %d rows", rc)
				if rc < numRows {
					err = fmt.Errorf("expected %d rows but got only %d", numRows, rc)
				} else {
					break
				}
			}
			if err != nil {
				log.Fatalf("Failed to receive row: %v", err)
			}

			rc++

			inRowApi <- rowMsg

			if rc >= numRows {
				log.Printf("all %d rows received", rc)
				err := stream.CloseSend()
				helperConn.Close()
				if err != nil {
					log.Fatalf("Failed to close send: %v", err)
				}
				break
			}
		}
		close(inRowApi)
	}()

	wg := sync.WaitGroup{}
	once := sync.Once{}
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			for inRowMsg := range inRowApi {
				inRow, err := api.GetEncRowWithHintFromMsg(inRowMsg)
				if err != nil {
					log.Fatalf("Failed to convert incoming row: %v", err)
				}
				inRows <- inRow
			}
			wg.Done()
			once.Do(func() { wg.Wait(); close(inRows) })
		}()
	}

	res, err := r.JoinTablesStream(inRows, len(sources))
	if err != nil {
		log.Fatalf("Failed to join tables: %v", err)
	}

	log.Printf("Result has %d rows", res.Len())
	common.PrintStats(statsHandler.GetStats(), time.Since(start), time.Since(startActive))

	w := csv.NewWriter(os.Stdout)
	if err := res.WriteTo(w); err != nil {
		log.Fatalf("Failed to write CSV: %v", err)
	}
}
