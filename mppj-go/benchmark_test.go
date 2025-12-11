package mppj

import (
	"fmt"
	"strconv"
	"testing"
)

type benchParam struct {
	numParties int
	numRows    int
	joinSize   int
}

func BenchmarkSourceProcessRow(b *testing.B) {
	sourceIDs := []SourceID{"source1", "source2"}
	sid := NewSessionID(2, "helper", "receiver", sourceIDs)
	receiver := NewReceiver(sid, sourceIDs)
	source := NewDataSource(sid, receiver.GetPK())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := source.ProcessRow("user1", "value1")
		if err != nil {
			b.Fatalf("ProcessRow failed: %v", err)
		}
	}
}

func BenchmarkHelperConvertRow(b *testing.B) {
	sourceIDs := []SourceID{"source1", "source2"}
	sid := NewSessionID(2, "helper", "receiver", sourceIDs)
	receiver := NewReceiver(sid, sourceIDs)
	source := NewDataSource(sid, receiver.GetPK())
	helper := NewHelper(sid, sourceIDs, 1)

	cuid, cval, err := source.ProcessRow("user1", "value1")
	if err != nil {
		b.Fatalf("ProcessRow failed: %v", err)
	}
	encRow := EncRow{Cuid: cuid, Cval: cval}

	pk := receiver.GetPK()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := helper.ConvertRow(pk, &encRow, 1)
		if err != nil {
			b.Fatalf("ConvertRow failed: %v", err)
		}
	}
}

func BenchmarkOps(b *testing.B) {
	sourceIDs := []SourceID{"source1", "source2"}
	sid := NewSessionID(2, "helper", "receiver", sourceIDs)

	receiver := NewReceiver(sid, sourceIDs)
	ds := NewDataSource(sid, receiver.GetPK())
	//helper := NewHelper(sid)

	tables := []TablePlain{make(TablePlain), make(TablePlain)}
	for i := range 10000 {
		tables[0][fmt.Sprintf("uid-%d", i)] = fmt.Sprintf("val-%d-1", i)
		tables[1][fmt.Sprintf("uid-%d", i)] = fmt.Sprintf("val-%d-0", i)
	}
	table := tables[0]
	pk := receiver.GetPK()

	b.ResetTimer()

	b.Run("PrepareStream", func(b *testing.B) {
		for b.Loop() {
			ds.PrepareStream(pk, table)
		}
	})

}

var benchParams = []benchParam{
	{numParties: 2, numRows: 1000, joinSize: 500},
	{numParties: 3, numRows: 1000, joinSize: 500},
	{numParties: 2, numRows: 10000, joinSize: 5000},
	{numParties: 3, numRows: 10000, joinSize: 5000},
}

func BenchmarkFullJoin(b *testing.B) {

	for _, param := range benchParams {
		b.Run(fmt.Sprintf("%dP-%dRows-%dJoin", param.numParties, param.numRows, param.joinSize), func(b *testing.B) {

			dsNames := make([]SourceID, param.numParties)
			for i := 0; i < param.numParties; i++ {
				dsNames[i] = SourceID("ds" + strconv.Itoa(i+1))
			}
			sid := NewSessionID(3, "helper", "receiver", dsNames)

			receiver := NewReceiver(sid, dsNames)
			ds := NewDataSource(sid, receiver.GetPK()) // technically, only one data source instance is needed
			helper := NewHelper(sid, dsNames, param.numRows)

			rpk := receiver.GetPK()

			tables := GenTestTables(dsNames, param.numRows, param.joinSize)

			b.ResetTimer()

			encTables := make(map[SourceID]EncTable, 0)

			for b.Loop() {

				// Data sources do this:

				for sourceID, table := range tables {
					encTable, err := ds.Prepare(receiver.GetPK(), table)
					encTables[sourceID] = encTable
					if err != nil {
						b.Errorf("Error in Prepare: %v", err)
					}
				}

				// Send tables to helper
				// Helper does this:

				joinedTables, err := helper.Convert(rpk, encTables)
				if err != nil {
					b.Errorf("Error in Convert")
				}

				// Send tables to receiver
				// Receiver does this:

				receiver.JoinTables(joinedTables, len(encTables))
			}
		})

	}

}
