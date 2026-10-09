package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestIndexer_MockEndpoints(t *testing.T) {
	srv, _ := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/readyz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/v1/health":
			json.NewEncoder(w).Encode(IndexerEnvelope[IndexerHealth]{
				Data: IndexerHealth{Status: "ok", Live: true, Ready: true},
				Freshness: IndexerFreshness{
					Checkpoint: 12345,
					TxDigest:   "tx-abc",
					UpdatedAt:  time.Now().UTC(),
				},
			})
		case "/v1/addresses/0xUser/storage":
			json.NewEncoder(w).Encode(IndexerEnvelope[UserStorage]{
				Data: UserStorage{StorageBytes: 1048576, FileCount: 12},
				Freshness: IndexerFreshness{
					Checkpoint: 12345,
					TxDigest:   "tx-abc",
				},
			})
		case "/v1/addresses/0xUser/balances":
			json.NewEncoder(w).Encode(IndexerEnvelope[[]WalletBalance]{
				Data: []WalletBalance{
					{Addr: "0xUser", CoinType: "0x2::sui::SUI", Balance: 5000000000},
				},
				Freshness: IndexerFreshness{
					Checkpoint: 12345,
				},
			})
		case "/v1/addresses/0xUser/history":
			json.NewEncoder(w).Encode(IndexerEnvelope[[]HistoryItem]{
				Data: []HistoryItem{
					{TxDigest: "0xTx1", Kind: "upload", Amount: 100},
				},
				Pagination: &IndexerPagination{HasMore: false},
				Freshness: IndexerFreshness{
					Checkpoint: 12345,
				},
			})
		case "/v1/holders/0xHolder/projects":
			json.NewEncoder(w).Encode(IndexerEnvelope[ActiveProjectsRecord]{
				Data: ActiveProjectsRecord{HolderID: "0xHolder", ActiveProjects: 3},
				Freshness: IndexerFreshness{
					Checkpoint: 12345,
				},
			})
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()

	idx := NewIndexerClient(srv.URL, srv.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check readiness
	if err := idx.CheckReadiness(ctx); err != nil {
		t.Fatalf("CheckReadiness failed: %v", err)
	}

	// Check health
	h, err := idx.CheckHealth(ctx)
	if err != nil || !h.Data.Ready || h.Freshness.Checkpoint != 12345 {
		t.Fatalf("CheckHealth mismatch: %+v err=%v", h, err)
	}

	// Address storage
	st, err := idx.GetAddressStorage(ctx, "0xUser")
	if err != nil || st.Data.StorageBytes != 1048576 || st.Data.FileCount != 12 {
		t.Fatalf("GetAddressStorage mismatch: %+v err=%v", st, err)
	}

	// Balances
	bals, err := idx.GetAddressBalances(ctx, "0xUser")
	if err != nil || len(bals.Data) != 1 || bals.Data[0].Balance != 5000000000 {
		t.Fatalf("GetAddressBalances mismatch: %+v err=%v", bals, err)
	}

	// History
	hist, err := idx.GetAddressHistory(ctx, "0xUser", "", 10)
	if err != nil || len(hist.Data) != 1 || hist.Data[0].TxDigest != "0xTx1" {
		t.Fatalf("GetAddressHistory mismatch: %+v err=%v", hist, err)
	}

	// Holder projects
	hp, err := idx.GetHolderProjects(ctx, "0xHolder")
	if err != nil || hp.Data.ActiveProjects != 3 {
		t.Fatalf("GetHolderProjects mismatch: %+v err=%v", hp, err)
	}
}

func TestIndexer_LiveClusterReachability(t *testing.T) {
	// Attempt check against local cluster service
	idx := NewIndexerClient("http://10.43.102.80:8081", &http.Client{Timeout: 3 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := idx.CheckReadiness(ctx); err != nil {
		t.Skipf("skipping live indexer check: %v", err)
	}

	h, err := idx.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("live indexer CheckHealth failed: %v", err)
	}
	if !h.Data.Ready {
		t.Fatalf("live indexer reports not ready: %+v", h)
	}

	storage, err := idx.GetAddressStorage(ctx, "0x0")
	if err != nil {
		t.Fatalf("live indexer GetAddressStorage failed: %v", err)
	}
	t.Logf("live indexer verified: checkpoint=%d, updated_at=%v, storage_bytes=%d",
		storage.Freshness.Checkpoint, storage.Freshness.UpdatedAt, storage.Data.StorageBytes)
}
