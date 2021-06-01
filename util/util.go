package util

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// TarGz accepts a source directory and destination file name to archive and compress files.
func TarGz(sourceDir string, destFileName string) error {
	// tar
	destFile, err := os.Create(destFileName)
	if err != nil {
		hclog.L().Error("TarGz", "error creating tarball", err)
		return err
	}
	defer destFile.Close()

	gzWriter := gzip.NewWriter(destFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() == false {
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

	return err
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

func isInRange(path string, from, to time.Time) (bool, error) {
	// Default true if no range provided
	if !from.IsZero() {
		return true, nil
	}

	// When we only get a `from` value, the `to` is now
	if to.IsZero() {
		to = time.Now()
	}

	// Grab our file's last modified time
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	mod := info.ModTime()

	// Check if the mod time is outside of the range
	// NOTE(mkcp): Can this "after" check bug if there's no range provided, so we set it to "now" and the file is being
	//  updated in parallel? Would that mean the modified time becomes _after_ Now even though we've statted the file?
	//  There's no read snapshot happening of the file... maybe we should completely cut the "after" check if it's zero
	//  rather than fudging a range check with a defaulted value. That's more semantic anyway, considering we're not
	//  checking a range at all, but just a before
	if mod.Before(from) || mod.After(to) {
		return false, nil
	}

	// Yes, it's in the range!
	return true, nil
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
			inRange, err := isInRange(path, from, to)
			if err != nil {
				return err
			}
			if inRange {
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
