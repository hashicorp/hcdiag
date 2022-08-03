package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDir          = "./testfrom"
	testFile         = "testfile"
	testFileContents = "hello"
	testDestination  = "./testdest"
)

func setupFiles(t *testing.T) func(t *testing.T) {
	// create "./testfrom"
	absTestDir, err := filepath.Abs(testDir)
	assert.NoError(t, err)
	err = os.MkdirAll(absTestDir, 0755)
	if err != nil {
		t.Error(err)
	}

	// create "./testfrom/testfile"
	absTestFile := filepath.Join(absTestDir, testFile)
	f, err := os.Create(absTestFile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	// write "hello" to testfile
	_, err = f.WriteString(testFileContents)
	if err != nil {
		t.Error(err)
	}

	// return a teardown function to delete what we just made
	return func(t *testing.T) {
		err := os.RemoveAll(absTestDir)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCopyDir(t *testing.T) {
	absTestDestination, err := filepath.Abs(testDestination)
	assert.NoError(t, err)

	absTestDir, err := filepath.Abs(testDir)
	assert.NoError(t, err)

	cleanup := setupFiles(t)

	defer cleanup(t)
	defer os.RemoveAll(absTestDestination)

	// this also implicitly tests CopyFile()
	err = CopyDir(absTestDestination, absTestDir, nil)
	if err != nil {
		t.Error(err)
	}

	// confirm the file exists in the right place and has the right contents
	expectedLocation := filepath.Join(absTestDir, testFile)
	content, err := ioutil.ReadFile(expectedLocation)
	if err != nil {
		t.Error(err)
	}
	if string(content) != testFileContents {
		t.Errorf("expected contents to be '%s'; got '%s'", testFileContents, content)
	}

}

func TestCopyDirNotExists(t *testing.T) {
	err := CopyDir(testDestination, "not-a-real-dir", nil)
	strErr := fmt.Sprintf("%s", err)
	expect := "no such file or directory"
	if !strings.Contains(strErr, expect) {
		t.Errorf("expected error to include '%s'; got: '%s'", expect, err)
	}
}

func TestCopyFileNotExists(t *testing.T) {
	err := CopyFile(testDestination, "not-a-real-file", nil)
	strErr := fmt.Sprintf("%s", err)
	expect := "no such file or directory"
	if !strings.Contains(strErr, expect) {
		t.Errorf("expected error to include '%s'; got: '%s'", expect, err)
	}
}
