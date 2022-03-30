//go:build functional

// end to end test
// expects `hcdiag` to be built and in PATH,
// along with consul, nomad, and vault
// which this test will run in the background.

package main_test

import (
	"context"
	"io/fs"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gulducat/go-run-programs/program"
	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

func TestFunctional(t *testing.T) {
	// run consul, nomad, and vault in the background,
	// and stop them when the tests are done.
	t.Log("starting consul, nomad, vault")
	stop, err := program.RunFromHCL(context.Background(), "go-run-programs.hcl")
	t.Cleanup(stop)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	testTable := map[string]struct {
		flags    []string // will be provided to hcdiag
		outFiles []string // we'll assert that these files exist
		skip     bool     // skip the sub-test or not
	}{
		"host": {
			flags:    []string{},
			outFiles: []string{},
			skip:     false,
		},
		"consul": {
			flags:    []string{"-consul"},
			outFiles: []string{"ConsulDebug.tar.gz"},
			skip:     false,
		},
		"nomad": {
			flags: []string{"-nomad"},
			// nomad is special and doesn't tar up its debug,
			// so we glob * for a file in its debug dir: "nomad*/index.json"
			outFiles: []string{filepath.Join("nomad*", "index.json")},
			skip:     false,
		},
		"vault-unix": {
			flags:    []string{"-vault"},
			outFiles: []string{"VaultDebug.tar.gz"},
			// TODO(gulducat): de-unique-ize when `vault debug` is fixed on windows
			// dave's pr: https://github.com/hashicorp/vault/pull/14399
			skip: runtime.GOOS == "windows",
		},
		"all-unix": {
			flags: []string{"-all"},
			outFiles: []string{
				"ConsulDebug.tar.gz",
				filepath.Join("nomad*", "index.json"),
				"VaultDebug.tar.gz",
			},
			skip: runtime.GOOS == "windows",
		},
		"vault-windows": {
			flags:    []string{"-vault", "-config", "exclude_debug.hcl"},
			outFiles: []string{},
			skip:     runtime.GOOS != "windows",
		},
	}

	for name, tc := range testTable {
		t.Run(name, func(t *testing.T) {
			// explicitly skipping here so the test output is not mysterious
			if tc.skip {
				t.SkipNow()
			}

			// get us a temp dir to put everything in, testing lib will clean it for us.
			tmpDir := t.TempDir()

			// run hcdiag
			output := runHCDiag(t, tmpDir, tc.flags)

			// ensure there was any output at all, "hcdiag" is semi-arbitrary
			assert.Contains(t, output, "hcdiag", "hcdiag output missing expected string 'hcdiag'")

			// for debugging, list files in the temp dir
			listFiles(t, tmpDir)

			// extract the .tar.gz file
			tarFile := findTar(t, tmpDir)
			extractedDir := unTar(t, tarFile, tmpDir)

			// the full filename should be in the command output
			assert.Contains(t, output, tarFile)

			listFiles(t, tmpDir)

			// ensure default and product-specific files are in our extracted directory
			// these files must always exist in the archive
			defaultFiles := []string{
				"Manifest.json",
				"Results.json",
			}
			files := append(defaultFiles, tc.outFiles...)
			assertFilesExist(t, extractedDir, files)

		})
	}
}

func listFiles(t *testing.T, tmpDir string) {
	t.Log("files in tmpDir:")
	err := filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			t.Log("  ", path)
		}
		return nil
	})
	assert.NoError(t, err)
}

func runHCDiag(t *testing.T, tmpDir string, flags []string) string {
	// assume "hcdiag" is already built and is in PATH
	// and always set -dest to keep the tests separate
	flags = append(flags, "-dest="+tmpDir)
	t.Log("running hcdiag:", flags)

	out, err := exec.Command("hcdiag", flags...).CombinedOutput()
	if !assert.NoError(t, err) {
		t.Fatalf("hcidag run failure, output:\n%s", out)
	}
	t.Logf("hcdiag output:\n%s", out)

	return string(out)
}

func findTar(t *testing.T, dir string) string {
	files, err := filepath.Glob(filepath.Join(dir, "*.tar.gz"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Len(t, files, 1, "expected one .tar.gz file") {
		t.FailNow()
	}
	return files[0]
}

func unTar(t *testing.T, file, dest string) string {
	t.Log("extracting archive:", file)
	tgz := archiver.NewTarGz()
	err := tgz.Unarchive(file, dest)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	// our extracted dir should be name of the file minus .tar.gz
	dir := strings.Replace(file, ".tar.gz", "", 1)
	t.Log("extracted to dir:", dir)
	return dir
}

func assertFilesExist(t *testing.T, dir string, files []string) {
	for _, file := range files {
		fullPath := filepath.Join(dir, file)

		// nomad dir has a timestamp in it, so we need to glob "nomad*"
		globFiles, err := filepath.Glob(fullPath)
		if !assert.NoError(t, err) {
			continue
		}
		assert.NotEmptyf(t, globFiles, "no files matching '%s'", file)

		for _, f := range globFiles {
			assert.FileExists(t, f)
		}
	}
}
