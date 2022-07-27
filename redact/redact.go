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
	newBts := x.matcher.ReplaceAll(bts, []byte(x.Replace))
	_, err = w.Write(newBts)
	if err != nil {
		return err
	}

	return nil
}
