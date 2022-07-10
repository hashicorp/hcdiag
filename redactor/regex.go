package redactor

import (
	"io"
	"io/ioutil"
	"regexp"

	"github.com/hashicorp/go-hclog"
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

	// The goroutine is required for piping from io.PipeWriter to io.PipeReader.
	go func() {
		defer func(w *io.PipeWriter) {
			err := w.Close()
			if err != nil {
				hclog.L().Warn("RegexRedactor failed to close PipeWriter", "writer", w)
			}
		}(w)

		content, _ := ioutil.ReadAll(in)

		redacted := reg.re.ReplaceAll(content, []byte(reg.Replacement))

		_, err := w.Write(redacted)
		if err != nil {
			hclog.L().Warn("RegexRedactor failed to write redacted data to PipeWriter", "writer", w)
		}
	}()

	return rr, nil
}
