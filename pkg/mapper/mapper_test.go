package mapper

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type srcInner struct {
	When sql.NullTime
	Tags []string
	Nest *nestedSrc
}

type nestedSrc struct {
	Value int
}

type src struct {
	ID        int
	Name      string
	Inner     srcInner
	List      []srcInner
	Ptr       *srcInner
	RawT      time.Time
	StrN      sql.NullString
	NullTime  sql.NullTime
	NullTime2 sql.NullTime
	Map       map[string]int
	SlicePtr  []*srcInner
}

type destInner struct {
	When time.Time
	Tags []string
	Nest nestedDest
}

type nestedDest struct {
	Value int
}

type dest struct {
	ID        int
	Name      string
	Inner     destInner
	List      []destInner
	Ptr       *destInner
	RawT      sql.NullTime
	StrN      string
	NullTime  string
	NullTime2 string
	Map       map[string]int
	SlicePtr  []destInner
}

func TestMapStruct_SimpleFields(t *testing.T) {
	now := time.Now()
	s := src{
		ID:   42,
		Name: "test",
		RawT: now,
		NullTime2: sql.NullTime{
			Time:  now,
			Valid: true,
		},
		StrN: sql.NullString{String: "hello", Valid: true},
		Map:  map[string]int{"a": 1},
	}
	got, err := MapStruct[dest](s)
	require.NoError(t, err)
	require.Equal(t, 42, got.ID)
	require.Equal(t, "test", got.Name)
	require.True(t, got.RawT.Valid)
	require.Equal(t, s.RawT, got.RawT.Time)
	require.Equal(t, "", got.NullTime)
	require.Equal(t, now.Format("2006-01-02 15:04:05"), got.NullTime2)
	require.Equal(t, "hello", got.StrN)
	require.True(t, reflect.DeepEqual(got.Map, map[string]int{"a": 1}))
}

func TestMapStruct_NestedAndSlice(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	s := src{
		ID:   1,
		Name: "nested",
		Inner: srcInner{
			When: sql.NullTime{Time: now, Valid: true},
			Tags: []string{"x", "y"},
			Nest: &nestedSrc{Value: 7},
		},
		List: []srcInner{
			{When: sql.NullTime{Time: now.Add(-24 * time.Hour), Valid: true}, Tags: []string{"a"}},
			{When: sql.NullTime{Valid: false}, Tags: nil},
		},
		Ptr: &srcInner{
			When: sql.NullTime{Time: now.Add(1 * time.Hour), Valid: true},
			Tags: []string{"p"},
		},
		SlicePtr: []*srcInner{
			{When: sql.NullTime{Time: now, Valid: true}, Tags: []string{"sp"}},
		},
	}
	got, err := MapStruct[dest](s)
	require.NoError(t, err)

	// Inner.When mapped to time.Time
	if !got.Inner.When.Equal(now) {
		t.Errorf("Inner.When: want %v got %v", now, got.Inner.When)
	}
	// Inner.Tags
	if !reflect.DeepEqual(got.Inner.Tags, []string{"x", "y"}) {
		t.Errorf("Inner.Tags mismatch: %+v", got.Inner.Tags)
	}
	// Inner.Nest.Value mapped
	if got.Inner.Nest.Value != 7 {
		t.Errorf("Inner.Nest.Value: want 7 got %d", got.Inner.Nest.Value)
	}
	// List length and first element When
	if len(got.List) != 2 {
		t.Fatalf("List length: want 2 got %d", len(got.List))
	}
	if !got.List[0].When.Equal(now.Add(-24 * time.Hour)) {
		t.Errorf("List[0].When mismatch: want %v got %v", now.Add(-24*time.Hour), got.List[0].When)
	}
	// Second element had NullTime invalid -> zero time
	if !got.List[1].When.IsZero() {
		t.Errorf("List[1].When expected zero, got %v", got.List[1].When)
	}
	// Ptr mapped into pointer dest.Ptr
	if got.Ptr == nil {
		t.Fatalf("Ptr expected non-nil")
	}
	if !got.Ptr.When.Equal(now.Add(1 * time.Hour)) {
		t.Errorf("Ptr.When mismatch: got %v", got.Ptr.When)
	}
	// SlicePtr ([]*srcInner) -> []destInner
	if len(got.SlicePtr) != 1 {
		t.Fatalf("SlicePtr length want 1 got %d", len(got.SlicePtr))
	}
	if got.SlicePtr[0].Tags == nil || got.SlicePtr[0].Tags[0] != "sp" {
		t.Errorf("SlicePtr element mismatch: %+v", got.SlicePtr[0])
	}
}

func TestMapStruct_TypesSliceToArrayLengthDiff(t *testing.T) {
	// Attempt mapping where destination array too small
	type S1 struct {
		A []int
	}
	type D1 struct {
		A [1]int
	}
	s := S1{A: []int{1, 2}}
	d, err := MapStruct[D1](s)
	require.NoError(t, err)
	require.Equal(t, 1, d.A[0])
}

func TestMapStruct_MapConversion(t *testing.T) {
	type S2 struct {
		M map[string]int
	}
	type D2 struct {
		M map[string]int
	}
	s := S2{M: map[string]int{"k": 9}}
	got, err := MapStruct[D2](s)
	if err != nil {
		t.Fatalf("MapStruct error: %v", err)
	}
	if got.M["k"] != 9 {
		t.Errorf("map value mismatch got %d", got.M["k"])
	}
}

type msSrc struct {
	ID   int
	Name string
	When sql.NullTime
}

type msDst struct {
	ID   int
	Name string
	When time.Time
}

func TestMapSlice_ValueElements(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	src := []msSrc{
		{ID: 1, Name: "a", When: sql.NullTime{Time: now, Valid: true}},
		{ID: 2, Name: "b", When: sql.NullTime{Valid: false}},
	}
	got, err := MapSlice[msDst](src)
	if err != nil {
		t.Fatalf("MapSlice error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len: want 2 got %d", len(got))
	}
	if got[0].ID != 1 || got[0].Name != "a" {
		t.Errorf("element0 mismatch: %+v", got[0])
	}
	if !got[0].When.Equal(now) {
		t.Errorf("element0 When: want %v got %v", now, got[0].When)
	}
	// second element had invalid NullTime -> zero time
	if !got[1].When.IsZero() {
		t.Errorf("element1 When expected zero, got %v", got[1].When)
	}
}

func TestMapSlice_PointerElements(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	src := []*msSrc{
		{ID: 3, Name: "p", When: sql.NullTime{Time: now, Valid: true}},
		nil,
	}
	got, err := MapSlice[msDst](src)
	require.NoError(t, err, fmt.Sprintf("MapSlice error: %v", err))
	require.Len(t, got, 2, fmt.Sprintf("len: want 2 got %d", len(got)))
	require.Equal(t, 3, got[0].ID, fmt.Sprintf("element0 mismatch: %+v", got[0]))
	require.Equal(t, "p", got[0].Name, fmt.Sprintf("element0 mismatch: %+v", got[0]))

	// nil source pointer -> zero value dest
	require.True(t, reflect.DeepEqual(got[1], msDst{}), fmt.Sprintf("element1 expected zero value, got %+v", got[1]))
}

func TestMapSlice_EmptyAndNil(t *testing.T) {
	var nilSlice []msSrc
	got, err := MapSlice[msDst](nilSlice)
	require.NoError(t, err, fmt.Sprintf("MapSlice error: %v", err))
	require.NotNil(t, got, "expected empty slice (not nil), got nil")

	empty := []msSrc{}
	got2, err := MapSlice[msDst](empty)
	require.NoError(t, err, fmt.Sprintf("MapSlice error: %v", err))
	require.Equal(t, 0, len(got2), fmt.Sprintf("expected length 0, got %d", len(got2)))
}

func TestMapSlice_NonSliceError(t *testing.T) {
	_, err := MapSlice[msDst](123)
	require.Error(t, err, "expected error for non-slice input")
}
