package api

import (
	"fmt"
	"mppj"
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestSerializeMessages(t *testing.T) {

	sourceIDs := []mppj.SourceID{"ds1", "ds2", "ds3"}

	sid := mppj.NewSessionID(3, "helper", "receiver", sourceIDs)

	// Setup

	helper := mppj.NewHelper(sid, sourceIDs, 1)

	receiver := mppj.NewReceiver(sid, sourceIDs)
	source := mppj.NewDataSource(sid, receiver.GetPK())

	// Data sources do this:

	cuid, cval, err := source.ProcessRow("user1", "value1")
	if err != nil {
		t.Fatalf("ProcessRow failed: %v", err)
	}

	bcuid, _ := cuid.Serialize()
	fmt.Println("cuid size", len(bcuid))

	encRow := mppj.EncRow{Cuid: cuid, Cval: cval}
	encRowMsg, err := GetEncRowMsg(encRow)
	if err != nil {
		t.Fatalf("GetEncRowMsg failed: %v", err)
	}

	fmt.Println("Size of EncRow message:", proto.Size(encRowMsg))

	encRow, err = GetEncRowFromMsg(encRowMsg)
	if err != nil {
		t.Fatalf("GetEncRowFromMsg failed: %v", err)
	}

	encRowWithHint, err := helper.ConvertRow(receiver.GetPK(), &encRow, 1)
	if err != nil {
		t.Fatalf("ConvertRow failed: %v", err)
	}

	fmt.Println("Size of CVal", len(encRowWithHint.CVal))

	encRowWithHintMsg, err := GetEncRowWithHintMsg(*encRowWithHint)
	if err != nil {
		t.Fatalf("GetEncRowWithHintMsg failed: %v", err)
	}

	fmt.Println("Size of EncRowWithHint message:", proto.Size(encRowWithHintMsg))

	_, err = GetEncRowWithHintFromMsg(encRowWithHintMsg)
	if err != nil {
		t.Fatalf("GetEncRowWithHintFromMsg failed: %v", err)
	}
}
