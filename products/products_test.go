package products

import (
	"github.com/hashicorp/host-diagnostics/seeker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilters(t *testing.T) {
	testTable := []struct {
		desc    string
		product *Product
		expect  []*seeker.Seeker
	}{
		{
			desc: "Handles empty seekers and empty filters",
			product: &Product{
				Seekers: []*seeker.Seeker{},
			},
			expect: []*seeker.Seeker{},
		},
		{
			desc: "Handles empty seekers with non-empty filters",
			product: &Product{
				Seekers:  []*seeker.Seeker{},
				excludes: []string{"hello"},
			},
			expect: []*seeker.Seeker{},
		},
		{
			desc: "Handles nil filters",
			product: &Product{
				Seekers: []*seeker.Seeker{{Identifier: "still here"}},
			},
			expect: []*seeker.Seeker{{Identifier: "still here"}},
		},
		{
			desc: "Handles nil seekers",
			product: &Product{
				excludes: []string{"nope"},
			},
			expect: []*seeker.Seeker{},
		},
		{
			desc: "Handles empty filters",
			product: &Product{
				Seekers: []*seeker.Seeker{
					{Identifier: "still here"},
				},
				excludes: []string{},
				selects:  []string{},
			},
			expect: []*seeker.Seeker{{Identifier: "still here"}},
		},
		{
			desc: "Applies matching excludes",
			product: &Product{
				Seekers: []*seeker.Seeker{
					{Identifier: "goodbye"},
				},
				excludes: []string{"goodbye"},
			},
			expect: []*seeker.Seeker{},
		},
		{
			desc: "Does not apply non-matching excludes",
			product: &Product{
				Seekers:  []*seeker.Seeker{{Identifier: "goodbye"}},
				excludes: []string{"hello"},
			},
			expect: []*seeker.Seeker{{Identifier: "goodbye"}},
		},
		{
			desc: "Applies matching selects",
			product: &Product{
				Seekers: []*seeker.Seeker{
					{Identifier: "goodbye"},
					{Identifier: "hello"},
				},
				selects: []string{"hello"},
			},
			expect: []*seeker.Seeker{{Identifier: "hello"}},
		},
		{
			desc: "Ignores excludes when selects are present, and ignores order",
			product: &Product{
				Seekers: []*seeker.Seeker{
					{Identifier: "select3"},
					{Identifier: "select1"},
					{Identifier: "goodbye"},
					{Identifier: "select2"},
				},
				excludes: []string{"select2", "select3"},
				selects:  []string{"select2", "select1", "select3"},
			},
			expect: []*seeker.Seeker{
				{Identifier: "select3"},
				{Identifier: "select1"},
				{Identifier: "select2"},
			},
		},
	}

	for _, tc := range testTable {
		tc.product.Filter()
		assert.NotNil(t, tc.product.Seekers)
		assert.Equal(t, tc.expect, tc.product.Seekers, tc.desc)
	}

}
