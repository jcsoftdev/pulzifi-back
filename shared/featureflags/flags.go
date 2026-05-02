package featureflags

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Reader struct{ db *sql.DB }

func NewReader(db *sql.DB) *Reader { return &Reader{db: db} }

// IsOn returns true if the dotted-path flag is present and exactly the bool true.
// Missing org → (false, nil). Missing path → (false, nil). DB error → (false, err).
// Non-bool value at the path → (false, nil) (default deny on type mismatch).
func (r *Reader) IsOn(ctx context.Context, orgID uuid.UUID, key string) (bool, error) {
	snap, err := r.Snapshot(ctx, orgID)
	if err != nil {
		return false, err
	}
	if snap == nil {
		return false, nil
	}
	parts := strings.Split(key, ".")
	cur := any(snap)
	for _, p := range parts {
		m, ok := cur.(map[string]any)
		if !ok {
			return false, nil
		}
		v, present := m[p]
		if !present {
			return false, nil
		}
		cur = v
	}
	b, ok := cur.(bool)
	if !ok {
		return false, nil
	}
	return b, nil
}

// Snapshot returns the full feature_flags blob for an org. Returns (nil, nil) if
// the org doesn't exist (default-deny semantics for the caller).
func (r *Reader) Snapshot(ctx context.Context, orgID uuid.UUID) (map[string]any, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT feature_flags FROM organizations WHERE id = $1 AND deleted_at IS NULL`, orgID,
	).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("featureflags: snapshot: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("featureflags: parse: %w", err)
	}
	return out, nil
}
