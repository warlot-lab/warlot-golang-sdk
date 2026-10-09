package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestProjects_Init_Issue_Resolve_Lifecycle(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects":
			json.NewEncoder(w).Encode(InitProjectResponse{
				ProjectID: "proj-123",
				DBID:      "0xDB",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/auth/issue":
			json.NewEncoder(w).Encode(IssueKeyResponse{APIKey: "key-abc", URL: "https://api.example.com/proj-123"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/resolve":
			json.NewEncoder(w).Encode(ResolveProjectResponse{ProjectID: "proj-123", DBID: "0xDB"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/proj-123/deactivate":
			json.NewEncoder(w).Encode(DeactivateProjectResponse{
				ProjectID: "proj-123",
				Status:    "deactivated",
				IsActive:  false,
			})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/projects/proj-123/reactivate":
			json.NewEncoder(w).Encode(ReactivateProjectResponse{
				ProjectID: "proj-123",
				Status:    "active",
				IsActive:  true,
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/projects/proj-123":
			json.NewEncoder(w).Encode(TerminateProjectResponse{
				ProjectID:        "proj-123",
				Status:           "terminated",
				TerminationState: "terminated",
				IsActive:         false,
			})
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Init
	initRes, err := cl.InitProject(ctx, InitProjectRequest{
		HolderID: "0xH", ProjectName: "p", OwnerAddress: "0xU",
	})
	if err != nil {
		t.Fatal(err)
	}
	if initRes.ProjectID != "proj-123" {
		t.Fatalf("got %s", initRes.ProjectID)
	}

	// Issue
	iss, err := cl.IssueAPIKey(ctx, IssueKeyRequest{
		ProjectID: "proj-123", ProjectHolder: "0xH", ProjectName: "p", User: "0xU",
	})
	if err != nil {
		t.Fatal(err)
	}
	if iss.APIKey == "" {
		t.Fatalf("no apikey")
	}

	// Resolve
	res, err := cl.ResolveProject(ctx, ResolveProjectRequest{HolderID: "0xH", ProjectName: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProjectID != "proj-123" {
		t.Fatalf("resolve mismatch: %s", res.ProjectID)
	}

	// Deactivate
	deact, err := cl.DeactivateProject(ctx, "proj-123")
	if err != nil {
		t.Fatal(err)
	}
	if deact.Status != "deactivated" || deact.IsActive {
		t.Fatalf("unexpected deactivation: %+v", deact)
	}

	// Reactivate
	react, err := cl.ReactivateProject(ctx, "proj-123")
	if err != nil {
		t.Fatal(err)
	}
	if react.Status != "active" || !react.IsActive {
		t.Fatalf("unexpected reactivation: %+v", react)
	}

	// Terminate
	term, err := cl.TerminateProject(ctx, "proj-123")
	if err != nil {
		t.Fatal(err)
	}
	if term.Status != "terminated" || term.TerminationState != "terminated" {
		t.Fatalf("unexpected termination: %+v", term)
	}
}
