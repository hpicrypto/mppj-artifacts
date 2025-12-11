package main

import (
	"fmt"
	"mppj"
)

func MPPJ() {

	numRows := 100
	joinSize := 10

	fmt.Println("----- MPPJ -----")
	fmt.Println("")

	sourceIDs := []mppj.SourceID{"ds1", "ds2", "ds3"}
	sid := mppj.NewSessionID(3, "helper", "receiver", sourceIDs)

	// Setup phase

	receiver := mppj.NewReceiver(sid, sourceIDs)
	ds := mppj.NewDataSource(sid, receiver.GetPK()) // technically, only one data source instance is needed
	converter := mppj.NewHelper(sid, sourceIDs, 100)

	// Data sources do this:
	tables := mppj.GenTestTables(sourceIDs, numRows, joinSize)

	// Encrypting the tables

	encTables := make(map[mppj.SourceID]mppj.EncTable, 0)

	for sourceID, table := range tables {
		encTable, _ := ds.Prepare(receiver.GetPK(), table)
		encTables[sourceID] = encTable
	}

	// Send tables to converter
	// Converter does this:

	joinedTables, _ := converter.Convert(receiver.GetPK(), encTables)

	// Send tables to receiver
	// Receive phase
	// Receiver does this:

	intersectionMPPJ, _ := receiver.JoinTables(joinedTables, len(encTables))

	fmt.Println("Tables after Join (Pseudonymized)")

	fmt.Println(intersectionMPPJ, "\n length ", intersectionMPPJ.Len())

	// Check results

	// Plaintext join
	joinedTablesPlain := mppj.IntersectSimple(tables, sourceIDs)

	fmt.Println("Intersection of tables (plaintext):")
	fmt.Println(joinedTablesPlain, "\n length ", joinedTablesPlain.Len())

	fmt.Println("Are tables' contents equal?", joinedTablesPlain.EqualContents(&intersectionMPPJ))
}

func main() {
	MPPJ()
}
