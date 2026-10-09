package warlot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CheckReadiness queries the cluster /readyz probe to evaluate dependency health.
// On HTTP 503 Service Unavailable, the response is decoded and returned alongside ErrClusterNotReady.
func (c *Client) CheckReadiness(ctx context.Context, opts ...CallOption) (*ReadinessResponse, error) {
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
	u := baseURL + "/readyz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create readiness request: %w", err)
	}

	for k, vs := range buildHeaders(nil, opts...) {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute readiness check: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read readiness response: %w", err)
	}

	var out ReadinessResponse
	if len(body) > 0 {
		if err := json.Unmarshal(body, &out); err != nil {
			return nil, fmt.Errorf("decode readiness response: %w (status=%d, body=%s)", err, res.StatusCode, string(body))
		}
	}

	if res.StatusCode != http.StatusOK {
		return &out, fmt.Errorf("%w (status=%d, readiness=%s)", ErrClusterNotReady, res.StatusCode, out.Status)
	}

	return &out, nil
}
