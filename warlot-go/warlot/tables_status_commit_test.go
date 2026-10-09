package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestTables_Status_Commit_Pager(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/tables"):
			json.NewEncoder(w).Encode(ListTablesResponse{Tables: []string{"products"}})
		case strings.Contains(r.URL.Path, "/tables/products/rows"):
			q := r.URL.Query()
			limit := q.Get("limit")
			offset := q.Get("offset")
			_ = limit
			_ = offset
			json.NewEncoder(w).Encode(BrowseRowsResponse{
				Limit:  2,
				Offset: 0,
				Table:  "products",
				Rows: []map[string]any{
					{"id": 1}, {"id": 2},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/schema"):
			json.NewEncoder(w).Encode(TableSchema{
				Table: "products",
				Columns: []TableColumn{
					{CID: 0, Name: "id", Type: "INTEGER", PrimaryPK: true},
					{CID: 1, Name: "name", Type: "TEXT", NotNull: true},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/count"):
			json.NewEncoder(w).Encode(TableCountResponse{ProjectID: "proj-123", TableCount: 1})
		case strings.HasSuffix(r.URL.Path, "/status"):
			json.NewEncoder(w).Encode(ProjectStatus{
				ProjectID:     "proj-123",
				IsActive:      true,
				Status:        "active",
				LastUploadSeq: 10,
				MaxSeq:        10,
				Synced:        true,
			})
		case strings.HasSuffix(r.URL.Path, "/commit"):
			json.NewEncoder(w).Encode(CommitResponse{
				CommittedOps:  5,
				TxDigest:      "0xABC",
				LastUploadSeq: 10,
				MaxSeq:        10,
			})
		case r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/tables/products"):
			json.NewEncoder(w).Encode(PatchTableResponse{OK: true, RowCount: 1})
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	proj := cl.Project("proj-123")

	// list
	lt, err := proj.Tables(ctx)
	if err != nil || len(lt.Tables) != 1 {
		t.Fatalf("tables: %+v err=%v", lt, err)
	}

	// browse + pager
	pgr := &Pager{Project: proj, Table: "products", Limit: 2}
	rows, err := pgr.Next(ctx)
	if err != nil || len(rows) != 2 {
		t.Fatalf("pager first: %v %v", len(rows), err)
	}
	// emulate end
	pgr.Done = true
	rows, err = pgr.Next(ctx)
	if err != nil || rows != nil {
		t.Fatalf("pager end: rows=%v err=%v", rows, err)
	}

	// schema
	sc, err := proj.Schema(ctx, "products")
	if err != nil || sc.Table != "products" || len(sc.Columns) != 2 || sc.Columns[0].Name != "id" {
		t.Fatalf("schema: %+v err=%v", sc, err)
	}

	// count
	cnt, err := proj.Count(ctx)
	if err != nil || cnt.TableCount != 1 {
		t.Fatalf("count: %+v err=%v", cnt, err)
	}

	// patch table
	patchRes, err := proj.Patch(ctx, "products", []TableEdit{
		{
			Set:   map[string]any{"name": "Updated"},
			Where: map[string]any{"id": 1},
		},
	})
	if err != nil || !patchRes.OK || patchRes.RowCount != 1 {
		t.Fatalf("patch: %+v err=%v", patchRes, err)
	}

	// status
	st, err := proj.Status(ctx)
	if err != nil || !st.IsActive || !st.Synced || st.Status != "active" {
		t.Fatalf("status: %+v err=%v", st, err)
	}

	// commit
	cm, err := proj.Commit(ctx)
	if err != nil || cm.CommittedOps != 5 || cm.TxDigest != "0xABC" {
		t.Fatalf("commit: %+v err=%v", cm, err)
	}
}
