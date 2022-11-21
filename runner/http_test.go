package runner

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTP(t *testing.T) {
	t.Parallel()

	c := getTestAPIClient(t)

	tt := []struct {
		desc      string
		cfg       HttpConfig
		expect    *HTTP
		expectErr bool
	}{
		{
			desc:      "empty config causes an error",
			cfg:       HttpConfig{},
			expectErr: true,
		},
		{
			desc: "test defaults",
			cfg:  HttpConfig{Client: c},
			expect: &HTTP{
				Path:       "",
				Client:     c,
				ctx:        context.Background(),
				Timeout:    0,
				Redactions: nil,
			},
		},
		{
			desc: "negative timeout duration causes an error",
			cfg: HttpConfig{
				Client:  c,
				Timeout: -10 * time.Second,
			},
			expectErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			c, err := NewHTTP(tc.cfg)
			if tc.expectErr {
				assert.ErrorAs(t, err, &HTTPConfigError{})
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, c)
			}
		})
	}
}

func TestHTTP_RunCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	h := HTTP{
		Client: getTestAPIClient(t),
		ctx:    ctx,
	}

	result := h.Run()
	assert.Equal(t, op.Canceled, result.Status)
	assert.ErrorIs(t, result.Error, context.Canceled)
}

func TestHTTP_RunTimeout(t *testing.T) {
	t.Parallel()

	// Set to a short timeout, and sleep briefly to ensure it passes before we try to run the command
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancelFunc()
	time.Sleep(1 * time.Nanosecond)

	h := HTTP{
		Client: getTestAPIClient(t),
		ctx:    ctx,
	}

	result := h.Run()
	assert.Equal(t, op.Timeout, result.Status)
	assert.ErrorIs(t, result.Error, context.DeadlineExceeded)
}

func getTestAPIClient(t *testing.T) *client.APIClient {
	t.Helper()

	c, err := client.NewAPIClient(client.APIConfig{
		Product:   "consul",
		BaseURL:   "https://someplace.local",
		TLSConfig: client.TLSConfig{},
	})
	require.NoError(t, err)

	return c
}
