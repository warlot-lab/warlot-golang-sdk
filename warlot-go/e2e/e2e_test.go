package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot"
)

func TestE2E_Live(t *testing.T) {
	if os.Getenv("WARLOT_E2E") != "1" {
		t.Skip("set WARLOT_E2E=1 to run live test")
	}

	holder := mustEnv(t, "WARLOT_HOLDER")
	owner := mustEnv(t, "WARLOT_OWNER")
	pname := mustEnv(t, "WARLOT_PNAME")
	base := os.Getenv("WARLOT_BASE_URL") // optional override
	apiKey := os.Getenv("WARLOT_API_KEY")

	opts := []warlot.Option{
		warlot.WithHolderID(holder),
		warlot.WithProjectName(pname),
		warlot.WithRetries(6),
		warlot.WithBackoff(2*time.Second, 16*time.Second),
		warlot.WithHTTPClient(&http.Client{Timeout: 120 * time.Second}),
		warlot.WithLogger(func(event string, meta map[string]any) { t.Logf("%s: %v", event, meta) }),
	}
	if base != "" {
		opts = append(opts, warlot.WithBaseURL(base))
	}
	if apiKey != "" {
		opts = append(opts, warlot.WithAPIKey(apiKey))
	}
	cl := warlot.New(opts...)
	ctx := context.Background()

	// warm-up (best-effort)
	{
		wctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		_, _ = cl.ResolveProject(wctx, warlot.ResolveProjectRequest{HolderID: holder, ProjectName: pname})
		cancel()
	}

	// resolve or init
	var projectID string
	{
		sctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		if res, err := cl.ResolveProject(sctx, warlot.ResolveProjectRequest{HolderID: holder, ProjectName: pname}); err == nil && res.ProjectID != "" {
			projectID = res.ProjectID
		} else {
			initRes, err := cl.InitProject(sctx, warlot.InitProjectRequest{
				HolderID: holder, ProjectName: pname, OwnerAddress: owner,
				IncludePass: true, Deletable: true,
			})
			if err != nil {
				t.Fatalf("InitProject failed: %v", err)
			}
			projectID = initRes.ProjectID
		}
	}

	// issue key if not already configured
	if apiKey == "" {
		sctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		iss, err := cl.IssueAPIKey(sctx, warlot.IssueKeyRequest{
			ProjectID: projectID, ProjectHolder: holder, ProjectName: pname, User: owner,
		})
		if err != nil {
			t.Fatalf("IssueAPIKey failed: %v", err)
		}
		apiKey = iss.APIKey
		cl.APIKey = apiKey
	}
	proj := cl.Project(projectID)

	// create table
	{
		sctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		if _, err := proj.SQL(sctx, `CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL)`, nil); err != nil {
			t.Fatalf("CREATE TABLE failed: %v", err)
		}
	}

	// insert
	{
		sctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		if _, err := proj.SQL(sctx, `INSERT INTO products (name,price) VALUES (?,?)`, []any{"X", 1.23}, warlot.WithIdempotencyKey("e2e-insert-1")); err != nil {
			t.Fatalf("INSERT failed: %v", err)
		}
	}

	// select typed
	type P struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	{
		sctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		ps, err := warlot.Query[P](sctx, proj, `SELECT * FROM products ORDER BY id DESC LIMIT 1`, nil)
		if err != nil {
			t.Fatalf("SELECT failed: %v", err)
		}
		if len(ps) == 0 {
			t.Fatalf("SELECT returned no rows")
		}
		t.Logf("row: %+v", ps[0])
	}

	// count
	{
		sctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		cnt, err := proj.Count(sctx)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}
		if cnt.TableCount < 1 {
			t.Fatalf("unexpected count: %+v", cnt)
		}
	}
}

func mustEnv(t *testing.T, k string) string {
	t.Helper()
	v := os.Getenv(k)
	if v == "" {
		t.Fatalf("missing env %s", k)
	}
	return v
}
