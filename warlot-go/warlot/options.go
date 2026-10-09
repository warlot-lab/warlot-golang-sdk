package warlot

import (
	"net/http"
	"strings"
	"time"
)

// Option customizes a Client at construction time.
type Option func(*Client)

func WithBaseURL(u string) Option {
	return func(c *Client) {
		if c != nil {
			c.BaseURL = strings.TrimRight(u, "/")
		}
	}
}
func WithAPIKey(k string) Option {
	return func(c *Client) {
		if c != nil {
			c.APIKey = k
		}
	}
}
func WithHolderID(h string) Option {
	return func(c *Client) {
		if c != nil {
			c.HolderID = h
		}
	}
}
func WithProjectName(n string) Option {
	return func(c *Client) {
		if c != nil {
			c.ProjectName = n
		}
	}
}
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if c != nil {
			c.HTTPClient = h
		}
	}
}
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if c != nil {
			c.UserAgent = ua
		}
	}
}
func WithRetries(max int) Option {
	return func(c *Client) {
		if c != nil {
			c.MaxRetries = max
		}
	}
}
func WithBackoff(init, max time.Duration) Option {
	return func(c *Client) {
		if c != nil {
			c.InitialBackoff = init
			c.MaxBackoff = max
		}
	}
}
func WithLogger(l Logger) Option {
	return func(c *Client) {
		if c != nil {
			c.Logger = l
		}
	}
}

// WithBeforeHook appends an HTTP request hook executed before each request attempt.
func WithBeforeHook(h func(*http.Request)) Option {
	return func(c *Client) {
		if c != nil {
			c.BeforeHooks = append(c.BeforeHooks, h)
		}
	}
}

// WithAfterHook appends an HTTP response hook executed after each request attempt.
func WithAfterHook(h func(*http.Response, []byte, error)) Option {
	return func(c *Client) {
		if c != nil {
			c.AfterHooks = append(c.AfterHooks, h)
		}
	}
}

// CallOption customizes a single API call (for example, idempotency keys).
type CallOption func(*callOptions)

type callOptions struct {
	headers http.Header
	label   string
}

// WithIdempotencyKey attaches an idempotency key for write operations.
func WithIdempotencyKey(k string) CallOption {
	return func(co *callOptions) {
		if co == nil {
			return
		}
		if co.headers == nil {
			co.headers = http.Header{}
		}
		co.headers.Set("x-idempotency-key", k)
	}
}

// WithHeader adds an arbitrary header to a single API call.
func WithHeader(key, value string) CallOption {
	return func(co *callOptions) {
		if co == nil {
			return
		}
		if co.headers == nil {
			co.headers = http.Header{}
		}
		co.headers.Add(key, value)
	}
}

// WithLabel sets an optional label for internal diagnostics.
func WithLabel(l string) CallOption {
	return func(co *callOptions) {
		if co == nil {
			return
		}
		co.label = l
	}
}
