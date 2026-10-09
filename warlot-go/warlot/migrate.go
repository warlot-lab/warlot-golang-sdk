package warlot

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"
)

// Migrator applies SQL migration files idempotently and records them in
// a ledger table named _migrations (id TEXT PRIMARY KEY, applied_at TEXT).
type Migrator struct{}

// Migrate is a package-level migrator instance for convenience.
var Migrate Migrator

// Deprecated: use Migrate.
var migrate = Migrate

// Up applies .sql files in fsys under dir, sorted by filename. Already-applied
// migration IDs are skipped based on the _migrations ledger.
func (Migrator) Up(ctx context.Context, p Project, fsys fs.FS, dir string) (applied []string, err error) {
	if p.Client == nil {
		return nil, ErrNilClient
	}
	if fsys == nil {
		return nil, errors.New("warlot: fsys is nil")
	}
	// Ensure ledger exists.
	if _, err = p.SQL(ctx, `
		CREATE TABLE IF NOT EXISTS _migrations (
			id TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`, nil); err != nil {
		return nil, fmt.Errorf("create _migrations: %w", err)
	}

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	// Load applied set.
	type migRow struct {
		ID string `json:"id"`
	}
	existing, err := Query[migRow](ctx, p, `SELECT id FROM _migrations`, nil)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	appliedSet := map[string]struct{}{}
	for _, r := range existing {
		appliedSet[r.ID] = struct{}{}
	}

	for _, name := range names {
		if _, done := appliedSet[name]; done {
			continue
		}
		b, err := fs.ReadFile(fsys, dir+"/"+name)
		if err != nil {
			return applied, fmt.Errorf("read %s: %w", name, err)
		}
		sqlText := string(b)

		if _, err := p.SQL(ctx, sqlText, nil, WithIdempotencyKey("mig-"+name)); err != nil {
			return applied, fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := p.SQL(ctx,
			`INSERT INTO _migrations (id, applied_at) VALUES (?, ?)`,
			[]any{name, time.Now().UTC().Format(time.RFC3339)},
		); err != nil {
			return applied, fmt.Errorf("record %s: %w", name, err)
		}
		applied = append(applied, name)
	}
	return applied, nil
}
