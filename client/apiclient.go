package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/hashicorp/go-rootcerts"
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

	// Insecure enables or disables SSL verification. Setting to `true` is highly
	// discouraged.
	Insecure bool
}

// NewAPIClient gets an API client for a product.
func NewAPIClient(product, baseURL string, headers map[string]string, t TLSConfig) (*APIClient, error) {
	transport := http.DefaultTransport.(*http.Transport)

	// If TLSConfig is set, then we need to configure the TLSClientConfig on the transport.
	if !reflect.DeepEqual(t, TLSConfig{}) {
		tlsClientConfig, err := createTLSClientConfig(t)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = tlsClientConfig
	}

	client := &http.Client{
		Transport: transport,
	}

	return &APIClient{
		Product: product,
		BaseURL: baseURL,
		headers: headers,
		http:    client,
	}, nil
}

func createTLSClientConfig(t TLSConfig) (*tls.Config, error) {
	tlsClientConfig := &tls.Config{}

	// If the input TLSConfig object is default, we do not need to update tlsClientConfig.
	if reflect.DeepEqual(t, TLSConfig{}) {
		return tlsClientConfig, nil
	}

	if t.CACert != "" || len(t.CACertBytes) != 0 || t.CAPath != "" {
		rootConfig := &rootcerts.Config{
			CAFile:        t.CACert,
			CACertificate: t.CACertBytes,
			CAPath:        t.CAPath,
		}
		if err := rootcerts.ConfigureTLS(tlsClientConfig, rootConfig); err != nil {
			return nil, err
		}
	}

	tlsClientConfig.InsecureSkipVerify = t.Insecure

	if t.TLSServerName != "" {
		tlsClientConfig.ServerName = t.TLSServerName
	}

	var clientCert tls.Certificate
	foundClientCert := false

	switch {
	case t.ClientCert != "" && t.ClientKey != "":
		var err error
		clientCert, err = tls.LoadX509KeyPair(t.ClientCert, t.ClientKey)
		if err != nil {
			return nil, err
		}
		foundClientCert = true
	case t.ClientCert != "" || t.ClientKey != "":
		return nil, fmt.Errorf("both client cert and client key must be provided")
	}

	if foundClientCert {
		// We use `GetClientCertificate` here because it works with Vault, along with other products. In Vault
		// client authentication can use a different CA than the one used for Vault's certificate. However, it
		// will indicate that its CA should be used when sending the client certificate request. If the client
		// certificate does not share this CA, Go will fail with a "remote error: tls: bad certificate" error.
		// By providing an override for `GetClientCertificate`, we ensure that we send the client certificate
		// anytime one is requested, even if the CA does not match.
		//
		// See GitHub issue https://github.com/hashicorp/vault/issues/2946 for more context on why Vault
		// uses this mechanism when building their own clients.
		tlsClientConfig.GetClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &clientCert, nil
		}
	}

	return tlsClientConfig, nil
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
