package warlot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// RowScanner streams rows for large SELECTs without loading the entire set.
type RowScanner struct {
	dec     *json.Decoder
	closer  io.Closer
	inRows  bool
	done    bool
	lastErr error
}

// Next decodes the next row into dst (map or struct pointer). Returns false
// on end of stream or on error. After false, Err should be checked.
func (s *RowScanner) Next(dst any) bool {
	if s == nil || s.done {
		return false
	}
	// Seek to "rows": [
	if !s.inRows {
		for {
			tok, err := s.dec.Token()
			if err != nil {
				s.lastErr = err
				_ = s.Close()
				return false
			}
			if key, ok := tok.(string); ok && key == "rows" {
				tok, err = s.dec.Token()
				if err != nil {
					s.lastErr = err
					_ = s.Close()
					return false
				}
				if delim, ok := tok.(json.Delim); !ok || delim != '[' {
					s.lastErr = fmt.Errorf("unexpected token after rows: %v", tok)
					_ = s.Close()
					return false
				}
				s.inRows = true
				break
			}
		}
	}
	// Decode next element or finish array.
	if s.dec.More() {
		if err := s.dec.Decode(dst); err != nil {
			s.lastErr = err
			_ = s.Close()
			return false
		}
		return true
	}
	// Consume closing ']' and finish.
	_, _ = s.dec.Token()
	s.done = true
	_ = s.Close()
	return false
}

// Err returns the last error encountered by the scanner, if any.
func (s *RowScanner) Err() error {
	if s == nil {
		return nil
	}
	return s.lastErr
}

// Close closes the underlying response body if still open.
func (s *RowScanner) Close() error {
	if s == nil {
		return nil
	}
	if s.closer != nil {
		err := s.closer.Close()
		s.closer = nil
		return err
	}
	return nil
}

// ExecSQLStream executes a SELECT and returns a RowScanner to iterate rows.
// The caller must Close the scanner when finished.
func (c *Client) ExecSQLStream(ctx context.Context, projectID string, req SQLRequest, opts ...CallOption) (*RowScanner, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	path := fmt.Sprintf("/v1/projects/%s/sql", url.PathEscape(projectID))
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	res, err := c.doRequest(ctx, http.MethodPost, path, h, req)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(res.Body)
	tok, err := dec.Token()
	if err != nil {
		_ = res.Body.Close()
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		_ = res.Body.Close()
		return nil, fmt.Errorf("unexpected response start: %v", tok)
	}
	return &RowScanner{dec: dec, closer: res.Body}, nil
}
