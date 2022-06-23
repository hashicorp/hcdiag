package op

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mockResult = "get mock'd"
	errFake    = errors.New("uh oh a fake error")
	testStatus = Status("test")
)

type MockRunner struct{}

func (r MockRunner) Run() (interface{}, Status, error) {
	return mockResult, testStatus, errFake
}

func TestOp_Run(t *testing.T) {
	r := MockRunner{}
	s := Op{Identifier: "mock", Runner: r}

	result, err := s.Run()

	// assert that return values are also being stored as struct fields
	if s.Result != result {
		t.Errorf("returned result (%s) does not match Op Result field (%s)", result, s.Result)
	}
	if s.Error != err {
		t.Errorf("returned err (%s) does not match Op Error field (%s)", err, s.Error)
	}
	errStr := fmt.Sprintf("%s", err)
	if s.ErrString != errStr {
		t.Errorf("Op ErrString (%s) not formatted as expected (%s)", s.ErrString, errStr)
	}

	// assert that values are what we expect from MockRunner.Run()
	if err != errFake {
		t.Errorf("err should be '%s', got: '%s'", errFake, err)
	}
	if result != mockResult {
		t.Errorf("resp should be '%s'; got '%s'", mockResult, result)
	}

}

func TestExclude(t *testing.T) {
	testTable := []struct {
		desc     string
		matchers []string
		ops      []*Op
		expect   int
	}{
		{
			desc:     "Can exclude none",
			matchers: []string{"hello"},
			ops: []*Op{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
			},
			expect: 3,
		},
		{
			desc:     "Can exclude one",
			matchers: []string{"hi"},
			ops:      []*Op{{Identifier: "hi"}},
			expect:   0,
		},
		{
			desc:     "Can exclude two",
			matchers: []string{"hi", "sup"},
			ops: []*Op{
				{Identifier: "hi"},
				{Identifier: "sup"},
			},
			expect: 0,
		},
		{
			desc:     "Can exclude many and and ignore one",
			matchers: []string{"exclude1", "exclude2", "exclude3"},
			ops: []*Op{
				{Identifier: "exclude1"},
				{Identifier: "exclude2"},
				{Identifier: "exclude3"},
				{Identifier: "keep"},
			},
			expect: 1,
		},
		{
			desc:     "Can exclude glob *",
			matchers: []string{"exclude*"},
			ops: []*Op{
				{Identifier: "exclude1"},
				{Identifier: "exclude2"},
				{Identifier: "keep"},
			},
			expect: 1,
		},
	}

	for _, tc := range testTable {
		res, err := Exclude(tc.matchers, tc.ops)
		assert.Nil(t, err)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}

func TestSelect(t *testing.T) {
	testTable := []struct {
		desc     string
		matchers []string
		ops      []*Op
		expect   int
	}{
		{
			desc:     "Can select none",
			matchers: []string{"hello"},
			ops: []*Op{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
			},
			expect: 0,
		},
		{
			desc:     "Can select one",
			matchers: []string{"match"},
			ops: []*Op{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
				{Identifier: "match"}},
			expect: 1,
		},
		{
			desc:     "Can select two",
			matchers: []string{"match1", "match2"},
			ops: []*Op{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
				{Identifier: "match1"},
				{Identifier: "match2"},
			},
			expect: 2,
		},
		{
			desc:     "Can select many regardless of order",
			matchers: []string{"select1", "select2", "select3"},
			ops: []*Op{
				{Identifier: "skip1"},
				{Identifier: "select2"},
				{Identifier: "skip2"},
				{Identifier: "skip3"},
				{Identifier: "select3"},
				{Identifier: "select1"},
			},
			expect: 3,
		},
		{
			desc:     "Can select glob *",
			matchers: []string{"select*"},
			ops: []*Op{
				{Identifier: "skip1"},
				{Identifier: "select2"},
				{Identifier: "skip2"},
				{Identifier: "skip3"},
				{Identifier: "select3"},
				{Identifier: "select1"},
			},
			expect: 3,
		},
	}

	for _, tc := range testTable {
		res, err := Select(tc.matchers, tc.ops)
		assert.Nil(t, err)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}
func Test_StatusCounts(t *testing.T) {
	testTable := []struct {
		desc   string
		ops    []*Op
		expect map[Status]int
	}{
		{
			desc: "Statuses sums statuses",
			ops: []*Op{
				{Status: Success},
				{Status: Unknown},
				{Status: Success},
				{Status: Fail},
				{Status: Success},
			},
			expect: map[Status]int{
				Success: 3,
				Unknown: 1,
				Fail:    1,
			},
		},
		{
			desc: "returns an error if a op doesn't have a status",
			ops: []*Op{
				{Status: Unknown},
				{Status: Success},
				{Status: ""},
				{Status: Success},
				{Status: Fail},
				{Status: Success},
			},
			expect: nil,
		},
	}

	for _, tc := range testTable {
		statuses, err := StatusCounts(tc.ops)
		assert.Equal(t, tc.expect, statuses)
		if tc.expect == nil {
			assert.Error(t, err)
			break
		}
		assert.NoError(t, err)
	}
}
