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
		case "/v1/addresses/0xUser/lifespan":
			if r.URL.Query().Get("group") == "band" {
				upper := 30.0
				json.NewEncoder(w).Encode(IndexerEnvelope[[]LifespanBand]{
					Data: []LifespanBand{
						{UpperDays: &upper, BlobCount: 5, TotalBytes: 1024},
					},
					Freshness: IndexerFreshness{
						Checkpoint: 12345,
					},
				})
			} else {
				json.NewEncoder(w).Encode(IndexerEnvelope[[]BlobLifespan]{
					Data: []BlobLifespan{
						{
							BlobObjID:                 "0xBlob1",
							ConfigID:                  "0xConfig1",
							SizeBytes:                 2048,
							ConservativeEndEpoch:      10,
							ConservativeDaysRemaining: 15.5,
							StoredAt:                  time.Now().UTC(),
							EpochSet:                  2,
						},
					},
					Freshness: IndexerFreshness{
						Checkpoint: 12345,
					},
				})
			}
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

	// Address lifespan
	ls, err := idx.GetAddressLifespan(ctx, "0xUser")
	if err != nil || len(ls.Data) != 1 || ls.Data[0].BlobObjID != "0xBlob1" {
		t.Fatalf("GetAddressLifespan mismatch: %+v err=%v", ls, err)
	}

	// Address lifespan bands
	bands, err := idx.GetAddressLifespanBands(ctx, "0xUser")
	if err != nil || len(bands.Data) != 1 || bands.Data[0].BlobCount != 5 {
		t.Fatalf("GetAddressLifespanBands mismatch: %+v err=%v", bands, err)
	}

	// Health and Readiness aliases
	hAlias, err := idx.GetHealth(ctx)
	if err != nil || !hAlias.Data.Ready {
		t.Fatalf("GetHealth mismatch: %+v err=%v", hAlias, err)
	}
	if err := idx.GetReadiness(ctx); err != nil {
		t.Fatalf("GetReadiness failed: %v", err)
	}
}

func TestIndexer_LiveClusterReachability(t *testing.T) {
	// Attempt check against local cluster service
	idx := NewIndexerClient("http://10.43.102.80:8081", &http.Client{Timeout: 3 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := idx.CheckReadiness(ctx); err != nil {
		t.Skipf("skipping live indexer check: %v", err)
	}

	// 1. Health check with checkpoint freshness metadata
	h, err := idx.CheckHealth(ctx)
	if err != nil {
		t.Fatalf("live indexer CheckHealth failed: %v", err)
	}
	if !h.Data.Ready || !h.Data.Live {
		t.Fatalf("live indexer reports not ready: %+v", h)
	}
	t.Logf("live indexer health: checkpoint=%d, tx_digest=%s, event_seq=%d, updated_at=%v",
		h.Freshness.Checkpoint, h.Freshness.TxDigest, h.Freshness.EventSeq, h.Freshness.UpdatedAt)

	// 2. Storage projection
	storage, err := idx.GetAddressStorage(ctx, "0x0")
	if err != nil {
		t.Fatalf("live indexer GetAddressStorage failed: %v", err)
	}
	t.Logf("live indexer storage: bytes=%d, files=%d", storage.Data.StorageBytes, storage.Data.FileCount)

	// 3. Balances projection
	balances, err := idx.GetAddressBalances(ctx, "0x0")
	if err != nil {
		t.Fatalf("live indexer GetAddressBalances failed: %v", err)
	}
	t.Logf("live indexer balances: count=%d", len(balances.Data))

	// 4. History projection
	history, err := idx.GetAddressHistory(ctx, "0x0", "", 5)
	if err != nil {
		t.Fatalf("live indexer GetAddressHistory failed: %v", err)
	}
	t.Logf("live indexer history: count=%d, has_more=%v", len(history.Data), history.Pagination != nil && history.Pagination.HasMore)

	// 5. Holders projection
	holders, err := idx.GetHoldersByAdmin(ctx, "0x0")
	if err != nil {
		t.Fatalf("live indexer GetHoldersByAdmin failed: %v", err)
	}
	t.Logf("live indexer holders: count=%d", len(holders.Data))

	// 6. Lifespan projection
	lifespan, err := idx.GetAddressLifespan(ctx, "0x0")
	if err != nil {
		t.Fatalf("live indexer GetAddressLifespan failed: %v", err)
	}
	t.Logf("live indexer lifespan: count=%d", len(lifespan.Data))

	// 7. Lifespan bands projection
	bands, err := idx.GetAddressLifespanBands(ctx, "0x0")
	if err != nil {
		t.Fatalf("live indexer GetAddressLifespanBands failed: %v", err)
	}
	t.Logf("live indexer lifespan bands: count=%d", len(bands.Data))
}
