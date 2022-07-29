package redact

import (
	"crypto/md5"
	"fmt"
	"io"
	"regexp"
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
