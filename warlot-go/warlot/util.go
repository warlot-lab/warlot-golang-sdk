package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// authHeaders builds the authentication headers from the Client configuration.
func (c *Client) authHeaders() http.Header {
	h := http.Header{}
	if c == nil {
		return h
	}
	if c.APIKey != "" {
		h.Set("Authorization", "Bearer "+c.APIKey)
	}
	return h
}

// buildHeaders collects headers from CallOptions into a header map.
func buildHeaders(h http.Header, opts ...CallOption) http.Header {
	co := &callOptions{}
	for _, o := range opts {
		o(co)
	}
	if h == nil {
		h = http.Header{}
	}
	mergeHeaders(h, co.headers)
	return h
}

// mergeHeaders appends values from src into dst.
func mergeHeaders(dst http.Header, src http.Header) {
	if src == nil {
		return
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

// statusOf returns the HTTP status code or zero if the response is nil.
func statusOf(res *http.Response) int {
	if res == nil {
		return 0
	}
	return res.StatusCode
}

// parseAPIError decodes an error body and captures message/code/details when available.
// It supports RFC 7807 problem details emitted by the gateway as well as flat envelopes.
func parseAPIError(code int, b []byte) *APIError {
	apiErr := &APIError{StatusCode: code, Body: string(b)}
	if len(b) == 0 {
		return apiErr
	}

	// 1. Try RFC 7807 problem structure: {"error": {"code": "...", "message": "...", "request_id": "..."}}
	var rfcProblem struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &rfcProblem); err == nil && (rfcProblem.Error.Message != "" || rfcProblem.Error.Code != "") {
		apiErr.Message = rfcProblem.Error.Message
		apiErr.Code = rfcProblem.Error.Code
		apiErr.RequestID = rfcProblem.Error.RequestID
		return apiErr
	}

	// 2. Try flat structure: {"message": "...", "error": "...", "code": "...", "details": ...}
	var flatMsg struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Code    string `json:"code"`
		Details any    `json:"details"`
	}
	if err := json.Unmarshal(b, &flatMsg); err == nil {
		if flatMsg.Message != "" {
			apiErr.Message = flatMsg.Message
		} else if flatMsg.Error != "" {
			apiErr.Message = flatMsg.Error
		}
		apiErr.Code = flatMsg.Code
		apiErr.Details = flatMsg.Details
	}

	return apiErr
}

// ClassOfRoute maps an HTTP method and path to the rate limit class enforced by the publisher.
// Chain covers routes that commit state or lease gas on-chain; Ordinary covers reads and SQL executions.
func ClassOfRoute(method, path string) RateLimitClass {
	if method == http.MethodPost {
		if path == "/v1/blobs" || path == "/v1/projects" || strings.HasSuffix(path, "/commit") {
			return RateLimitClassChain
		}
	}
	return RateLimitClassOrdinary
}

// parseRateLimits extracts rate ceiling headers (X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset).
func parseRateLimits(h http.Header) *RateLimit {
	if h == nil {
		return nil
	}
	limitStr := strings.TrimSpace(h.Get("X-RateLimit-Limit"))
	if limitStr == "" {
		return nil
	}
	limitVal, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		return nil
	}
	remVal, _ := strconv.ParseUint(strings.TrimSpace(h.Get("X-RateLimit-Remaining")), 10, 64)
	var resetTime time.Time
	if resetStr := strings.TrimSpace(h.Get("X-RateLimit-Reset")); resetStr != "" {
		if resetEpoch, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
			resetTime = time.Unix(resetEpoch, 0)
		}
	}
	return &RateLimit{
		Limit:     limitVal,
		Remaining: remVal,
		Reset:     resetTime,
	}
}

// parseRetryAfter interprets Retry-After header values (seconds or HTTP-date).
func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs >= 0 {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}


// redactHeaders masks sensitive header values for logging.
func redactHeaders(h http.Header) http.Header {
	if h == nil {
		return h
	}
	cp := http.Header{}
	for k, vs := range h {
		for _, v := range vs {
			lowerK := strings.ToLower(k)
			if lowerK == "authorization" || lowerK == "x-api-key" || lowerK == "cookie" {
				if len(v) > 12 {
					cp.Add(k, v[:6]+"…"+v[len(v)-4:])
				} else {
					cp.Add(k, "********")
				}
			} else {
				cp.Add(k, v)
			}
		}
	}
	return cp
}

// randFloat64 returns a pseudo-random value in [0,1).
// It is based on the monotonic clock to avoid a global RNG.
func randFloat64() float64 {
	n := time.Now().UnixNano()
	n ^= n << 13
	n ^= n >> 7
	n ^= n << 17
	if n < 0 {
		n = -n
	}
	return float64(n%1000) / 1000.0
}

// normalizeBackoff ensures sane defaults for backoff windows.
func normalizeBackoff(initial, max time.Duration) (time.Duration, time.Duration) {
	if initial <= 0 {
		initial = 200 * time.Millisecond
	}
	if max <= 0 {
		max = 2 * time.Second
	}
	return initial, max
}

// normalizeRetries ensures non-negative retry counts.
func normalizeRetries(r int) int {
	if r < 0 {
		return 0
	}
	return r
}

// jitterSleep sleeps for a randomized duration based on the current backoff.
// Context cancellation is respected.
func jitterSleep(ctx context.Context, backoff, maxBack time.Duration) {
	jitter := time.Duration(float64(backoff) * (0.5 + 0.5*randFloat64()))
	if jitter > maxBack {
		jitter = maxBack
	}
	timer := time.NewTimer(jitter)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

// nextBackoff doubles backoff up to maxBack.
func nextBackoff(backoff, maxBack time.Duration) time.Duration {
	backoff *= 2
	if backoff > maxBack {
		backoff = maxBack
	}
	return backoff
}
