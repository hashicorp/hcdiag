// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
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
	err := json.Unmarshal(bts, &iface)
	if err != nil {
		t.Error("failed to unmarshal", "error", err.Error())
	}

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

func TestIsInRange(t *testing.T) {
	testTable := []struct {
		desc                 string
		target, since, until time.Time
		expect               bool
	}{
		{
			desc:   "Target within range is valid",
			target: time.Date(2000, 0, 0, 0, 0, 0, 0, &time.Location{}),
			since:  time.Date(1977, 0, 0, 0, 0, 0, 0, &time.Location{}),
			until:  time.Date(2200, 0, 0, 0, 0, 0, 0, &time.Location{}),
			expect: true,
		},
		{
			desc:   "Target after range is invalid",
			target: time.Date(2300, 0, 0, 0, 0, 0, 0, &time.Location{}),
			since:  time.Date(1977, 0, 0, 0, 0, 0, 0, &time.Location{}),
			until:  time.Date(2200, 0, 0, 0, 0, 0, 0, &time.Location{}),
			expect: false,
		},
		{
			desc:   "Target before range is invalid",
			target: time.Date(1800, 0, 0, 0, 0, 0, 0, &time.Location{}),
			since:  time.Date(1977, 0, 0, 0, 0, 0, 0, &time.Location{}),
			until:  time.Date(2200, 0, 0, 0, 0, 0, 0, &time.Location{}),
			expect: false,
		},
		{
			desc:   "Zeroed `since` is always in range",
			target: time.Now(),
			expect: true,
		},
		{
			desc:   "Zeroed `until` includes recent and/or actively-written-until target",
			target: time.Now(),
			since:  time.Date(1977, 0, 0, 0, 0, 0, 0, &time.Location{}),
			expect: true,
		},
		{
			desc:   "Zeroed `until` does not include target before `since`",
			target: time.Date(1800, 0, 0, 0, 0, 0, 0, &time.Location{}),
			since:  time.Date(1977, 0, 0, 0, 0, 0, 0, &time.Location{}),
			expect: false,
		},
	}

	for _, c := range testTable {
		res := IsInRange(c.target, c.since, c.until)
		assert.Equal(t, res, c.expect, c.desc)
	}
}

func Test_getTarRelativePathName(t *testing.T) {
	type arguments struct {
		baseName string
		filePath string
		fileRoot string
	}
	testCases := []struct {
		name     string
		args     arguments
		expected string
	}{
		{
			name: "Test Source Files that are not Nested in File Root",
			args: arguments{
				baseName: "hcdiag0123456",
				filePath: "/tmp/a/b/c/results.json",
				fileRoot: "/tmp/a/b/c",
			},
			expected: "hcdiag0123456/results.json",
		},
		{
			name: "Test Source Files that are Nested in File Root",
			args: arguments{
				baseName: "hcdiag0123456",
				filePath: "/tmp/a/b/c/d/e/f/results.json",
				fileRoot: "/tmp/a/b/c",
			},
			expected: "hcdiag0123456/d/e/f/results.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getTarRelativePathName(tc.args.baseName, tc.args.filePath, tc.args.fileRoot)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHostCommandExists(t *testing.T) {
	tt := []struct {
		desc    string
		command string
		expect  bool
	}{
		{
			desc:    "test ls command",
			command: "ls",
			expect:  true,
		},
		{
			desc:    "ensure additional args are not tested",
			command: "ls fooblarbalurg sdlfkj",
			expect:  true,
		},
		{
			desc:    "ensure additional args are not tested 2",
			command: "fooblarbalurg ls pwd",
			expect:  false,
		},
		{
			desc:    "nonexistent commands should return false",
			command: "fooblarbalurg",
			expect:  false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			result, _ := HostCommandExists(tc.command)
			assert.Equal(t, result, tc.expect)
		})
	}
}

func TestCreateAndCleanupTemporaryDirectory(t *testing.T) {
	tmp, cleanup, err := CreateTemp(".")
	if err != nil {
		t.Errorf("failed to create temporary directory, err=%s", err)
	}

	fileInfo, err := os.Stat(tmp)
	if err != nil {
		t.Errorf("error checking for temp dir: %s", err)
	}
	if !fileInfo.IsDir() {
		t.Error("temporary directory was not created")
	}

	// test cleanup
	cleanup(hclog.NewNullLogger())
	if err != nil {
		t.Errorf("error while cleaning up temporary directory: %s", err)
	}

	_, err = os.Stat(tmp)
	if err == nil {
		t.Errorf("error checking for temp dir: %s", err)
	}
}

// FIXME(mkcp): Ensure the since and until works with modtime properly
// func TestFilterWalk(t *testing.T) {
// 	testTable := []struct{
// 		filter   string
// 		since     time.Time
// 		until       time.Time
// 		testCase func(t *testing.T)
// 		expect   what
// 	}{
// 		{
// 			filter: "",
// 			since:   time.Time{},
// 			until:     time.Time{},
// 		},
// 	}
// 	for _, test := range testTable {
// 		path := GenerateAbsolutePathIntoTestsResources()
// 		res, err := FilterWalk(path, test.filter, test.since, test.until)
// 		assert.NoError(t, err)
// 		for _, r := range res {
// 			expect == r
// 		}
// 	}
// }
