package warlot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// ListTables returns the list of tables for a project.
func (c *Client) ListTables(ctx context.Context, projectID string, opts ...CallOption) (*ListTablesResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/tables", url.PathEscape(projectID))
	var out ListTablesResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodGet, path, h, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// BrowseRows returns a page of table rows.
func (c *Client) BrowseRows(ctx context.Context, projectID, table string, limit, offset int, opts ...CallOption) (*BrowseRowsResponse, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}
	path := fmt.Sprintf("/v1/projects/%s/tables/%s/rows", url.PathEscape(projectID), url.PathEscape(table))
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var out BrowseRowsResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodGet, path, h, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTableSchema returns the schema of a table.
func (c *Client) GetTableSchema(ctx context.Context, projectID, table string, opts ...CallOption) (*TableSchema, error) {
	path := fmt.Sprintf("/v1/projects/%s/tables/%s/schema", url.PathEscape(projectID), url.PathEscape(table))
	var out TableSchema
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodGet, path, h, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTableCount returns the number of tables in a project.
func (c *Client) GetTableCount(ctx context.Context, projectID string, opts ...CallOption) (*TableCountResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/tables/count", url.PathEscape(projectID))
	var out TableCountResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodGet, path, h, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PatchTable applies a batch of structured edits to a table.
func (c *Client) PatchTable(ctx context.Context, projectID, table string, req PatchTableRequest, opts ...CallOption) (*PatchTableResponse, error) {
	path := fmt.Sprintf("/v1/projects/%s/tables/%s", url.PathEscape(projectID), url.PathEscape(table))
	var out PatchTableResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPatch, path, h, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
