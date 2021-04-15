package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const (
	testDir          = "./testfrom"
	testFile         = testDir + "/testfile"
	testFileContents = "hello"
	testDestination  = "./testdest"
)

func setupFiles(t *testing.T) func(t *testing.T) {
	// create "./testfrom"
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Error(err)
	}

	// create "./testfrom/testfile"
	f, err := os.Create(testFile)
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
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCopyDir(t *testing.T) {
	cleanup := setupFiles(t)
	defer cleanup(t)
	defer os.RemoveAll(testDestination)

	// this also implicitly tests CopyFile()
	err := CopyDir(testDestination, testDir)
	if err != nil {
		t.Error(err)
	}

	// confirm the file exists in the right place and has the right contents
	expectedLocation := "./testdest/testfrom/testfile"
	content, err := ioutil.ReadFile(expectedLocation)
	if err != nil {
		t.Error(err)
	}
	if string(content) != testFileContents {
		t.Errorf("expected contents to be '%s'; got '%s'", testFileContents, content)
	}

}

func TestCopyDirNotExists(t *testing.T) {
	err := CopyDir(testDestination, "not-a-real-dir")
	strErr := fmt.Sprintf("%s", err)
	expect := "no such file or directory"
	if !strings.Contains(strErr, expect) {
		t.Errorf("expected error to include '%s'; got: '%s'", expect, err)
	}
}

func TestCopyFileNotExists(t *testing.T) {
	err := CopyFile(testDestination, "not-a-real-file")
	strErr := fmt.Sprintf("%s", err)
	expect := "no such file or directory"
	if !strings.Contains(strErr, expect) {
		t.Errorf("expected error to include '%s'; got: '%s'", expect, err)
	}
}
