package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetry_WithRetryAfter_ThenSuccess(t *testing.T) {
	var attempts int32

	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/projects/x/sql" {
			if atomic.AddInt32(&attempts, 1) == 1 {
				w.Header().Set("Retry-After", "1")
				w.Header().Set("X-RateLimit-Limit", "120")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", "1728489600")
				http.Error(w, `{"error":{"code":"rate_limited","message":"too many requests"}}`, http.StatusTooManyRequests)
				return
			}
			w.Header().Set("X-RateLimit-Limit", "120")
			w.Header().Set("X-RateLimit-Remaining", "119")
			json.NewEncoder(w).Encode(SQLResponse{OK: true, RowCount: intPtr(1)})
			return
		}
		http.NotFound(w, r)
	})
	defer srv.Close()

	// Observe hooks
	var sawBefore, sawAfter bool
	cl.BeforeHooks = append(cl.BeforeHooks, func(*http.Request) { sawBefore = true })
	cl.AfterHooks = append(cl.AfterHooks, func(*http.Response, []byte, error) { sawAfter = true })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	_, err := cl.ExecSQL(ctx, "x", SQLRequest{SQL: "INSERT INTO t VALUES (1)"})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&attempts) < 2 {
		t.Fatalf("expected retry, attempts=%d", attempts)
	}
	// Verify that client slept for at least the Retry-After duration (1s).
	if elapsed < 1*time.Second {
		t.Fatalf("expected retry sleep of at least 1s honoring Retry-After, got %v", elapsed)
	}
	if !sawBefore || !sawAfter {
		t.Fatalf("hooks not triggered: before=%v after=%v", sawBefore, sawAfter)
	}

	// Verify rate limit headers were captured.
	rl := cl.LastRateLimit()
	if rl == nil {
		t.Fatalf("expected LastRateLimit to be populated")
	}
	if rl.Limit != 120 {
		t.Fatalf("expected limit 120, got %d", rl.Limit)
	}
}

func TestClassOfRoute(t *testing.T) {
	// Chain routes: blobs, project creation, commit
	if c := ClassOfRoute(http.MethodPost, "/v1/blobs"); c != RateLimitClassChain {
		t.Fatalf("expected chain for POST /v1/blobs, got %s", c)
	}
	if c := ClassOfRoute(http.MethodPost, "/v1/projects"); c != RateLimitClassChain {
		t.Fatalf("expected chain for POST /v1/projects, got %s", c)
	}
	if c := ClassOfRoute(http.MethodPost, "/v1/projects/proj-123/commit"); c != RateLimitClassChain {
		t.Fatalf("expected chain for commit route, got %s", c)
	}

	// Ordinary routes: reads, SQL queries, table inspection
	if c := ClassOfRoute(http.MethodGet, "/v1/projects"); c != RateLimitClassOrdinary {
		t.Fatalf("expected ordinary for GET /v1/projects, got %s", c)
	}
	if c := ClassOfRoute(http.MethodPost, "/v1/projects/proj-123/sql"); c != RateLimitClassOrdinary {
		t.Fatalf("expected ordinary for SQL POST, got %s", c)
	}
	if c := ClassOfRoute(http.MethodGet, "/v1/projects/proj-123/tables"); c != RateLimitClassOrdinary {
		t.Fatalf("expected ordinary for tables GET, got %s", c)
	}
}

func TestParseRetryAfter(t *testing.T) {
	if d := parseRetryAfter("60"); d != 60*time.Second {
		t.Fatalf("expected 60s, got %v", d)
	}
	if d := parseRetryAfter("0"); d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
	if d := parseRetryAfter(""); d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}
