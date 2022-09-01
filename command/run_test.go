package command

import (
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/agent"
	"github.com/stretchr/testify/assert"
)

func TestSetTime(t *testing.T) {
	// This test is kind of a tautology because we're replacing the impl. to build the expected struct. But it holds
	// some value to protect against regressions.
	testDur, err := time.ParseDuration("48h")
	if !assert.NoError(t, err) {
		return
	}
	cfg := agent.Config{}
	now := time.Now()
	newCfg := setTime(cfg, now, testDur)

	expect := agent.Config{
		Since: now.Add(-testDur),
		Until: time.Time{}, // This is the default zero value, but we're just being extra explicit.
	}

	assert.Equal(t, newCfg, expect)
}
