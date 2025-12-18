// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package client

// https://www.terraform.io/docs/cloud/api/index.html

import (
	"os"
)

const DefaultTFEAddr = "https://127.0.0.1"

// NewTFEAPI returns an APIClient for TFE.
func NewTFEAPI() (*APIClient, error) {
	product := "terraform-ent"

	addr := os.Getenv("TFE_HTTP_ADDR")
	if addr == "" {
		addr = DefaultTFEAddr
	}

	headers := map[string]string{}
	token := os.Getenv("TFE_TOKEN")
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	cfg := APIConfig{
		Product:   product,
		BaseURL:   addr,
		TLSConfig: TLSConfig{}, // Custom TLS Configuration is not yet supported for Terraform Enterprise
		headers:   headers,
	}

	apiClient, err := NewAPIClient(cfg)
	if err != nil {
		return nil, err
	}
	return apiClient, nil
}
