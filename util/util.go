package util

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// TarGz func to archive and compress
func TarGz(sourceFilePath string, destFilePathTar string, destFilePathTarGz string) error {
	// tar
	destFile, err := os.Create(destFilePathTar)
	if err != nil {
		fmt.Printf("Error creating tarball, got '%s'", err.Error())
	}
	defer destFile.Close()

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		fmt.Printf("Error opening source file, got '%s'", err.Error())
	}
	defer sourceFile.Close()

	stat, err := sourceFile.Stat()
	if err != nil {
		fmt.Printf("Error on stat for source file, got '%s'", err.Error())
	}

	header := &tar.Header{
		Name:    sourceFilePath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		fmt.Printf("Error writing header for tar file, got '%s'", err.Error())
	}

	_, err = io.Copy(tarWriter, sourceFile)
	if err != nil {
		fmt.Printf("Error copying file to tarball, got '%s'", err.Error())
	}

	// gzip
	reader, err := os.Open(destFilePathTar)
	if err != nil {
		fmt.Printf("Error opening tar file, got '%s'", err.Error())
	}

	writer, err := os.Create(destFilePathTarGz)
	if err != nil {
		fmt.Printf("Error creating tar gz file, got '%s'", err.Error())
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = destFilePathTarGz
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)

	return err
}

// MapToJSON stuff
func MapToJSON(mapVar map[string]interface{}) []byte {
	InfoJSON, err := json.MarshalIndent(mapVar, "", "    ")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return InfoJSON
}

// JSONToFile stuff
func JSONToFile(JSON []byte, outFile string) {
	err := ioutil.WriteFile(outFile, JSON, 0644)
	if err != nil {
		fmt.Println(err)
	}
	return
}
