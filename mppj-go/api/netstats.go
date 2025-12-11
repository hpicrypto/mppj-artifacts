package api

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

// NetStats contains the network statistics of a connection.
type NetStats struct {
	DataSent, DataRecv uint64
}

// String returns a string representation of the network statistics.
func (s NetStats) String() string {
	return fmt.Sprintf("Sent: %s, Received: %s", byteCountSI(s.DataSent), byteCountSI(s.DataRecv))
}

type statsHandler struct {
	mu sync.Mutex
	ns NetStats
}

func NewStatsHandler() *statsHandler {
	return new(statsHandler)
}

// TagRPC can attach some information to the given context.
// The context used for the rest lifetime of the RPC will be derived from
// the returned context.
func (s *statsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

// HandleRPC processes the RPC stats.
func (s *statsHandler) HandleRPC(ctx context.Context, sta stats.RPCStats) {

	s.mu.Lock()
	defer s.mu.Unlock()
	switch sta := sta.(type) {
	case *stats.InPayload:
		s.ns.DataRecv += uint64(sta.WireLength)
	case *stats.OutPayload:
		s.ns.DataSent += uint64(sta.WireLength)
	}
}

// TagConn can attach some information to the given context.
// The returned context will be used for stats handling.
// For conn stats handling, the context used in HandleConn for this
// connection will be derived from the context returned.
// For RPC stats handling,
//   - On server side, the context used in HandleRPC for all RPCs on this
//     connection will be derived from the context returned.
//   - On client side, the context is not derived from the context returned.
func (s *statsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn processes the Conn stats.
func (s *statsHandler) HandleConn(_ context.Context, sta stats.ConnStats) {}

func (s *statsHandler) GetStats() NetStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ns
}

// byteCountSI returns a string representation of a byte count b,
// by formatting it as a SI value.
func byteCountSI(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
