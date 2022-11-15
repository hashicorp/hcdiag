package runner

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/hcdiag/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTP(t *testing.T) {
	t.Parallel()

	c, err := client.NewAPIClient(client.APIConfig{
		Product:   "consul",
		BaseURL:   "https://someplace.local",
		TLSConfig: client.TLSConfig{},
	})
	require.NoError(t, err)

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
