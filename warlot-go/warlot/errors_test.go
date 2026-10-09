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
		"X-Custom":      {"public-value"},
	}
	redacted := redactHeaders(h)
	authVal := redacted.Get("Authorization")
	if strings.Contains(authVal, "secret_token") {
		t.Fatalf("authorization header was not redacted: %s", authVal)
	}
	if redacted.Get("X-Custom") != "public-value" {
		t.Fatalf("non-sensitive header was altered: %s", redacted.Get("X-Custom"))
	}
}
