package warlot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// doJSON sends an HTTP request with a JSON encoded body and decodes a JSON response.
// Retries are performed for 429 and 5xx responses using jittered backoff and
// honoring Retry-After when present.
func (c *Client) doJSON(ctx context.Context, method, path string, hdr http.Header, in, out any) error {
	if c == nil {
		return ErrNilClient
	}
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.warlot.stevenhert.xyz"
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := baseURL + path

	makeBody := func() (io.ReadCloser, []byte, error) {
		if in == nil {
			return nil, nil, nil
		}
		b, err := json.Marshal(in)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request: %w", err)
		}
		return io.NopCloser(bytes.NewReader(b)), b, nil
	}

	var lastErr error
	backoff, maxBack := normalizeBackoff(c.InitialBackoff, c.MaxBackoff)
	retries := normalizeRetries(c.MaxRetries)

	for attempt := 0; attempt <= retries; attempt++ {
		rc, raw, err := makeBody()
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, method, u, rc)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.UserAgent != "" {
			req.Header.Set("User-Agent", c.UserAgent)
		}
		for k, vs := range hdr {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
		if c.Logger != nil {
			c.Logger("request", map[string]any{
				"method": method, "url": u, "headers": redactHeaders(req.Header), "attempt": attempt,
			})
		}
		for _, h := range c.BeforeHooks {
			h(req)
		}

		res, err := httpClient.Do(req)
		var body []byte
		if err == nil {
			defer res.Body.Close()
			body, _ = io.ReadAll(res.Body)
		}
		if c.Logger != nil {
			c.Logger("response", map[string]any{
				"method": method, "url": u, "status": statusOf(res), "attempt": attempt,
			})
		}
		for _, h := range c.AfterHooks {
			h(res, body, err)
		}

		if err != nil {
			lastErr = fmt.Errorf("%s %s: %w", method, u, err)
		} else if res.StatusCode/100 == 2 {
			if out != nil && len(body) > 0 {
				if err := json.Unmarshal(body, out); err != nil {
					lastErr = fmt.Errorf("decode response: %w (body=%s)", err, string(body))
				} else {
					return nil
				}
			} else {
				return nil
			}
		} else {
			apiErr := parseAPIError(res.StatusCode, body)
			if res.StatusCode == http.StatusTooManyRequests || res.StatusCode/100 == 5 {
				lastErr = fmt.Errorf("%s %s: %w", method, u, apiErr)
				if ra := parseRetryAfter(res.Header.Get("Retry-After")); ra > 0 && ra > backoff {
					backoff = ra
				}
			} else {
				return apiErr
			}
		}

		// Stop retrying on non-retriable errors.
		if e, ok := lastErr.(*APIError); ok && e.StatusCode != http.StatusTooManyRequests && e.StatusCode/100 != 5 {
			return lastErr
		}

		// Backoff with jitter.
		if attempt < retries {
			jitterSleep(ctx, backoff, maxBack)
			backoff = nextBackoff(backoff, maxBack)
		}
		_ = raw
	}
	return fmt.Errorf("warlot request failed after %d attempts: %w", retries+1, lastErr)
}

// doRequest is similar to doJSON but returns a raw response for streaming.
// The caller must close the response body.
func (c *Client) doRequest(ctx context.Context, method, path string, hdr http.Header, in any) (*http.Response, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = "https://api.warlot.stevenhert.xyz"
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u := baseURL + path

	makeBody := func() (io.ReadCloser, []byte, error) {
		if in == nil {
			return nil, nil, nil
		}
		b, err := json.Marshal(in)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request: %w", err)
		}
		return io.NopCloser(bytes.NewReader(b)), b, nil
	}

	var lastErr error
	backoff, maxBack := normalizeBackoff(c.InitialBackoff, c.MaxBackoff)
	retries := normalizeRetries(c.MaxRetries)

	for attempt := 0; attempt <= retries; attempt++ {
		rc, _, err := makeBody()
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, method, u, rc)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.UserAgent != "" {
			req.Header.Set("User-Agent", c.UserAgent)
		}
		for k, vs := range hdr {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
		if c.Logger != nil {
			c.Logger("request", map[string]any{
				"method": method, "url": u, "headers": redactHeaders(req.Header), "attempt": attempt,
			})
		}
		for _, h := range c.BeforeHooks {
			h(req)
		}

		res, err := httpClient.Do(req)
		if err == nil && res.StatusCode/100 == 2 {
			return res, nil
		}

		var body []byte
		if err == nil {
			body, _ = io.ReadAll(res.Body)
			res.Body.Close()
		}
		if err != nil {
			lastErr = fmt.Errorf("%s %s: %w", method, u, err)
		} else {
			apiErr := parseAPIError(res.StatusCode, body)
			if res.StatusCode == http.StatusTooManyRequests || res.StatusCode/100 == 5 {
				lastErr = fmt.Errorf("%s %s: %w", method, u, apiErr)
				if ra := parseRetryAfter(res.Header.Get("Retry-After")); ra > 0 && ra > backoff {
					backoff = ra
				}
			} else {
				return nil, apiErr
			}
		}

		if attempt < retries {
			jitterSleep(ctx, backoff, maxBack)
			backoff = nextBackoff(backoff, maxBack)
		}
	}
	return nil, fmt.Errorf("warlot request failed after %d attempts: %w", retries+1, lastErr)
}
