package featureflags_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/shared/featureflags"
)

func openTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// seedOrg inserts a fresh user + org with the given feature_flags JSONB and registers cleanup.
func seedOrg(t *testing.T, db *sql.DB, flags string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	orgID := uuid.New()
	sub := "ff-test-" + orgID.String()[:8]
	email := "ff-test-" + orgID.String()[:8] + "@example.com"

	_, err := db.Exec(`INSERT INTO users (id, email, password_hash, created_at, updated_at)
                       VALUES ($1::uuid, $2, 'x', NOW(), NOW())`,
		userID, email)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`INSERT INTO organizations (id, name, subdomain, schema_name, owner_user_id, feature_flags, created_at, updated_at)
                       VALUES ($1::uuid, $2, $3, $3, $4::uuid, $5::jsonb, NOW(), NOW())`,
		orgID, orgID.String(), sub, userID, flags)
	if err != nil {
		db.Exec(`DELETE FROM users WHERE id = $1`, userID)
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Exec(`DELETE FROM organizations WHERE id = $1`, orgID)
		db.Exec(`DELETE FROM users WHERE id = $1`, userID)
	})
	return orgID
}

func TestIsOn_TrueWhenPresent(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db, `{"integrations":{"discord":true}}`)
	r := featureflags.NewReader(db)

	got, err := r.IsOn(context.Background(), orgID, "integrations.discord")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Errorf("expected true, got false")
	}
}

func TestIsOn_FalseWhenMissingPath(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db, `{"integrations":{"discord":true}}`)
	r := featureflags.NewReader(db)

	got, err := r.IsOn(context.Background(), orgID, "integrations.twilio")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("expected false for missing path, got true")
	}
}

func TestIsOn_FalseWhenMissingOrg(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	randomID := uuid.New()
	r := featureflags.NewReader(db)

	got, err := r.IsOn(context.Background(), randomID, "integrations.discord")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("expected false for missing org, got true")
	}
}

func TestIsOn_FalseWhenNonBoolValue(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db, `{"integrations":{"discord":"yes"}}`)
	r := featureflags.NewReader(db)

	got, err := r.IsOn(context.Background(), orgID, "integrations.discord")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Errorf("expected false for non-bool value, got true")
	}
}

func TestSnapshot_ReturnsFullBlob(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	orgID := seedOrg(t, db, `{"integrations":{"discord":true,"twilio":false},"other":"x"}`)
	r := featureflags.NewReader(db)

	snap, err := r.Snapshot(context.Background(), orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil snapshot")
	}

	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	s := string(raw)
	if !strings.Contains(s, "discord") {
		t.Errorf("expected snapshot to contain 'discord', got: %s", s)
	}
	if !strings.Contains(s, "other") {
		t.Errorf("expected snapshot to contain 'other', got: %s", s)
	}
}
