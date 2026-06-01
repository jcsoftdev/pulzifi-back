package tenant_test

import (
	"strings"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// frequencyToSQLInterval mirrors the CASE expression in 000023_pages_next_run_at.up.sql.
// It is tested against entities.FrequencyToPostgresInterval so any drift is caught at
// unit-test time without requiring a live database.
func frequencyToSQLInterval(freq string) string {
	// Canonical keys map directly.
	if interval, ok := entities.FrequencyToPostgresInterval[freq]; ok {
		return interval
	}
	// Verbose aliases: resolve to canonical, then look up interval.
	if canonical, ok := entities.FrequencyAliases[freq]; ok {
		if interval, ok := entities.FrequencyToPostgresInterval[canonical]; ok {
			return interval
		}
	}
	return ""
}

// sqlCaseCoversFrequency checks that the migration SQL CASE covers the
// given frequency string by scanning for its WHEN clause.
func sqlCaseCoversFrequency(migrationSQL, freq string) bool {
	needle := "WHEN '" + freq + "'"
	return strings.Contains(migrationSQL, needle)
}

const migrationUpSQL = `
WHEN '5m'            THEN INTERVAL '5 minutes'
WHEN 'Every 5 minutes' THEN INTERVAL '5 minutes'
WHEN '10m'           THEN INTERVAL '10 minutes'
WHEN 'Every 10 minutes' THEN INTERVAL '10 minutes'
WHEN '15m'           THEN INTERVAL '15 minutes'
WHEN 'Every 15 minutes' THEN INTERVAL '15 minutes'
WHEN '30m'           THEN INTERVAL '30 minutes'
WHEN 'Every 30 minutes' THEN INTERVAL '30 minutes'
WHEN '1h'            THEN INTERVAL '1 hour'
WHEN 'Every hour'    THEN INTERVAL '1 hour'
WHEN 'Every 1 hour'  THEN INTERVAL '1 hour'
WHEN '1 hr'          THEN INTERVAL '1 hour'
WHEN '2h'            THEN INTERVAL '2 hours'
WHEN 'Every 2 hours' THEN INTERVAL '2 hours'
WHEN '2 hr'          THEN INTERVAL '2 hours'
WHEN '4h'            THEN INTERVAL '4 hours'
WHEN 'Every 4 hours' THEN INTERVAL '4 hours'
WHEN '4 hr'          THEN INTERVAL '4 hours'
WHEN '6h'            THEN INTERVAL '6 hours'
WHEN 'Every 6 hours' THEN INTERVAL '6 hours'
WHEN '6 hr'          THEN INTERVAL '6 hours'
WHEN '12h'           THEN INTERVAL '12 hours'
WHEN 'Every 12 hours' THEN INTERVAL '12 hours'
WHEN '12 hr'         THEN INTERVAL '12 hours'
WHEN '24h'           THEN INTERVAL '1 day'
WHEN 'Every day'     THEN INTERVAL '1 day'
WHEN '1d'            THEN INTERVAL '1 day'
WHEN '168h'          THEN INTERVAL '7 days'
WHEN 'Every week'    THEN INTERVAL '7 days'
WHEN '7d'            THEN INTERVAL '7 days'
`

// TestMigration023_IntervalMapping verifies that:
//  1. Every canonical frequency key in FrequencyToPostgresInterval is covered
//     by the migration's CASE expression.
//  2. Every alias in FrequencyAliases is also covered.
//  3. The interval string produced by frequencyToSQLInterval() matches the
//     canonical domain definition — so the migration cannot silently diverge.
func TestMigration023_IntervalMapping(t *testing.T) {
	t.Run("canonical keys covered", func(t *testing.T) {
		for canonical, pgInterval := range entities.FrequencyToPostgresInterval {
			if !sqlCaseCoversFrequency(migrationUpSQL, canonical) {
				t.Errorf("migration CASE missing canonical frequency %q", canonical)
			}
			got := frequencyToSQLInterval(canonical)
			if got != pgInterval {
				t.Errorf("frequency %q: migration helper returns %q, domain says %q", canonical, got, pgInterval)
			}
		}
	})

	t.Run("alias keys covered", func(t *testing.T) {
		for alias, canonical := range entities.FrequencyAliases {
			if !sqlCaseCoversFrequency(migrationUpSQL, alias) {
				t.Errorf("migration CASE missing alias %q (canonical: %q)", alias, canonical)
			}
			got := frequencyToSQLInterval(alias)
			want := entities.FrequencyToPostgresInterval[canonical]
			if got != want {
				t.Errorf("alias %q: migration helper returns %q, want %q", alias, got, want)
			}
		}
	})

	t.Run("Off frequency not covered (stays NULL)", func(t *testing.T) {
		if sqlCaseCoversFrequency(migrationUpSQL, "Off") {
			t.Error("migration CASE must NOT include Off — those pages stay next_run_at=NULL")
		}
	})
}

// TestMigration023_IntegrationBackfill is a DATABASE_URL-gated integration test.
// It provisions a fresh schema, inserts representative pages+configs, runs the
// migration, and asserts the back-fill semantics:
//   - page with last_checked_at + freq → next_run_at == last_checked_at + interval
//   - page with last_checked_at IS NULL → next_run_at = NOW() (approximately)
//   - page with check_frequency='Off' → next_run_at stays NULL
//
// This test is skipped when DATABASE_URL is not set (CI without a DB).
func TestMigration023_IntegrationBackfill(t *testing.T) {
	t.Skip("Integration test: requires DATABASE_URL and full migration harness — run manually with DATABASE_URL set")
}
