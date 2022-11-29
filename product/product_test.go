// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package product

import (
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/hcdiag/runner"
	"github.com/stretchr/testify/assert"
)

var _ runner.Runner = mockRunner{}

type mockRunner struct {
	id string
}

func (m mockRunner) ID() string { return m.id }
func (m mockRunner) Run() op.Op { return op.Op{} }

func TestFilters(t *testing.T) {
	testTable := []struct {
		desc    string
		product *Product
		expect  []runner.Runner
	}{
		{
			desc: "Handles empty ops and empty filters",
			product: &Product{
				Runners: []runner.Runner{},
			},
			expect: []runner.Runner{},
		},
		{
			desc: "Handles empty ops with non-empty filters",
			product: &Product{
				Runners:  []runner.Runner{},
				Excludes: []string{"hello"},
			},
			expect: []runner.Runner{},
		},
		{
			desc: "Handles nil filters",
			product: &Product{
				Runners: []runner.Runner{mockRunner{id: "still here"}},
			},
			expect: []runner.Runner{mockRunner{id: "still here"}},
		},
		{
			desc: "Handles nil runners",
			product: &Product{
				Excludes: []string{"nope"},
			},
			expect: []runner.Runner{},
		},
		{
			desc: "Handles empty filters",
			product: &Product{
				Runners: []runner.Runner{
					mockRunner{id: "still here"},
				},
				Excludes: []string{},
				Selects:  []string{},
			},
			expect: []runner.Runner{mockRunner{id: "still here"}},
		},
		{
			desc: "Applies matching excludes",
			product: &Product{
				Runners: []runner.Runner{
					mockRunner{id: "goodbye"},
				},
				Excludes: []string{"goodbye"},
			},
			expect: []runner.Runner{},
		},
		{
			desc: "Does not apply non-matching excludes",
			product: &Product{
				Runners:  []runner.Runner{mockRunner{id: "goodbye"}},
				Excludes: []string{"hello"},
			},
			expect: []runner.Runner{mockRunner{id: "goodbye"}},
		},
		{
			desc: "Applies matching Selects",
			product: &Product{
				Runners: []runner.Runner{
					mockRunner{id: "goodbye"},
					mockRunner{id: "hello"},
				},
				Selects: []string{"hello"},
			},
			expect: []runner.Runner{mockRunner{id: "hello"}},
		},
		{
			desc: "Ignores excludes when Selects are present, and ignores order",
			product: &Product{
				Runners: []runner.Runner{
					mockRunner{id: "select3"},
					mockRunner{id: "select1"},
					mockRunner{id: "goodbye"},
					mockRunner{id: "select2"},
				},
				Excludes: []string{"select2", "select3"},
				Selects:  []string{"select2", "select1", "select3"},
			},
			expect: []runner.Runner{
				mockRunner{id: "select3"},
				mockRunner{id: "select1"},
				mockRunner{id: "select2"},
			},
		},
	}

	for _, tc := range testTable {
		err := tc.product.Filter()
		assert.Nil(t, err)
		assert.NotNil(t, tc.product.Runners)
		assert.Equal(t, tc.expect, tc.product.Runners, tc.desc)
	}

}

func TestFilterErrors(t *testing.T) {
	testTable := []struct {
		desc    string
		product *Product
		expect  string
	}{
		{
			desc: "Select returns error when pattern is malformed",
			product: &Product{
				Runners: []runner.Runner{mockRunner{id: "ignoreme"}},
				Selects: []string{"mal[formed"},
			},
			expect: "filter error: 'syntax error in pattern'",
		},
		{
			desc: "Exclude returns error when pattern is malformed",
			product: &Product{
				Runners:  []runner.Runner{mockRunner{id: "ignoreme"}},
				Excludes: []string{"mal[formed"},
			},
			expect: "filter error: 'syntax error in pattern'",
		},
	}

	for _, tc := range testTable {
		err := tc.product.Filter()
		if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), tc.expect)
		}
	}
}
