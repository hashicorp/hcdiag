package redactor

import (
	"io"
	"io/ioutil"
	"regexp"
)

var _ Redactor = &RegexRedactor{}

type RegexRedactor struct {
	RegEx       string
	Replacement string

	re *regexp.Regexp
}

func NewRegexRedactor(reg string, repl string) RegexRedactor {
	// TODO (nwchandler): Should we panic on this (MustCompile)? Redaction seems critical enough to me that we should...
	re := regexp.MustCompile(reg)

	return RegexRedactor{
		RegEx:       reg,
		Replacement: repl,
		re:          re,
	}
}

func (reg RegexRedactor) Redact(in io.Reader) (RedactedReader, error) {
	if reg.re == nil {
		reg.re = regexp.MustCompile(reg.RegEx)
	}

	r, w := io.Pipe()
	rr := RedactedReader{
		rdr: r,
	}
	go func() {
		defer w.Close()

		content, _ := ioutil.ReadAll(in)

		redacted := reg.re.ReplaceAll(content, []byte(reg.Replacement))

		w.Write(redacted)
	}()

	return rr, nil
}
