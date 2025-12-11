package common

import (
	"encoding/json"
	"fmt"
	"log"
	"mppj"
	"mppj/api"
	"mppj/cmd/config"
	"time"
)

// this function simulates a public key distribution mechanism
func GetRPK(sid []byte) mppj.PublicKeyTuple {
	_, rpk := mppj.GetTestKeys(sid)
	return rpk
}

func PrintStats(s api.NetStats, total, active time.Duration) {
	var stats string
	switch config.LogNetworkStats {
	case config.None:
		// do nothing
	case config.StringFormat:
		stats = fmt.Sprintf("%s, total time: %v, active time: %v", s, total, active)
	case config.JsonFormat:
		json, err := json.Marshal(struct {
			Sent   uint64        `json:"data_sent"`
			Recv   uint64        `json:"data_recv"`
			Total  time.Duration `json:"time_total"`
			Active time.Duration `json:"time_active"`
		}{
			Sent:   s.DataSent,
			Recv:   s.DataRecv,
			Total:  total,
			Active: active,
		})
		if err != nil {
			log.Printf("Failed to marshal stats to json: %v", err)
		}
		stats = string(json)
	}
	if stats != "" {
		log.Printf("Stats: %s", stats)
	}
}
