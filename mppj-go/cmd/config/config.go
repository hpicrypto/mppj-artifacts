package config

const DEFAULT_PORT = 40000

var SessionID = []byte("session-id-12345") // this should be randomly generated in real usage, then distributed (see mppj.NewSessionID())
var MaxValLen = 30
var SourceIDContextKey = "source-id"

type NetStatsFormat int

const (
	None NetStatsFormat = iota
	StringFormat
	JsonFormat
)

var LogNetworkStats = JsonFormat
