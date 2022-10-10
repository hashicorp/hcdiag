package op

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StatusCounts(t *testing.T) {
	testTable := []struct {
		desc   string
		ops    map[string]Op
		expect map[Status]int
	}{
		{
			desc: "Statuses sums statuses",
			ops: map[string]Op{
				"1": {Status: Success},
				"2": {Status: Unknown},
				"3": {Status: Success},
				"4": {Status: Fail},
				"5": {Status: Success},
			},
			expect: map[Status]int{
				Success: 3,
				Unknown: 1,
				Fail:    1,
			},
		},
		{
			desc: "returns an error if an op doesn't have a status",
			ops: map[string]Op{
				"1": {Status: Unknown},
				"2": {Status: Success},
				"3": {Status: ""},
				"4": {Status: Success},
				"5": {Status: Fail},
				"6": {Status: Success},
			},
			expect: nil,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.desc, func(t *testing.T) {
			statuses, err := StatusCounts(tc.ops)
			assert.Equal(t, tc.expect, statuses)
			if tc.expect == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_WalkStatuses(t *testing.T) {
	testTable := []struct {
		desc   string
		opMap  map[string]any
		expect map[Status]int
	}{
		{
			desc:   "Empty case",
			opMap:  map[string]any{},
			expect: map[Status]int{},
		},
		{
			desc: "Walk a single nested Op",
			opMap: map[string]any{
				"1": map[string]any{"1-nested": Op{Status: Success, Result: map[string]any{"foo": "bar"}}}},
			expect: map[Status]int{Success: 1},
		},
		{
			desc: "Walk multiple nested Ops",
			opMap: map[string]any{
				"0-flat": Op{Status: Success, Result: map[string]any{"already": "flat"}},
				"1": map[string]any{
					"1-nested": Op{Status: Success, Result: map[string]any{"foo": "bar"}},
					"1-nested-2": Op{Status: Success, Result: map[string]any{
						"2-nested":   Op{Status: Success, Result: map[string]any{"foo": "bar"}},
						"2-nested-2": Op{Status: Success, Result: map[string]any{"qux": "schnarg"}},
					}},
				},
			},
			expect: map[Status]int{Success: 5},
		},
	}

	for _, tc := range testTable {
		t.Run(tc.desc, func(t *testing.T) {
			result := WalkStatuses(tc.opMap)
			assert.Equal(t, tc.expect, result)
		})
	}
}
