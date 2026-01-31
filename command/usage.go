// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package command

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/kr/text"
)

// maxLineLength is the maximum width of any line.
const maxLineLength int = 72

func Usage(txt string, flags *flag.FlagSet) string {
	u := &Usager{
		Usage: txt,
		Flags: flags,
	}
	return u.String()
}

type Usager struct {
	Usage string
	Flags *flag.FlagSet
}

func (u *Usager) String() string {
	out := new(bytes.Buffer)

	// Write out the usage slug.
	out.WriteString(strings.TrimSpace(u.Usage))
	out.WriteString("\n")
	out.WriteString("\n")

	if u.Flags != nil {
		printTitle(out, "Command Options")

		u.Flags.VisitAll(func(f *flag.Flag) {
			printFlag(out, f)
		})
	}

	return strings.TrimRight(out.String(), "\n")
}

// printTitle prints a consistently-formatted title to the given writer.
func printTitle(w io.Writer, s string) {
	_, _ = fmt.Fprintf(w, "%s\n\n", s)
}

// printFlag prints a single flag to the given writer.
func printFlag(w io.Writer, f *flag.Flag) {
	_, _ = fmt.Fprintf(w, "  -%s\n", f.Name)

	indented := wrapAtLength(f.Usage, 5)
	_, _ = fmt.Fprintf(w, "%s\n\n", indented)
}

// wrapAtLength wraps the given text at the maxLineLength, taking into account
// any provided left padding.
func wrapAtLength(s string, pad int) string {
	wrapped := text.Wrap(s, maxLineLength-pad)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", pad) + line
	}
	return strings.Join(lines, "\n")
}
