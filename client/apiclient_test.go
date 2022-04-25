package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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
