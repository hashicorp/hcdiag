package agent

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/products"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

// TODO: abstract away filesystem-related actions,
// so mocks can be used instead of actually writing files?
// that would also allow us to run these tests in parallel if we wish.

func TestNewAgent(t *testing.T) {
	a := NewAgent(Config{}, hclog.Default())
	assert.NotNil(t, a)
}

func TestStartAndEnd(t *testing.T) {
	a := Agent{l: hclog.Default()}

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
	defer a.Cleanup()

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
	d := Agent{l: hclog.Default()}

	if err = d.CreateTemp(); err != nil {
		t.Errorf("Error creating tmpDir: %s", err)
	}

	if _, err = os.Stat(d.tmpDir); err != nil {
		t.Errorf("Error checking for temp dir: %s", err)
	}

	if err = d.Cleanup(); err != nil {
		t.Errorf("Cleanup error: %s", err)
	}

	_, err = os.Stat(d.tmpDir)
	if !os.IsNotExist(err) {
		t.Errorf("Got unexpected error when validating that tmpDir was removed: %s", err)
	}
}

func TestCopyIncludes(t *testing.T) {
	t.Skip("There's a bug in include that we need to fix and re-enable this. See ENGSYS-1251")
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
		path, err := filepath.Abs(path)
		assert.NoError(t, err)
		println(path)
		includeStr = append(includeStr, path)
	}

	cfg := Config{Includes: includeStr}
	a := NewAgent(cfg, hclog.Default())
	err := a.CreateTemp()
	assert.NoError(t, err, "Error creating tmpDir")
	// defer a.Cleanup()

	// execute what we're aiming to test
	err = a.CopyIncludes()
	assert.NoError(t, err, "could not copy includes")

	// verify expected file locations
	for _, data := range testTable {
		expect := filepath.Join(a.tmpDir, "includes", "tests", "resources", data["expect"])
		if _, err := os.Stat(expect); err != nil {
			t.Errorf("Expect %s to exist, got error: %s", expect, err)
		}
	}
}

func TestRunProducts(t *testing.T) {
	pCfg := products.Config{OS: "auto"}
	p := make(map[string]*products.Product)
	p["host"] = products.NewHost(pCfg)
	a := Agent{
		l:        hclog.Default(),
		products: p,
		results:  make(map[string]map[string]interface{}),
	}

	err := a.RunProducts()
	assert.NoError(t, err)
	assert.Len(t, a.products, 1, "has one product")
	assert.NotNil(t, a.products["host"], "product is under \"host\" key")
}

func TestWriteOutput(t *testing.T) {
	a := Agent{
		l:       hclog.Default(),
		results: make(map[string]map[string]interface{}),
	}

	testOut := "test"
	resultsDest := a.DestinationFileName()
	a.Config.Outfile = testOut
	err := a.CreateTemp()
	if err != nil {
		t.Errorf("failed to create tempDir, err=%s", err)
	}
	defer a.Cleanup()
	defer os.Remove(resultsDest)

	if err := a.WriteOutput(resultsDest); err != nil {
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
