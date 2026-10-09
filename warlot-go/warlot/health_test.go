package warlot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"
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
		"status": "terminating",
		"termination_state": "terminating"
	}`

	var st ProjectStatus
	if err := json.Unmarshal([]byte(rawJSON), &st); err != nil {
		t.Fatalf("unmarshal status failed: %v", err)
	}

	if st.ProjectID != "0xproj999" || st.DBID != "0xdb123" || st.WriterPassID != "0xpass456" {
		t.Fatalf("identifiers mismatch: %+v", st)
	}
	if st.LastUploadSeq != 150 || st.MaxSeq != 175 || st.Synced != false || !st.IsActive {
		t.Fatalf("sequence/state fields mismatch: %+v", st)
	}
	if st.Status != "terminating" || st.TerminationState != "terminating" {
		t.Fatalf("status/termination mismatch: %+v", st)
	}
	if st.LastUploadAt == nil {
		t.Fatalf("expected non-nil LastUploadAt")
	}

	// Test ReplicationLag: 175 - 150 = 25 (value receiver and pointer receiver)
	lag := st.ReplicationLag()
	if lag != 25 {
		t.Errorf("expected lag 25, got %d", lag)
	}
	if ptrLag := (&st).ReplicationLag(); ptrLag != 25 {
		t.Errorf("expected pointer receiver lag 25, got %d", ptrLag)
	}

	// Test Freshness (value receiver and pointer receiver)
	freshness := st.Freshness()
	if freshness <= 0 {
		t.Errorf("expected positive freshness duration, got %v", freshness)
	}
	if ptrFreshness := (&st).Freshness(); ptrFreshness <= 0 {
		t.Errorf("expected positive pointer receiver freshness, got %v", ptrFreshness)
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
	if f := noTimeStatus.Freshness(); f != 0 {
		t.Errorf("expected freshness 0 on nil timestamp, got %v", f)
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
	if !errors.Is(err, ErrClusterNotReady) {
		t.Errorf("expected ErrClusterNotReady, got %v", err)
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

func TestEndpoints_ProjectStatusAndCommit(t *testing.T) {
	uploadTime := time.Now().Add(-10 * time.Minute).UTC().Truncate(time.Second)
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/projects/proj-alpha-001/status":
			json.NewEncoder(w).Encode(ProjectStatus{
				ProjectID:        "proj-alpha-001",
				DBID:             "db-alpha-001",
				WriterPassID:     "pass-alpha-001",
				LastUploadSeq:    42,
				LastUploadAt:     &uploadTime,
				MaxSeq:           50,
				Synced:           false,
				IsActive:         true,
				Status:           "active",
				TerminationState: "none",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/proj-alpha-001/commit":
			json.NewEncoder(w).Encode(CommitResponse{
				CommittedOps:  8,
				TxDigest:      "0xDigest999",
				LastUploadSeq: 50,
				MaxSeq:        50,
			})
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()

	ctx := context.Background()

	// 1. Direct Client.GetProjectStatus
	status, err := cl.GetProjectStatus(ctx, "proj-alpha-001")
	if err != nil {
		t.Fatalf("GetProjectStatus failed: %v", err)
	}
	if status.ProjectID != "proj-alpha-001" || status.DBID != "db-alpha-001" {
		t.Errorf("unexpected identifiers: %+v", status)
	}
	if status.ReplicationLag() != 8 {
		t.Errorf("expected replication lag 8, got %d", status.ReplicationLag())
	}
	if status.Freshness() < 9*time.Minute {
		t.Errorf("expected freshness >= 9m, got %v", status.Freshness())
	}

	// 2. Direct Client.CommitProject
	commit, err := cl.CommitProject(ctx, "proj-alpha-001")
	if err != nil {
		t.Fatalf("CommitProject failed: %v", err)
	}
	if commit.CommittedOps != 8 || commit.TxDigest != "0xDigest999" || commit.LastUploadSeq != 50 {
		t.Errorf("unexpected commit response: %+v", commit)
	}

	// 3. Project helper methods
	p := cl.Project("proj-alpha-001")
	pStatus, err := p.Status(ctx)
	if err != nil || pStatus.ProjectID != "proj-alpha-001" {
		t.Fatalf("Project.Status failed: %+v err=%v", pStatus, err)
	}
	pCommit, err := p.Commit(ctx)
	if err != nil || pCommit.CommittedOps != 8 {
		t.Fatalf("Project.Commit failed: %+v err=%v", pCommit, err)
	}
}
