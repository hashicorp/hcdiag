// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/hcl"
	"github.com/hashicorp/hcdiag/util"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: abstract away filesystem-related actions,
// so mocks can be used instead of actually writing files?
// that would also allow us to run these tests in parallel if we wish.

// helperCreateTmpDir creates a tmpdir for testing
func helperCreateTmpDir(t *testing.T) (string, error) {
	tmp, err := util.CreateTemp(".")
	if err != nil {
		t.Errorf("failed to create tempDir, err=%s", err)
	}
	return tmp, nil
}

// Wraps NewAgent(Config{}, hclog.Default()) for testing
func newTestAgent(t *testing.T) *Agent {
	t.Helper()
	tmp, _ := helperCreateTmpDir(t)
	a, err := NewAgent(Config{TmpDir: tmp}, hclog.Default())
	require.NoError(t, err, "Error new test Agent")
	require.NotNil(t, a)
	return a
}

func TestNewAgentIncludesBackgroundContext(t *testing.T) {
	a := newTestAgent(t)
	assert.Equal(t, context.Background(), a.ctx)
}

func TestNewAgentWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tmp, _ := helperCreateTmpDir(t)
	a, err := NewAgentWithContext(ctx, Config{TmpDir: tmp}, hclog.Default())
	require.NoError(t, err, "Error new test Agent with context")
	require.NotNil(t, a)
	require.Equal(t, ctx, a.ctx)
}

func TestStartAndEnd(t *testing.T) {
	a := newTestAgent(t)

	// Start and End fields should be zero at first,
	// and Duration should be empty ""
	assert.Zero(t, a.Start, "Start value non-zero before start")
	assert.Zero(t, a.End, "End value non-zero before start")
	assert.Equal(t, "", a.Duration, "Duration value not an empty string before start")

	// recordEnd should set a time and calculate a duration
	a.recordEnd()

	assert.NotZero(t, a.End, "End value still zero after recordEnd()")
	assert.NotEqual(t, "", a.Duration, "Duration value still an empty string after recordEnd()")
}

func TestCreateTemp(t *testing.T) {
	a := newTestAgent(t)
	defer cleanupHelper(t, a)

	tmp, err := util.CreateTemp(".")
	if err != nil {
		t.Errorf("failed to create tempDir, err=%s", err)
	}
	a.tmpDir = tmp

	fileInfo, err := os.Stat(a.tmpDir)
	if err != nil {
		t.Errorf("Error checking for temp dir: %s", err)
	}
	if !fileInfo.IsDir() {
		t.Error("tmpDir is not a directory")
	}
}

func TestCreateTempAndCleanup(t *testing.T) {
	var err error
	a := newTestAgent(t)

	tmp, err := util.CreateTemp(".")
	if err != nil {
		t.Errorf("failed to create tempDir, err=%s", err)
	}
	a.tmpDir = tmp

	if _, err = os.Stat(a.tmpDir); err != nil {
		t.Errorf("Error checking for temp dir: %s", err)
	}

	cleanupHelper(t, a)

	_, err = os.Stat(a.tmpDir)
	if os.IsExist(err) {
		t.Errorf("Got unexpected error when validating that tmpDir was removed: %s", err)
	}
}

func TestRunProducts(t *testing.T) {
	l := hclog.Default()
	pCfg := product.Config{OS: "auto"}
	p := make(map[product.Name]*product.Product)
	a := newTestAgent(t)

	a.products = p
	h, err := product.NewHostWithContext(context.Background(), l, pCfg, &hcl.Host{})
	assert.NoError(t, err)
	p[product.Host] = h

	err1 := a.RunProducts()
	assert.NoError(t, err1)
	assert.Len(t, a.products, 1, "has one product")
	assert.NotNil(t, a.products["host"], "product is under \"host\" key")
}

func TestAgent_RecordManifest(t *testing.T) {
	t.Run("adds to ManifestOps when ops exist", func(t *testing.T) {
		// Setup
		testProduct := product.Host
		testResults := map[string]op.Op{
			"": {},
		}
		a := newTestAgent(t)

		a.results[testProduct] = testResults
		assert.NotEmptyf(t, a.results[testProduct], "test setup failure, no ops available")

		// Record and check
		a.RecordManifest()
		assert.NotEmptyf(t, a.ManifestOps, "no ops metadata added to manifest")
	})
}

func TestWriteOutput(t *testing.T) {
	a := newTestAgent(t)
	defer func() {
		cleanupHelper(t, a)
	}()

	if err := a.WriteOutput(); err != nil {
		t.Errorf("Error writing outputs: %s", err)
	}

	expectFiles := []string{
		filepath.Join(a.tmpDir, "manifest.json"),
		filepath.Join(a.tmpDir, "results.json"),
	}
	for _, f := range expectFiles {
		// NOTE: OS X is case insensitive, so this test will never correctly check filename case on a dev machine
		_, err := os.Stat(f)
		assert.NoError(t, err, "Missing file %s", f)
	}
}

func TestSetup(t *testing.T) {
	tmp, _ := helperCreateTmpDir(t)

	testCases := []struct {
		name        string
		cfg         Config
		expectedLen int
	}{
		{
			name: "Should only get host if no products enabled",
			cfg: Config{
				OS:     "auto",
				TmpDir: tmp,
			},
			expectedLen: 1,
		},
		{
			name: "Should have host and nomad enabled",
			cfg: Config{
				Nomad:  true,
				OS:     "auto",
				TmpDir: tmp,
			},
			expectedLen: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a, err := NewAgent(tc.cfg, hclog.Default())
			assert.NoError(t, err, "Error creating agent")

			err = a.Setup()
			assert.NoError(t, err)
			assert.Len(t, a.products, tc.expectedLen)
		})
	}
}

func cleanupHelper(t *testing.T, a *Agent) {
	err := os.RemoveAll(a.tmpDir)
	if err != nil {
		t.Errorf("Failed to clean up")
	}
}
