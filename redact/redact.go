package redact

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const DefaultReplace = "<REDACTED>"

type Redact struct {
	ID      string `json:"ID"`
	matcher *regexp.Regexp
	Replace string `json:"replace"`
}

type Config struct {
	ID      string
	Matcher string
	Replace string
}

// New takes the matcher as a string and returned a compiled and ready-to-use redactor. ID and Replace are
// optional and can be left empty.
func New(cfg Config) (*Redact, error) {
	id := cfg.ID
	replace := cfg.Replace

	r, err := regexp.Compile(cfg.Matcher)
	if err != nil {
		return nil, err
	}
	if id == "" {
		genID := md5.Sum([]byte(cfg.Matcher))
		id = fmt.Sprint(genID)
	}
	if replace == "" {
		replace = DefaultReplace
	}
	return &Redact{id, r, replace}, nil
}

// Apply takes a slice of redactions and a writer + reader, reading everything in and applying redactions in
// sequential order before writing. Therefore, each Redact that appears earlier in the list takes precedence over later
// Redacts. It is possible for redactions to collide with one another if a matcher can match with the Replace string
// of an earlier Redact.
func Apply(redactions []*Redact, w io.Writer, r io.Reader) error {
	var bts []byte
	var err error

	bts, err = io.ReadAll(r)
	if err != nil {
		return err
	}
	if len(bts) == 0 {
		_, err = w.Write(bts)
		if err != nil {
			return err
		}
		return nil
	}
	for _, redact := range redactions {
		bts = redact.matcher.ReplaceAll(bts, []byte(redact.Replace))
	}
	_, err = w.Write(bts)
	if err != nil {
		return err
	}
	return nil
}

// String takes a string result and a slice of redactions, and wraps it with a reader and writer to apply the
// redactions, returning a string back.
// TODO(mkcp): Speed improvement & out of memory error: JSON responses can be really big, so we're going to have to
//  chunk extremely large strings down.
func String(result string, redactions []*Redact) (string, error) {
	r := strings.NewReader(result)
	buf := new(bytes.Buffer)
	err := Apply(redactions, buf, r)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Bytes takes a byte slice and a slice of redactions, and wraps it with a reader and writer to apply the
// redactions, returning a string back.
// TODO(mkcp): Speed improvement & out of memory error: JSON responses can be really big, so we're going to have to
//  chunk extremely large byte arrays down.
func Bytes(b []byte, redactions []*Redact) ([]byte, error) {
	r := bytes.NewReader(b)
	buf := new(bytes.Buffer)
	err := Apply(redactions, buf, r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// JSON accepts a json map or array and traverses the collections and redacts any strings we find.
func JSON(a any, redactions []*Redact) (any, error) {
	switch coll := a.(type) {
	case map[string]any:
		r, err := redactMap(coll, redactions)
		if err != nil {
			return nil, err
		}
		return r, nil
	case []any:
		r, err := redactSlice(coll, redactions)
		if err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, fmt.Errorf("json collection to redact is neither a map nor an array, coll=%v", coll)
	}
}

func redactSlice(a []any, redactions []*Redact) ([]any, error) {
	for i, v := range a {
		switch val := v.(type) {
		case map[string]any:
			res, err := redactMap(val, redactions)
			if err != nil {
				return nil, err
			}
			a[i] = res
		case []any:
			res, err := redactSlice(val, redactions)
			if err != nil {
				return nil, err
			}
			a[i] = res
		case string:
			res, err := String(val, redactions)
			if err != nil {
				return nil, err
			}
			a[i] = res
		default:
			continue
		}
	}
	return a, nil
}

func redactMap(m map[string]any, redactions []*Redact) (map[string]any, error) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			res, err := redactMap(val, redactions)
			if err != nil {
				return nil, err
			}
			m[k] = res
		case []any:
			res, err := redactSlice(val, redactions)
			if err != nil {
				return nil, err
			}
			m[k] = res
		case string:
			res, err := String(val, redactions)
			if err != nil {
				return nil, err
			}
			m[k] = res
		default:
			continue
		}
	}
	return m, nil
}

// Flatten takes any number of slices of redacts and returns one slice containing all redacts in argument order
func Flatten(redacts ...[]*Redact) []*Redact {
	flattened := make([]*Redact, 0)
	for _, rslice := range redacts {
		flattened = append(flattened, rslice...)
	}
	return flattened
}

// MapNew takes a slice of RedactConfig and return a slice of redact.Redact
func MapNew(configs []Config) ([]*Redact, error) {
	var redactions = make([]*Redact, len(configs))
	for i, cfg := range configs {
		r, err := New(cfg)
		if err != nil {
			return nil, err
		}
		redactions[i] = r
	}
	return redactions, nil
}
