package mppj

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/big"
	"slices"
	"strconv"
	"strings"

	"crypto/rand"

	"github.com/google/uuid"
)

type TablePlain map[string]string

type TableRow struct {
	uid string
	val string
}

type EncRow struct {
	Cuid *Ciphertext
	Cval []*Ciphertext
}

type EncTable []EncRow

type EncRowWithHint struct {
	Cnyme   Ciphertext
	CVal    SymmetricCiphertext
	CValKey Ciphertext
	CHint   Ciphertext
}

type EncTableWithHint []EncRowWithHint

type EncValueWithHint struct {
	val        SymmetricCiphertext
	blindedkey Message
	hint       Message
}

type JoinTable struct {
	sourceids []SourceID
	values    [][]string
}

func (t JoinTable) Len() int {
	return len(t.values)
}

func (er EncRow) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	cuidBytes, err := er.Cuid.Serialize()
	if err != nil {
		return nil, err
	}
	if len(er.Cval) > 1 {
		return nil, fmt.Errorf("multiple ciphertext values not supported")
	}
	cvalBytes, err := er.Cval[0].Serialize()
	if err != nil {
		return nil, err
	}
	buf.Write(cuidBytes)
	buf.Write(cvalBytes)
	return buf.Bytes(), nil
}

// NewTablePlain creates a new Table from a UID list and optional values.
func NewTablePlain(uids []string, values []string) TablePlain {

	var newTable = make(map[string]string, len(uids))

	for i, uid := range uids {
		if i < len(values) {
			newTable[uid] = values[i]
		}
	}
	return TablePlain(newTable)
}

func NewJoinTable(sourceIDs []SourceID) JoinTable {

	var newTable = JoinTable{
		sourceids: make([]SourceID, len(sourceIDs)),
		values:    make([][]string, 0),
	}
	copy(newTable.sourceids, sourceIDs)
	return newTable
}

func (t *JoinTable) Insert(values map[SourceID]string) error {
	row := make([]string, len(t.sourceids))
	for sourceID, value := range values {
		col := slices.Index(t.sourceids, sourceID)
		if col == -1 {
			return fmt.Errorf("source ID %s not found", sourceID)
		}
		row[col] = value
	}
	t.values = append(t.values, row)
	return nil
}

func (t JoinTable) WriteTo(w *csv.Writer) error {
	sourceIDsStr := make([]string, len(t.sourceids))
	for i, sid := range t.sourceids {
		sourceIDsStr[i] = string(sid)
	}
	if err := w.Write(sourceIDsStr); err != nil {
		return err
	}
	for _, row := range t.values {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return nil
}

// Equality for plain tables checks both the keys and the values
func (t1 *TablePlain) Equal(t2 *TablePlain) bool {
	if len(*t1) != len(*t2) {
		return false
	}

	for key, value1 := range *t1 {
		value2, exists := (*t2)[key]
		if !exists {
			return false
		}
		if value1 != value2 {
			return false
		}
	}

	return true
}

// Equality for joined tables only checks the values, the keys may be different
func (t1 *JoinTable) EqualContents(t2 *JoinTable) bool {

	if t1.Len() != t2.Len() {
		return false
	}

	for sid := range t1.sourceids {
		if t1.sourceids[sid] != t2.sourceids[sid] {
			return false
		}
	}

	t1Vals := make(map[string]struct{})
	for _, row := range t1.values {
		rowKey := strings.Join(row, "|") // TODO: more robust way to determine equality
		t1Vals[rowKey] = struct{}{}
	}

	for _, row := range t2.values {
		rowKey := strings.Join(row, "|")
		if _, exists := t1Vals[rowKey]; !exists {
			return false
		}
	}

	return true
}

// GenUIDs generates a specified amount of unique UUIDs.
func GenUIDs(amount int) []string {
	uids := make(map[string]struct{})
	for len(uids) < amount {
		uid := uuid.New().String()
		uids[uid] = struct{}{}
	}

	uidList := make([]string, 0, len(uids))
	for uid := range uids {
		uidList = append(uidList, uid)
	}
	return uidList
}

// ExpandUIDs expands a list of UIDs to a specified amount by adding random UIDs.
func ExpandUIDs(uids []string, amount int) []string {
	uidSet := make(map[string]struct{})
	for _, uid := range uids {
		uidSet[uid] = struct{}{}
	}

	for len(uidSet) < amount {
		coin, err := rand.Int(rand.Reader, big.NewInt(2))
		if err != nil {
			panic(fmt.Sprintf("RNG error: %v", err)) // other functions like RandomPoint also panic when RNG fails
		}
		if coin.Int64() == 0 && len(uids) > 0 {
			uidSet[uids[0]] = struct{}{}
			uids = uids[1:]
		}
		uidSet[uuid.New().String()] = struct{}{}
	}

	expandedUIDs := make([]string, 0, len(uidSet))
	for uid := range uidSet {
		expandedUIDs = append(expandedUIDs, uid)
	}
	return expandedUIDs
}

// IntersectSimple performs a join on plain tables
func IntersectSimple(tables map[SourceID]TablePlain, sources []SourceID) JoinTable {

	// groups the values by uids
	partJoin := make(map[string]map[SourceID]string)
	for sourceID, table := range tables {
		for uid, val := range table {
			if _, exists := partJoin[uid]; !exists {
				partJoin[uid] = make(map[SourceID]string)
			}
			partJoin[uid][sourceID] = val
		}
	}

	joined := NewJoinTable(sources)
	for _, vals := range partJoin {
		if len(vals) == len(tables) {
			joined.Insert(vals)
		}
	}
	return joined
}

func (t TablePlain) String() string {
	var s string
	s += "UID " + " " + " Value\n"
	s += "---------------------\n"

	for uid, value := range t {
		s += uid + " "
		s += value + " "

		s += "\n"
	}

	return s
}

func GenTestTables(sourceIDs []SourceID, nRows, intersectionSize int) map[SourceID]TablePlain {
	intersection := make([]string, intersectionSize)
	for i := 0; i < intersectionSize; i++ {
		intersection[i] = fmt.Sprintf("join_key_%d", i)
	}
	tables := make(map[SourceID]TablePlain)
	for _, sourceID := range sourceIDs {
		tables[sourceID] = GenTestTable(sourceID, nRows, intersection)
	}
	return tables
}

func GenTestTable(sourceId SourceID, nRows int, intersection []string) TablePlain {
	table := make(TablePlain)
	v := 0
	// Add intersection rows
	for _, uid := range intersection {
		table[uid] = "value_" + strconv.Itoa(v)
		v++
	}
	// Add non-intersection rows
	for i := 0; i < nRows-len(intersection); i++ {
		uid := fmt.Sprintf("%s_%d", sourceId, i)
		table[uid] = "non_join_value_" + strconv.Itoa(v)
		v++
	}
	return table
}
