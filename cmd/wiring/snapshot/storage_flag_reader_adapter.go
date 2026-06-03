package snapshotwiring

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	snapRepos "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/featureflags"
)

// storageFlagReaderAdapter implements snapshot's StorageFlagReader port by
// bridging schema-name → org-ID lookup and the shared featureflags.Reader.
// Default-deny on any error: returns (false, nil) so callers see "prefixing off"
// and the system remains safe/public by default.
type storageFlagReaderAdapter struct {
	db      *sql.DB
	ffReader *featureflags.Reader
}

// NewStorageFlagReader builds a StorageFlagReader adapter backed by the shared DB pool.
// The adapter resolves schemaName → orgID via public.organizations and then
// reads the "snapshot.private_storage" flag from the org's feature_flags column.
func NewStorageFlagReader(db *sql.DB) snapRepos.StorageFlagReader {
	return &storageFlagReaderAdapter{
		db:       db,
		ffReader: featureflags.NewReader(db),
	}
}

// IsPrivateStorageEnabled returns true when the "snapshot.private_storage" feature
// flag is set to true for the organization that owns the given schema. On any
// error (org not found, DB failure, flag missing/non-bool) it returns false so
// callers fall back to the safe public behaviour (no key prefixing).
func (a *storageFlagReaderAdapter) IsPrivateStorageEnabled(ctx context.Context, schemaName string) (bool, error) {
	orgID, err := a.orgIDBySchemaName(ctx, schemaName)
	if err != nil {
		// Schema not found or DB error → default deny (no prefix).
		return false, nil //nolint:nilerr // intentional default-deny
	}

	enabled, err := a.ffReader.IsOn(ctx, orgID, "snapshot.private_storage")
	if err != nil {
		// Flag reader error → default deny.
		return false, nil //nolint:nilerr
	}

	return enabled, nil
}

func (a *storageFlagReaderAdapter) orgIDBySchemaName(ctx context.Context, schemaName string) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := a.db.QueryRowContext(ctx,
		`SELECT id FROM public.organizations WHERE schema_name = $1 AND deleted_at IS NULL`,
		schemaName,
	).Scan(&orgID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("storageFlagReader: org lookup: %w", err)
	}
	return orgID, nil
}
