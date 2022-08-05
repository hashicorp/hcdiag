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

func TestApplyMany(t *testing.T) {
	var redactions []*Redact
	matchers := []string{"myRegex", "test", "does not apply"}
	for _, matcher := range matchers {
		redact, err := New(matcher, "", "")
		assert.NoError(t, err)
		redactions = append(redactions, redact)
	}
	tcs := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "empty input",
			input:  "",
			expect: "",
		},
		{
			name:   "redacts once",
			input:  "myRegex",
			expect: "<REDACTED>",
		},
		{
			name:   "redacts many",
			input:  "test test_test+test-test\n!test ??test",
			expect: "<REDACTED> <REDACTED>_<REDACTED>+<REDACTED>-<REDACTED>\n!<REDACTED> ??<REDACTED>",
		},
	}
	for _, tc := range tcs {
		r := strings.NewReader(tc.input)
		buf := new(bytes.Buffer)
		err := ApplyMany(redactions, buf, r)
		assert.NoError(t, err, tc.name)

		result := buf.String()
		assert.Equal(t, tc.expect, result, tc.name)
	}
}

func TestJSON(t *testing.T) {
	tcs := []struct {
		name    string
		json    any
		redacts func() ([]*Redact, error)
		expect  any
	}{
		{
			name:    "empty json",
			json:    map[string]any{},
			redacts: func() ([]*Redact, error) { return nil, nil },
			expect:  map[string]any{},
		},
		{
			name: "no redacts passes json map through undisturbed",
			json: map[string]any{
				"hello": "there",
				"1":     2,
				"array": []any{"one", "two", "three"},
				"m":     map[string]any{"ello": "hthere"},
			},
			redacts: func() ([]*Redact, error) { return nil, nil },
			expect: map[string]any{
				"hello": "there",
				"1":     2,
				"array": []any{"one", "two", "three"},
				"m":     map[string]any{"ello": "hthere"},
			},
		},
		{
			name: "no redacts passes json array through undisturbed",
			json: []any{
				"there",
				2,
				[]any{"one", "two", "three"},
				map[string]any{"ello": "hthere"},
			},
			redacts: func() ([]*Redact, error) { return nil, nil },
			expect: []any{
				"there",
				2,
				[]any{"one", "two", "three"},
				map[string]any{"ello": "hthere"},
			},
		},
		{
			name: "can redact arbitrarily nested strings in map",
			json: map[string]any{
				"hello": "there",
				"1":     2,
				"array": []any{"one", "two", "three", "there"},
				"m":     map[string]any{"ello": "hthere"},
			},
			redacts: func() ([]*Redact, error) {
				one, err := New("there", "", "")
				if err != nil {
					return nil, err
				}
				return []*Redact{
					one,
				}, nil
			},
			expect: map[string]any{
				"hello": "<REDACTED>",
				"1":     2,
				"array": []any{"one", "two", "three", "<REDACTED>"},
				"m":     map[string]any{"ello": "h<REDACTED>"},
			},
		},
		{
			name: "can redact arbitrarily nested strings in array",
			json: []any{
				"there",
				2,
				[]any{"one", "two", "three", "there"},
				map[string]any{"ello": "hthere"},
			},
			redacts: func() ([]*Redact, error) {
				one, err := New("there", "", "")
				if err != nil {
					return nil, err
				}
				return []*Redact{
					one,
				}, nil
			},
			expect: []any{
				"<REDACTED>",
				2,
				[]any{"one", "two", "three", "<REDACTED>"},
				map[string]any{"ello": "h<REDACTED>"},
			},
		},
		{
			name: "can redact arbitrarily nested arrays",
			json: []any{
				[]any{"one", "two", "three", []any{"there"}},
			},
			redacts: func() ([]*Redact, error) {
				one, err := New("there", "", "")
				if err != nil {
					return nil, err
				}
				return []*Redact{
					one,
				}, nil
			},
			expect: []any{
				[]any{"one", "two", "three", []any{"<REDACTED>"}},
			},
		},
	}

	for _, tc := range tcs {
		redactions, err := tc.redacts()
		assert.NoError(t, err, tc.name)
		result, err := JSON(tc.json, redactions)
		assert.NoError(t, err, tc.name)
		assert.Equal(t, tc.expect, result, tc.name)
	}
}
