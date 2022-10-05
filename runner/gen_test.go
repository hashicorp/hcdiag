package runner

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/stretchr/testify/assert"
)

func Test_GenRun(t *testing.T) {
	gen := NewGen("test generator",
		[]string{"echo hi", "echo hi2"},
		Commander{
			Format: "string",
		}, "just testing stuff")

	result := gen.Run()
	assert.Len(t, result.Result, 2)
	spew.Dump(result)
}
