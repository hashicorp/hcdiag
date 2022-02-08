package seeker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	mockResult = "get mock'd"
	errFake    = errors.New("uh oh a fake error")
)

func NewMockRunner() *MockRunner {
	return &MockRunner{
		result: mockResult,
	}
}

type MockRunner struct {
	result string
	err    error
}

func (r MockRunner) Run() (interface{}, error) {
	return r.result, r.err
}

func TestSeekerRun(t *testing.T) {
	r := NewMockRunner()
	s := Seeker{Identifier: "mock", Runner: r}

	// no error
	result, err := s.Run()
	assert.Equal(t, mockResult, result)
	assert.Equal(t, mockResult, s.Result)
	assert.NoError(t, err)
	assert.NoError(t, s.Error)
	assert.Equal(t, Success, s.Status)

	// with error
	r.err = errFake
	_, err = s.Run()
	assert.ErrorIs(t, err, errFake)
	assert.ErrorIs(t, s.Error, errFake)
	assert.Equal(t, errFake.Error(), s.ErrString)
	assert.Equal(t, Fail, s.Status)
}

func TestExclude(t *testing.T) {
	testTable := []struct {
		desc     string
		matchers []string
		seekers  []*Seeker
		expect   int
	}{
		{
			desc:     "Can exclude none",
			matchers: []string{"hello"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
			},
			expect: 3,
		},
		{
			desc:     "Can exclude one",
			matchers: []string{"hi"},
			seekers:  []*Seeker{{Identifier: "hi"}},
			expect:   0,
		},
		{
			desc:     "Can exclude two",
			matchers: []string{"hi", "sup"},
			seekers: []*Seeker{
				{Identifier: "hi"},
				{Identifier: "sup"},
			},
			expect: 0,
		},
		{
			desc:     "Can exclude many and and ignore one",
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
			desc:     "Can exclude glob *",
			matchers: []string{"exclude*"},
			seekers: []*Seeker{
				{Identifier: "exclude1"},
				{Identifier: "exclude2"},
				{Identifier: "keep"},
			},
			expect: 1,
		},
	}

	for _, tc := range testTable {
		res, err := Exclude(tc.matchers, tc.seekers)
		assert.Nil(t, err)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}

func TestSelect(t *testing.T) {
	testTable := []struct {
		desc     string
		matchers []string
		seekers  []*Seeker
		expect   int
	}{
		{
			desc:     "Can select none",
			matchers: []string{"hello"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
			},
			expect: 0,
		},
		{
			desc:     "Can select one",
			matchers: []string{"match"},
			seekers: []*Seeker{
				{Identifier: "nope"},
				{Identifier: "nah"},
				{Identifier: "sry"},
				{Identifier: "match"}},
			expect: 1,
		},
		{
			desc:     "Can select two",
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
			desc:     "Can select many regardless of order",
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
			desc:     "Can select glob *",
			matchers: []string{"select*"},
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
	}

	for _, tc := range testTable {
		res, err := Select(tc.matchers, tc.seekers)
		assert.Nil(t, err)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}
