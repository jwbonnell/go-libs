package mapper

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
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

type MapperTestSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, you need to add a Test function
// that calls suite.Run.
func TestMapperTestSuite(t *testing.T) {
	suite.Run(t, new(MapperTestSuite))
}

// SetupSuite adds all included custom mappers
func (su *MapperTestSuite) SetupSuite() {
	// sql.NullTime -> time.Time
	RegisterConverter(reflect.TypeOf(sql.NullTime{}), reflect.TypeOf(time.Time{}), sqlNullTimeToTime)

	// *sql.NullTime -> time.Time
	RegisterConverter(reflect.TypeOf(&sql.NullTime{}).Elem(), reflect.TypeOf(time.Time{}), sqlNullPtrTimeToTime)

	// time.Time -> sql.NullTime
	RegisterConverter(reflect.TypeOf(time.Time{}), reflect.TypeOf(sql.NullTime{}), timeToSqlNullTime)

	// sql.NullString -> string
	RegisterConverter(reflect.TypeOf(sql.NullString{}), reflect.TypeOf(""), nullStringToString)

	// sql.NullTime -> string
	RegisterConverter(reflect.TypeOf(sql.NullTime{}), reflect.TypeOf(""), sqlNullTimeToString(func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	}))
}

func (su *MapperTestSuite) TestMapStruct_SimpleFields() {
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
	su.Require().NoError(err)
	su.Require().Equal(42, got.ID)
	su.Require().Equal("test", got.Name)
	su.Require().True(got.RawT.Valid)
	su.Require().Equal(s.RawT, got.RawT.Time)
	su.Require().Equal("", got.NullTime)
	su.Require().Equal(now.Format("2006-01-02 15:04:05"), got.NullTime2)
	su.Require().Equal("hello", got.StrN)
	su.Require().True(reflect.DeepEqual(got.Map, map[string]int{"a": 1}))
}

func (su *MapperTestSuite) TestMapStruct_NestedAndSlice() {
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
	su.Require().NoError(err)

	// Inner.When mapped to time.Time
	su.Require().True(got.Inner.When.Equal(now), fmt.Sprintf("Inner.When: want %v got %v", now, got.Inner.When))

	// Inner.Tags
	su.Require().True(reflect.DeepEqual(got.Inner.Tags, []string{"x", "y"}), fmt.Sprintf("Inner.Tags mismatch: %+v", got.Inner.Tags))
	su.Require().Equal(7, got.Inner.Nest.Value, fmt.Sprintf("Inner.Nest.Value: want 7 got %d", got.Inner.Nest.Value))
	su.Require().Len(got.List, 2)
	su.Require().True(got.List[0].When.Equal(now.Add(-24*time.Hour)), fmt.Sprintf("List[0].When mismatch: want %v got %v", now.Add(-24*time.Hour), got.List[0].When))
	su.Require().True(got.List[1].When.IsZero(), fmt.Sprintf("List[1].When expected zero, got %v", got.List[1].When))
	su.Require().NotNil(got.Ptr)
	su.Require().True(got.Ptr.When.Equal(now.Add(1*time.Hour)), fmt.Sprintf("Ptr.When mismatch: got %v", got.Ptr.When))
	su.Require().Equal(1, len(got.SlicePtr), fmt.Sprintf("SlicePtr length want 1 got %d", len(got.SlicePtr)))
	su.Require().False(got.SlicePtr[0].Tags == nil || got.SlicePtr[0].Tags[0] != "sp", fmt.Sprintf("SlicePtr element mismatch: %+v", got.SlicePtr[0]))
}

func (su *MapperTestSuite) TestMapStruct_TypesSliceToArrayLengthDiff() {
	// Attempt mapping where destination array too small
	type S1 struct {
		A []int
	}
	type D1 struct {
		A [1]int
	}
	s := S1{A: []int{1, 2}}
	d, err := MapStruct[D1](s)
	su.Require().NoError(err)
	su.Require().Equal(1, d.A[0])
}

func (su *MapperTestSuite) TestMapStruct_MapConversion() {
	type S2 struct {
		M map[string]int
	}
	type D2 struct {
		M map[string]int
	}
	s := S2{M: map[string]int{"k": 9}}
	got, err := MapStruct[D2](s)
	su.Require().NoError(err, fmt.Sprintf("MapStruct[D2]: %v", err))
	su.Require().Equal(9, got.M["k"], fmt.Sprintf("MapStruct[D2]: want %v got %v", 9, got.M["k"]))
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

func (su *MapperTestSuite) TestMapSlice_ValueElements() {
	now := time.Now().Truncate(time.Second)
	src := []msSrc{
		{ID: 1, Name: "a", When: sql.NullTime{Time: now, Valid: true}},
		{ID: 2, Name: "b", When: sql.NullTime{Valid: false}},
	}
	got, err := MapSlice[msDst](src)
	su.Require().NoError(err)
	su.Require().Len(got, 2)
	su.Require().False(got[0].ID != 1 || got[0].Name != "a", fmt.Sprintf("element0 mismatch: %+v", got[0]))
	su.Require().True(got[0].When.Equal(now), fmt.Sprintf("element0 When: want %v got %v", now, got[0].When))
	su.Require().True(got[1].When.IsZero(), fmt.Sprintf("element1 When expected zero, got %v", got[1].When))
}

func (su *MapperTestSuite) TestMapSlice_PointerElements() {
	now := time.Now().Truncate(time.Second)
	src := []*msSrc{
		{ID: 3, Name: "p", When: sql.NullTime{Time: now, Valid: true}},
		nil,
	}
	got, err := MapSlice[msDst](src)
	su.Require().NoError(err, fmt.Sprintf("MapSlice error: %v", err))
	su.Require().Len(got, 2, fmt.Sprintf("len: want 2 got %d", len(got)))
	su.Require().Equal(3, got[0].ID, fmt.Sprintf("element0 mismatch: %+v", got[0]))
	su.Require().Equal("p", got[0].Name, fmt.Sprintf("element0 mismatch: %+v", got[0]))

	// nil source pointer -> zero value dest
	su.Require().True(reflect.DeepEqual(got[1], msDst{}), fmt.Sprintf("element1 expected zero value, got %+v", got[1]))
}

func (su *MapperTestSuite) TestMapSlice_EmptyAndNil() {
	var nilSlice []msSrc
	got, err := MapSlice[msDst](nilSlice)
	su.Require().NoError(err, fmt.Sprintf("MapSlice error: %v", err))
	su.Require().NotNil(got, "expected empty slice (not nil), got nil")

	var empty []msSrc
	got2, err := MapSlice[msDst](empty)
	su.Require().NoError(err, fmt.Sprintf("MapSlice error: %v", err))
	su.Require().Equal(0, len(got2), fmt.Sprintf("expected length 0, got %d", len(got2)))
}

func (su *MapperTestSuite) TestMapSlice_NonSliceError() {
	_, err := MapSlice[msDst](123)
	su.Require().Error(err, "expected error for non-slice input")
}
