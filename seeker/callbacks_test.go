package seeker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeekerCallback(t *testing.T) {
	r := NewMockRunner()
	s := Seeker{Identifier: "mock", Runner: r}

	s.Run()
	assert.NoError(t, s.Error)
	assert.Equal(t, Success, s.Status)

	s.AddCallback(func() error { return errFake })
	_, err := s.Run()

	// correct error in desired places
	assert.ErrorIs(t, err, errFake)
	assert.ErrorIs(t, s.Error, errFake)
	assert.Equal(t, err.Error(), s.ErrString)
	assert.Equal(t, Fail, s.Status)
	// and still same result, in case it has useful debug info
	assert.Equal(t, mockResult, s.Result)
}

func TestResultNotContains(t *testing.T) {
	r := NewMockRunner()
	s := &Seeker{Identifier: "mock", Runner: r}

	e := errors.New("aw shucks")
	s.AddCallback(ResultNotContains(s, "mock", e))
	s.Run()

	assert.ErrorIs(t, s.Error, e)
}
