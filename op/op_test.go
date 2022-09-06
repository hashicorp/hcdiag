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
		statuses, err := StatusCounts(tc.ops)
		assert.Equal(t, tc.expect, statuses)
		if tc.expect == nil {
			assert.Error(t, err)
			break
		}
		assert.NoError(t, err)
	}
}
