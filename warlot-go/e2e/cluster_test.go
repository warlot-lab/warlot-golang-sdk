package e2e

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot"
)

func TestCluster_LiveProbesAndProjections(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Publisher cluster readiness probe
	pubClient := warlot.New(
		warlot.WithBaseURL("http://10.43.158.223:8080"),
		warlot.WithHTTPClient(&http.Client{Timeout: 3 * time.Second}),
	)
	readiness, err := pubClient.CheckReadiness(ctx)
	if err != nil {
		t.Fatalf("publisher /readyz failed: %v", err)
	}
	if !readiness.IsReady() {
		t.Fatalf("publisher reported unready status: %+v", readiness)
	}
	t.Logf("verified publisher live readiness: status=%s, deps=%+v", readiness.Status, readiness.Dependencies)

	// 2. Indexer cluster readiness probe & health
	idxClient := warlot.NewIndexerClient("http://10.43.102.80:8081", &http.Client{Timeout: 3 * time.Second})
	if err := idxClient.CheckReadiness(ctx); err != nil {
		t.Fatalf("indexer /readyz failed: %v", err)
	}

	health, err := idxClient.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("indexer /v1/health failed: %v", err)
	}
	if !health.Data.Ready {
		t.Fatalf("indexer reported unready status: %+v", health)
	}
	t.Logf("verified indexer live health: checkpoint=%d, tx_digest=%s", health.Freshness.Checkpoint, health.Freshness.TxDigest)

	// 3. Indexer read projection query
	storage, err := idxClient.GetAddressStorage(ctx, "0x0")
	if err != nil {
		t.Fatalf("indexer /v1/addresses/0x0/storage failed: %v", err)
	}
	t.Logf("verified indexer storage projection: bytes=%d, files=%d", storage.Data.StorageBytes, storage.Data.FileCount)
}

func TestCluster_EndToEndPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pubBase := os.Getenv("WARLOT_BASE_URL")
	if pubBase == "" {
		pubBase = "http://10.43.158.223:8080"
	}
	idxBase := os.Getenv("WARLOT_INDEXER_URL")
	if idxBase == "" {
		idxBase = "http://10.43.102.80:8081"
	}
	apiKey := os.Getenv("WARLOT_API_KEY")
	if apiKey == "" {
		apiKey = "wlt.1.ASB3v9D0_v8Gkwa82vuU1TMAAAAAaskp1wAAAAAAAAAAQjB4ZmRjMDRlNjdlYmQ1OGJjZDQ1YTAzNjNmYzBhZDlmNmMzYTFmNDVmZmY4NTViODVkYTQzNmMwOTdmZWQ1ODExMg.-WrY8elW6VZE0HA3U9m_mUEhqk_r30InDIu_4OzDnIs"
	}
	holder := os.Getenv("WARLOT_HOLDER")
	if holder == "" {
		holder = "0xd6f0ace363a7d44c13da1215bab31c733f890df36fd8083b31633bba3166c6bd"
	}
	owner := os.Getenv("WARLOT_OWNER")
	if owner == "" {
		owner = "0xfdc04e67ebd58bcd45a0363fc0ad9f6c3a1f45fff855b85da436c097fed58112"
	}
	pname := os.Getenv("WARLOT_PNAME")
	if pname == "" {
		pname = "warlot-e2e"
	}

	cl := warlot.New(
		warlot.WithBaseURL(pubBase),
		warlot.WithAPIKey(apiKey),
		warlot.WithHolderID(holder),
		warlot.WithProjectName(pname),
		warlot.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
	)

	idx := warlot.NewIndexerClient(idxBase, &http.Client{Timeout: 10 * time.Second})

	// 1. Indexer readiness & health queries (returning checkpoint freshness metadata)
	if err := idx.CheckReadiness(ctx); err != nil {
		t.Fatalf("indexer readiness probe failed: %v", err)
	}
	idxHealth, err := idx.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("indexer health check failed: %v", err)
	}
	if !idxHealth.Data.Ready || !idxHealth.Data.Live {
		t.Fatalf("indexer not live or ready: %+v", idxHealth)
	}
	if idxHealth.Freshness.Checkpoint == 0 {
		t.Fatalf("indexer reported zero checkpoint: %+v", idxHealth.Freshness)
	}
	t.Logf("verified indexer health & checkpoint: checkpoint=%d, tx=%s, seq=%d, updated_at=%v",
		idxHealth.Freshness.Checkpoint, idxHealth.Freshness.TxDigest, idxHealth.Freshness.EventSeq, idxHealth.Freshness.UpdatedAt)

	// 2. Indexer typed projection queries
	balances, err := idx.GetAddressBalances(ctx, owner)
	if err != nil {
		t.Fatalf("indexer GetAddressBalances failed: %v", err)
	}
	t.Logf("verified address balances: count=%d, checkpoint=%d", len(balances.Data), balances.Freshness.Checkpoint)

	history, err := idx.GetAddressHistory(ctx, owner, "", 10)
	if err != nil {
		t.Fatalf("indexer GetAddressHistory failed: %v", err)
	}
	t.Logf("verified address history: count=%d, checkpoint=%d", len(history.Data), history.Freshness.Checkpoint)

	adminHolders, err := idx.GetHoldersByAdmin(ctx, owner)
	if err != nil {
		t.Fatalf("indexer GetHoldersByAdmin failed: %v", err)
	}
	if len(adminHolders.Data) == 0 {
		t.Fatalf("expected holders for admin %s, got 0", owner)
	}
	t.Logf("verified admin holders: count=%d, first=%s", len(adminHolders.Data), adminHolders.Data[0].HolderID)

	storage, err := idx.GetAddressStorage(ctx, owner)
	if err != nil {
		t.Fatalf("indexer GetAddressStorage failed: %v", err)
	}
	t.Logf("verified address storage: bytes=%d, files=%d", storage.Data.StorageBytes, storage.Data.FileCount)

	lifespans, err := idx.GetAddressLifespan(ctx, owner)
	if err != nil {
		t.Fatalf("indexer GetAddressLifespan failed: %v", err)
	}
	t.Logf("verified address lifespan: count=%d", len(lifespans.Data))

	holderProj, err := idx.GetHolderProjects(ctx, holder)
	if err != nil {
		t.Fatalf("indexer GetHolderProjects failed: %v", err)
	}
	if holderProj.Data.ActiveProjects < 1 {
		t.Fatalf("expected >= 1 active projects for holder %s, got %d", holder, holderProj.Data.ActiveProjects)
	}
	t.Logf("verified holder projects: holder=%s, active=%d", holder, holderProj.Data.ActiveProjects)

	// 3. Publisher: Resolve project
	resolveRes, err := cl.ResolveProject(ctx, warlot.ResolveProjectRequest{
		HolderID:    holder,
		ProjectName: pname,
	})
	if err != nil {
		t.Fatalf("ResolveProject failed: %v", err)
	}
	if !resolveRes.ExistsMeta || !resolveRes.ExistsChain || resolveRes.ProjectID == "" {
		t.Fatalf("ResolveProject unexpected outcome: %+v", resolveRes)
	}
	t.Logf("verified project resolve: project_id=%s, db_id=%s, action=%s",
		resolveRes.ProjectID, resolveRes.DBID, resolveRes.Action)
	projectID := resolveRes.ProjectID

	// 4. Publisher: Initialization (ensure initialized)
	initRes, err := cl.InitProject(ctx, warlot.InitProjectRequest{
		HolderID:     holder,
		ProjectName:  pname,
		OwnerAddress: owner,
	})
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}
	if initRes.ProjectID == "" {
		t.Fatalf("InitProject returned empty ProjectID: %+v", initRes)
	}
	t.Logf("verified project init: project_id=%s, db_id=%s", initRes.ProjectID, initRes.DBID)

	proj := cl.Project(projectID)

	// 5. Publisher: Table listing
	tablesRes, err := proj.Tables(ctx)
	if err != nil {
		t.Fatalf("Tables failed: %v", err)
	}
	if len(tablesRes.Tables) == 0 {
		t.Fatalf("expected at least 1 table, got 0")
	}
	t.Logf("verified table listing: count=%d, tables=%v", len(tablesRes.Tables), tablesRes.Tables)

	// 6. Publisher: SQL writes (DDL and DML)
	createSQL := `CREATE TABLE IF NOT EXISTS e2e_cluster_audit (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		audit_note TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`
	if _, err := proj.SQL(ctx, createSQL, nil); err != nil {
		t.Fatalf("SQL CREATE TABLE failed: %v", err)
	}

	insertSQL := `INSERT INTO e2e_cluster_audit (audit_note, created_at) VALUES (?, ?)`
	nowStr := time.Now().UTC().Format(time.RFC3339)
	insertRes, err := proj.SQL(ctx, insertSQL, []any{"cluster-e2e-run", nowStr})
	if err != nil {
		t.Fatalf("SQL INSERT failed: %v", err)
	}
	if !insertRes.OK {
		t.Fatalf("SQL INSERT reported not OK: %+v", insertRes)
	}
	t.Logf("verified SQL writes: row_count=%v", insertRes.RowCount)

	// 7. Publisher: Table browsing
	browseRes, err := proj.Browse(ctx, "e2e_cluster_audit", 5, 0)
	if err != nil {
		t.Fatalf("Browse failed: %v", err)
	}
	if len(browseRes.Rows) == 0 {
		t.Fatalf("Browse returned 0 rows for e2e_cluster_audit")
	}
	t.Logf("verified table browsing: table=%s, rows_returned=%d, first_row=%v",
		browseRes.Table, len(browseRes.Rows), browseRes.Rows[len(browseRes.Rows)-1])

	// 8. Publisher: Status SLIs
	status, err := proj.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if !status.IsActive || status.Status != "active" {
		t.Fatalf("unexpected project status: %+v", status)
	}
	lag := status.ReplicationLag()
	freshness := status.Freshness()
	t.Logf("verified status SLIs: active=%v, status=%s, max_seq=%d, last_upload_seq=%d, lag=%d, freshness=%v",
		status.IsActive, status.Status, status.MaxSeq, status.LastUploadSeq, lag, freshness)

	// 9. Publisher: Commit
	commitRes, err := proj.Commit(ctx)
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}
	t.Logf("verified commit: committed_ops=%d, last_upload_seq=%d, max_seq=%d, tx_digest=%s",
		commitRes.CommittedOps, commitRes.LastUploadSeq, commitRes.MaxSeq, commitRes.TxDigest)
}
