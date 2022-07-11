package redactor

import "io"

var _ io.Reader = &RedactedReader{}

// RedactedReader contains details about an input that has already had some redactions applied to it. It is the
// primary return type for other redaction types in the redactor package. Because it implements the io.Reader
// interface, it can be used as input into other redactors, which allows for chaining redactions.
type RedactedReader struct {
	reader io.Reader
}

func (rr RedactedReader) Read(p []byte) (int, error) {
	return rr.reader.Read(p)
}
