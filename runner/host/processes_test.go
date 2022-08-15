package host

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/require"
)

// NOTE: The external library defines an interface (ps.Process) for listing processes instead of a struct; we implement
// that interface in mockProc with Pid, PPid, and Executable.
type mockProc struct {
	pid  int
	ppid int
	exe  string
}

func (m mockProc) Pid() int           { return m.pid }
func (m mockProc) PPid() int          { return m.ppid }
func (m mockProc) Executable() string { return m.exe }

func TestProcess_convertProcessInfo(t *testing.T) {
	testCases := []struct {
		name       string
		proc       Process
		inputProcs []ps.Process
		expected   []Proc
		expectErr  bool
	}{
		{
			name:       "Test No Redactions",
			proc:       Process{},
			inputProcs: []ps.Process{mockProc{exe: "application-1"}, mockProc{exe: "secret-application"}},
			expected: []Proc{
				{
					Name: "application-1",
				},
				{
					Name: "secret-application",
				},
			},
		},
		{
			name:       "Test Redactions",
			proc:       Process{Redactions: createRedactionSlice(t, redact.Config{Matcher: "secret-application"})},
			inputProcs: []ps.Process{mockProc{exe: "application-1"}, mockProc{exe: "secret-application"}},
			expected: []Proc{
				{
					Name: "application-1",
				},
				{
					Name: "<REDACTED>",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			procs, err := tc.proc.convertProcessInfo(tc.inputProcs)
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not returned")
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tc.expected, procs),
					"result did not match the expected result:\nactual=%#v\nexpected=%#v", procs, tc.expected)
			}
		})
	}
}
