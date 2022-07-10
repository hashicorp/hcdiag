// Package redactor includes a set of tools for redacting potentially sensitive data.
package redactor

import "io"

// Redactor indicates a type implements a Redact method, which both takes and returns an io.Reader.
// Because a source may need to go through multiple Redactors, returning an io.Reader makes it easier to chain
// them together.
type Redactor interface {
	Redact(reader io.Reader) (RedactedReader, error)
}
