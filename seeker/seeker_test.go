package seeker

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	mockResult = "get mock'd"
	errFake    = errors.New("uh oh a fake error")
)

type MockRunner struct{}

func (r MockRunner) Run() (interface{}, error) {
	return mockResult, errFake
}

func TestSeekerRun(t *testing.T) {
	r := MockRunner{}
	s := Seeker{Identifier: "mock", Runner: r}

	result, err := s.Run()

	// assert that return values are also being stored as struct fields
	if s.Result != result {
		t.Errorf("returned result (%s) does not match Seeker Result field (%s)", result, s.Result)
	}
	if s.Error != err {
		t.Errorf("returned err (%s) does not match Seeker Error field (%s)", err, s.Error)
	}
	errStr := fmt.Sprintf("%s", err)
	if s.ErrString != errStr {
		t.Errorf("Seeker ErrString (%s) not formatted as expected (%s)", s.ErrString, errStr)
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
	testTable := []struct{
		desc     string
		matchers []string
		seekers  []*Seeker
		expect   int
	}{
		{
			desc: "Can exclude none",
			matchers: []string{"hello"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
			},
			expect: 3,
		},
		{
			desc: "Can exclude one",
			matchers: []string{"hi"},
			seekers: []*Seeker{{Identifier: "hi"}},
			expect: 0,
		},
		{
			desc: "Can exclude two",
			matchers: []string{"hi", "sup"},
			seekers: []*Seeker{
				{Identifier: "hi"},
				{Identifier: "sup"},
			},
			expect: 0,
		},
		{
			desc: "Can exclude many and and ignore one",
			matchers: []string{"exclude1", "exclude2", "exclude3"},
			seekers: []*Seeker{
				{Identifier: "exclude1"},
				{Identifier: "exclude2"},
				{Identifier: "exclude3"},
				{Identifier: "keep"},
			},
			expect: 1,
		},
		{
			desc:     "Can exclude with glob",
			matchers: []string{"hello*friend? hi"},
			seekers: []*Seeker{
				{Identifier: "hello there friend? hi"},
			},
			expect: 0,
		},
	}

	for _, tc := range testTable {
		res := Exclude(tc.matchers, tc.seekers)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}

func TestSelect(t *testing.T) {
	testTable := []struct{
		desc     string
		matchers []string
		seekers  []*Seeker
		expect   int
	}{
		{
			desc: "Can select none",
			matchers: []string{"hello"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
			},
			expect: 0,
		},
		{
			desc: "Can select one",
			matchers: []string{"match"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
				{Identifier: "match"}},
			expect: 1,
		},
		{
			desc: "Can select two",
			matchers: []string{"match1", "match2"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
				{Identifier: "match1"},
				{Identifier: "match2"},
			},
			expect: 2,
		},
		{
			desc: "Can select many regardless of order",
			matchers: []string{"select1", "select2", "select3"},
			seekers: []*Seeker{
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
			desc:     "Can select with glob",
			matchers: []string{"hello*friend? hi"},
			seekers: []*Seeker{
				{Identifier: "hello there friend? hi"},
				{Identifier: "not me though"},
			},
			expect: 1,
		},
	}

	for _, tc := range testTable {
		res := Select(tc.matchers, tc.seekers)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}
