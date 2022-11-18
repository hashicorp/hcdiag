package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterArgs(t *testing.T) {
	tcs := []struct {
		name     string
		flagname string
		filters  []string
		expect   string
	}{
		{
			name:     "to avoid duplicating product logic, invalid filters should work",
			flagname: "floob",
			filters:  []string{"one", "two"},
			expect:   " -floob=one -floob=two",
		},
		{
			name:     "test some vault targets",
			flagname: "target",
			filters:  []string{"pprof", "metrics"},
			expect:   " -target=pprof -target=metrics",
		},
		{
			name:     "test empty filters",
			flagname: "capture",
			filters:  []string{},
			expect:   "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result := filterArgs(tc.flagname, tc.filters)
			assert.Equal(t, tc.expect, result)
		})
	}
}
