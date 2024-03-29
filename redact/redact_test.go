// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redact

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		cfg := Config{
			Matcher: tc.matcher,
			ID:      tc.id,
			Replace: tc.replace,
		}
		reg, err := New(cfg)
		assert.NoError(t, err, tc.name)
		assert.NotEqual(t, "", reg.ID, tc.name)
		assert.NotEqual(t, "", reg.Replace, tc.name)
	}
}

func TestApply(t *testing.T) {
	var redactions []*Redact
	matchers := []string{"myRegex", "test", "does not apply"}
	for _, matcher := range matchers {
		cfg := Config{
			Matcher: matcher,
			ID:      "",
			Replace: "",
		}
		redact, err := New(cfg)
		assert.NoError(t, err)
		redactions = append(redactions, redact)
	}
	tcs := []struct {
		name       string
		input      string
		expect     string
		redactions []*Redact
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
		{
			name: "redacts with grouping",
			redactions: []*Redact{
				newTestRedact(t, `(SECRET=)[^ ]+`, "${1}REDACTED"),
			},
			input:  "SECRET=my-secret-password",
			expect: "SECRET=REDACTED",
		},
		{
			name: "redacts with named grouping",
			redactions: []*Redact{
				newTestRedact(t, `(?P<MyGroup>SECRET=)[^ ]+`, "${MyGroup}REDACTED"),
			},
			input:  "SECRET=my-secret-password",
			expect: "SECRET=REDACTED",
		},
		{
			name: "case-insensitive redaction with grouping",
			redactions: []*Redact{
				newTestRedact(t, `(?i)(SECRET=)[^ ]+`, "${1}REDACTED"),
			},
			input:  "secret=my-secret-password",
			expect: "secret=REDACTED",
		},
		{
			name: "multi-group surround redaction",
			redactions: []*Redact{
				newTestRedact(t, `(\s+")[a-zA-Z0-9]{8}("\s+)`, "${1}REDACTED${2}"),
			},
			input:  "\"begin \"12345678\" end\"",
			expect: "\"begin \"REDACTED\" end\"",
		},
	}

	for _, tc := range tcs {
		r := strings.NewReader(tc.input)
		buf := new(bytes.Buffer)

		tcRedactions := redactions
		if tc.redactions != nil {
			tcRedactions = tc.redactions
		}
		err := Apply(tcRedactions, buf, r)
		assert.NoError(t, err, tc.name)

		result := buf.String()
		assert.Equal(t, tc.expect, result, tc.name)
	}
}

func TestString(t *testing.T) {
	tcs := []struct {
		name    string
		in      string
		redacts []*Redact
		expect  string
	}{
		{
			name:   "Test no redactions leaves string unchanged",
			in:     "an input string with a secret value",
			expect: "an input string with a secret value",
		},
		{
			name:   "Test no redactions leaves string unchanged with unicode",
			in:     "an input string with a secret value 😬",
			expect: "an input string with a secret value 😬",
		},
		{
			name: "Test redactions",
			redacts: []*Redact{
				newTestRedact(t, "secret", "REDACTED"),
			},
			in:     "an input string with a secret value",
			expect: "an input string with a REDACTED value",
		},
		{
			name: "Test redactions with unicode",
			redacts: []*Redact{
				newTestRedact(t, "😬", "REDACTED"),
			},
			in:     "an input string with a secret value 😬",
			expect: "an input string with a secret value REDACTED",
		},
		{
			name: "Test email redaction",
			redacts: []*Redact{
				newTestRedact(t, EmailPattern, EmailReplace),
			},
			in:     "lorem ipsum abc.def+_@ghi.jkl.mnop lorem ipsum",
			expect: "lorem ipsum REDACTED@REDACTED lorem ipsum",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result, err := String(tc.in, tc.redacts)
			assert.NoError(t, err, "encountered unexpected error in redaction")
			assert.Equal(t, tc.expect, result, tc.name)
		})
	}
}

func TestBytes(t *testing.T) {
	tcs := []struct {
		name    string
		in      []byte
		redacts []*Redact
		expect  []byte
	}{
		{
			name:   "Test no redactions leaves string unchanged",
			in:     []byte("an input string with a secret value"),
			expect: []byte("an input string with a secret value"),
		},
		{
			name:   "Test no redactions leaves string unchanged with unicode",
			in:     []byte("an input string with a secret value 😬"),
			expect: []byte("an input string with a secret value 😬"),
		},
		{
			name: "Test redactions",
			redacts: []*Redact{
				newTestRedact(t, "secret", "REDACTED"),
			},
			in:     []byte("an input string with a secret value"),
			expect: []byte("an input string with a REDACTED value"),
		},
		{
			name: "Test redactions with unicode",
			redacts: []*Redact{
				newTestRedact(t, "😬", "REDACTED"),
			},
			in:     []byte("an input string with a secret value 😬"),
			expect: []byte("an input string with a secret value REDACTED"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Bytes(tc.in, tc.redacts)
			assert.NoError(t, err, "encountered unexpected error in redaction")
			assert.Equal(t, tc.expect, result, tc.name)
		})
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
				one, err := New(Config{"", "there", ""})
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
				one, err := New(Config{"", "there", ""})
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
				one, err := New(Config{"", "there", ""})
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

// test redact.Flatten()
func TestFlatten(t *testing.T) {
	// Set up test redacts
	var nilSlice []*Redact
	var emptySlice = []*Redact{}
	singleRedact := []*Redact{newTestRedact(t, "matchredact", "foobar")}
	multiRedact := []*Redact{
		newTestRedact(t, "foobar", "baz"),
		newTestRedact(t, "baz", "<REDACTED>"),
	}

	tcs := []struct {
		name   string
		input  [][]*Redact
		expect []*Redact
	}{
		// Nil slice alone, first, last, middle
		{
			name:   "Flatten should return empty redact slice for nil slice input",
			input:  [][]*Redact{nilSlice},
			expect: make([]*Redact, 0),
		},
		{
			name:   "Flatten should treat a nil slice (first arg) correctly",
			input:  [][]*Redact{nilSlice, singleRedact},
			expect: singleRedact,
		},
		{
			name:   "Flatten should treat mixed args (multiRedact, singleRedact, nil slice) correctly",
			input:  [][]*Redact{multiRedact, singleRedact, nilSlice},
			expect: []*Redact{multiRedact[0], multiRedact[1], singleRedact[0]},
		},
		{
			name:   "Flatten should treat mixed args (multiRedact, nil slice, singleRedact) correctly",
			input:  [][]*Redact{multiRedact, nilSlice, singleRedact},
			expect: []*Redact{multiRedact[0], multiRedact[1], singleRedact[0]},
		},
		// Single arg
		{
			name:   "Flatten should treat a single redact input correctly",
			input:  [][]*Redact{singleRedact},
			expect: singleRedact,
		},
		// Multi-arg
		{
			name:   "Flatten should treat a multi-redact slice input correctly",
			input:  [][]*Redact{multiRedact},
			expect: multiRedact,
		},
		{
			name:   "Flatten should treat mixed-length inputs correctly (1)",
			input:  [][]*Redact{multiRedact, singleRedact},
			expect: []*Redact{multiRedact[0], multiRedact[1], singleRedact[0]},
		},
		{
			name:   "Flatten should treat mixed-length inputs correctly (2)",
			input:  [][]*Redact{singleRedact, multiRedact},
			expect: []*Redact{singleRedact[0], multiRedact[0], multiRedact[1]},
		},
		// empty slice alone, first, last, middle
		{
			name:   "Flatten should return empty redact slice for empty redact slice input",
			input:  [][]*Redact{emptySlice},
			expect: make([]*Redact, 0),
		},
		{
			name:   "Flatten should treat an empty slice (first arg) correctly",
			input:  [][]*Redact{emptySlice, singleRedact},
			expect: singleRedact,
		},
		{
			name:   "Flatten should treat mixed args (multiRedact, singleRedact, empty slice) correctly",
			input:  [][]*Redact{multiRedact, singleRedact, emptySlice},
			expect: []*Redact{multiRedact[0], multiRedact[1], singleRedact[0]},
		},
		{
			name:   "Flatten should treat mixed args (multiRedact, empty slice, singleRedact) correctly",
			input:  [][]*Redact{multiRedact, emptySlice, singleRedact},
			expect: []*Redact{multiRedact[0], multiRedact[1], singleRedact[0]},
		},
		{
			name:   "Flatten should treat mixed args (nil slice, multiRedact, empty slice, singleRedact) correctly",
			input:  [][]*Redact{nilSlice, multiRedact, emptySlice, singleRedact},
			expect: []*Redact{multiRedact[0], multiRedact[1], singleRedact[0]},
		},
	}
	// Run assertions
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result := Flatten(tc.input...)
			assert.Equal(t, tc.expect, result, tc.name)
		})
	}
}

// test redact.MapNew()
func TestMapNew(t *testing.T) {
	// Set up test redacts
	var nilSlice []Config
	var emptySlice = []Config{}

	tcs := []struct {
		name      string
		input     []Config
		expectLen int
	}{
		{
			name:      "MapNew should return empty redact slice for empty slice input",
			input:     emptySlice,
			expectLen: 0,
		},
		{
			name:      "MapNew should return empty redact slice for nil slice input",
			input:     nilSlice,
			expectLen: 0,
		},
		{
			name:      "MapNew should treat single-config slices correctly",
			input:     []Config{{"", "something", "repl"}},
			expectLen: 1,
		},
		{
			name:      "MapNew should treat multi-config slices correctly",
			input:     []Config{{"", "something", "repl"}, {"", "otherthing", ""}},
			expectLen: 2,
		},
	}
	// Run assertions
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := MapNew(tc.input)
			resultLen := len(result)
			assert.Equal(t, tc.expectLen, resultLen, tc.name)
		})
	}
}

// newTestRedact wraps redaction creation and fails the test if there's an error
func newTestRedact(t *testing.T, matcher string, replace string) *Redact {
	t.Helper()
	cfg := Config{
		Matcher: matcher,
		ID:      "",
		Replace: replace,
	}
	r, err := New(cfg)
	require.NoError(t, err, "error creating test redaction")
	return r
}

func BenchmarkStringUnchanged(b *testing.B) {
	input := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed fringilla sodales dolor, quis eleifend."
	var redactions []*Redact

	for n := 0; n < b.N; n++ {
		result, err := String(input, redactions)
		if err != nil {
			b.Errorf("redaction caused error: %#v", err)
		}

		if result != input {
			b.Errorf("string was changed unexpectedly;\ninput: %s\nresult: %s", input, result)
		}
	}
}

func BenchmarkStringRedacted(b *testing.B) {
	input := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed fringilla sodales dolor, quis eleifend."
	redactions := []*Redact{
		func() *Redact {
			red, err := New(Config{Matcher: `.*`, Replace: "REDACTED"})
			if err != nil {
				b.Fatalf("error creating redaction: %#v", err)
			}
			return red
		}(),
	}

	for n := 0; n < b.N; n++ {
		result, err := String(input, redactions)
		if err != nil {
			b.Errorf("redaction caused error: %#v", err)
		}

		if result != "REDACTED" {
			b.Errorf("string was changed unexpectedly;\ninput: %s\nresult: %s", input, result)
		}
	}
}

func BenchmarkBytesUnchanged(b *testing.B) {
	input := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed fringilla sodales dolor, quis eleifend.")
	var redactions []*Redact

	for n := 0; n < b.N; n++ {
		result, err := Bytes(input, redactions)
		if err != nil {
			b.Errorf("redaction caused error: %#v", err)
		}

		if string(result) != string(input) {
			b.Errorf("input was changed unexpectedly;\ninput: %s\nresult: %s", string(input), result)
		}
	}
}

func BenchmarkBytesRedacted(b *testing.B) {
	input := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed fringilla sodales dolor, quis eleifend.")
	redactions := []*Redact{
		func() *Redact {
			red, err := New(Config{Matcher: `.*`, Replace: "REDACTED"})
			if err != nil {
				b.Fatalf("error creating redaction: %#v", err)
			}
			return red
		}(),
	}

	for n := 0; n < b.N; n++ {
		result, err := Bytes(input, redactions)
		if err != nil {
			b.Errorf("redaction caused error: %#v", err)
		}

		if string(result) != "REDACTED" {
			b.Errorf("input was changed unexpectedly;\ninput: %s\nresult: %s", string(input), result)
		}
	}
}
