//go:build e2e

// Package e2e contains black-box smoke tests that exercise the FULL
// notification pipeline end to end against a live `make dev` stack:
//
//	snapshot worker (change detection) -> EventBus "change.detected"
//	-> integration dispatch_event (resolve/seed destinations, enqueue delivery)
//	-> integration delivery worker -> email provider -> delivered
//
// This chain has broken silently twice with zero test coverage: once because
// the in-memory EventBus had no subscriber wired, once because an adapter
// queried a nonexistent `pages.title` column. This test exists so it can
// never regress again without a loud, fast failure.
//
// Prerequisites
//
//   - `make dev` running (postgres, scraper/extractor, monolith API on :3000).
//   - `make dev-web` running (Next.js on :3001; the Go monolith on :3000
//     proxies unmatched routes to it, including the `/lecture-ai` demo page
//     and its `/api/demo/lecture-ai` toggle endpoint).
//   - The worker needs a way to actually send the email. Either:
//     (a) EMAIL_PROVIDER=log set on the worker container, so it logs instead
//     of calling a real provider, or
//     (b) a real RESEND_API_KEY configured, in which case this test targets
//     delivered@resend.dev — Resend's official safe test address that is
//     always accepted and never actually delivered to an inbox.
//
// Run with:
//
//	make e2e-notifications
//
// Configuration (env vars, all optional — defaults match `make dev`):
//
//	E2E_DB_DSN   - postgres DSN reachable from the HOST (default matches
//	               docker-compose's postgres port mapping + docker-compose.yml
//	               DB_USER/DB_PASSWORD/DB_NAME defaults)
//	E2E_API_URL  - Go monolith URL reachable from the HOST (default
//	               http://localhost:3000)
//	E2E_PAGE_URL - URL the WORKER container will fetch to render the demo
//	               page. The worker runs inside the docker-compose network,
//	               where the Go monolith is reachable at the `monolith`
//	               service hostname (see docker-compose.yml), NOT localhost.
//	               (default http://monolith:3000/lecture-ai)
package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

const (
	tenantSchema    = "e2e_notif"
	tenantSubdomain = "e2e-notif"
	seedEmail       = "e2e-notif@pulzifi.test"
	testRecipient   = "delivered@resend.dev"
)

// repoRoot resolves the repository root from this test file's own location,
// independent of the process working directory. `go test` sets cwd to the
// package directory (tests/e2e/), so anything relying on a repo-root-relative
// path (like shared/database.ProvisionTenantSchema, whose migration source
// URL is built from os.Getwd()) would silently look in the wrong place if
// called in-process from here. That is exactly why tenant provisioning below
// shells out to `go run ./cmd/migrate` with an explicit Dir instead of
// calling database.ProvisionTenantSchema directly.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to resolve test file path")
	}
	// tests/e2e/notification_pipeline_test.go -> repo root is two levels up.
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func getenvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// TestNotificationPipeline drives the full change-detection -> email-delivery
// pipeline against a live dev stack. See the package doc comment for
// prerequisites and configuration.
func TestNotificationPipeline(t *testing.T) {
	dbDSN := getenvDefault("E2E_DB_DSN", "postgres://pulzifi_user:pulzifi_password@localhost:5434/pulzifi?sslmode=disable")
	apiURL := getenvDefault("E2E_API_URL", "http://localhost:3000")
	pageURL := getenvDefault("E2E_PAGE_URL", "http://monolith:3000/lecture-ai")

	root := repoRoot(t)

	// (a) Connect to Postgres, reachable from the HOST.
	t.Log("connecting to database:", dbDSN)
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		t.Skipf("database unreachable at %s (run `make dev` first): %v", dbDSN, err)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	// (b) Toggle the demo page to version 1 (baseline content).
	t.Log("setting demo page to version 1")
	if err := postLectureAIVersion(httpClient, apiURL, "1"); err != nil {
		t.Skipf("could not reach %s/api/demo/lecture-ai (run `make dev-web` first): %v", apiURL, err)
	}

	// (c) Seed tenant e2e_notif. Not deferred/cleaned up on failure so a
	// broken run leaves inspectable state in the DB.
	t.Log("dropping any stale e2e_notif schema/org/user")
	if err := database.DeprovisionTenantSchema(db, tenantSchema); err != nil {
		t.Fatalf("DeprovisionTenantSchema(%q): %v", tenantSchema, err)
	}
	if _, err := db.Exec(`DELETE FROM public.organizations WHERE schema_name = $1 OR subdomain = $2`, tenantSchema, tenantSubdomain); err != nil {
		t.Fatalf("delete stale organization: %v", err)
	}
	if _, err := db.Exec(`DELETE FROM public.users WHERE email = $1`, seedEmail); err != nil {
		t.Fatalf("delete stale user: %v", err)
	}

	t.Log("seeding public.users + public.organizations")
	var userID uuid.UUID
	if err := db.QueryRow(`
		INSERT INTO public.users (email, password_hash, email_notifications_enabled, status)
		VALUES ($1, 'not-a-real-hash', TRUE, 'approved')
		RETURNING id`, seedEmail).Scan(&userID); err != nil {
		t.Fatalf("insert seed user: %v", err)
	}

	var orgID uuid.UUID
	if err := db.QueryRow(`
		INSERT INTO public.organizations (name, subdomain, schema_name, owner_user_id)
		VALUES ('E2E Notification Pipeline', $1, $2, $3)
		RETURNING id`, tenantSubdomain, tenantSchema, userID).Scan(&orgID); err != nil {
		t.Fatalf("insert seed organization: %v", err)
	}

	t.Log("provisioning tenant schema + running tenant migrations via `go run ./cmd/migrate`")
	if err := runTenantMigrations(t, root, dbDSN, tenantSchema); err != nil {
		t.Fatalf("tenant migration failed: %v", err)
	}

	// (d) Seed workspace, member, page, monitoring config.
	t.Log("seeding workspace/workspace_member/page/monitoring_config")
	var workspaceID, pageID uuid.UUID
	err = database.WithTenant(context.Background(), db, tenantSchema, func(tx *sql.Tx) error {
		if err := tx.QueryRow(`
			INSERT INTO workspaces (name, type, created_by)
			VALUES ('E2E Workspace', 'general', $1)
			RETURNING id`, userID).Scan(&workspaceID); err != nil {
			return fmt.Errorf("insert workspace: %w", err)
		}

		if _, err := tx.Exec(`
			INSERT INTO workspace_members (workspace_id, user_id, role)
			VALUES ($1, $2, 'owner')`, workspaceID, userID); err != nil {
			return fmt.Errorf("insert workspace_member: %w", err)
		}

		if err := tx.QueryRow(`
			INSERT INTO pages (workspace_id, name, url, created_by)
			VALUES ($1, 'E2E Lecture AI Demo Page', $2, $3)
			RETURNING id`, workspaceID, pageURL, userID).Scan(&pageID); err != nil {
			return fmt.Errorf("insert page: %w", err)
		}

		// check_frequency='5m' (a valid, non-'Off' frequency key — see
		// modules/monitoring/domain/entities/frequency.go). The scheduler's
		// poll-mode due condition is:
		//   check_frequency != 'Off' AND (last_checked_at IS NULL OR
		//     last_checked_at < NOW() - INTERVAL <freq>)
		// enabled_alert_conditions defaults to '["any_changes"]', which is
		// required for the snapshot worker to call createAlert/publish
		// change.detected at all.
		if _, err := tx.Exec(`
			INSERT INTO monitoring_configs (page_id, check_frequency)
			VALUES ($1, '5m')`, pageID); err != nil {
			return fmt.Errorf("insert monitoring_config: %w", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed workspace/page/config: %v", err)
	}

	// (e) Seed an explicit email destination at workspace scope. A unique
	// partial index (uq_integration_destinations_enabled_email_scope) allows
	// at most one *enabled* email destination per scope, so this also
	// prevents dispatch_event's own default-email fallback from creating a
	// duplicate later.
	t.Log("seeding integration_destinations (email, workspace scope)")
	destinationID := uuid.New()
	targetJSON, _ := json.Marshal(map[string]any{"emails": []string{testRecipient}})
	err = database.WithTenant(context.Background(), db, tenantSchema, func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			INSERT INTO integration_destinations
				(id, service_type, scope_type, scope_id, target, events, enabled)
			VALUES ($1, 'email', 'workspace', $2, $3::jsonb, ARRAY['change.detected','alert.created'], TRUE)`,
			destinationID, workspaceID, targetJSON)
		return err
	})
	if err != nil {
		t.Fatalf("seed email destination: %v", err)
	}

	// (f) Ensure the page is due, then poll for the baseline check.
	t.Log("marking page due and waiting for baseline check")
	if err := markPageDue(db, tenantSchema, pageID); err != nil {
		t.Fatalf("markPageDue (baseline): %v", err)
	}

	baselineOK := pollUntil(t, 90*time.Second, 3*time.Second, func() (bool, error) {
		var count int
		err := database.WithTenant(context.Background(), db, tenantSchema, func(tx *sql.Tx) error {
			return tx.QueryRow(`
				SELECT COUNT(*) FROM checks
				WHERE page_id = $1 AND status = 'success'`, pageID).Scan(&count)
		})
		return count > 0, err
	})
	if !baselineOK {
		dumpDebugState(t, db, tenantSchema, pageID)
		t.Fatal("timed out waiting for baseline check to complete (90s) — is the worker running with ENABLE_WORKERS=true?")
	}

	// (g) Toggle content to version 2, mark the page due again, then poll
	// for the delivered change.detected email.
	t.Log("setting demo page to version 2 (content change)")
	if err := postLectureAIVersion(httpClient, apiURL, "2"); err != nil {
		t.Fatalf("toggle demo page to version 2: %v", err)
	}

	t.Log("marking page due again and waiting for change.detected -> delivered")
	if err := markPageDue(db, tenantSchema, pageID); err != nil {
		t.Fatalf("markPageDue (post-change): %v", err)
	}

	deliveredOK := pollUntil(t, 120*time.Second, 3*time.Second, func() (bool, error) {
		var count int
		err := database.WithTenant(context.Background(), db, tenantSchema, func(tx *sql.Tx) error {
			return tx.QueryRow(`
				SELECT COUNT(*) FROM integration_deliveries d
				JOIN integration_destinations dest ON dest.id = d.destination_id
				WHERE dest.id = $1 AND d.event_type = 'change.detected' AND d.status = 'delivered'`,
				destinationID).Scan(&count)
		})
		return count > 0, err
	})
	if !deliveredOK {
		dumpDebugState(t, db, tenantSchema, pageID)
		t.Fatal("timed out waiting for change.detected delivery to reach status=delivered (120s)")
	}

	t.Log("notification pipeline E2E smoke test passed")
}

// postLectureAIVersion POSTs {"version": v} to the demo toggle endpoint.
func postLectureAIVersion(client *http.Client, apiURL, version string) error {
	body, _ := json.Marshal(map[string]string{"version": version})
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(apiURL, "/")+"/api/demo/lecture-ai", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from /api/demo/lecture-ai", resp.StatusCode)
	}
	return nil
}

// runTenantMigrations shells out to `go run ./cmd/migrate` instead of calling
// database.ProvisionTenantSchema directly. ProvisionTenantSchema builds its
// migration source URL from os.Getwd()+"shared/database/migrations/tenant",
// assuming the process cwd is the repo root — true for a compiled binary or
// `go run` invoked from root, but NOT true for `go test`, which always runs
// with cwd set to the test's own package directory (tests/e2e/). Calling it
// in-process here would silently look for migrations under
// tests/e2e/shared/database/migrations/tenant, which does not exist.
// Shelling out with an explicit Dir sidesteps that coupling entirely and
// matches how the Makefile itself runs migrations.
func runTenantMigrations(t *testing.T, repoRootDir, dbDSN, tenant string) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/migrate",
		"-scope", "tenant",
		"-tenant", tenant,
		"-cmd", "up",
		"-db", dbDSN,
	)
	cmd.Dir = repoRootDir
	out, err := cmd.CombinedOutput()
	t.Logf("migrate output:\n%s", out)
	if err != nil {
		return fmt.Errorf("go run ./cmd/migrate: %w", err)
	}
	return nil
}

// markPageDue resets last_checked_at/next_run_at so the page is immediately
// due under BOTH scheduler modes:
//   - poll (default, SCHEDULER_MODE unset): due when last_checked_at IS NULL
//     or older than the configured frequency interval.
//   - queue (SCHEDULER_MODE=queue): due when next_run_at IS NOT NULL and
//     <= NOW().
func markPageDue(db *sql.DB, tenant string, pageID uuid.UUID) error {
	return database.WithTenant(context.Background(), db, tenant, func(tx *sql.Tx) error {
		_, err := tx.Exec(`
			UPDATE pages
			SET last_checked_at = NULL, next_run_at = NOW() - INTERVAL '1 minute'
			WHERE id = $1`, pageID)
		return err
	})
}

// pollUntil calls check repeatedly until it returns true, an error, or the
// timeout elapses. Returns whether check ever reported true.
func pollUntil(t *testing.T, timeout, interval time.Duration, check func() (bool, error)) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ok, err := check()
		if err != nil {
			t.Logf("poll check error (continuing): %v", err)
		} else if ok {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// dumpDebugState logs alerts and integration_deliveries rows for the tenant
// to aid debugging a timed-out run. Best-effort: logs and swallows its own
// query errors rather than failing the test a second time.
func dumpDebugState(t *testing.T, db *sql.DB, tenant string, pageID uuid.UUID) {
	t.Helper()
	_ = database.WithTenant(context.Background(), db, tenant, func(tx *sql.Tx) error {
		rows, err := tx.Query(`
			SELECT id, type, title, change_summary, created_at
			FROM alerts WHERE page_id = $1 ORDER BY created_at DESC LIMIT 10`, pageID)
		if err != nil {
			t.Logf("dumpDebugState: query alerts: %v", err)
		} else {
			defer func() { _ = rows.Close() }()
			t.Log("-- alerts for page --")
			for rows.Next() {
				var id, alertType, title, summary string
				var createdAt time.Time
				if err := rows.Scan(&id, &alertType, &title, &summary, &createdAt); err == nil {
					t.Logf("alert id=%s type=%s title=%q summary=%q created_at=%s", id, alertType, title, summary, createdAt)
				}
			}
		}

		dRows, err := tx.Query(`
			SELECT id, destination_id, event_type, status, attempts, error_message, created_at
			FROM integration_deliveries ORDER BY created_at DESC LIMIT 10`)
		if err != nil {
			t.Logf("dumpDebugState: query integration_deliveries: %v", err)
		} else {
			defer func() { _ = dRows.Close() }()
			t.Log("-- integration_deliveries (tenant-wide) --")
			for dRows.Next() {
				var id, destID, eventType, status string
				var attempts int
				var errMsg sql.NullString
				var createdAt time.Time
				if err := dRows.Scan(&id, &destID, &eventType, &status, &attempts, &errMsg, &createdAt); err == nil {
					t.Logf("delivery id=%s dest=%s event=%s status=%s attempts=%d error=%q created_at=%s",
						id, destID, eventType, status, attempts, errMsg.String, createdAt)
				}
			}
		}
		return nil
	})
}
