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

func TestCountInSets(t *testing.T) {
	sets := map[string][]*Seeker {
		"zero": {},
		"one": {{}},
		"two": {{}, {}},
		"three": {{}, {}, {}},
	}
	expected := 6
	count    := CountInSets(sets)
	assert.Equal(t, count, expected)
}
