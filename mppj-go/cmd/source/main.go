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
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_VAL_LEN = 30

var (
	nodeID     = flag.String("id", "", "the id of the source")
	helperAddr = flag.String("helper_address", fmt.Sprintf(":%d", config.DEFAULT_PORT), "the address of the helper node")
	input      = flag.String("input", "stdin", "the input CSV file (or 'stdin' for standard input)")
	nCPU       = flag.Int("n_cpu", 0, "number of CPUs to use (default is all available CPUs)")
)

func init() {
	log.SetFlags(log.Flags() &^ log.Ldate)
	log.SetPrefix("> ")
}

func main() {

	flag.Parse()

	if *nodeID == "" {
		log.Fatal("Source ID is required")
	}

	if *nCPU < 0 {
		*nCPU = runtime.NumCPU()
	}

	log.Println("MPPJ Source", *nodeID)

	// reads csv from stdin
	table, err := readFile(*input)
	if err != nil {
		log.Fatalf("Failed to read input file: %v", err)
	}

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

	rpk := common.GetRPK(config.SessionID)
	ds := mppj.NewDataSource(config.SessionID, rpk)

	start := time.Now()

	ctx := mppj.SourceIDToOutgoingContext(context.Background(), mppj.SourceID(*nodeID))
	stream, err := helperClient.PushRows(ctx)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	log.Printf("preparing and sending %d rows using %d CPU(s)...", len(*table), *nCPU)

	encRows, err := ds.PrepareStream(rpk, *table, *nCPU)
	if err != nil {
		log.Fatalf("Failed to prepare stream: %v", err)
	}

	startActive := time.Now() // measured time from helper connect
	for encRow := range encRows {
		encRowMsg, err := api.GetEncRowMsg(encRow)
		if err != nil {
			log.Fatalf("Failed to marshal enc row: %v", err)
		}
		if err := stream.Send(encRowMsg); err != nil {
			log.Printf("Failed to send enc row: %v", err)
			break
		}
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Stream resulted in error: %v", err)
	}

	log.Printf("done sending %d rows", len(*table))
	total := time.Since(start)
	active := time.Since(startActive)
	common.PrintStats(statsHandler.GetStats(), total, active)
}

func readFile(filename string) (*mppj.TablePlain, error) {

	var r io.Reader
	switch filename {
	case "stdin":
		r = os.Stdin
	default:
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}
	return readCSV(csv.NewReader(r))
}

func readCSV(csvReader *csv.Reader) (*mppj.TablePlain, error) {
	table := make(mppj.TablePlain)
	first := true
	for line, err := csvReader.Read(); err == nil; line, err = csvReader.Read() {
		if first {
			first = false
			continue
		}
		val := strings.Join(line[1:], string(csvReader.Comma))
		if len(val) > MAX_VAL_LEN {
			return nil, fmt.Errorf("value too long: %s", val)
		}
		table[line[0]] = val
	}
	return &table, nil
}
