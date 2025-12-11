package mppj

import (
	"testing"
)

const TABLE_AMOUNT = 3
const ROW_AMOUNT = 10
const INTERSECTION_SIZE = 3

func TestMPPJ(t *testing.T) {

	sourceIDs := []SourceID{"ds1", "ds2", "ds3"}

	sid := NewSessionID(3, "helper", "receiver", sourceIDs)

	// Setup

	helper := NewHelper(sid, sourceIDs, ROW_AMOUNT)
	receiver := NewReceiver(sid, sourceIDs)

	// Data sources do this:

	tables := GenTestTables(sourceIDs, ROW_AMOUNT, INTERSECTION_SIZE)
	encTables := make(map[SourceID]EncTable, TABLE_AMOUNT)

	for sourceID, table := range tables {
		ds := NewDataSource(sid, receiver.GetPK()) // technically, only one data source instance is needed

		prepTable, err := ds.Prepare(receiver.GetPK(), table)
		if err != nil {
			t.Errorf("Error in PrepareTable")
		}
		encTables[sourceID] = prepTable
	}

	// Send tables to helper
	// Helper does this:

	joinedTables, err := helper.Convert(receiver.recvPK, encTables)
	if err != nil {
		t.Errorf("Error in ConvertTablesMPPJ")
	}

	// Send tables to receiver
	// Receiver does this:

	intersectionMPPJ, err := receiver.JoinTables(joinedTables, len(encTables))
	if err != nil {
		t.Errorf("Error in JoinTablesMPPJ")
	}

	// Check results
	joinedTablesPlain := IntersectSimple(tables, sourceIDs)

	if !joinedTablesPlain.EqualContents(&intersectionMPPJ) {
		t.Errorf("Expected tables' contents to be equal, but they are not: \n Plain: \n%s \n MPPJ: \n%s", joinedTablesPlain, intersectionMPPJ)
	}

}
