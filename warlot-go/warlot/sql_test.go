package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSQL_DDL_DML_Select_QueryTyped_Idempotency(t *testing.T) {
	var sawIdem string

	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sql") {
			// capture idempotency
			sawIdem = r.Header.Get("x-idempotency-key")

			var req SQLRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			defer r.Body.Close()

			sqlLower := strings.ToLower(strings.TrimSpace(req.SQL))
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasPrefix(sqlLower, "create table"):
				json.NewEncoder(w).Encode(SQLResponse{OK: true, RowCount: intPtr(0)})
			case strings.HasPrefix(sqlLower, "insert"):
				json.NewEncoder(w).Encode(SQLResponse{OK: true, RowCount: intPtr(1)})
			case strings.HasPrefix(sqlLower, "select"):
				json.NewEncoder(w).Encode(SQLResponse{
					OK: true,
					Rows: []map[string]any{
						{"id": 1, "name": "A", "price": 9.99, "category": "Electronics"},
						{"id": 2, "name": "B", "price": 19.99, "category": "Books"},
					},
				})
			default:
				json.NewEncoder(w).Encode(SQLResponse{OK: true})
			}
			return
		}
		http.NotFound(w, r)
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proj := cl.Project("proj-123")

	// DDL
	_, err := proj.SQL(ctx, `CREATE TABLE IF NOT EXISTS products (id INTEGER)`, nil)
	if err != nil {
		t.Fatal(err)
	}

	// DML with idempotency
	_, err = proj.SQL(ctx, `INSERT INTO products (id) VALUES (?)`, []any{1}, WithIdempotencyKey("insert-1"))
	if err != nil {
		t.Fatal(err)
	}
	if sawIdem != "insert-1" {
		t.Fatalf("idempotency header not forwarded: %q", sawIdem)
	}

	// SELECT raw
	sel, err := proj.SQL(ctx, `SELECT * FROM products`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sel.Rows) != 2 {
		t.Fatalf("rows=%d", len(sel.Rows))
	}

	// Typed Query
	type Product struct {
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		Price    float64 `json:"price"`
		Category string  `json:"category"`
	}
	ps, err := Query[Product](ctx, proj, `SELECT * FROM products`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ps[0].Name != "A" || ps[1].Price != 19.99 {
		t.Fatalf("mapped %+v", ps)
	}
}

func intPtr(i int) *int { return &i }

func TestSQL_Select_NormalizationWithoutOK(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Simulate warlot-publisher datahttp/mutations.go SELECT response:
		// c.JSON(http.StatusOK, gin.H{"rows": result.Rows, "row_count": result.RowCount})
		// omitting "ok" field entirely.
		_, _ = w.Write([]byte(`{"rows":[{"id":1,"name":"Widget"}],"row_count":1}`))
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proj := cl.Project("proj-123")
	res, err := proj.SQL(ctx, "SELECT id, name FROM items", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.OK {
		t.Fatalf("expected res.OK to be normalized to true, got false")
	}
	if res.RowCount == nil || *res.RowCount != 1 {
		t.Fatalf("expected row_count to be 1, got %v", res.RowCount)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(res.Rows))
	}
}
