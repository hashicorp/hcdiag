package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
)

const directoryPerms = 0755

// CopyDir copies a directory and all of its contents into a target directory.
func CopyDir(to, src string) error {
	// get the absolute path, so we can remove it
	// to avoid copying the entire directory structure into the dest
	absPath, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for '%s': %s", src, err)
	}
	// we'll remove this from the target
	absBase := filepath.Dir(absPath)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		target := filepath.Join(to, strings.Replace(path, absBase, "", 1))
		if info != nil && info.IsDir() {
			hclog.L().Info("copying", "path", path, "to", target)
			return os.MkdirAll(target, directoryPerms)
		}
		return CopyFile(target, path)
	})
}

// CopyFile copies a file to a target file path.
func CopyFile(to, src string) error {
	hclog.L().Info("copying", "path", src, "to", to)

	// Ensure directories
	dir, _ := filepath.Split(to)
	err := os.MkdirAll(dir, directoryPerms)
	if err != nil {
		return err
	}

	// Open source file
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create destination file
	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()

	// Write source contents to destination
	_, err = io.Copy(w, r)
	return err
}
