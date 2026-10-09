package warlot

import "time"

// ---- Project Models ----

type InitProjectRequest struct {
	HolderID      string `json:"holder_id"`
	ProjectName   string `json:"project_name"`
	OwnerAddress  string `json:"owner_address"`
	EpochSet      int    `json:"epoch_set"`
	CycleEnd      int    `json:"cycle_end"`
	WritersLen    int    `json:"writers_len"`
	TrackBackLen  int    `json:"track_back_len"`
	DraftEpochDur int    `json:"draft_epoch_dur"`
	IncludePass   bool   `json:"include_pass"`
	Deletable     bool   `json:"deletable"`
}

type InitProjectResponse struct {
	ProjectID    string `json:"ProjectID"`
	DBID         string `json:"DBID"`
	WriterPassID string `json:"WriterPassID"`
	BlobID       string `json:"BlobID"`
	TxDigest     string `json:"TxDigest"`
	CSVHashHex   string `json:"CSVHashHex"`
	DigestHex    string `json:"DigestHex"`
	SignatureHex string `json:"SignatureHex"`
}

type IssueKeyRequest struct {
	ProjectID     string `json:"projectId"`
	ProjectHolder string `json:"projectHolder"`
	ProjectName   string `json:"projectName"`
	User          string `json:"user"`
}

type IssueKeyResponse struct {
	APIKey string `json:"apiKey"`
	URL    string `json:"url"`
}

type ResolveProjectRequest struct {
	HolderID    string `json:"holder_id"`
	ProjectName string `json:"project_name"`
}

// ResolveProjectResponse accepts both modern snake_case and legacy PascalCase.
type ResolveProjectResponse struct {
	// Current/observed fields.
	ExistsMeta  bool   `json:"exists_meta"`
	ExistsChain bool   `json:"exists_chain"`
	ProjectID   string `json:"project_id"`
	DBID        string `json:"db_id"`
	Action      string `json:"action"`
	// Legacy fields.
	LegacyProjectID string `json:"ProjectID,omitempty"`
	LegacyDBID      string `json:"DBID,omitempty"`
}

type TableCountResponse struct {
	ProjectID  string `json:"project_id"`
	TableCount int    `json:"table_count"`
}

// ---- SQL Models ----

type SQLRequest struct {
	SQL    string        `json:"sql"`
	Params []interface{} `json:"params"`
}

// SQLResponse supports both DDL/DML and SELECT shapes.
type SQLResponse struct {
	OK       bool                     `json:"ok"`
	RowCount *int                     `json:"row_count,omitempty"`
	Rows     []map[string]interface{} `json:"rows,omitempty"`
	Error    string                   `json:"error,omitempty"`
}

// ---- Tables and Status Models ----

type ListTablesResponse struct {
	Tables []string `json:"tables"`
}

type BrowseRowsResponse struct {
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
	Table  string                   `json:"table"`
	Rows   []map[string]interface{} `json:"rows"`
}

// TableColumn describes a single column in a table schema.
type TableColumn struct {
	CID       int         `json:"cid"`
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	NotNull   bool        `json:"not_null"`
	Default   interface{} `json:"default"`
	PrimaryPK bool        `json:"primary_key"`
}

// TableSchema describes a table and its columns.
type TableSchema struct {
	Table   string        `json:"table"`
	Columns []TableColumn `json:"columns"`
}

// TableEdit represents an insert, update, or delete modification on a table.
type TableEdit struct {
	Where  map[string]any `json:"where,omitempty"`
	Set    map[string]any `json:"set,omitempty"`
	Insert map[string]any `json:"insert,omitempty"`
	Delete bool           `json:"delete,omitempty"`
}

// PatchTableRequest represents a batch of structured edits to apply to a table.
type PatchTableRequest struct {
	Edits []TableEdit `json:"edits"`
}

// PatchTableResponse represents the result of applying a batch of table edits.
type PatchTableResponse struct {
	OK       bool  `json:"ok"`
	RowCount int64 `json:"row_count"`
}

// ProjectStatus describes the replication, synchronization, and lifecycle state of a project.
type ProjectStatus struct {
	ProjectID                  string  `json:"project_id"`
	DBID                       string  `json:"db_id"`
	WriterPassID               string  `json:"writer_pass_id"`
	LastUploadSeq              int64   `json:"last_upload_seq"`
	LastUploadAt               *string `json:"last_upload_at,omitempty"`
	MaxSeq                     int64   `json:"max_seq"`
	Synced                     bool    `json:"synced"`
	IsActive                   bool    `json:"is_active"`
	Status                     string  `json:"status"`
	TerminationState           string  `json:"termination_state,omitempty"`
	TerminationFilesTotal      int     `json:"termination_files_total,omitempty"`
	TerminationFilesCompleted  int     `json:"termination_files_completed,omitempty"`
	TerminatedAt               *string `json:"terminated_at,omitempty"`
	DeactivatedAt              *string `json:"deactivated_at,omitempty"`
}

// ReplicationLag calculates the count of mutations recorded locally awaiting on-chain commit.
func (s *ProjectStatus) ReplicationLag() int64 {
	if s == nil {
		return 0
	}
	lag := s.MaxSeq - s.LastUploadSeq
	if lag < 0 {
		return 0
	}
	return lag
}

// Freshness returns how long ago the project's state was last anchored on-chain.
// Returns 0, false if LastUploadAt is nil or unparseable.
func (s *ProjectStatus) Freshness() (time.Duration, bool) {
	if s == nil || s.LastUploadAt == nil || *s.LastUploadAt == "" {
		return 0, false
	}
	t, err := time.Parse(time.RFC3339Nano, *s.LastUploadAt)
	if err != nil {
		t, err = time.Parse(time.RFC3339, *s.LastUploadAt)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05.999999-07:00", *s.LastUploadAt)
			if err != nil {
				return 0, false
			}
		}
	}
	return time.Since(t), true
}

// DependencyStatus describes the reachability of a cluster dependency.
type DependencyStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// ReadinessResponse describes the overall readiness state of the cluster.
type ReadinessResponse struct {
	Status       string             `json:"status"`
	Dependencies []DependencyStatus `json:"dependencies"`
}

// IsReady reports whether all probed dependencies are reachable and healthy.
func (r *ReadinessResponse) IsReady() bool {
	if r == nil || r.Status != "ready" {
		return false
	}
	for _, d := range r.Dependencies {
		if d.Status != "ok" {
			return false
		}
	}
	return true
}

// CommitResponse describes the result of anchoring pending operations on-chain.
type CommitResponse struct {
	CommittedOps  int64  `json:"committed_ops"`
	TxDigest      string `json:"tx_digest"`
	LastUploadSeq int64  `json:"last_upload_seq"`
	MaxSeq        int64  `json:"max_seq"`
}

// DeactivateProjectResponse describes the outcome of deactivating a project.
type DeactivateProjectResponse struct {
	ProjectID                  string `json:"project_id"`
	Status                     string `json:"status"`
	IsActive                   bool   `json:"is_active"`
	AlreadyDeactivated         bool   `json:"already_deactivated"`
	DeactivatedAt              string `json:"deactivated_at"`
	DirectChainCallsPermitted  bool   `json:"direct_chain_calls_permitted"`
	PendingOperationsAnchored  bool   `json:"pending_operations_anchored"`
	PendingOperationsCommitted int    `json:"pending_operations_committed"`
	Effect                     string `json:"effect"`
	NotDeleted                 string `json:"not_deleted"`
	ChainUnblocked             string `json:"chain_unblocked"`
	DataRetained               string `json:"data_retained"`
	NotPaused                  string `json:"not_paused"`
	PendingOperations          string `json:"pending_operations,omitempty"`
}

// ReactivateProjectResponse describes the outcome of reactivating a project.
type ReactivateProjectResponse struct {
	ProjectID     string `json:"project_id"`
	Status        string `json:"status"`
	IsActive      bool   `json:"is_active"`
	AlreadyActive bool   `json:"already_active"`
	ReactivatedAt string `json:"reactivated_at"`
	Effect        string `json:"effect"`
}

// TerminateProjectResponse describes the outcome of terminating a project.
type TerminateProjectResponse struct {
	ProjectID                  string `json:"project_id"`
	Status                     string `json:"status"`
	TerminationState           string `json:"termination_state"`
	TerminationFilesTotal      int    `json:"termination_files_total"`
	TerminationFilesCompleted  int    `json:"termination_files_completed"`
	KeysDestroyed              int    `json:"keys_destroyed,omitempty"`
	IsActive                   bool   `json:"is_active"`
	TerminatedAt               string `json:"terminated_at,omitempty"`
	AlreadyTerminated          bool   `json:"already_terminated,omitempty"`
	Effect                     string `json:"effect"`
	NotDeleted                 string `json:"not_deleted"`
}

// RateLimit captures the rate limit ceiling status parsed from response headers.
type RateLimit struct {
	Limit     uint64    `json:"limit"`
	Remaining uint64    `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

// RateLimitClass identifies the rate limit route tier in Warlot.
type RateLimitClass string

const (
	RateLimitClassOrdinary RateLimitClass = "ordinary"
	RateLimitClassChain    RateLimitClass = "chain"
)

