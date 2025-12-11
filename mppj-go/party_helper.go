package mppj

import (
	"errors"
	"math/big"
	"math/rand/v2"
	"runtime"
	"sync"
)

type Helper struct {
	sid           []byte
	sourceIndices map[SourceID]int

	convK        *OPRFKey
	padKeyShares []*Scalar
	padKey       *Scalar
	rowPerm      []int
}

// NewHelper creates a new Helper with the given key.
func NewHelper(sid []byte, sources []SourceID, nRows int) *Helper {
	c := &Helper{sid: sid, sourceIndices: make(map[SourceID]int)}
	for i, source := range sources {
		c.sourceIndices[source] = i
	}
	c.resetKey()
	c.genNonces(len(sources))
	c.rowPerm = rand.Perm(nRows * len(sources)) // TODO: proper RNG
	return c
}

// resetKey generates a new  random key for the Helper.
func (h *Helper) resetKey() {
	k := OPRFKeyGen()
	h.convK = k
}

func (h *Helper) getK() *OPRFKey {
	return h.convK
}

// ResetKey generates a new  random key for the Helper.
func (h *Helper) genNonces(tableAmount int) {
	nonceSum := NewScalar(big.NewInt(0))

	nonces := make([]*Scalar, tableAmount)
	for i := range tableAmount {
		s := RandomScalar()

		nonces[i] = s
		nonceSum = nonceSum.Add(s)
	}

	h.padKeyShares = nonces
	h.padKey = nonceSum

}

// blindAndHint produces an "ad" ciphertext, a blinded key, and a hint
func (h *Helper) blindAndHint(rpk PublicKeyTuple, joinid *Ciphertext, value []*Ciphertext, tindex int) ([]byte, *Ciphertext, *Ciphertext, error) {

	rp, key := RandomKeyFromPoint(h.sid)

	serialized, err := SerializeCiphertexts(ReRandVector(rpk.epk, value))
	if err != nil {
		return nil, nil, nil, err
	}

	ad, err := SymmetricEncrypt(key, append([]byte{byte(tindex)}, serialized...)) // append the table pos for in order reconstruction
	if err != nil {
		return nil, nil, nil, err
	}

	blindkey := OPRFEval((*OPRFKey)(h.padKey), rpk.bpk, joinid) // ReRand internally
	blindkey.c1 = Mul(blindkey.c1, rp)                          // blind the ephemeral point using joinid ^ s

	hint := OPRFEval((*OPRFKey)(h.padKeyShares[tindex]), rpk.bpk, joinid) // ReRand internally

	return ad, blindkey, hint, nil
}

// Convert performs DH-PRF on the hashed identifiers, blinds the data, then rerandomizes and shuffles all ciphertexts. GenNonces does not neet to be run before this function.
func (h *Helper) Convert(rpk PublicKeyTuple, tables map[SourceID]EncTable) (EncTableWithHint, error) {

	encRowsTasks := make(chan ConvertRowTask, 0)

	go func() {
		for sourceID, table := range tables {
			for _, row := range table {
				encRowsTasks <- ConvertRowTask{
					EncRowMsg:  EncRow{Cuid: row.Cuid, Cval: row.Cval},
					TableIndex: TableIndex(h.sourceIndices[sourceID]),
				}
			}
		}
		close(encRowsTasks)
	}()

	return h.ConvertTablesStream(rpk, encRowsTasks)
}

type TableIndex int

type ConvertRowTask struct {
	EncRowMsg  EncRow
	TableIndex TableIndex
}

func (h *Helper) ConvertTablesStream(rpk PublicKeyTuple, encRowsTasks chan ConvertRowTask) (EncTableWithHint, error) {

	if h.padKey == nil || h.padKeyShares == nil {
		return nil, errors.New("nonceerr, Nonces not generated. Please call GenNonces() before calling this function")
	}

	res := make(EncTableWithHint, len(h.rowPerm))
	permIndex := 0
	mu := new(sync.Mutex)

	var wg sync.WaitGroup
	for range runtime.NumCPU() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for encRow := range encRowsTasks {
				convRow, err := h.ConvertRow(rpk, &encRow.EncRowMsg, int(encRow.TableIndex))
				if err != nil {
					panic(err)
				}
				mu.Lock()
				if permIndex >= len(h.rowPerm) {
					panic("conversion received more rows than expected")
				}
				res[h.rowPerm[permIndex]] = *convRow
				permIndex++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	return res, nil
}

func (h *Helper) ConvertRow(rpk PublicKeyTuple, r *EncRow, rid int) (*EncRowWithHint, error) {

	joinid := *OPRFEval(h.convK, rpk.bpk, r.Cuid) // ReRand internally

	ad, blindedkey, hint, err := h.blindAndHint(rpk, &joinid, r.Cval, rid)
	if err != nil {
		panic(err)
	}
	return &EncRowWithHint{Cnyme: joinid, CVal: ad, CValKey: *blindedkey, CHint: *hint}, nil
}
