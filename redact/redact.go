package redact

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const DefaultReplace = "<REDACTED>"

type Redact struct {
	ID      string `json:"ID"`
	matcher *regexp.Regexp
	Replace string `json:"replace"`
}

// New takes the matcher as a string and returned a compiled and ready-to-use redactor. ID and Replace are
// optional and can be left empty.
func New(matcher, id, replace string) (*Redact, error) {
	r, err := regexp.Compile(matcher)
	if err != nil {
		return nil, err
	}
	if id == "" {
		genID := md5.Sum([]byte(matcher))
		id = fmt.Sprint(genID)
	}
	if replace == "" {
		replace = DefaultReplace
	}
	return &Redact{id, r, replace}, nil
}

func (x Redact) Apply(w io.Writer, r io.Reader) error {
	bts, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(bts) == 0 {
		_, err = w.Write(bts)
		if err != nil {
			return err
		}
		return nil
	}
	newBts := x.matcher.ReplaceAll(bts, []byte(x.Replace))
	_, err = w.Write(newBts)
	if err != nil {
		return err
	}

	return nil
}

// ApplyMany takes a slice of redactions and a writer + reader, reading everything in and applying redactions in
// sequential order before writing. Therefore, each Redact that appears earlier in the list takes precedence over later
// Redacts. It is possible for redactions to collide with one another if a matcher can match with the Replace string
// of an earlier Redact.
func ApplyMany(redactions []*Redact, w io.Writer, r io.Reader) error {
	var bts []byte
	var err error

	bts, err = io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(bts) == 0 {
		_, err = w.Write(bts)
		if err != nil {
			return err
		}
		return nil
	}
	for _, redact := range redactions {
		bts = redact.matcher.ReplaceAll(bts, []byte(redact.Replace))
	}
	_, err = w.Write(bts)
	if err != nil {
		return err
	}
	return nil
}

// String takes a string result and a slice of redactions, and wraps it with a reader and writer to apply the
// redactions, returning a string back.
func String(result string, redactions []*Redact) (string, error) {
	r := strings.NewReader(result)
	buf := new(bytes.Buffer)
	err := ApplyMany(redactions, buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// File takes src, dest paths and a slice of redactions. It applies redactions line by line, reading from the source and
// writing to the destination
// redactions, returning a string back. Returns nil on success, otherwise an error.
func File(src, dest string, redactions []*Redact) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	destFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer destFile.Close()
	scanner := bufio.NewScanner(srcFile)
	// Scan, redact, and write each line of the src file
	for scanner.Scan() {
		res, err := String(scanner.Text(), redactions)
		if err != nil {
			return err
		}
		_, err = destFile.Write([]byte(res))
		if err != nil {
			return err
		}
	}
	return nil
}
