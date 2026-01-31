// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/agent"
	"github.com/hashicorp/hcdiag/op"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")

func TestSetTime(t *testing.T) {
	// This test is kind of a tautology because we're replacing the impl. to build the expected struct. But it holds
	// some value to protect against regressions.
	testDur, err := time.ParseDuration("48h")
	if !assert.NoError(t, err) {
		return
	}
	cfg := agent.Config{}
	now := time.Now()
	newCfg := setTime(cfg, now, testDur)

	expect := agent.Config{
		Since: now.Add(-testDur),
		Until: time.Time{}, // This is the default zero value, but we're just being extra explicit.
	}

	assert.Equal(t, newCfg, expect)
}

func Test_writeSummary(t *testing.T) {
	// NOTE: If you make changes to WriteSummary, you may break existing unit tests until the golden files are updated
	// to reflect your changes. To update them, run `go test ./command -update`, and then manually verify that the new
	// files under testdata/writeSummary look like you expect. If so, commit them to source control, and future
	// test executions should succeed.
	testCases := []struct {
		name        string
		resultsFile string
		manifestOps map[string][]agent.ManifestOp
	}{
		{
			name: "Test Header Only",
		},
		{
			name:        "Test with Products",
			resultsFile: "/this/is/a/test/path/bundle.tar.gz",
			manifestOps: map[string][]agent.ManifestOp{
				"consul": {
					{
						Status: op.Success,
					},
					{
						Status: op.Success,
					},
					{
						Status: op.Fail,
					},
					{
						Status: op.Skip,
					},
					{
						Status: op.Unknown,
					},
					{
						Status: op.Timeout,
					},
					{
						Status: op.Timeout,
					},
				},
				"nomad": {
					{
						Status: op.Fail,
					},
					{
						Status: op.Unknown,
					},
					{
						Status: op.Skip,
					},
					{
						Status: op.Canceled,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b := new(bytes.Buffer)

			err := writeSummary(b, tc.resultsFile, tc.manifestOps)

			assert.NoError(t, err)
			golden := filepath.Join("testdata/writeSummary", tc.name+".golden")

			if *update {
				writeErr := os.WriteFile(golden, b.Bytes(), 0644)
				if writeErr != nil {
					t.Errorf("Error writing golden file (%s): %s", golden, writeErr)
				}
			}

			expected, readErr := os.ReadFile(golden)
			if readErr != nil {
				t.Errorf("Error reading golden file (%s): %s", golden, readErr)
			}
			assert.Equal(t, expected, b.Bytes())
		})
	}
}

func Test_formatReportLine(t *testing.T) {
	testCases := []struct {
		name   string
		cells  []string
		expect string
	}{
		{
			name:   "Test Nil Input",
			cells:  nil,
			expect: "\n",
		},
		{
			name:   "Test Empty Input",
			cells:  []string{},
			expect: "\n",
		},
		{
			name:   "Test Sample Header Row",
			cells:  []string{"product", "success", "failed", "skip", "unknown", "total"},
			expect: "product\tsuccess\tfailed\tskip\tunknown\ttotal\t\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := formatReportLine(tc.cells...)
			assert.Equal(t, tc.expect, res, tc.name)
		})
	}
}
