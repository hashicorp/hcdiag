package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcdiag/hcl"

	"github.com/hashicorp/hcdiag/op"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: abstract away filesystem-related actions,
// so mocks can be used instead of actually writing files?
// that would also allow us to run these tests in parallel if we wish.

// Wraps NewAgent(Config{}, hclog.Default()) for testing
func newTestAgent(t *testing.T) *Agent {
	t.Helper()
	a, err := NewAgent(Config{}, hclog.Default())
	require.NoError(t, err, "Error new test Agent")
	require.NotNil(t, a)
	return a
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

	if err := a.CreateTemp(); err != nil {
		t.Errorf("Failed creating temp dir: %s", err)
	}

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

	if err = a.CreateTemp(); err != nil {
		t.Errorf("Error creating tmpDir: %s", err)
	}

	if _, err = os.Stat(a.tmpDir); err != nil {
		t.Errorf("Error checking for temp dir: %s", err)
	}

	if err = a.Cleanup(); err != nil {
		t.Errorf("Cleanup error: %s", err)
	}

	_, err = os.Stat(a.tmpDir)
	if !os.IsNotExist(err) {
		t.Errorf("Got unexpected error when validating that tmpDir was removed: %s", err)
	}
}

func TestCopyIncludes(t *testing.T) {
	// set up a table of test cases
	// these dirs/files are checked in to this repo under tests/resources/
	testTable := []map[string]string{
		{
			"path":   "file.0",
			"expect": "file.0",
		},
		{
			"path":   "dir1",
			"expect": filepath.Join("dir1", "file.1"),
		},
		{
			"path":   filepath.Join("dir2", "file*"),
			"expect": filepath.Join("dir2", "file.2"),
		},
	}

	var includeStr []string
	for _, data := range testTable {
		path := filepath.Join("../", "tests", "resources", data["path"])
		absPath, err := filepath.Abs(path)
		assert.NoError(t, err)
		includeStr = append(includeStr, absPath)
	}

	cfg := Config{Includes: includeStr}
	a, err := NewAgent(cfg, hclog.Default())
	assert.NoError(t, err, "Error creating agent")
	err = a.CreateTemp()
	assert.NoError(t, err, "Error creating tmpDir")
	defer cleanupHelper(t, a)

	// execute what we're aiming to test
	err = a.CopyIncludes()
	assert.NoError(t, err, "Could not copy includes")

	// verify expected file locations
	for _, data := range testTable {
		path := filepath.Join("../", "tests", "resources", data["expect"])
		absPath, err := filepath.Abs(path)
		assert.NoError(t, err)
		expect := filepath.Join(a.tmpDir, "includes", absPath)
		if _, err := os.Stat(expect); err != nil {
			t.Errorf("Expect %s to exist, got error: %s", expect, err)
		}
	}
}

func TestRunProducts(t *testing.T) {
	l := hclog.Default()
	pCfg := product.Config{OS: "auto"}
	p := make(map[product.Name]*product.Product)
	a := newTestAgent(t)

	a.products = p
	h, err := product.NewHost(l, pCfg, &hcl.Host{})
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

	testOut := "."
	resultsDest := a.TempDir() + ".tar.gz"
	a.Config.Destination = testOut
	err := a.CreateTemp()
	if err != nil {
		t.Errorf("failed to create tempDir, err=%s", err)
	}

	defer func() {
		if err := a.Cleanup(); err != nil {
			a.l.Error("Failed to cleanup", "error", err)
		}
	}()

	defer func() {
		err := os.Remove(resultsDest)
		if err != nil {
			// Simply log this case because it's not an error in the function we're testing
			t.Logf("Error removing test results file: %s", resultsDest)
		}
	}()

	if err := a.WriteOutput(); err != nil {
		t.Errorf("Error writing outputs: %s", err)
	}

	expectFiles := []string{
		filepath.Join(a.tmpDir, "manifest.json"),
		filepath.Join(a.tmpDir, "results.json"),
		resultsDest,
	}
	for _, f := range expectFiles {
		// NOTE: OS X is case insensitive, so this test will never correctly check filename case on a dev machine
		_, err := os.Stat(f)
		assert.NoError(t, err, "Missing file %s", f)
	}
}

func TestSetup(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         Config
		expectedLen int
	}{
		{
			name: "Should only get host if no products enabled",
			cfg: Config{
				OS: "auto",
			},
			expectedLen: 1,
		},
		{
			name: "Should have host and nomad enabled",
			cfg: Config{
				Nomad: true,
				OS:    "auto",
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
	err := a.Cleanup()
	if err != nil {
		t.Errorf("Failed to clean up")
	}
}
