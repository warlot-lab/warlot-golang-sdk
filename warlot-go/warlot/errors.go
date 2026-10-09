package warlot

import (
	"errors"
	"fmt"
)

// ErrNilClient indicates that an operation was attempted on an uninitialized or nil Client.
var ErrNilClient = errors.New("warlot: client is nil")

// APIError represents a non-success HTTP response from the API.
// It supports RFC 7807 problem details emitted by the gateway.
type APIError struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Body       string `json:"body,omitempty"`
	Details    any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.Body
	}
	if e.Code != "" && e.RequestID != "" {
		return fmt.Sprintf("warlot API %d (%s, req: %s): %s", e.StatusCode, e.Code, e.RequestID, msg)
	}
	if e.Code != "" {
		return fmt.Sprintf("warlot API %d (%s): %s", e.StatusCode, e.Code, msg)
	}
	return fmt.Sprintf("warlot API %d: %s", e.StatusCode, msg)
}
