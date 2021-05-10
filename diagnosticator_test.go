package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/host-diagnostics/seeker"
)

// TODO: abstract away filesystem-related actions,
// so mocks can be used instead of actually writing files?
// that would also allow us to run these tests in parallel if we wish.

func TestNewDiagnosticator(t *testing.T) {
	d := NewDiagnosticator(hclog.Default())
	if d.Start.IsZero() {
		t.Errorf("Start value still zero after start(): %s", d.Start)
	}
}

func TestParsesFlags(t *testing.T) {
	// not testing all flags, just that one is parsed appropriately
	d := NewDiagnosticator(hclog.Default())
	d.ParseFlags([]string{"-dryrun"})
	if !d.Dryrun {
		t.Error("-dryrun should enable Dryrun")
	}
}

func TestStartAndEnd(t *testing.T) {
	d := Diagnosticator{l: hclog.Default()}

	// Start and End fields should be nil at first,
	// and Duration should be empty ""
	if !d.Start.IsZero() {
		t.Errorf("Start value non-zero before start(): %s", d.Start)
	}
	if !d.End.IsZero() {
		t.Errorf("End value non-zero before start(): %s", d.Start)
	}
	if d.Duration != "" {
		t.Errorf("Duration value not an empty string before start(): %s", d.Duration)
	}

	// after start() and end(), the above values should be set to something
	d.start()
	if d.Start.IsZero() {
		t.Errorf("Start value still zero after start(): %s", d.Start)
	}
	d.end()
	if d.End.IsZero() {
		t.Errorf("End value still zero after start(): %s", d.Start)
	}
	if d.Duration == "" {
		t.Error("Duration value still an empty string after start()")
	}
}

func TestCreateTemp(t *testing.T) {
	d := NewDiagnosticator(hclog.Default())
	defer d.Cleanup()

	if err := d.CreateTemp(); err != nil {
		t.Errorf("Failed creating temp dir: %s", err)
	}

	fileInfo, err := os.Stat(d.tmpDir)
	if err != nil {
		t.Errorf("Error checking for temp dir: %s", err)
	}
	if !fileInfo.IsDir() {
		t.Error("tmpDir is not a directory")
	}
}

func TestCreateTempAndCleanup(t *testing.T) {
	var err error
	d := Diagnosticator{l: hclog.Default()}

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
	d := Diagnosticator{l: hclog.Default()}
	d.ParseFlags([]string{"-includes", "main.go,products,seeker/*"})
	d.CreateTemp()
	defer d.Cleanup()

	if err := d.CopyIncludes(); err != nil {
		t.Errorf("Error copying includes: %s", err)
	}

	expectFiles := []string{
		"main.go",
		filepath.Join("products", "products.go"),
		filepath.Join("seeker", "seeker.go"),
	}
	for _, f := range expectFiles {
		path := filepath.Join(d.tmpDir, "includes", f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Expect %s to exist, got error: %s", path, err)
		}
	}
}

func TestGetSeekers(t *testing.T) {
	d := Diagnosticator{l: hclog.Default()}

	// no product Seekers, host only
	d.GetSeekers()
	if len(d.seekers) != 1 {
		t.Errorf("Expected 1 Seeker; got: %d", len(d.seekers))
	}

	// include a product's Seekers
	d.Nomad = true
	d.GetSeekers() // replaces d.seekers, does not append.
	if len(d.seekers) <= 1 {
		t.Errorf("Expected >1 Seeker; got: %d", len(d.seekers))
	}
}

func TestRunSeekers(t *testing.T) {
	d := Diagnosticator{
		l:       hclog.Default(),
		results: make(map[string]interface{}),
	}

	if err := d.RunSeekers(); err != nil {
		t.Errorf("Error running Seekers: %s", err)
	}
	r, ok := d.results["host"]
	if !ok {
		t.Error("Expected 'host' in results, not found")
	}
	if _, ok := r.(*seeker.Seeker); !ok {
		t.Errorf("Expected 'host' result to be a Seeker; got: %#v", r)
	}
}

func TestWriteOutput(t *testing.T) {
	d := Diagnosticator{
		l:       hclog.Default(),
		results: make(map[string]interface{}),
	}

	testOut := "test.tar.gz"
	d.Outfile = testOut // ordinarily would come from ParseFlags() but see bottom of this file...
	d.CreateTemp()
	defer d.Cleanup()
	defer os.Remove(testOut)

	if err := d.WriteOutput(); err != nil {
		t.Errorf("Error writing outputs: %s", err)
	}

	expectFiles := []string{
		filepath.Join(d.tmpDir, "Manifest.json"),
		filepath.Join(d.tmpDir, "Results.json"),
		testOut,
	}
	for _, f := range expectFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("Missing file %s: %s", f, err)
		}
	}
}
