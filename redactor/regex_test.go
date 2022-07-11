package redactor

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexRedactor_Redact(t *testing.T) {
	testCases := []struct {
		name      string
		redactor  RegexRedactor
		input     *strings.Reader
		expected  []byte
		expectErr bool
	}{
		{
			name: "Basic Redaction",
			redactor: RegexRedactor{
				RegEx:       `(SECRET=)[^ ]+`,
				Replacement: "${1}REDACTED",
			},
			input:    strings.NewReader("SECRET=my-secret-password"),
			expected: []byte("SECRET=REDACTED"),
		},
		{
			name: "Literal Redaction",
			redactor: RegexRedactor{
				RegEx:       "hello",
				Replacement: "REDACTED",
			},
			input:    strings.NewReader("hello world"),
			expected: []byte("REDACTED world"),
		},
		{
			name: "Middle Redaction",
			redactor: RegexRedactor{
				RegEx:       `(SECRET=)[^ ]+`,
				Replacement: "${1}REDACTED",
			},
			input:    strings.NewReader("Other text SECRET=my-secret-password Other text"),
			expected: []byte("Other text SECRET=REDACTED Other text"),
		},
		{
			name: "Beginning Only Redaction",
			redactor: RegexRedactor{
				RegEx:       `^(SECRET=)[^ ]+`,
				Replacement: "${1}REDACTED",
			},
			input:    strings.NewReader("Other text SECRET=my-secret-password Other text"),
			expected: []byte("Other text SECRET=my-secret-password Other text"),
		},
		{
			name: "Unicode Redaction",
			redactor: RegexRedactor{
				RegEx:       `(🫣=)[^ ]+`,
				Replacement: "${1}REDACTED",
			},
			input:    strings.NewReader("🫣=🤫"),
			expected: []byte("🫣=REDACTED"),
		},
		{
			name: "Multiple Group Surround Redaction",
			redactor: RegexRedactor{
				RegEx:       `(\s+")[a-zA-Z0-9]{8}("\s+)`,
				Replacement: "${1}REDACTED${2}",
			},
			input:    strings.NewReader("begin \"12345678\" end"),
			expected: []byte("begin \"REDACTED\" end"),
		},
		{
			name: "More than One Redaction",
			redactor: RegexRedactor{
				RegEx:       `(SECRET=)[^\s]+`,
				Replacement: "${1}REDACTED",
			},
			input: strings.NewReader(`
SECRET=my-secret-password
other text
SECRET=my-other-password
`),
			expected: []byte(`
SECRET=REDACTED
other text
SECRET=REDACTED
`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lr := tc.redactor
			rr, err := lr.Redact(tc.input)
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, rr)
			} else {
				assert.NoError(t, err)

				b, err := io.ReadAll(rr)
				assert.Equal(t, string(tc.expected), string(b))
				assert.NoError(t, err)
			}
		})
	}
}

func FuzzRegexRedactor_Redact(f *testing.F) {
	testCases := []string{
		"hello",
		"world",
		"12345",
		" ",
	}
	for _, tc := range testCases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// We aren't doing too much here, but when fuzzing, we can at least make sure that weird input
		// doesn't cause errors, and we don't make unexpected changes to input.
		redactor := NewRegexRedactor("", "")
		rr, err := redactor.Redact(strings.NewReader(input))
		if err != nil {
			t.Errorf("encountered error in test: %#v\n", err)
		}
		b, err := io.ReadAll(rr)
		if string(b) != input {
			t.Errorf("input was unexpectedly altered;\nINPUT = %q\nOUTPUT = %q\n", input, string(b))
		}
	})
}

func BenchmarkRegexRedactor_Redact1(b *testing.B)   { benchmarkRegexHelper(1, b) }
func BenchmarkRegexRedactor_Redact2(b *testing.B)   { benchmarkRegexHelper(2, b) }
func BenchmarkRegexRedactor_Redact3(b *testing.B)   { benchmarkRegexHelper(3, b) }
func BenchmarkRegexRedactor_Redact10(b *testing.B)  { benchmarkRegexHelper(10, b) }
func BenchmarkRegexRedactor_Redact20(b *testing.B)  { benchmarkRegexHelper(20, b) }
func BenchmarkRegexRedactor_Redact30(b *testing.B)  { benchmarkRegexHelper(30, b) }
func BenchmarkRegexRedactor_Redact100(b *testing.B) { benchmarkRegexHelper(100, b) }

func benchmarkRegexHelper(num int, b *testing.B) {
	input := strings.NewReader(strings.Repeat("hello world", num))
	redactor := RegexRedactor{
		RegEx:       "hello",
		Replacement: "REDACTED",
	}

	for i := 0; i < b.N; i++ {
		_, _ = redactor.Redact(input)
	}
}
