package warlot

import (
	"strings"
	"testing"
)

func TestParseAPIError_RFC7807(t *testing.T) {
	raw := []byte(`{
		"error": {
			"code": "invalid_request",
			"message": "a sql statement is required",
			"request_id": "84a7e94e-b5c6-4d0f-a912-32a588b56d32"
		}
	}`)

	err := parseAPIError(400, raw)
	if err.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", err.StatusCode)
	}
	if err.Code != "invalid_request" {
		t.Fatalf("expected code invalid_request, got %q", err.Code)
	}
	if err.Message != "a sql statement is required" {
		t.Fatalf("expected message 'a sql statement is required', got %q", err.Message)
	}
	if err.RequestID != "84a7e94e-b5c6-4d0f-a912-32a588b56d32" {
		t.Fatalf("expected request_id '84a7e94e-b5c6-4d0f-a912-32a588b56d32', got %q", err.RequestID)
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "invalid_request") || !strings.Contains(errStr, "84a7e94e-b5c6") {
		t.Fatalf("formatted error missing code or req ID: %s", errStr)
	}
}

func TestParseAPIError_FlatFallback(t *testing.T) {
	raw := []byte(`{"error": "table not found", "code": "not_found"}`)
	err := parseAPIError(404, raw)
	if err.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", err.StatusCode)
	}
	if err.Message != "table not found" {
		t.Fatalf("expected message 'table not found', got %q", err.Message)
	}
	if err.Code != "not_found" {
		t.Fatalf("expected code 'not_found', got %q", err.Code)
	}
}

func TestRedactHeaders(t *testing.T) {
	h := map[string][]string{
		"Authorization": {"Bearer secret_token_123456789"},
		"X-API-Key":     {"secret_api_key_abcdef123"},
		"Cookie":        {"session=secret_cookie_token_987"},
		"X-Custom":      {"public-value"},
	}
	redacted := redactHeaders(h)
	authVal := redacted.Get("Authorization")
	if strings.Contains(authVal, "secret_token_123456789") {
		t.Fatalf("authorization header was not redacted: %s", authVal)
	}
	apiVal := redacted.Get("X-API-Key")
	if strings.Contains(apiVal, "secret_api_key_abcdef123") {
		t.Fatalf("x-api-key header was not redacted: %s", apiVal)
	}
	cookieVal := redacted.Get("Cookie")
	if strings.Contains(cookieVal, "secret_cookie_token_987") {
		t.Fatalf("cookie header was not redacted: %s", cookieVal)
	}
	if redacted.Get("X-Custom") != "public-value" {
		t.Fatalf("non-sensitive header was altered: %s", redacted.Get("X-Custom"))
	}
}

func TestParseAPIError_PublisherVariants(t *testing.T) {
	// Test 401 unauthenticated problem details
	raw401 := []byte(`{
		"error": {
			"code": "unauthenticated",
			"message": "a credential is required",
			"request_id": "c1f7b022-901e-4501-8ef1-987812903abc"
		}
	}`)
	err401 := parseAPIError(401, raw401)
	if err401.StatusCode != 401 || err401.Code != "unauthenticated" || err401.RequestID != "c1f7b022-901e-4501-8ef1-987812903abc" {
		t.Fatalf("unexpected 401 error: %+v", err401)
	}

	// Test 429 rate_limited problem details
	raw429 := []byte(`{
		"error": {
			"code": "rate_limited",
			"message": "a ceiling of 120 requests per minute applies to this credential; retry in 60 seconds",
			"request_id": "06009d76-be10-44ec-a6c8-c1c471437a46"
		}
	}`)
	err429 := parseAPIError(429, raw429)
	if err429.StatusCode != 429 || err429.Code != "rate_limited" || err429.RequestID != "06009d76-be10-44ec-a6c8-c1c471437a46" {
		t.Fatalf("unexpected 429 error: %+v", err429)
	}
}
