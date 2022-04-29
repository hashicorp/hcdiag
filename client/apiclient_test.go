package client

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockHTTP struct {
	called []*http.Request
	resp   string

	Transport *http.Transport
}

func (m *mockHTTP) Do(r *http.Request) (*http.Response, error) {
	m.called = append(m.called, r)
	return &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(m.resp)),
		StatusCode: 200,
	}, nil
}

// This is not a super thorough test, but it's something
func TestAPIClientGet(t *testing.T) {
	// set up mock
	testBaseURL := "test://local"
	testResp := `{"hello":"there"}`
	headers := map[string]string{
		"special": "headeroni",
	}
	mock := &mockHTTP{resp: testResp}
	c, err := NewAPIClient("test", testBaseURL, headers, TLSConfig{})
	if err != nil {
		t.Errorf("NewAPIClient returned error: %s", err)
	}
	c.http = mock

	// make the request
	testPath := "/test/path"
	resp, _ := c.Get(testPath)

	// only one request expected
	if len(mock.called) != 1 {
		t.Errorf("expected 1 httpClient.Do's; got %d", len(mock.called))
	}

	// convenience
	req := mock.called[0]

	// ensure we tried to hit the right URL
	expectURL := testBaseURL + testPath
	actualURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path
	if expectURL != actualURL {
		t.Errorf("expected url '%s'; got '%s'", expectURL, actualURL)
	}

	// check request headers
	// this one is default for all requests
	if req.Header["Content-Type"][0] != "application/json" {
		t.Errorf("expected 'Content-Type' header 'application/json'; got '%s'", req.Header["Content-Type"][0])
	}
	// this is a special headeroni
	if req.Header["Special"][0] != "headeroni" {
		t.Errorf("expected 'special' header 'headeroni'; got '%s'", req.Header["Special"][0])
	}

	// ensure response is an interface (be Marshal-able) and matches our testResp
	bodyBts, _ := json.Marshal(resp)
	if string(bodyBts) != testResp {
		t.Errorf("expected url '%s'; got '%s'", testResp, bodyBts)
	}
}

func TestAPIClientGetStringValue(t *testing.T) {
	// this also implicily tests APIClient.GetValue()

	mock := &mockHTTP{resp: `{"one": {"two": "three"}}`}
	c, err := NewAPIClient("test", "test://local", map[string]string{}, TLSConfig{})
	if err != nil {
		t.Errorf("NewAPIClient returned error: %s", err)
	}
	c.http = mock

	resp, err := c.GetStringValue("/fake/path", "one", "two")
	if err != nil {
		t.Errorf("error making mock API call: %s", err)
	}
	if resp != "three" {
		t.Errorf("expected resp='three', got: '%s'", resp)
	}

}

func TestCreateTLSClientConfig(t *testing.T) {
	testCases := []struct {
		name      string
		expectErr bool
		input     TLSConfig
	}{
		{
			name:  "Test Empty TLSConfig Returns Empty tls.Config",
			input: TLSConfig{},
		},
		{
			name: "Test InsecureSkipVerify",
			input: TLSConfig{
				Insecure: true,
			},
		},
		{
			name: "Test Server Name",
			input: TLSConfig{
				TLSServerName: "server.domain",
			},
		},
		{
			name: "Test Missing ClientKey with ClientCert Returns Error",
			input: TLSConfig{
				ClientKey: "testdata/signed.key",
			},
			expectErr: true,
		},
		{
			name: "Test Missing ClientCert with ClientKey Returns Error",
			input: TLSConfig{
				ClientCert: "testdata/signed.crt",
			},
			expectErr: true,
		},
		{
			name: "Test Client Cert and Key",
			input: TLSConfig{
				ClientCert: "testdata/signed.crt",
				ClientKey:  "testdata/signed.key",
			},
		},
		{
			name: "Test CA File",
			input: TLSConfig{
				CACert: "testdata/ca/ca.crt",
			},
		},
		{
			name: "Test CA Path",
			input: TLSConfig{
				CAPath: "testdata/ca/",
			},
		},
		{
			name: "Test Bad CA File Path Returns Error",
			input: TLSConfig{
				CACert: "/this/file/does/not/exist/ca.crt",
			},
			expectErr: true,
		},
		{
			name: "Test All Field Types Set",
			input: TLSConfig{
				CACert:        "testdata/ca/ca.crt",
				ClientCert:    "testdata/signed.crt",
				ClientKey:     "testdata/signed.key",
				TLSServerName: "server.domain",
				Insecure:      true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o, err := createTLSClientConfig(tc.input)
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not raised")
				require.Nil(t, o)
			} else {
				// Basic Struct Validation
				require.NoError(t, err)
				require.Equal(t, tc.input.Insecure, o.InsecureSkipVerify)
				require.Equal(t, tc.input.TLSServerName, o.ServerName)

				// CA Validation
				if tc.input.CACert != "" || tc.input.CAPath != "" || tc.input.CACertBytes != nil {
					require.NotNil(t, o.RootCAs)
				}

				// Client Cert & Key Validation
				if tc.input.ClientCert != "" && tc.input.ClientKey != "" {
					require.NotNil(t, o.GetClientCertificate)
					c, e := o.GetClientCertificate(&tls.CertificateRequestInfo{})
					require.NoError(t, e)
					require.NotNil(t, c)
				}
			}
		})
	}
}
