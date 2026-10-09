package warlot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// IndexerFreshness represents projection sync checkpoint and timestamp.
type IndexerFreshness struct {
	Checkpoint int64     `json:"checkpoint"`
	TxDigest   string    `json:"tx_digest"`
	EventSeq   int64     `json:"event_seq"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// IndexerPagination carries cursor metadata for paginated list endpoints.
type IndexerPagination struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// IndexerEnvelope wraps all successful API responses with data, optional pagination, and freshness metadata.
type IndexerEnvelope[T any] struct {
	Data       T                  `json:"data"`
	Pagination *IndexerPagination `json:"pagination,omitempty"`
	Freshness  IndexerFreshness   `json:"freshness"`
}

// UserStorage describes active storage bytes and uploaded file counts for an address.
type UserStorage struct {
	StorageBytes int64 `json:"storage_bytes"`
	FileCount    int64 `json:"file_count"`
}

// WalletBalance describes the observed contract wallet balance.
type WalletBalance struct {
	Addr       string    `json:"addr"`
	CoinType   string    `json:"coin_type"`
	Balance    int64     `json:"balance"`
	ObservedTx string    `json:"observed_tx"`
	ObservedAt time.Time `json:"observed_at"`
}

// HistoryItem describes an entry in an address's on-chain activity feed.
type HistoryItem struct {
	TxDigest  string    `json:"tx_digest"`
	EventSeq  int64     `json:"event_seq"`
	Kind      string    `json:"kind"`
	CoinType  string    `json:"coin_type"`
	Amount    int64     `json:"amount"`
	Direction string    `json:"direction"`
	At        time.Time `json:"at"`
	ConfigID  *string   `json:"config_id,omitempty"`
	BlobObjID *string   `json:"blob_obj_id,omitempty"`
	ProjectID *string   `json:"project_id,omitempty"`
}

// HolderRecord describes a project holder object and administering address.
type HolderRecord struct {
	HolderID  string    `json:"holder_id"`
	AdminAddr string    `json:"admin_addr"`
	CreatedTx string    `json:"created_tx"`
	CreatedAt time.Time `json:"created_at"`
}

// ActiveProjectsRecord describes active project counts for a holder.
type ActiveProjectsRecord struct {
	HolderID       string `json:"holder_id"`
	ActiveProjects int64  `json:"active_projects"`
}

// BlobLifespan serializes conservative blob expiry information.
type BlobLifespan struct {
	BlobObjID                 string    `json:"blob_obj_id"`
	ConfigID                  string    `json:"config_id"`
	SizeBytes                 int64     `json:"size_bytes"`
	ConservativeEndEpoch      int64     `json:"conservative_end_epoch"`
	ConservativeDaysRemaining float64   `json:"conservative_days_remaining"`
	StoredAt                  time.Time `json:"stored_at"`
	EpochSet                  int       `json:"epoch_set"`
}

// LifespanBand serializes the totals of one lifespan band of an address's live blobs.
type LifespanBand struct {
	UpperDays  *float64 `json:"upper_days"`
	BlobCount  int64    `json:"blob_count"`
	TotalBytes int64    `json:"total_bytes"`
}

// IndexerHealth describes the indexer service liveness and readiness state.
type IndexerHealth struct {
	Status string `json:"status"`
	Live   bool   `json:"live"`
	Ready  bool   `json:"ready"`
}

// IndexerClient interacts with the Warlot Indexer read projections.
type IndexerClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewIndexerClient initializes an IndexerClient with default cluster configuration.
func NewIndexerClient(baseURL string, httpClient *http.Client) *IndexerClient {
	if baseURL == "" {
		baseURL = "https://indexer.warlot.stevenhert.xyz"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &IndexerClient{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
	}
}

// Indexer returns an IndexerClient bound to this Client's HTTP client or indexer URL.
func (c *Client) Indexer(indexerBaseURL string) *IndexerClient {
	if indexerBaseURL == "" {
		indexerBaseURL = "https://indexer.warlot.stevenhert.xyz"
	}
	var httpClient *http.Client
	if c != nil {
		httpClient = c.HTTPClient
	}
	return NewIndexerClient(indexerBaseURL, httpClient)
}

func (idx *IndexerClient) get(ctx context.Context, path string, out any) error {
	if idx == nil {
		return ErrNilClient
	}
	httpClient := idx.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL := idx.BaseURL
	if baseURL == "" {
		baseURL = "https://indexer.warlot.stevenhert.xyz"
	}
	u := baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("create indexer request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("indexer get %s: %w", u, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read indexer response: %w", err)
	}

	if res.StatusCode/100 != 2 {
		return parseAPIError(res.StatusCode, body)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode indexer response: %w (body=%s)", err, string(body))
	}
	return nil
}

// GetAddressStorage reports active storage bytes and file counts for an address.
func (idx *IndexerClient) GetAddressStorage(ctx context.Context, address string) (*IndexerEnvelope[UserStorage], error) {
	path := fmt.Sprintf("/v1/addresses/%s/storage", url.PathEscape(address))
	var env IndexerEnvelope[UserStorage]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetAddressBalances reports wallet balances for an address across coin types.
func (idx *IndexerClient) GetAddressBalances(ctx context.Context, address string) (*IndexerEnvelope[[]WalletBalance], error) {
	path := fmt.Sprintf("/v1/addresses/%s/balances", url.PathEscape(address))
	var env IndexerEnvelope[[]WalletBalance]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetAddressHistory retrieves the activity history feed for an address.
func (idx *IndexerClient) GetAddressHistory(ctx context.Context, address string, cursor string, limit int) (*IndexerEnvelope[[]HistoryItem], error) {
	q := url.Values{}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := fmt.Sprintf("/v1/addresses/%s/history", url.PathEscape(address))
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var env IndexerEnvelope[[]HistoryItem]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetHoldersByAdmin lists project holders administered by an address.
func (idx *IndexerClient) GetHoldersByAdmin(ctx context.Context, address string) (*IndexerEnvelope[[]HolderRecord], error) {
	path := fmt.Sprintf("/v1/addresses/%s/holders", url.PathEscape(address))
	var env IndexerEnvelope[[]HolderRecord]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetHolderProjects reports active project counts under a holder.
func (idx *IndexerClient) GetHolderProjects(ctx context.Context, holderID string) (*IndexerEnvelope[ActiveProjectsRecord], error) {
	path := fmt.Sprintf("/v1/holders/%s/projects", url.PathEscape(holderID))
	var env IndexerEnvelope[ActiveProjectsRecord]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetAddressLifespan retrieves conservative blob expiry information for an address.
func (idx *IndexerClient) GetAddressLifespan(ctx context.Context, address string) (*IndexerEnvelope[[]BlobLifespan], error) {
	path := fmt.Sprintf("/v1/addresses/%s/lifespan", url.PathEscape(address))
	var env IndexerEnvelope[[]BlobLifespan]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetAddressLifespanBands retrieves lifespan totals grouped by band for an address.
func (idx *IndexerClient) GetAddressLifespanBands(ctx context.Context, address string) (*IndexerEnvelope[[]LifespanBand], error) {
	path := fmt.Sprintf("/v1/addresses/%s/lifespan?group=band", url.PathEscape(address))
	var env IndexerEnvelope[[]LifespanBand]
	if err := idx.get(ctx, path, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// CheckHealth queries the indexer /v1/health endpoint for projection sync checkpoint.
func (idx *IndexerClient) CheckHealth(ctx context.Context) (*IndexerEnvelope[IndexerHealth], error) {
	var env IndexerEnvelope[IndexerHealth]
	if err := idx.get(ctx, "/v1/health", &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// GetHealth is an alias for CheckHealth returning health and checkpoint freshness metadata.
func (idx *IndexerClient) GetHealth(ctx context.Context) (*IndexerEnvelope[IndexerHealth], error) {
	return idx.CheckHealth(ctx)
}

// CheckReadiness checks the indexer /readyz probe.
func (idx *IndexerClient) CheckReadiness(ctx context.Context) error {
	if idx == nil {
		return ErrNilClient
	}
	httpClient := idx.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL := idx.BaseURL
	if baseURL == "" {
		baseURL = "https://indexer.warlot.stevenhert.xyz"
	}
	u := baseURL + "/readyz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("create readiness request: %w", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("indexer readiness request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("indexer readiness returned status %d", res.StatusCode)
	}
	return nil
}

// GetReadiness is an alias for CheckReadiness.
func (idx *IndexerClient) GetReadiness(ctx context.Context) error {
	return idx.CheckReadiness(ctx)
}
