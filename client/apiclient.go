package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/hcdiag/util"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// TLSConfig contains the parameters needed to configure TLS on the HTTP client
// used to communicate with an API.
type TLSConfig struct {
	// CACert is the path to a PEM-encoded CA cert file to use to verify the
	// server SSL certificate. It takes precedence over CACertBytes
	// and CAPath.
	CACert string

	// CACertBytes is a PEM-encoded certificate or bundle. It takes precedence
	// over CAPath.
	CACertBytes []byte

	// CAPath is the path to a directory of PEM-encoded CA cert files to verify
	// the server SSL certificate.
	CAPath string

	// ClientCert is the path to the certificate for communication
	ClientCert string

	// ClientKey is the path to the private key for communication
	ClientKey string

	// TLSServerName, if set, is used to set the SNI host when connecting via
	// TLS.
	TLSServerName string

	// Insecure enables or disables SSL verification. Setting to `false` is highly
	// discouraged.
	Insecure bool
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

// Get makes a GET request to a given path.
func (c *APIClient) Get(path string) (interface{}, error) {
	return c.request("GET", path, []byte{})
}

// GetValue runs Get() then looks through the response for nested mapKeys.
func (c *APIClient) GetValue(path string, mapKeys ...string) (interface{}, error) {
	i, err := c.Get(path)
	if err != nil {
		return nil, err
	}
	return util.FindInInterface(i, mapKeys...)
}

// GetStringValue runs GetValue() then casts the result to a string.
func (c *APIClient) GetStringValue(path string, mapKeys ...string) (string, error) {
	i, err := c.GetValue(path, mapKeys...)
	if err != nil {
		return "", err
	}
	v, ok := i.(string)
	if !ok {
		return "", fmt.Errorf("unable to cast '%#v' to string", i)
	}
	return v, nil
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

	// Error-return the status code if it's not 200 OK
	if resp.StatusCode != http.StatusOK {
		return iface, errors.New(resp.Status)
	}

	return iface, err
}
