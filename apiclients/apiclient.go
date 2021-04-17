package apiclients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// NewAPIClient gets an API client for a product.
func NewAPIClient(product, baseURL string, headers map[string]string) *APIClient {
	return &APIClient{
		Product: product,
		BaseURL: baseURL,
		headers: headers,
		http:    http.DefaultClient,
	}
}

// APIClient can make API calls.
type APIClient struct {
	Product string `json:"product"`
	BaseURL string `json:"baseurl"`
	// headers may contain secrets, so *do not export*
	headers map[string]string
	http    HTTPClient
}

type APIResponse struct {
	Errors []string `json:"errors"`
}

// We *only* want to retrieve information, never put/post/delete
func (c *APIClient) Get(path string) (interface{}, error) {
	return c.request("GET", path, []byte{})
}

func (c *APIClient) request(method string, path string, data []byte) (interface{}, error) {
	// Build request
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Make request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Grab response contents
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert to interface{}
	var iface interface{}
	err = json.Unmarshal(body, &iface)

	return iface, err
}
