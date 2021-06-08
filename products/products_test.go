package products

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetup(t *testing.T) {
	t.Run("Should only get host if no products enabled", func(t *testing.T) {
		cfg := Config{OS: "auto"}
		p, err := Setup(cfg)
		assert.NoError(t, err)
		assert.Len(t, p, 1)
	})
	t.Run("Should have host and nomad enabled", func(t *testing.T) {
		cfg := Config{
			Nomad: true,
			OS:    "auto",
		}
		p, err := Setup(cfg)
		assert.NoError(t, err)
		assert.Len(t, p, 2)
	})
}
