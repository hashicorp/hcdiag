package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TarGz(sourceDir string, destFileName string) error

// InterfaceToJSON(mapVar map[string]interface{}) ([]byte, error)

// JSONToFile(JSON []byte, outFile string) error

func TestSplitFilepath(t *testing.T) {
	// set up a table of test cases
	abs, err := filepath.Abs(".")
	if err != nil {
		t.Errorf("error getting absolute path: %s", err)
		return
	}
	testTable := []map[string]string{
		{
			"path": "coolfile", // input
			"dir":  ".",        // expected dir
			"file": "coolfile", // expected file
		},
		{
			"path": "cooldir/coolfile",
			"dir":  "cooldir/",
			"file": "coolfile",
		},
		{
			"path": filepath.Join(abs, "cooldir/coolfile"),
			"dir":  filepath.Join(abs, "cooldir") + "/",
			"file": "coolfile",
		},
	}

	// SplitFilepath() does os.Stat to determine if something is a directory,
	// so we need to actually create the file(s).
	err = os.Mkdir("cooldir", 0755)
	if err != nil {
		t.Errorf("error creating cooldir/coolfile: %s", err)
		return
	}
	defer os.RemoveAll("cooldir")
	for _, data := range testTable {
		f := data["path"]
		_, err = os.Create(f)
		if err != nil {
			t.Errorf("error creating %s: %s", f, err)
			return
		}
		defer os.Remove(f)
	}

	// Validate our test results
	for _, data := range testTable {
		dir, file := SplitFilepath(data["path"])
		if err != nil {
			t.Errorf("error from SplitFilepath: %s", err)
		}
		if dir != data["dir"] {
			t.Errorf("Expected dir: '%s'; got: '%s'", data["dir"], dir)
		}
		if file != data["file"] {
			t.Errorf("Expected file: '%s'; got: '%s'", data["file"], file)
		}
	}
}

func TestFindInInterface(t *testing.T) {
	bts := []byte(`{"one": {"two": {"three": "cool_value"}}}`)
	var iface interface{}
	json.Unmarshal(bts, &iface)

	i, err := FindInInterface(iface, "one", "two", "three")
	if err != nil {
		t.Errorf("Failed to find 'cool_value' in %#v: %s", iface, err)
	}
	str, ok := i.(string)
	if !ok {
		t.Errorf("Failed to cast '%#v' as string", i)
	}
	if str != "cool_value" {
		t.Errorf("Expected 'cool_value'; got: '%s'", str)
	}
}
