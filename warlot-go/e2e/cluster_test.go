package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
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
