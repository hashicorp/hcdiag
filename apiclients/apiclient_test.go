package apiclients

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

const (
	testProduct = "test"
	testBaseURL = "test://local"
	testBody    = `{"hello":"there"}`
)

type mockHTTP struct {
	called []*http.Request
}

func (m *mockHTTP) Do(r *http.Request) (*http.Response, error) {
	m.called = append(m.called, r)
	return &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(testBody)),
		StatusCode: 200,
	}, nil
}

// This is not a super thorough test, but it's something
func TestAPIClientGet(t *testing.T) {
	// set up mock
	headers := map[string]string{
		"special": "headeroni",
	}
	mock := &mockHTTP{}
	c := NewAPIClient(testProduct, testBaseURL, headers)
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

	// ensure response is an interface (be Marshal-able) and matches our testBody
	bodyBts, _ := json.Marshal(resp)
	if string(bodyBts) != testBody {
		t.Errorf("expected url '%s'; got '%s'", testBody, bodyBts)
	}
}