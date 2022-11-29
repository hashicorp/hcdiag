// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package host

import (
	"testing"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/stretchr/testify/require"
)

func createRedactionSlice(t *testing.T, config ...redact.Config) []*redact.Redact {
	t.Helper()

	var result []*redact.Redact
	for _, cfg := range config {
		result = append(result, createRedaction(t, cfg))
	}
	return result
}

func createRedaction(t *testing.T, config redact.Config) *redact.Redact {
	t.Helper()

	redaction, err := redact.New(config)
	if err != nil {
		require.NoError(t, err)
	}
	return redaction
}
