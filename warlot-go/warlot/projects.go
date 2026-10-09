package warlot

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// InitProject initializes a new project and returns its identifiers.
func (c *Client) InitProject(ctx context.Context, req InitProjectRequest, opts ...CallOption) (*InitProjectResponse, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	var out InitProjectResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, "/v1/projects", h, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IssueAPIKey creates an API key for a project, returning the key and URL.
func (c *Client) IssueAPIKey(ctx context.Context, req IssueKeyRequest, opts ...CallOption) (*IssueKeyResponse, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	var out IssueKeyResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, "/auth/issue", h, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResolveProject resolves a project by holder and name.
// Legacy fields are normalized to the modern shape if necessary.
func (c *Client) ResolveProject(ctx context.Context, req ResolveProjectRequest, opts ...CallOption) (*ResolveProjectResponse, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	var out ResolveProjectResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, "/v1/projects/resolve", h, req, &out); err != nil {
		return nil, err
	}
	if out.ProjectID == "" && out.LegacyProjectID != "" {
		out.ProjectID = out.LegacyProjectID
	}
	if out.DBID == "" && out.LegacyDBID != "" {
		out.DBID = out.LegacyDBID
	}
	return &out, nil
}

// DeactivateProject deactivates a project, refusing future writes while keeping reads available.
func (c *Client) DeactivateProject(ctx context.Context, projectID string, opts ...CallOption) (*DeactivateProjectResponse, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	path := fmt.Sprintf("/v1/projects/%s/deactivate", url.PathEscape(projectID))
	var out DeactivateProjectResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, path, h, struct{}{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReactivateProject reactivates a previously deactivated project, restoring write capabilities.
func (c *Client) ReactivateProject(ctx context.Context, projectID string, opts ...CallOption) (*ReactivateProjectResponse, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	path := fmt.Sprintf("/v1/projects/%s/reactivate", url.PathEscape(projectID))
	var out ReactivateProjectResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodPost, path, h, struct{}{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TerminateProject permanently terminates a project, shreds wrapped keys, and deletes on-chain state.
func (c *Client) TerminateProject(ctx context.Context, projectID string, opts ...CallOption) (*TerminateProjectResponse, error) {
	if c == nil {
		return nil, ErrNilClient
	}
	path := fmt.Sprintf("/v1/projects/%s", url.PathEscape(projectID))
	var out TerminateProjectResponse
	h := c.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))
	if err := c.doJSON(ctx, http.MethodDelete, path, h, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
