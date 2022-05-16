package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTFEAPI(t *testing.T) {
	api, err := NewTFEAPI()
	assert.NoError(t, err)
	assert.Equal(t, "terraform-ent", api.Product)
	assert.Equal(t, api.BaseURL, DefaultTFEAddr)
}
