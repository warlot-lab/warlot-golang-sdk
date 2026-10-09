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
				http.Error(w, `{"error":"too many"}`, http.StatusTooManyRequests)
				return
			}
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

	_, err := cl.ExecSQL(ctx, "x", SQLRequest{SQL: "INSERT INTO t VALUES (1)"})
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&attempts) < 2 {
		t.Fatalf("expected retry, attempts=%d", attempts)
	}
	if !sawBefore || !sawAfter {
		t.Fatalf("hooks not triggered: before=%v after=%v", sawBefore, sawAfter)
	}
}
