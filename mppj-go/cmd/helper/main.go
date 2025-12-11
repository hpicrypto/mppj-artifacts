package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mppj"
	"mppj/api"
	"mppj/api/pb"
	"mppj/cmd/common"
	"mppj/cmd/config"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var sources mppj.SourceList
var (
	nodeId   = flag.String("id", "", "the id of the node")
	bindAddr = flag.String("bind_address", fmt.Sprintf(":%d", config.DEFAULT_PORT), "the address to bind")
	nRows    = flag.Int("n_rows", 0, "the number of rows per source")
)

func init() {
	flag.Var((*mppj.SourceList)(&sources), "sources", "the sources' ids as a comma-separated list")
	log.SetFlags(log.Flags() &^ log.Ldate)
	log.SetPrefix("> ")
}

type mppjHelperServer struct {
	incomingEncRows chan mppj.ConvertRowTask

	convTables chan mppj.EncTableWithHint

	expected map[mppj.SourceID]mppj.TableIndex
	mu       sync.Mutex
	once     sync.Once

	start, stop chan struct{} // signals for start and stop of processing

	pb.UnimplementedMPPJHelperServer
}

func newHelperServer() *mppjHelperServer {

	h := mppj.NewHelper(config.SessionID, sources, *nRows)

	rpk := common.GetRPK(config.SessionID)

	srv := &mppjHelperServer{
		incomingEncRows: make(chan mppj.ConvertRowTask),
		convTables:      make(chan mppj.EncTableWithHint, 1),
		expected:        make(map[mppj.SourceID]mppj.TableIndex, len(sources)),
		start:           make(chan struct{}),
		stop:            make(chan struct{}),
	}

	for i, id := range sources {
		srv.expected[id] = mppj.TableIndex(i)
	}

	go func() {
		log.Printf("waiting for %d sources: %v", len(srv.expected), sources)
		convTables, err := h.ConvertTablesStream(rpk, srv.incomingEncRows)
		if err != nil {
			log.Fatalf("failed to convert tables: %v", err)
		}
		srv.convTables <- convTables
		log.Println("conversion done")
	}()

	return srv
}

func (s *mppjHelperServer) PushRows(stream pb.MPPJHelper_PushRowsServer) error {

	sourceID, ok := mppj.SourceIDFromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "missing source ID")
	}

	s.mu.Lock()
	tindex, ok := s.expected[sourceID] // TODO: this doesn't check for multiple connections from the same source
	if !ok {
		s.mu.Unlock()
		return status.Error(codes.NotFound, "unexpected source ID")
	}
	s.once.Do(func() { close(s.start) })
	s.mu.Unlock()

	log.Printf("starting to receive rows for source %s", sourceID)

	var rc int
	for {
		encRowMsg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		encRow, err := api.GetEncRowFromMsg(encRowMsg)
		if err != nil {
			return err
		}
		s.incomingEncRows <- mppj.ConvertRowTask{EncRowMsg: encRow, TableIndex: tindex}
		rc++
	}

	log.Printf("%d rows received for source %s", rc, sourceID)
	if err := stream.SendAndClose(&pb.Void{}); err != nil {
		return err
	}

	// Close the incoming channel if all tables have been received
	s.mu.Lock()
	delete(s.expected, sourceID)
	if len(s.expected) == 0 {
		close(s.incomingEncRows)
	}
	s.mu.Unlock()

	return nil
}

func (s *mppjHelperServer) PullRows(_ *pb.Void, stream grpc.ServerStreamingServer[pb.EncRowWithHint]) error {
	convTables := <-s.convTables

	log.Printf("sending %d rows to receiver", len(convTables))
	if err := stream.SetHeader(metadata.New(map[string]string{
		"num_rows": fmt.Sprintf("%d", len(convTables)),
	})); err != nil {
		return err
	}

	var i int
	for ; i < len(convTables); i++ {
		row := convTables[i]
		rowMsg, err := api.GetEncRowWithHintMsg(row)
		if err != nil {
			return err
		}

		if err := stream.Send(rowMsg); err != nil {
			log.Printf("error sending row: %v", err)
			return err
		}
	}

	log.Printf("done sending %d rows to receiver", i)
	<-stream.Context().Done() // wait for receiver to close
	close(s.stop)

	return nil
}

func main() {

	flag.Parse()

	log.Printf("MPPJ Helper %s", *nodeId)

	if len(*nodeId) == 0 {
		log.Fatal("an id should be provided")
	}

	if len(sources) < 2 { // TODO: does not check for duplicates
		log.Fatal("at least two sources ids must be provided")
	}

	if *nRows <= 0 {
		log.Fatal("number of rows per source must be positive")
	}

	lis, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	statsHandler := api.NewStatsHandler()
	opts = append(opts, grpc.StatsHandler(statsHandler))
	grpcServer := grpc.NewServer(opts...)
	helper := newHelperServer()
	pb.RegisterMPPJHelperServer(grpcServer, helper)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("error during serve: %v", err)
		}
	}()

	log.Printf("helper listening at %v", lis.Addr())
	start := time.Now()
	<-helper.start
	startActive := time.Now() // measured time from first source connection
	<-helper.stop
	log.Println("done processing")
	total := time.Since(start)
	active := time.Since(startActive)
	common.PrintStats(statsHandler.GetStats(), total, active)
	<-time.After(time.Second) // leaves some time for streams to close as GracefulStop seems insufficient
	log.Println("shutting down")
	grpcServer.GracefulStop()
}
