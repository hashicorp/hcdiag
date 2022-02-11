package util

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
)

// TarGz accepts a source directory and destination file name to archive and compress files.
func TarGz(sourceDir string, destFileName string) error {
	// tar
	destFile, err := os.Create(destFileName)
	if err != nil {
		hclog.L().Error("TarGz", "error creating tarball", err)
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			hclog.L().Warn(err.Error())
		}
	}(destFile)

	gzWriter := gzip.NewWriter(destFile)
	defer func(gzWriter *gzip.Writer) {
		err := gzWriter.Close()
		if err != nil {
			hclog.L().Warn(err.Error())
		}
	}(gzWriter)

	tarWriter := tar.NewWriter(gzWriter)
	defer func(tarWriter *tar.Writer) {
		err := tarWriter.Close()
		if err != nil {
			hclog.L().Warn(err.Error())
		}
	}(tarWriter)

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sourceFile, err := os.Open(path)
			if err != nil {
				hclog.L().Error("TarGz", "error opening source file", err)
				return err
			}
			defer sourceFile.Close()

			stat, err := sourceFile.Stat()
			if err != nil {
				hclog.L().Error("TarGz", "error checking source file", err)
				return err
			}

			header := &tar.Header{
				Name:    path,
				Size:    stat.Size(),
				Mode:    int64(stat.Mode()),
				ModTime: stat.ModTime(),
			}

			if err := tarWriter.WriteHeader(header); err != nil {
				hclog.L().Error("TarGz", "error writing header for tar", err)
				return err
			}

			if _, err := io.Copy(tarWriter, sourceFile); err != nil {
				hclog.L().Error("TarGz", "error copying file to tarball", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// WriteJSON converts an interface{} to JSON then writes to filePath.
func WriteJSON(iface interface{}, filePath string) error {
	// TODO: these funcs have their own logging in em, perhaps we don't want that?
	jsonBts, err := InterfaceToJSON(iface)
	if err != nil {
		return err
	}
	err = JSONToFile(jsonBts, filePath)
	if err != nil {
		return err
	}
	return nil
}

// InterfaceToJSON converts an interface{} to JSON.
func InterfaceToJSON(mapVar interface{}) ([]byte, error) {
	InfoJSON, err := json.MarshalIndent(mapVar, "", "    ")
	if err != nil {
		hclog.L().Error("InterfaceToJSON", "message", err)
		return InfoJSON, err
	}

	return InfoJSON, err
}

// JSONToFile accepts JSON and an output file path to create a JSON file.
func JSONToFile(JSON []byte, outFile string) error {
	err := ioutil.WriteFile(outFile, JSON, 0644)
	if err != nil {
		hclog.L().Error("JSONToFile", "error during json to file", err)
	}

	return err
}

// SplitFilepath takes a full path string and turns it into directory and file parts.
// In particular, it's useful for passing into seekers.NewCopier()
func SplitFilepath(path string) (dir string, file string) {
	dir, file = filepath.Split(path)
	// this enables a path like "somedir" (which filepath.Split() would call the "file")
	// to be interpreted as a relative path "./somedir" to align with normal CLI expectations
	if dir == "" {
		dir = "."
	}

	// try to discover whether our path is actually a directory
	stat, err := os.Stat(path)
	if err != nil {
		// since path may include "*" wildcards, which don't really exist, just return what we've managed so far.
		return dir, file
	}
	// if it is a directory, file=* to match everything in the dir
	if stat.IsDir() {
		dir = path
		file = "*"
	}

	return dir, file
}

func IsInRange(target, from, to time.Time) bool {
	// Default true if no range provided
	if from.IsZero() {
		return true
	}

	// Check if the end of our range is zero
	if to.IsZero() {
		// Anything after the start time is valid
		return target.After(from)
	}

	// Check if the fileTime is within range
	return target.After(from) && target.Before(to)
}

// FilterWalk accepts a source directory, filter string, and from and to Times to return a list of matching files.
func FilterWalk(srcDir, filter string, from, to time.Time) ([]string, error) {
	var fileMatches []string

	// Filter the files
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check for files that match the filter then check for time matches
		match, err := filepath.Match(filter, filepath.Base(path))
		if match && err == nil {
			// grab our file's last modified time
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			mod := info.ModTime()
			if IsInRange(mod, from, to) {
				fileMatches = append(fileMatches, path)
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	return fileMatches, nil
}

// FindInInterface treats an interface{} like a (nested) map,
// and searches through its contents for a given list of mapKeys.
// For example, given an interface{} containing a map like
// iface ~ interface{}{"top": {"mid": {"bottom": "desired_value"}}}
// one could fetch an interface{} of "desired_value" with
// FindInInterface(iface, "top", "mid", "bottom")
// then afterwards cast it to a string, or whatever type you're expecting.
func FindInInterface(iface interface{}, mapKeys ...string) (interface{}, error) {
	var (
		mapped map[string]interface{}
		ok     bool
		val    interface{}
	)
	mapped, ok = iface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unable to cast to map[string]interface{}; iface: %#v", iface)
	}
	for _, k := range mapKeys {
		val, ok = mapped[k]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found in mapped iface: %#v", k, mapped)
		}
		mapped, ok = val.(map[string]interface{})
		if !ok {
			// cannot map-ify any further, so assume this is what they're looking for
			return val, nil
		}
	}
	return mapped, nil
}
