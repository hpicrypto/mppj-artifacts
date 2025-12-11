package mppj

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"
)

type DataSource struct {
	sid []byte
	rpk PublicKeyTuple
}

func NewDataSource(sid []byte, rpk PublicKeyTuple) *DataSource {
	return &DataSource{sid: sid, rpk: rpk}
}

// Prepare prepares a table for joining by adding hashing the UIDs and encrypting its contents towards the receiver.
func (s *DataSource) Prepare(rpk PublicKeyTuple, table TablePlain) (EncTable, error) {

	preparedTable := make(EncTable, len(table))

	encRows, err := s.PrepareStream(rpk, table)
	if err != nil {
		return nil, err
	}

	i := 0
	for encRow := range encRows {
		preparedTable[i] = encRow
		i++
	}
	if i != len(table) {
		return nil, fmt.Errorf("number of prepared elements do not match")
	}

	return preparedTable, nil
}

// Prepare prepares a table for joining by adding hashing the UIDs and encrypting its contents towards the receiver.
func (s *DataSource) PrepareStream(rpk PublicKeyTuple, table TablePlain, ncpu ...int) (encRows <-chan EncRow, err error) {
	var wg sync.WaitGroup

	rows := make(chan TableRow, len(table))

	encRowsChan := make(chan EncRow, len(table))
	//fmt.Printf("tasks: %d\n", len(table))

	n := runtime.NumCPU()
	if len(ncpu) > 0 && ncpu[0] > 0 {
		n = ncpu[0]
	}

	for range n {
		wg.Add(1)
		go func() {

			i := 0
			for task := range rows {

				cuid, cval, err := s.ProcessRow(task.uid, task.val)
				if err != nil {
					return
				}
				encRowsChan <- EncRow{Cuid: cuid, Cval: cval}
				i++
			}
			//fmt.Printf("worker processed %d\n", i)
			wg.Done()
		}()
	}

	go func() {
		uids := make([]string, 0, len(table))
		for uid := range table {
			uids = append(uids, uid)
		}
		perm := rand.Perm(len(uids)) // TODO: use secure random source
		for _, uid := range perm {
			rows <- TableRow{uid: uids[uid], val: table[uids[uid]]}
		}
		close(rows)
		wg.Wait()
		close(encRowsChan)
	}()

	return encRowsChan, nil
}

func (s *DataSource) ProcessRow(uid, val string) (cuid *Ciphertext, cval []*Ciphertext, err error) {
	cuid = OPRFBlind(s.rpk.bpk, []byte(uid), s.sid)
	cval, err = PKEEncryptVector(s.rpk.epk, []byte(val))
	return
}
