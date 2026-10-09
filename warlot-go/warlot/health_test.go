package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestProjectStatus_ReplicationLagAndFreshness(t *testing.T) {
	// 1. Status JSON as emitted by warlot-publisher datahttp/project.go:187-218
	rawJSON := `{
		"project_id": "0xproj999",
		"db_id": "0xdb123",
		"writer_pass_id": "0xpass456",
		"last_upload_seq": 150,
		"last_upload_at": "2026-10-09T10:00:00Z",
		"max_seq": 175,
		"synced": false,
		"is_active": true,
		"status": "active"
	}`

	var st ProjectStatus
	if err := json.Unmarshal([]byte(rawJSON), &st); err != nil {
		t.Fatalf("unmarshal status failed: %v", err)
	}

	if st.ProjectID != "0xproj999" || st.LastUploadSeq != 150 || st.MaxSeq != 175 {
		t.Fatalf("deserialized fields mismatch: %+v", st)
	}

	// Test ReplicationLag: 175 - 150 = 25
	lag := st.ReplicationLag()
	if lag != 25 {
		t.Errorf("expected lag 25, got %d", lag)
	}

	// Test Freshness
	freshness, ok := st.Freshness()
	if !ok {
		t.Errorf("expected freshness to be parsed successfully")
	}
	if freshness <= 0 {
		t.Errorf("expected positive freshness duration, got %v", freshness)
	}

	// 2. Synced state: max_seq == last_upload_seq
	syncedStatus := ProjectStatus{
		LastUploadSeq: 200,
		MaxSeq:        200,
		Synced:        true,
	}
	if syncedStatus.ReplicationLag() != 0 {
		t.Errorf("expected synced lag 0, got %d", syncedStatus.ReplicationLag())
	}

	// 3. Edge case: max_seq < last_upload_seq (clock skew or reset)
	skewStatus := ProjectStatus{
		LastUploadSeq: 200,
		MaxSeq:        190,
	}
	if skewStatus.ReplicationLag() != 0 {
		t.Errorf("expected non-negative lag 0, got %d", skewStatus.ReplicationLag())
	}

	// 4. Edge case: nil LastUploadAt
	noTimeStatus := ProjectStatus{
		LastUploadAt: nil,
	}
	if _, ok := noTimeStatus.Freshness(); ok {
		t.Errorf("expected freshness to fail on nil timestamp")
	}
}

func TestCheckReadiness_Healthy(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"status": "ready",
			"dependencies": [
				{"name": "postgres", "status": "ok"},
				{"name": "aggregator", "status": "ok"}
			]
		}`))
	})
	defer srv.Close()

	res, err := cl.CheckReadiness(context.Background())
	if err != nil {
		t.Fatalf("unexpected error checking readiness: %v", err)
	}
	if !res.IsReady() {
		t.Fatalf("expected cluster to be ready: %+v", res)
	}
	if len(res.Dependencies) != 2 || res.Dependencies[0].Status != "ok" {
		t.Fatalf("dependencies mismatch: %+v", res.Dependencies)
	}
}

func TestCheckReadiness_Degraded(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{
			"status": "not ready",
			"dependencies": [
				{"name": "postgres", "status": "ok"},
				{"name": "aggregator", "status": "unreachable"}
			]
		}`))
	})
	defer srv.Close()

	res, err := cl.CheckReadiness(context.Background())
	if err == nil {
		t.Fatalf("expected error on 503 degraded status, got nil")
	}
	if res == nil {
		t.Fatalf("expected response object despite error, got nil")
	}
	if res.IsReady() {
		t.Fatalf("expected cluster IsReady to be false")
	}
	if len(res.Dependencies) != 2 || res.Dependencies[1].Status != "unreachable" {
		t.Fatalf("expected aggregator to be unreachable: %+v", res.Dependencies)
	}
}
