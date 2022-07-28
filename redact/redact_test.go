package redact

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegex(t *testing.T) {
	tcs := []struct {
		name    string
		matcher string
		id      string
		replace string
	}{
		{
			name:    "empty optional fields",
			matcher: "/some regex/",
		},
		{
			name:    "set optional fields",
			matcher: "/some other regex/",
			id:      "COOLCOOL",
			replace: "WOWOW",
		},
	}

	for _, tc := range tcs {
		reg, err := New(tc.matcher, tc.id, tc.replace)
		assert.NoError(t, err, tc.name)
		assert.NotEqual(t, "", reg.ID, tc.name)
		assert.NotEqual(t, "", reg.Replace, tc.name)
	}
}

func TestRedact_Apply(t *testing.T) {
	tcs := []struct {
		name    string
		matcher string
		input   string
		expect  string
	}{
		{
			name:    "empty input",
			matcher: "/myRegex/",
			input:   "",
			expect:  "",
		},
		{
			name:    "redacts once",
			matcher: "myRegex",
			input:   "myRegex",
			expect:  "<REDACTED>",
		},
		{
			name:    "redacts many",
			matcher: "test",
			input:   "test test_test+test-test\n!test ??test",
			expect:  "<REDACTED> <REDACTED>_<REDACTED>+<REDACTED>-<REDACTED>\n!<REDACTED> ??<REDACTED>",
		},
	}
	for _, tc := range tcs {
		redactor, err := New(tc.matcher, "", "")
		assert.NoError(t, err, tc.name)

		r := strings.NewReader(tc.input)
		buf := new(bytes.Buffer)
		err = redactor.Apply(buf, r)
		assert.NoError(t, err, tc.name)

		result := buf.String()

		assert.Equal(t, tc.expect, result, tc.name)
	}
}

func TestLiteral_Redact(t *testing.T) {

}
