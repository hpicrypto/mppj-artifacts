package mppj

import (
	"fmt"
	"log"
	"runtime"
	"sync"
)

type Receiver struct {
	sid       []byte
	sourceIDs []SourceID
	recvSK    SecretKeyTuple
	recvPK    PublicKeyTuple
}

// NewReceiver creates a new receiver
func NewReceiver(sid []byte, sourceIDs []SourceID) *Receiver { // TODO probably we don't want this one
	bsk, bpk := PKEKeyGen()
	esk, epk := PKEKeyGen()

	rsk := SecretKeyTuple{
		bsk: bsk,
		esk: esk,
	}

	rpk := PublicKeyTuple{
		bpk: bpk,
		epk: epk,
	}

	return NewReceiverWithKeys(sid, sourceIDs, rsk, rpk)
}

func NewReceiverWithKeys(sid []byte, sourceIDs []SourceID, rsk SecretKeyTuple, rpk PublicKeyTuple) *Receiver {
	r := &Receiver{
		sid:       sid,
		sourceIDs: make([]SourceID, len(sourceIDs)),
		recvSK:    rsk,
		recvPK:    rpk,
	}
	copy(r.sourceIDs, sourceIDs)
	return r
}

func (r *Receiver) GetPK() PublicKeyTuple {
	return r.recvPK
}

func (r *Receiver) getSK() SecretKeyTuple {
	return r.recvSK
}

// JoinTables joins the tables using the MPPJ protocol.
func (r *Receiver) JoinTables(joinedTables EncTableWithHint, tableAmount int) (JoinTable, error) {

	encrows := make(chan EncRowWithHint, len(joinedTables))

	go func() {
		defer close(encrows)
		for _, ct := range joinedTables {
			encrows <- ct
		}
	}()

	return r.JoinTablesStream(encrows, tableAmount)

}

func (r *Receiver) JoinTablesStream(in chan EncRowWithHint, numTable int) (JoinTable, error) {

	groups := make(map[string][]EncRowWithHint)

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ciphertexts := range in {
				msgPRF, err := OPRFUnblind(r.recvSK.bsk, &ciphertexts.Cnyme).GetMessageBytes()
				if err != nil {
					log.Fatalf("Decryption error: %v", err)
				}

				mu.Lock()
				groups[string(msgPRF)] = append(groups[string(msgPRF)], ciphertexts)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return r.intersectHint(groups)
}

func (r *Receiver) decryptGroup(group []EncRowWithHint) (map[SourceID]string, error) {
	decGroup := make([]EncValueWithHint, len(group))

	for i, ge := range group {
		decGroup[i] = EncValueWithHint{
			val:        ge.CVal,
			blindedkey: *OPRFUnblind(r.recvSK.bsk, &ge.CValKey),
			hint:       *OPRFUnblind(r.recvSK.bsk, &ge.CHint),
		}
	}

	mask := Identity()
	for _, dge := range decGroup {
		mask = Mul(mask, &dge.hint.m)
	}
	invMask := mask.Invert()

	out := make(map[SourceID]string, len(group))
	for _, dge := range decGroup {
		keyp := Mul(&dge.blindedkey.m, invMask)
		key, err := KeyFromPoint(keyp, r.sid)
		if err != nil {
			panic(err)
		}

		encAttridValBytes, err := SymmetricDecrypt(key, dge.val)
		if err != nil {
			panic(err)
		}

		if len(encAttridValBytes) == 0 {
			panic(fmt.Errorf("incorrect encrypted attribute value"))
		}

		sourceIndex, encValBytes := int(encAttridValBytes[0]), encAttridValBytes[1:]
		if sourceIndex < 0 || sourceIndex >= len(r.sourceIDs) {
			panic(fmt.Errorf("invalid source index: %d", sourceIndex))
		}
		sourceID := r.sourceIDs[sourceIndex]

		encVal, err := DeserializeCiphertexts(encValBytes)
		if err != nil {
			panic(err)
		}

		plantext_data, err := PKEDecryptVector(r.recvSK.esk, encVal)
		if err != nil {
			panic(err)
		}

		out[sourceID] = string(plantext_data)
	}
	return out, nil
}

func (r *Receiver) intersectHint(groups map[string][]EncRowWithHint) (JoinTable, error) {

	decryptTasks := make(chan []EncRowWithHint)

	join := NewJoinTable(r.sourceIDs)
	mu := sync.Mutex{}

	wg := sync.WaitGroup{}
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for dectask := range decryptTasks {
				vals, err := r.decryptGroup(dectask)
				if err != nil {
					panic(err)
				}
				mu.Lock()
				if err := join.Insert(vals); err != nil {
					panic(err)
				}
				mu.Unlock()
			}
		}()
	}

	for _, group := range groups {
		if len(group) == len(r.sourceIDs) {
			decryptTasks <- group
		}
	}
	close(decryptTasks)

	wg.Wait()

	return join, nil
}
