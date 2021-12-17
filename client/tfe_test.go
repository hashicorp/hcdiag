package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTFEAPI(t *testing.T) {
	api := NewTFEAPI()
	assert.Equal(t, "terraform-ent", api.Product)
	assert.Equal(t, api.BaseURL, DefaultTFEAddr)
}
