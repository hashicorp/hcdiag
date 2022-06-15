package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcdiag/product"
	"github.com/stretchr/testify/assert"
)

// TODO: abstract away filesystem-related actions,
// so mocks can be used instead of actually writing files?
// that would also allow us to run these tests in parallel if we wish.

func TestNewAgent(t *testing.T) {
	a := NewAgent(Config{}, hclog.Default())
	assert.NotNil(t, a)
}

func TestStartAndEnd(t *testing.T) {
	a := NewAgent(Config{}, hclog.Default())

	// Start and End fields should be nil at first,
	// and Duration should be empty ""
	if !a.Start.IsZero() {
		t.Errorf("Start value non-zero before start(): %s", a.Start)
	}
	if !a.End.IsZero() {
		t.Errorf("End value non-zero before start(): %s", a.Start)
	}
	if a.Duration != "" {
		t.Errorf("Duration value not an empty string before start(): %s", a.Duration)
	}

	// recordEnd should set a time and calculate a duration
	a.recordEnd()
	if a.End.IsZero() {
		t.Errorf("End value still zero after start(): %s", a.Start)
	}
	if a.Duration == "" {
		t.Error("Duration value still an empty string after start()")
	}
}

func TestCreateTemp(t *testing.T) {
	a := NewAgent(Config{}, hclog.Default())
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
	a := NewAgent(Config{}, hclog.Default())

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
	a := NewAgent(cfg, hclog.Default())
	err := a.CreateTemp()
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
	pCfg := product.Config{OS: "auto"}
	p := make(map[string]*product.Product)
	p["host"] = product.NewHost(pCfg)

	a := NewAgent(Config{}, hclog.Default())
	a.products = p

	err := a.RunProducts()
	assert.NoError(t, err)
	assert.Len(t, a.products, 1, "has one product")
	assert.NotNil(t, a.products["host"], "product is under \"host\" key")
}

func TestAgent_RecordManifest(t *testing.T) {
	t.Run("adds to MetadataSeekers when seekers exist", func(t *testing.T) {
		// Setup
		testProduct := "host"
		a := NewAgent(Config{}, hclog.Default())
		pCfg := product.Config{OS: "auto"}
		p := make(map[string]*product.Product)
		p[testProduct] = product.NewHost(pCfg)
		a.products = p
		assert.NotEmptyf(t, a.products[testProduct].Seekers, "test setup failure, no seekers available")

		// Record and check
		a.RecordManifest()
		assert.NotEmptyf(t, a.ManifestSeekers, "no seekers metadata added to manifest")
	})
}

func TestWriteOutput(t *testing.T) {
	a := NewAgent(Config{}, hclog.Default())

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
		filepath.Join(a.tmpDir, "Manifest.json"),
		filepath.Join(a.tmpDir, "Results.json"),
		resultsDest,
	}
	for _, f := range expectFiles {
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
			a := NewAgent(tc.cfg, hclog.Default())
			p, err := a.Setup()
			assert.NoError(t, err)
			assert.Len(t, p, tc.expectedLen)
		})
	}
}

func TestParseHCL(t *testing.T) {
	testCases := []struct {
		name   string
		path   string
		expect Config
	}{
		{
			name:   "Empty config is valid",
			path:   "../tests/resources/config/empty.hcl",
			expect: Config{},
		},
		{
			name: "Host with no attributes is valid",
			path: "../tests/resources/config/host_no_seekers.hcl",
			expect: Config{
				Host: &HostConfig{},
			},
		},
		{
			name: "Host with one of each seeker is valid",
			path: "../tests/resources/config/host_each_seeker.hcl",
			expect: Config{
				Host: &HostConfig{
					Commands: []CommandConfig{
						{Run: "testing", Format: "string"},
					},
					Shells: []ShellConfig{
						{Run: "testing"},
					},
					GETs: []GETConfig{
						{Path: "/v1/api/lol"},
					},
					Copies: []CopyConfig{
						{Path: "./*", Since: "10h"},
					},
				},
			},
		},
		{
			name: "Host with multiple of a seeker type is valid",
			path: "../tests/resources/config/multi_seekers.hcl",
			expect: Config{
				Host: &HostConfig{
					Commands: []CommandConfig{
						{
							Run:    "testing",
							Format: "string",
						},
						{
							Run:    "another one",
							Format: "string",
						},
						{
							Run:    "do a thing",
							Format: "json",
						},
					},
				},
			},
		},
		{
			name: "Config with a host and one product with everything is valid",
			path: "../tests/resources/config/config.hcl",
			expect: Config{
				Host: &HostConfig{
					Commands: []CommandConfig{
						{Run: "ps aux", Format: "string"},
					},
				},
				Products: []*ProductConfig{
					{
						Name: "consul",
						Commands: []CommandConfig{
							{Run: "consul version", Format: "json"},
							{Run: "consul operator raft list-peers", Format: "json"},
						},
						Shells: []ShellConfig{
							{Run: "consul members | grep ."},
						},
						GETs: []GETConfig{
							{Path: "/v1/api/metrics?format=prometheus"},
						},
						Copies: []CopyConfig{
							{Path: "/another/test/log", Since: "240h"},
						},
						Excludes: []string{"consul some-awfully-long-command"},
						Selects: []string{
							"consul just this",
							"consul and this",
						},
					},
				},
			},
		},
		{
			name: "Config with multiple products is valid",
			path: "../tests/resources/config/multi_product.hcl",
			expect: Config{
				Products: []*ProductConfig{
					{
						Name:     "consul",
						Commands: []CommandConfig{{Run: "consul version", Format: "string"}},
					},
					{
						Name:     "nomad",
						Commands: []CommandConfig{{Run: "nomad version", Format: "string"}},
					},
					{
						Name:     "vault",
						Commands: []CommandConfig{{Run: "vault version", Format: "string"}},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ParseHCL(tc.path)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, res, tc.name)
		})
	}
}

func cleanupHelper(t *testing.T, a *Agent) {
	err := a.Cleanup()
	if err != nil {
		t.Errorf("Failed to clean up")
	}
}
