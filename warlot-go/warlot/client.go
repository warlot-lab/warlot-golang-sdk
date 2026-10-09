// Package warlot provides a typed Go client for the Warlot SQL Database API.
// The client wraps HTTP transport, retries, authentication headers, and
// response decoding with strongly-typed helpers for common operations.
//
// This package is designed for production usage. Public interfaces include
// thorough documentation and avoid unstable implementation details.
package warlot

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// Logger defines an optional structured logging hook. Implementations should
// avoid recording sensitive values. The SDK already redacts API keys in headers.
type Logger func(event string, metadata map[string]any)

// Client contains shared configuration and HTTP plumbing for the SDK.
// Once configured, a Client is safe for concurrent use by multiple goroutines.
type Client struct {
	// BaseURL is the API origin (defaults to https://api.warlot.stevenhert.xyz).
	BaseURL string

	// APIKey is sent as a Bearer token in the Authorization header for authenticated operations.
	// It can be set during construction or after key issuance.
	APIKey string

	// HolderID is an optional default holder identifier for project provisioning operations.
	HolderID string

	// ProjectName is an optional default project name for project provisioning operations.
	ProjectName string

	// HTTPClient is the underlying HTTP client. A tuned default is provided
	// and can be replaced via WithHTTPClient.
	HTTPClient *http.Client

	// UserAgent is added to each request.
	UserAgent string

	// Retry configuration controls jittered exponential backoff for 429/5xx.
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration

	// Observability hooks.
	Logger      Logger
	BeforeHooks []func(*http.Request)
	AfterHooks  []func(*http.Response, []byte, error)

	// Rate limiting state observed from response headers.
	mu            sync.RWMutex
	lastRateLimit *RateLimit
}

// LastRateLimit returns the most recently observed rate limit ceiling status, if available.
func (c *Client) LastRateLimit() *RateLimit {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lastRateLimit == nil {
		return nil
	}
	cp := *c.lastRateLimit
	return &cp
}


// New constructs a Client with safe defaults. Options can override defaults.
func New(opts ...Option) *Client {
	c := &Client{
		BaseURL: "https://api.warlot.stevenhert.xyz",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
			},
		},
		UserAgent:      "warlot-go/0.2 (+https://github.com/yourorg/warlot-go)",
		MaxRetries:     3,
		InitialBackoff: 300 * time.Millisecond,
		MaxBackoff:     3 * time.Second,
	}
	for _, f := range opts {
		f(c)
	}
	return c
}
