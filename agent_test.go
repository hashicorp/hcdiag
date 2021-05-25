package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/seeker"
)

// TODO: abstract away filesystem-related actions,
// so mocks can be used instead of actually writing files?
// that would also allow us to run these tests in parallel if we wish.

func TestNewAgent(t *testing.T) {
	a := NewAgent(hclog.Default())
	assert.NotNil(t, a)
}

func TestParsesFlags(t *testing.T) {
	// not testing all flags, just that one is parsed appropriately
	a := NewAgent(hclog.Default())
	err := a.ParseFlags([]string{"-dryrun"})
	if err != nil {
		t.Error(err)
	}
	if !a.Dryrun {
		t.Error("-dryrun should enable Dryrun")
	}
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
	a := NewAgent(hclog.Default())
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

	// build -includes string
	var includeStr []string
	for _, data := range testTable {
		path := filepath.Join("tests", "resources", data["path"])
		includeStr = append(includeStr, path)
	}

	// basic Agent setup
	d := NewAgent(hclog.Default())
	// the args here now amount to:
	// -includes 'tests/resources/file.0,tests/resources/dir1/file.1,tests/resources/dir2/file*'
	includes := []string{"-includes", strings.Join(includeStr, ",")}
	if err := d.ParseFlags(includes); err != nil {
		t.Errorf("Error parsing flags: %s", err)
	}
	if err := d.CreateTemp(); err != nil {
		t.Errorf("Error creating tmpDir: %s", err)
	}
	defer d.Cleanup()

	// execute what we're aiming to test
	if err := d.CopyIncludes(); err != nil {
		t.Errorf("Error copying includes: %s", err)
	}

	// verify expected file locations
	for _, data := range testTable {
		expect := filepath.Join(d.tmpDir, "includes", "tests", "resources", data["expect"])
		if _, err := os.Stat(expect); err != nil {
			t.Errorf("Expect %s to exist, got error: %s", expect, err)
		}
	}
}

func TestGetSeekers(t *testing.T) {
	a := Agent{l: hclog.Default()}

	// no product Seekers, host only
	err := a.GetSeekers()
	if err != nil {
		t.Errorf("Error getting seekers: #{err}")
	}
	if len(a.seekers) != 1 {
		t.Errorf("Expected 1 Seeker; got: %d", len(a.seekers))
	}

	// include a product's Seekers
	a.Nomad = true
	err = a.GetSeekers() // replaces a.seekers, does not append.
	if err != nil {
		t.Errorf("Error getting seekers: #{err}")
	}
	if len(a.seekers) <= 1 {
		t.Errorf("Expected >1 Seeker; got: %d", len(a.seekers))
	}
}

func TestRunSeekers(t *testing.T) {
	a := Agent{
		l:       hclog.Default(),
		results: make(map[string]map[string]interface{}),
	}

	if err := a.RunSeekers(); err != nil {
		t.Errorf("Error running Seekers: %s", err)
	}
	// FIXME(mkcp): This host-host key is super awkward, need to work on the host some more
	r, ok := a.results["host"]["host"]
	if !ok {
		t.Error("Expected 'host' in results, not found")
	}
	if _, ok := r.(*seeker.Seeker); !ok {
		t.Errorf("Expected 'host' result to be a Seeker; got: %#v", r)
	}
}

func TestWriteOutput(t *testing.T) {
	a := Agent{
		l:       hclog.Default(),
		results: make(map[string]map[string]interface{}),
	}

	testOut := "test.tar.gz"
	a.Outfile = testOut // ordinarily would come from ParseFlags() but see bottom of this file...
	a.CreateTemp()
	defer a.Cleanup()
	defer os.Remove(testOut)

	if err := a.WriteOutput(); err != nil {
		t.Errorf("Error writing outputs: %s", err)
	}

	expectFiles := []string{
		filepath.Join(a.tmpDir, "Manifest.json"),
		filepath.Join(a.tmpDir, "Results.json"),
		testOut,
	}
	for _, f := range expectFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("Missing file %s: %s", f, err)
		}
	}
}
