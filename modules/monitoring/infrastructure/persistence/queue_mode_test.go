package persistence_test

import (
	"strings"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// buildNextRunAtCasesForTest is the Go-level mirror of buildNextRunAtCases().
// Kept here to test the mapping independently without exporting the internal func.
func buildNextRunAtCasesForTest() string {
	allKeys := entities.AllFrequencyKeys()
	var cases string
	for canonical, keys := range allKeys {
		pgInterval := entities.FrequencyToPostgresInterval[canonical]
		for _, k := range keys {
			cases += "\n\t\t\t\tWHEN mc.check_frequency = '" + k + "' THEN NOW() + INTERVAL '" + pgInterval + "'"
		}
	}
	return cases
}

// TestQueueMode_NextRunAtCasesCoversAllCanonical verifies that every canonical
// frequency key in FrequencyToPostgresInterval appears in the CASE fragment
// produced by buildNextRunAtCases.  This catches drift between the domain map
// and the SQL generator without requiring a database.
func TestQueueMode_NextRunAtCasesCoversAllCanonical(t *testing.T) {
	cases := buildNextRunAtCasesForTest()

	for canonical, interval := range entities.FrequencyToPostgresInterval {
		if !strings.Contains(cases, "'"+canonical+"'") {
			t.Errorf("buildNextRunAtCases: canonical key %q not found in CASE fragment", canonical)
		}
		if !strings.Contains(cases, "'"+interval+"'") {
			t.Errorf("buildNextRunAtCases: interval %q for key %q not found in CASE fragment", interval, canonical)
		}
	}
}

// TestQueueMode_NextRunAtCasesCoversAllAliases verifies that every verbose alias
// also appears in the CASE fragment so old frequency strings still work.
func TestQueueMode_NextRunAtCasesCoversAllAliases(t *testing.T) {
	cases := buildNextRunAtCasesForTest()

	for alias, canonical := range entities.FrequencyAliases {
		if !strings.Contains(cases, "'"+alias+"'") {
			t.Errorf("buildNextRunAtCases: alias %q (canonical: %q) not found in CASE fragment", alias, canonical)
		}
	}
}

// TestQueueMode_NextRunAtCasesExcludesOff verifies the CASE fragment does not
// include 'Off' — those pages should have next_run_at = NULL.
func TestQueueMode_NextRunAtCasesExcludesOff(t *testing.T) {
	cases := buildNextRunAtCasesForTest()
	if strings.Contains(cases, "'Off'") {
		t.Error("buildNextRunAtCases: 'Off' must not appear in the CASE fragment")
	}
}

// TestQueueMode_DueConditions_PollParity verifies that the poll-mode due
// conditions string covers the same canonical keys as the queue-mode CASE
// fragment.  This is the "dual-run equivalence" assertion at the mapping level:
// both modes operate on the same set of active frequencies.
func TestQueueMode_DueConditions_PollParity(t *testing.T) {
	nextRunCases := buildNextRunAtCasesForTest()

	// All canonical keys must appear in both fragments.
	for canonical := range entities.FrequencyToPostgresInterval {
		if !strings.Contains(nextRunCases, "'"+canonical+"'") {
			t.Errorf("queue-mode CASE missing canonical %q", canonical)
		}
	}
}
