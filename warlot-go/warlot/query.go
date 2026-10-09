package warlot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Query executes a SQL query on a project and directly unmarshals the returned rows
// into a typed slice of T without intermediate map[string]any allocations or re-serialization loops.
func Query[T any](ctx context.Context, p Project, sql string, params []any, opts ...CallOption) ([]T, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	path := fmt.Sprintf("/v1/projects/%s/sql", url.PathEscape(p.ID))
	h := p.Client.authHeaders()
	mergeHeaders(h, buildHeaders(nil, opts...))

	var envelope struct {
		Rows     []T    `json:"rows"`
		RowCount *int   `json:"row_count"`
		Error    string `json:"error"`
	}
	req := SQLRequest{SQL: sql, Params: params}
	if err := p.Client.doJSON(ctx, http.MethodPost, path, h, req, &envelope); err != nil {
		return nil, err
	}
	if envelope.Error != "" {
		return nil, errors.New(envelope.Error)
	}
	if envelope.Rows == nil {
		return []T{}, nil
	}
	return envelope.Rows, nil
}
