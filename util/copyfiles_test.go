package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/hcdiag/redact"

	"github.com/stretchr/testify/assert"
)

const (
	testDir         = "./testfrom"
	testFile        = "testfile"
	testContent     = "hello"
	testDestination = "./testdest"
)

func setupFiles(t *testing.T, dir, file, content string) func(t *testing.T) error {
	// create directory
	absDir, err := filepath.Abs(dir)
	assert.NoError(t, err)
	err = os.MkdirAll(absDir, 0755)
	assert.NoError(t, err)

	// touch file
	absFile := filepath.Join(absDir, file)
	f, err := os.Create(absFile)
	assert.NoError(t, err)
	defer f.Close()

	// write to file
	_, err = f.WriteString(content)
	assert.NoError(t, err)

	// return a teardown function
	return func(t *testing.T) error {
		return os.RemoveAll(absDir)
	}
}

// TODO(mkcp): This code should be concerned with multi-file operations
// TODO(mkcp): Dynamically generate the write locations that are hard coded as package vars so tcs can run in parallel
func TestCopyDir(t *testing.T) {
	tcs := []struct {
		name        string
		readDir     string
		writeDir    string
		file        string
		content     string
		contentFile string
		redacts     func(*testing.T) []*redact.Redact
	}{
		{
			name:     "can copy directory",
			readDir:  testDir,
			writeDir: testDestination,
			file:     testFile,
			content:  testContent,
		},
	}

	for _, tc := range tcs {
		var reds []*redact.Redact
		// Init our params
		absDest, err := filepath.Abs(tc.writeDir)
		assert.NoError(t, err)
		absSrc, err := filepath.Abs(tc.readDir)
		assert.NoError(t, err)

		if tc.redacts != nil {
			reds = tc.redacts(t)
		}

		// Initialize our test directory, file, and its content
		// TODO(mkcp): Decouple directory setup, from file setup, to content writing so these all can be dynamically
		//  generated with a broader variety of data and run in parallel without writes colliding.
		cleanup := setupFiles(t, absSrc, tc.file, tc.content)

		// Copy the directory contents
		ce := CopyDir(absDest, absSrc, reds)
		assert.NoError(t, ce, tc.name)

		// TODO(mkcp): Compare files in the future, not strings
		if tc.content != "" {
			// confirm the file exists in the right place and has the right contents
			expectedLocation := filepath.Join(absSrc, tc.file)
			result, err := ioutil.ReadFile(expectedLocation)
			assert.NoError(t, err, tc.name)
			assert.Equal(t, tc.content, string(result), tc.name)
		}

		// Cleanup our test data
		cle := cleanup(t)
		assert.NoError(t, cle, tc.name)
	}
}

// TODO(mkcp): test tables and more cases
func TestCopyDirErrors(t *testing.T) {
	tcs := []struct {
		name string
		dest string
		src  string
	}{
		{
			name: "empty src and dest",
		},
		{
			name: "dir does not exist",
			dest: testDestination,
			src:  "dir-doesnt-exist2347890-12348079-",
		},
	}
	for _, tc := range tcs {
		err := CopyDir(tc.dest, tc.src, nil)
		assert.Error(t, err, tc.name)
	}
}

// TODO(mkcp): This code is exercised by the CopyDir cases above, but it should be tested in isolation with a variety
//  of cases.
func TestCopyFile(t *testing.T) {
	t.Skip()
}

// TODO(mkcp): more cases
func TestCopyFileErrors(t *testing.T) {
	tcs := []struct {
		name       string
		dest       string
		src        string
		redactions func(*testing.T) []*redact.Redact
	}{
		{
			name: "empty src and dest",
		},
		{
			name: "file does not exist",
			dest: testDestination,
			src:  "file-doesnt-exist1234712347091234",
		},
	}
	for _, tc := range tcs {
		var reds []*redact.Redact
		if tc.redactions != nil {
			reds = tc.redactions(t)
		}
		err := CopyFile(tc.dest, tc.src, reds)
		assert.Error(t, err, tc.name)
	}
}
