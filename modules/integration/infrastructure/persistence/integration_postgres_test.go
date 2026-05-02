//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/shared/crypto"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			t.Skip("DATABASE_URL (or DB_HOST) not set; skipping integration test")
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		name := os.Getenv("DB_NAME")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

func TestIntegrationRepo_RoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openTestDB(t)
	defer db.Close()

	enc, err := crypto.NewAESGCM(make([]byte, 32)) // zero key OK in tests
	if err != nil {
		t.Fatalf("NewAESGCM: %v", err)
	}

	repo := persistence.NewIntegrationPostgresRepository(db, enc)
	ctx := context.Background()

	// --- seed a user and organization for FK constraints ---
	userID := uuid.New()
	orgID := uuid.New()

	_, err = db.ExecContext(ctx, `
		INSERT INTO public.users (id, email, password_hash, status)
		VALUES ($1, $2, 'testhash', 'approved')
		ON CONFLICT DO NOTHING`,
		userID, fmt.Sprintf("testuser-%s@example.com", userID),
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		orgID,
		"Test Org "+orgID.String(),
		"test-"+orgID.String(),
		"test_"+orgID.String()[:8],
		userID,
	)
	if err != nil {
		t.Fatalf("insert org: %v", err)
	}

	t.Cleanup(func() {
		db.ExecContext(ctx, `DELETE FROM public.integrations WHERE org_id = $1`, orgID)
		db.ExecContext(ctx, `DELETE FROM public.organizations WHERE id = $1`, orgID)
		db.ExecContext(ctx, `DELETE FROM public.users WHERE id = $1`, userID)
	})

	// --- 1. Create ---
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Millisecond)
	integ := &entities.Integration{
		ID:          uuid.New(),
		OrgID:       orgID,
		ServiceType: "slack",
		Status:      entities.IntegrationActive,
		AccessToken: "access-secret-123",
		RefreshToken: "refresh-secret-456",
		TokenExpiresAt: &expires,
		ProviderMeta: map[string]any{"workspace": "my-workspace", "bot_user_id": "U12345"},
		CreatedBy:   userID,
	}

	if err := repo.Create(ctx, integ); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// --- 2. GetByID ---
	got, err := repo.GetByID(ctx, integ.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil")
	}
	if got.AccessToken != "access-secret-123" {
		t.Errorf("AccessToken mismatch: got %q", got.AccessToken)
	}
	if got.RefreshToken != "refresh-secret-456" {
		t.Errorf("RefreshToken mismatch: got %q", got.RefreshToken)
	}
	if got.ProviderMeta["workspace"] != "my-workspace" {
		t.Errorf("ProviderMeta[workspace] mismatch: got %v", got.ProviderMeta["workspace"])
	}
	if got.Status != entities.IntegrationActive {
		t.Errorf("Status mismatch: got %q", got.Status)
	}

	// --- 3. GetByOrgAndService ---
	got2, err := repo.GetByOrgAndService(ctx, orgID, "slack")
	if err != nil {
		t.Fatalf("GetByOrgAndService: %v", err)
	}
	if got2 == nil {
		t.Fatal("GetByOrgAndService returned nil")
	}
	if got2.ID != integ.ID {
		t.Errorf("ID mismatch: want %s got %s", integ.ID, got2.ID)
	}

	// GetByOrgAndService for unknown service returns nil, nil
	notFound, err := repo.GetByOrgAndService(ctx, orgID, "nonexistent")
	if err != nil {
		t.Fatalf("GetByOrgAndService(nonexistent): %v", err)
	}
	if notFound != nil {
		t.Errorf("expected nil for nonexistent service, got %+v", notFound)
	}

	// --- 4. ListByOrg ---
	list, err := repo.ListByOrg(ctx, orgID)
	if err != nil {
		t.Fatalf("ListByOrg: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("ListByOrg length: want 1 got %d", len(list))
	}
	if list[0].ID != integ.ID {
		t.Errorf("ListByOrg[0].ID mismatch")
	}

	// --- 5. Update ---
	got.AccessToken = "new-access-token"
	got.RefreshToken = ""
	got.Status = entities.IntegrationExpired
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := repo.GetByID(ctx, integ.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if updated.AccessToken != "new-access-token" {
		t.Errorf("updated AccessToken mismatch: got %q", updated.AccessToken)
	}
	if updated.RefreshToken != "" {
		t.Errorf("updated RefreshToken should be empty, got %q", updated.RefreshToken)
	}
	if updated.Status != entities.IntegrationExpired {
		t.Errorf("updated Status mismatch: got %q", updated.Status)
	}

	// --- 6. SoftDelete ---
	if err := repo.SoftDelete(ctx, integ.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	// GetByID after soft delete should return nil
	afterDelete, err := repo.GetByID(ctx, integ.ID)
	if err != nil {
		t.Fatalf("GetByID after SoftDelete: %v", err)
	}
	if afterDelete != nil {
		t.Errorf("expected nil after SoftDelete, got %+v", afterDelete)
	}

	// ListByOrg after soft delete should be empty
	listAfter, err := repo.ListByOrg(ctx, orgID)
	if err != nil {
		t.Fatalf("ListByOrg after SoftDelete: %v", err)
	}
	if len(listAfter) != 0 {
		t.Errorf("ListByOrg after SoftDelete: want 0 got %d", len(listAfter))
	}
}
