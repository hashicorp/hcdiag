package apiclients

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTFEAPI(t *testing.T) {
	api := NewTFEAPI()
	assert.Equal(t, "terraform-ent", api.Product)
	assert.Equal(t, api.BaseURL, DefaultTFEAddr)
}
