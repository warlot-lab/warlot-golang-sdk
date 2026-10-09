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
	if _, err := idx.CheckHealth(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.CheckHealth, got: %v", err)
	}
	if err := idx.CheckReadiness(ctx); !errors.Is(err, ErrNilClient) {
		t.Fatalf("expected ErrNilClient from nil idx.CheckReadiness, got: %v", err)
	}
}

func TestDefensive_NilPagerAndScanner(t *testing.T) {
	ctx := context.Background()
	var p *Pager
	rows, err := p.Next(ctx)
	if err != nil || rows != nil {
		t.Fatalf("expected nil rows and nil err from nil Pager, got %v, %v", rows, err)
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
