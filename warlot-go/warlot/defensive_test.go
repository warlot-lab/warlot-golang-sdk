package warlot

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestDefensive_NilProjectClient(t *testing.T) {
	ctx := context.Background()
	p := Project{ID: "proj-uninit", Client: nil}

	if _, err := p.SQL(ctx, "SELECT 1", nil); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.SQL, got: %v", err)
	}
	if _, err := p.Tables(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Tables, got: %v", err)
	}
	if _, err := p.Browse(ctx, "users", 10, 0); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Browse, got: %v", err)
	}
	if _, err := p.Schema(ctx, "users"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Schema, got: %v", err)
	}
	if _, err := p.Count(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Count, got: %v", err)
	}
	if _, err := p.Patch(ctx, "users", nil); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Patch, got: %v", err)
	}
	if _, err := p.Status(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Status, got: %v", err)
	}
	if _, err := p.Commit(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Commit, got: %v", err)
	}
	if _, err := p.Deactivate(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Deactivate, got: %v", err)
	}
	if _, err := p.Reactivate(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Reactivate, got: %v", err)
	}
	if _, err := p.Terminate(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from p.Terminate, got: %v", err)
	}
}

func TestDefensive_NilClientMethods(t *testing.T) {
	ctx := context.Background()
	var cl *Client

	if err := cl.doJSON(ctx, http.MethodGet, "/test", nil, nil, nil); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.doJSON, got: %v", err)
	}
	if _, err := cl.doRequest(ctx, http.MethodGet, "/test", nil, nil); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.doRequest, got: %v", err)
	}
	if _, err := cl.CheckReadiness(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.CheckReadiness, got: %v", err)
	}
	if rl := cl.LastRateLimit(); rl != nil {
		t.Fatalf("expected nil from nil cl.LastRateLimit, got: %v", rl)
	}
	p := cl.Project("test-proj")
	if p.ID != "test-proj" || p.Client != nil {
		t.Fatalf("unexpected project handle from nil client: %+v", p)
	}
	idx := cl.Indexer("")
	if idx == nil {
		t.Fatalf("expected non-nil IndexerClient from nil cl.Indexer")
	}

	// SQL execution
	if _, err := cl.ExecSQL(ctx, "proj-1", SQLRequest{SQL: "SELECT 1"}); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.ExecSQL, got: %v", err)
	}
	if _, err := cl.ExecSQLStream(ctx, "proj-1", SQLRequest{SQL: "SELECT 1"}); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.ExecSQLStream, got: %v", err)
	}

	// Table operations
	if _, err := cl.ListTables(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.ListTables, got: %v", err)
	}
	if _, err := cl.BrowseRows(ctx, "proj-1", "tbl", 10, 0); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.BrowseRows, got: %v", err)
	}
	if _, err := cl.GetTableSchema(ctx, "proj-1", "tbl"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.GetTableSchema, got: %v", err)
	}
	if _, err := cl.GetTableCount(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.GetTableCount, got: %v", err)
	}
	if _, err := cl.PatchTable(ctx, "proj-1", "tbl", PatchTableRequest{}); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.PatchTable, got: %v", err)
	}

	// Project status & commit
	if _, err := cl.GetProjectStatus(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.GetProjectStatus, got: %v", err)
	}
	if _, err := cl.CommitProject(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.CommitProject, got: %v", err)
	}

	// Lifecycle operations
	if _, err := cl.InitProject(ctx, InitProjectRequest{}); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.InitProject, got: %v", err)
	}
	if _, err := cl.IssueAPIKey(ctx, IssueKeyRequest{}); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.IssueAPIKey, got: %v", err)
	}
	if _, err := cl.ResolveProject(ctx, ResolveProjectRequest{}); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.ResolveProject, got: %v", err)
	}
	if _, err := cl.DeactivateProject(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.DeactivateProject, got: %v", err)
	}
	if _, err := cl.ReactivateProject(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.ReactivateProject, got: %v", err)
	}
	if _, err := cl.TerminateProject(ctx, "proj-1"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil cl.TerminateProject, got: %v", err)
	}

	// Internal authHeaders safety
	if h := cl.authHeaders(); len(h) != 0 {
		t.Fatalf("expected empty header from nil cl.authHeaders, got: %v", h)
	}
}

func TestDefensive_NilIndexerClient(t *testing.T) {
	ctx := context.Background()
	var idx *IndexerClient

	if _, err := idx.GetAddressStorage(ctx, "0x123"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetAddressStorage, got: %v", err)
	}
	if _, err := idx.GetAddressBalances(ctx, "0x123"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetAddressBalances, got: %v", err)
	}
	if _, err := idx.GetAddressHistory(ctx, "0x123", "", 10); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetAddressHistory, got: %v", err)
	}
	if _, err := idx.GetHoldersByAdmin(ctx, "0x123"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetHoldersByAdmin, got: %v", err)
	}
	if _, err := idx.GetHolderProjects(ctx, "0x123"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetHolderProjects, got: %v", err)
	}
	if _, err := idx.GetAddressLifespan(ctx, "0x123"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetAddressLifespan, got: %v", err)
	}
	if _, err := idx.GetAddressLifespanBands(ctx, "0x123"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetAddressLifespanBands, got: %v", err)
	}
	if _, err := idx.CheckHealth(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.CheckHealth, got: %v", err)
	}
	if _, err := idx.GetHealth(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetHealth, got: %v", err)
	}
	if err := idx.CheckReadiness(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.CheckReadiness, got: %v", err)
	}
	if err := idx.GetReadiness(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.GetReadiness, got: %v", err)
	}
}

func TestDefensive_NilPagerAndScanner(t *testing.T) {
	ctx := context.Background()
	var p *Pager
	rows, err := p.Next(ctx)
	if err != nil || rows != nil {
		t.Fatalf("expected nil rows and nil err from nil Pager, got %v, %v", rows, err)
	}

	// Pager with uninitialized Client in Project and zero Limit
	pUninit := &Pager{
		Project: Project{ID: "proj-1", Client: nil},
		Table:   "notes",
		Limit:   0,
	}
	rows, err = pUninit.Next(ctx)
	if !errors.Is(err, ErrNilClient) || rows != nil {
		t.Fatalf("expected ErrNilClient from uninitialized Pager, got rows=%v err=%v", rows, err)
	}

	var s *RowScanner
	if s.Next(nil) {
		t.Fatalf("expected false from nil RowScanner.Next")
	}
	if s.Err() != nil {
		t.Fatalf("expected nil from nil RowScanner.Err, got: %v", s.Err())
	}
	if err := s.Close(); err != nil {
		t.Fatalf("expected nil from nil RowScanner.Close, got: %v", err)
	}
}

func TestDefensive_QueryAndMigrate(t *testing.T) {
	ctx := context.Background()
	p := Project{ID: "proj-uninit", Client: nil}

	if _, err := Query[map[string]any](ctx, p, "SELECT 1", nil); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from Query with nil Client, got: %v", err)
	}

	if _, err := Migrate.Up(ctx, p, nil, "migrations"); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from Migrate.Up with nil Client, got: %v", err)
	}

	var apiErr *APIError
	if msg := apiErr.Error(); msg != "<nil>" {
		t.Fatalf("expected <nil> from nil APIError.Error(), got: %q", msg)
	}

	var readyResp *ReadinessResponse
	if readyResp.IsReady() {
		t.Fatalf("expected false from nil ReadinessResponse.IsReady()")
	}
}

func TestDefensive_NilOptionsAndCallOptions(t *testing.T) {
	// Options applied to nil Client must not panic
	WithBaseURL("https://example.com")(nil)
	WithAPIKey("test-key")(nil)
	WithHolderID("holder-1")(nil)
	WithProjectName("proj-1")(nil)
	WithHTTPClient(nil)(nil)
	WithUserAgent("custom-ua")(nil)
	WithRetries(2)(nil)
	WithBackoff(10, 20)(nil)
	WithLogger(nil)(nil)
	WithBeforeHook(nil)(nil)
	WithAfterHook(nil)(nil)

	// CallOptions applied to nil callOptions must not panic
	WithIdempotencyKey("idem-1")(nil)
	WithHeader("k", "v")(nil)
	WithLabel("test-label")(nil)
}

func TestOptions_WithHooks(t *testing.T) {
	var beforeCalled, afterCalled bool
	cl := New(
		WithBeforeHook(func(r *http.Request) { beforeCalled = true }),
		WithAfterHook(func(res *http.Response, body []byte, err error) { afterCalled = true }),
	)

	if len(cl.BeforeHooks) != 1 || len(cl.AfterHooks) != 1 {
		t.Fatalf("expected 1 before hook and 1 after hook, got %d and %d", len(cl.BeforeHooks), len(cl.AfterHooks))
	}
	cl.BeforeHooks[0](nil)
	cl.AfterHooks[0](nil, nil, nil)
	if !beforeCalled || !afterCalled {
		t.Fatalf("hooks were not invoked properly")
	}
}
