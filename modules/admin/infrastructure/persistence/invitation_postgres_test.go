package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/persistence"
)

// openTestDB opens a connection to the dev Postgres instance for integration
// tests. The test is skipped (not failed) when DB_HOST is unset or the database
// is unreachable, so unit-only `go test ./...` runs continue to pass.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	host := os.Getenv("DB_HOST")
	if host == "" {
		t.Skip("DB_HOST not set; skipping postgres integration test")
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), host, os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Skip("postgres unreachable: ", err)
	}
	return db
}

// seedInviter inserts a unique inviter user and returns its ID. The email
// embeds a fresh UUID so concurrent test runs do not collide on the unique
// users.email constraint.
func seedInviter(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := db.Exec(
		`INSERT INTO public.users (id, email, password_hash, first_name, last_name, email_verified) VALUES ($1, $2, 'x', 'a', 'b', TRUE)`,
		id, "inviter+"+id.String()+"@example.com",
	); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestCreate_Success(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewInvitationPostgres(db)
	inviter := seedInviter(t, db)
	ctx := context.Background()

	inv, err := repo.Create(ctx, repositories.CreateInvitationInput{
		Email:              "newuser+" + uuid.NewString() + "@example.com",
		Token:              uuid.NewString(),
		InvitedBy:          inviter,
		ExpiresAt:          time.Now().Add(72 * time.Hour),
		DailyCapPerInviter: 50,
		DailyCapGlobal:     200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if inv.Status != "pending" {
		t.Fatalf("expected pending, got %s", inv.Status)
	}
}

func TestCreate_ConcurrentSameEmail_ExactlyOneSucceeds(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewInvitationPostgres(db)
	inviter := seedInviter(t, db)
	ctx := context.Background()

	email := "concurrent+" + uuid.NewString() + "@example.com"

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.Create(ctx, repositories.CreateInvitationInput{
				Email:              email,
				Token:              uuid.NewString(),
				InvitedBy:          inviter,
				ExpiresAt:          time.Now().Add(72 * time.Hour),
				DailyCapPerInviter: 50,
				DailyCapGlobal:     200,
			})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	var ok, dup int
	for err := range results {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, domerrs.ErrCannotInviteEmail):
			dup++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if ok != 1 || dup != 1 {
		t.Fatalf("expected 1 success + 1 duplicate, got ok=%d dup=%d", ok, dup)
	}
}

func TestCreate_DailyCapExceeded(t *testing.T) {
	db := openTestDB(t)
	repo := persistence.NewInvitationPostgres(db)
	inviter := seedInviter(t, db)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_, err := repo.Create(ctx, repositories.CreateInvitationInput{
			Email:              "cap" + uuid.NewString() + "@example.com",
			Token:              uuid.NewString(),
			InvitedBy:          inviter,
			ExpiresAt:          time.Now().Add(72 * time.Hour),
			DailyCapPerInviter: 2,
			DailyCapGlobal:     200,
		})
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	_, err := repo.Create(ctx, repositories.CreateInvitationInput{
		Email:              "cap-overflow" + uuid.NewString() + "@example.com",
		Token:              uuid.NewString(),
		InvitedBy:          inviter,
		ExpiresAt:          time.Now().Add(72 * time.Hour),
		DailyCapPerInviter: 2,
		DailyCapGlobal:     200,
	})
	if !errors.Is(err, domerrs.ErrDailyCapExceeded) {
		t.Fatalf("expected ErrDailyCapExceeded, got %v", err)
	}
}
