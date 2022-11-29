// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package runner

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/hcdiag/op"

	"github.com/stretchr/testify/assert"
)

var (
	mockID            = "mock"
	mockResult        = map[string]any{"mock": "get mock'd"}
	errFake           = errors.New("uh oh a fake error")
	testStatus        = op.Status("test")
	_          Runner = MockRunner{}
)

type MockRunner struct {
	id    string
	Param bool `json:"param"`
}

func NewMockRunner(id string) *MockRunner {
	return &MockRunner{id: id}
}

func (r MockRunner) ID() string {
	return r.id
}

func (r MockRunner) Run() op.Op {
	return op.Op{
		Identifier: r.ID(),
		Result:     mockResult,
		ErrString:  errFake.Error(),
		Error:      errFake,
		Status:     testStatus,
	}
}

func TestRunner_Run(t *testing.T) {
	r := NewMockRunner(mockID)
	o := r.Run()
	assert.Equal(t, mockResult, o.Result, "returned result does not match Op Result field")
	if o.Identifier != mockID {
		t.Errorf("returned result (%s) does not match Op Result field (%s)", mockID, o.Identifier)
	}
	if o.Error != errFake {
		t.Errorf("returned err (%s) does not match Op Error field (%s)", errFake, o.Error)
	}
	errStr := fmt.Sprintf("%s", o.Error)
	if o.ErrString != errStr {
		t.Errorf("Op ErrString (%s) not formatted as expected (%s)", o.ErrString, errStr)
	}
}

func TestExclude(t *testing.T) {
	testTable := []struct {
		desc     string
		matchers []string
		runners  []Runner
		expect   int
	}{
		{
			desc:     "Can exclude none",
			matchers: []string{"hello"},
			runners: []Runner{
				NewMockRunner("nope"),
				NewMockRunner("nah"),
				NewMockRunner("sry"),
			},
			expect: 3,
		},
		{
			desc:     "Can exclude one",
			matchers: []string{"hi"},
			runners: []Runner{
				NewMockRunner("hi"),
			},
			expect: 0,
		},
		{
			desc:     "Can exclude two",
			matchers: []string{"hi", "sup"},
			runners: []Runner{
				NewMockRunner("hi"),
				NewMockRunner("sup"),
			},
			expect: 0,
		},
		{
			desc:     "Can exclude many and and ignore one",
			matchers: []string{"exclude1", "exclude2", "exclude3"},
			runners: []Runner{
				NewMockRunner("exclude1"),
				NewMockRunner("exclude2"),
				NewMockRunner("exclude3"),
				NewMockRunner("keep"),
			},
			expect: 1,
		},
		{
			desc:     "Can exclude glob *",
			matchers: []string{"exclude*"},
			runners: []Runner{
				NewMockRunner("exclude1"),
				NewMockRunner("exclude2"),
				NewMockRunner("keep"),
			},
			expect: 1,
		},
	}

	for _, tc := range testTable {
		res, err := Exclude(tc.matchers, tc.runners)
		assert.Nil(t, err)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}

func TestSelect(t *testing.T) {
	testTable := []struct {
		desc     string
		matchers []string
		runners  []Runner
		expect   int
	}{
		{
			desc:     "Can select none",
			matchers: []string{"hello"},
			runners: []Runner{
				NewMockRunner("nope"),
				NewMockRunner("nah"),
				NewMockRunner("sry"),
			},
			expect: 0,
		},
		{
			desc:     "Can select one",
			matchers: []string{"match"},
			runners: []Runner{
				NewMockRunner("nope"),
				NewMockRunner("nah"),
				NewMockRunner("sry"),
				NewMockRunner("match"),
			},
			expect: 1,
		},
		{
			desc:     "Can select two",
			matchers: []string{"match1", "match2"},
			runners: []Runner{
				NewMockRunner("nope"),
				NewMockRunner("nah"),
				NewMockRunner("sry"),
				NewMockRunner("match1"),
				NewMockRunner("match2"),
			},
			expect: 2,
		},
		{
			desc:     "Can select many regardless of order",
			matchers: []string{"select1", "select2", "select3"},
			runners: []Runner{
				NewMockRunner("skip1"),
				NewMockRunner("select2"),
				NewMockRunner("skip2"),
				NewMockRunner("skip3"),
				NewMockRunner("select3"),
				NewMockRunner("select1"),
			},
			expect: 3,
		},
		{
			desc:     "Can select glob *",
			matchers: []string{"select*"},
			runners: []Runner{
				NewMockRunner("skip1"),
				NewMockRunner("select2"),
				NewMockRunner("skip2"),
				NewMockRunner("skip3"),
				NewMockRunner("select3"),
				NewMockRunner("select1"),
			},
			expect: 3,
		},
	}

	for _, tc := range testTable {
		res, err := Select(tc.matchers, tc.runners)
		assert.Nil(t, err)
		assert.Len(t, res, tc.expect, tc.desc)
	}
}

func TestParams(t *testing.T) {
	r := MockRunner{id: "test", Param: true}
	key := "param"
	v := Params(r)
	assert.Len(t, v, 1)
	assert.True(t, v[key].(bool))
}
