package warlot

import "context"

// Project is a light-weight handle bound to a specific project ID.
// It exposes ergonomic helpers that forward to Client methods.
type Project struct {
	ID     string
	Client *Client
}

// Project returns a handle for a given project ID.
func (c *Client) Project(id string) Project { return Project{ID: id, Client: c} }

// SQL executes a SQL statement in the bound project.
func (p Project) SQL(ctx context.Context, sql string, params []any, opts ...CallOption) (*SQLResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.ExecSQL(ctx, p.ID, SQLRequest{SQL: sql, Params: params}, opts...)
}

// Tables lists all tables in the bound project.
func (p Project) Tables(ctx context.Context, opts ...CallOption) (*ListTablesResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.ListTables(ctx, p.ID, opts...)
}

// Browse returns a page of rows for a table in the bound project.
func (p Project) Browse(ctx context.Context, table string, limit, offset int, opts ...CallOption) (*BrowseRowsResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.BrowseRows(ctx, p.ID, table, limit, offset, opts...)
}

// Schema returns a table schema from the bound project.
func (p Project) Schema(ctx context.Context, table string, opts ...CallOption) (*TableSchema, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.GetTableSchema(ctx, p.ID, table, opts...)
}

// Count returns the table count for the bound project.
func (p Project) Count(ctx context.Context, opts ...CallOption) (*TableCountResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.GetTableCount(ctx, p.ID, opts...)
}

// Patch applies structured edits to a table in the bound project.
func (p Project) Patch(ctx context.Context, table string, edits []TableEdit, opts ...CallOption) (*PatchTableResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.PatchTable(ctx, p.ID, table, PatchTableRequest{Edits: edits}, opts...)
}

// Status returns the project status, replication state, and lifecycle metadata.
func (p Project) Status(ctx context.Context, opts ...CallOption) (*ProjectStatus, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.GetProjectStatus(ctx, p.ID, opts...)
}

// Commit triggers a commit for the bound project to anchor operations on-chain.
func (p Project) Commit(ctx context.Context, opts ...CallOption) (*CommitResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.CommitProject(ctx, p.ID, opts...)
}

// Deactivate deactivates the bound project, pausing writes while retaining reads.
func (p Project) Deactivate(ctx context.Context, opts ...CallOption) (*DeactivateProjectResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.DeactivateProject(ctx, p.ID, opts...)
}

// Reactivate reactivates the bound project, restoring write operations.
func (p Project) Reactivate(ctx context.Context, opts ...CallOption) (*ReactivateProjectResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.ReactivateProject(ctx, p.ID, opts...)
}

// Terminate permanently terminates the bound project, shredding data keys.
func (p Project) Terminate(ctx context.Context, opts ...CallOption) (*TerminateProjectResponse, error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	return p.Client.TerminateProject(ctx, p.ID, opts...)
}
