package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/hcdiag/client"
	"github.com/hashicorp/hcdiag/op"
	"github.com/hashicorp/hcdiag/redact"
)

var _ Runner = HTTP{}

// HTTP hits APIs.
type HTTP struct {
	// Parameters that are not shared/common
	Path   string            `json:"path"`
	Client *client.APIClient `json:"client"`

	// Parameters that are common across runner types
	ctx context.Context

	Timeout    Timeout          `json:"timeout"`
	Redactions []*redact.Redact `json:"redactions"`
}

// HttpConfig is the configuration object passed into NewHTTP or NewHTTPWithContext. It includes
// the fields that those constructors will use to configure the HTTP object that they return.
type HttpConfig struct {
	// Client is the client.APIClient that will be used to make HTTP requests.
	Client *client.APIClient

	// Path is the path portion of the URL that the runner will hit.
	Path string

	// Timeout specifies the amount of time that the runner should be allowed to execute before cancellation.
	Timeout time.Duration

	// Redactions includes any redactions to apply to the output of the runner.
	Redactions []*redact.Redact
}

func NewHTTP(cfg HttpConfig) (*HTTP, error) {
	return NewHTTPWithContext(context.Background(), cfg)
}

func NewHTTPWithContext(ctx context.Context, cfg HttpConfig) (*HTTP, error) {
	if cfg.Client == nil {
		return nil, HTTPConfigError{
			config: cfg,
			err:    fmt.Errorf("client must be non-nil when creating an HTTP runner"),
		}
	}

	timeout := cfg.Timeout
	if timeout < 0 {
		return nil, HTTPConfigError{
			config: cfg,
			err:    fmt.Errorf("timeout must be a nonnegative value, but got '%s'", timeout.String()),
		}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return &HTTP{
		ctx:        ctx,
		Client:     cfg.Client,
		Path:       cfg.Path,
		Timeout:    Timeout(cfg.Timeout),
		Redactions: cfg.Redactions,
	}, nil
}

func (h HTTP) ID() string {
	return "GET" + " " + h.Path
}

// Run executes a GET request to the Path using the Client
func (h HTTP) Run() op.Op {
	// protect from accidental nil reference panics
	if h.ctx == nil {
		h.ctx = context.Background()
	}

	runCtx := h.ctx
	var runCancelFunc context.CancelFunc
	if h.Timeout > 0 {
		runCtx, runCancelFunc = context.WithTimeout(h.ctx, time.Duration(h.Timeout))
		defer runCancelFunc()
	}

	startTime := time.Now()

	redactedResponse, err := h.Client.RedactGetWithContext(runCtx, h.Path, h.Redactions)
	result := map[string]any{"response": redactedResponse}
	if err != nil {
		var failureType op.Status
		if errors.Is(err, context.DeadlineExceeded) {
			failureType = op.Timeout
		} else if errors.Is(err, context.Canceled) {
			failureType = op.Canceled
		} else {
			failureType = op.Fail
		}
		return op.New(h.ID(), result, failureType, err, Params(h), startTime, time.Now())
	}

	return op.New(h.ID(), result, op.Success, nil, Params(h), startTime, time.Now())
}

var _ error = HTTPConfigError{}

type HTTPConfigError struct {
	config HttpConfig
	err    error
}

func (e HTTPConfigError) Error() string {
	message := "invalid HTTP Config"
	if e.err != nil {
		return fmt.Sprintf("%s: %s", message, e.err.Error())
	}
	return message
}

func (e HTTPConfigError) Unwrap() error {
	return e.err
}
