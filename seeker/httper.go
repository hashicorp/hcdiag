package seeker

import "fmt"

func NewHTTPer(url string, mustSucceed bool) *Seeker {
	return &Seeker{
		Identifier:  url,
		Runner:      HTTPer{URL: url},
		MustSucceed: mustSucceed,
	}
}

// HTTPer hits APIs.
type HTTPer struct {
	URL string `json:"url"`
}

func (h HTTPer) Run() (interface{}, error) {
	c := Commander{ // TODO: make it actually do apis
		Command: fmt.Sprintf(`echo {"url":"%s","TODO":"actually hit api"}`, h.URL),
		format:  "json",
	}
	return c.Run()
}
