package op

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_StatusCounts(t *testing.T) {
	testTable := []struct {
		desc   string
		ops    map[string][]Op
		expect map[Status]int
	}{
		{
			desc: "Statuses sums statuses",
			ops: map[string][]Op{
				"1": {{Status: Success}},
				"2": {{Status: Unknown}},
				"3": {{Status: Success}},
				"4": {{Status: Fail}},
				"5": {{Status: Success}},
			},
			expect: map[Status]int{
				Success: 3,
				Unknown: 1,
				Fail:    1,
			},
		},
		{
			desc: "returns an error if an op doesn't have a status",
			ops: map[string][]Op{
				"1": {{Status: Unknown}},
				"2": {{Status: Success}},
				"3": {{Status: ""}},
				"4": {{Status: Success}},
				"5": {{Status: Fail}},
				"6": {{Status: Success}},
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

func Test_IsAllStatus(t *testing.T) {
	testTable := []struct {
		desc   string
		ops    []Op
		status Status
		expect bool
	}{
		{
			desc: "All Successes is true",
			ops: []Op{
				{Status: Success},
				{Status: Success},
				{Status: Success},
				{Status: Success},
			},
			status: Success,
			expect: true,
		},
		{
			desc: "All Skipped is true",
			ops: []Op{
				{Status: Skip},
				{Status: Skip},
				{Status: Skip},
			},
			status: Skip,
			expect: true,
		},
		{
			desc: "A single Failure is true",
			ops: []Op{
				{Status: Fail},
			},
			status: Fail,
			expect: true,
		},
		{
			desc: "All of wrong Status is false",
			ops: []Op{
				{Status: Skip},
				{Status: Skip},
				{Status: Skip},
			},
			status: Success,
			expect: false,
		},
		{
			desc: "Mixed Statuses are false",
			ops: []Op{
				{Status: Success},
				{Status: Unknown},
			},
			status: Fail,
			expect: false,
		},
		{
			desc: "Mixed Statuses are false 2",
			ops: []Op{
				{Status: Fail},
				{Status: Skip},
				{Status: Unknown},
			},
			status: Skip,
			expect: false,
		},
	}

	for _, tc := range testTable {
		t.Run(tc.desc, func(t *testing.T) {
			status := IsAllStatus(tc.status, tc.ops)
			assert.Equal(t, tc.expect, status)
		})
	}
}
