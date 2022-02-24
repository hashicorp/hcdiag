package main

import (
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/agent"
	"github.com/stretchr/testify/assert"
)

func TestMergeAgentConfig(t *testing.T) {
	testDur, err := time.ParseDuration("48h")
	if !assert.NoError(t, err) {
		return
	}
	flags := Flags{
		OS:          "auto",
		AllProducts: true,
		Since:       testDur,
	}

	newCfg := mergeAgentConfig(agent.Config{}, flags)

	testCfg := agent.Config{
		OS:     "auto",
		Consul: true,
		Nomad:  true,
		TFE:    true,
		Vault:  true,
		Since:  time.Time{},
		Until:  time.Time{},
	}

	assert.Equal(t, newCfg, testCfg)
}
