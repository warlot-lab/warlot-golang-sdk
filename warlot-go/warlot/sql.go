package warlot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// ExecSQL executes a parameterized SQL statement within a project.
// Both DDL/DML and SELECT responses are supported.
func (c *Client) ExecSQL(ctx context.Context, projectID string, req SQLRequest, opts ...CallOption) (*SQLResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/sql", url.PathEscape(projectID))
	var out SQLResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, path, h, req, &out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		return &out, errors.New(out.Error)
	}
	// Normalize response: successful execution sets OK to true even when
	// the upstream server omits the boolean on SELECT queries.
	out.OK = true
	if out.RowCount == nil && out.Rows != nil {
		n := len(out.Rows)
		out.RowCount = &n
	}
	return &out, nil
}
