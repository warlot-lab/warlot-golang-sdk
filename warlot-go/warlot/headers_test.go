package warlot

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHeaders_Auth_Forwarded(t *testing.T) {
	var gotAuth, gotAPI, gotHolder, gotProject, gotX string

	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAPI = r.Header.Get("x-api-key")
		gotHolder = r.Header.Get("x-holder-id")
		gotProject = r.Header.Get("x-project-name")
		gotX = r.Header.Get("x-extra")
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	cl.APIKey = "test_key_wlt_123"
	cl.HolderID = "holder_abc"
	cl.ProjectName = "project_xyz"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, _ = cl.doRequest(ctx, http.MethodGet, "/w", cl.authHeaders(), nil)

	h := cl.authHeaders()
	h.Set("x-extra", "1")
	_, _ = cl.doRequest(ctx, http.MethodGet, "/w", h, nil)

	if gotAuth != "Bearer test_key_wlt_123" {
		t.Fatalf("expected Authorization header Bearer test_key_wlt_123, got: %q", gotAuth)
	}
	if gotAPI != "" {
		t.Fatalf("expected x-api-key header to be empty, got: %q", gotAPI)
	}
	if gotHolder != "" {
		t.Fatalf("expected x-holder-id header to be empty, got: %q", gotHolder)
	}
	if gotProject != "" {
		t.Fatalf("expected x-project-name header to be empty, got: %q", gotProject)
	}
	if gotX != "1" {
		t.Fatalf("expected x-extra header to be 1, got: %q", gotX)
	}
}
