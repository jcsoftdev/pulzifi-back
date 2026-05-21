package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// TrialPlanReader reports whether the current tenant's trial has lapsed.
// The middleware depends on this small port so unit tests can inject a fake.
type TrialPlanReader interface {
	// TrialExpired returns true when the tenant's active plan is the trial
	// plan AND trial_ends_at < now AND converted_at IS NULL.
	TrialExpired(ctx context.Context, subdomain string) (bool, error)
}

// TrialGuard returns HTTP 402 Payment Required for any write request issued by
// a tenant whose trial has expired without conversion. GET/HEAD/OPTIONS pass
// through, and writes to /billing/checkout are always allowed so the user can
// pay their way out.
type TrialGuard struct {
	reader TrialPlanReader
	// pathPrefix lists URL prefixes (after the API base) that are always
	// allowed even when the trial is expired. Default: ["/billing/"].
	allowedPrefixes []string
}

// NewTrialGuard constructs a TrialGuard middleware.
func NewTrialGuard(reader TrialPlanReader) *TrialGuard {
	return &TrialGuard{
		reader:          reader,
		allowedPrefixes: []string{"/billing/"},
	}
}

// Handler returns the chi-compatible middleware function.
func (g *TrialGuard) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always allow safe methods.
		if isReadMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// Always allow checkout — that's how users escape the gate.
		if g.isAllowedPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		subdomain := GetSubdomainFromContext(r.Context())
		if subdomain == "" {
			// No tenant resolved — let downstream middleware decide. This
			// keeps the guard from misfiring on public/auth endpoints.
			next.ServeHTTP(w, r)
			return
		}

		expired, err := g.reader.TrialExpired(r.Context(), subdomain)
		if err != nil {
			// Fail open: trial enforcement is not security-critical and we
			// must not 500 every write because the DB hiccupped.
			next.ServeHTTP(w, r)
			return
		}

		if expired {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":  "trial_expired",
				"action": "upgrade",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (g *TrialGuard) isAllowedPath(path string) bool {
	for _, p := range g.allowedPrefixes {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

func isReadMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// ── Postgres adapter ─────────────────────────────────────────────────────────

// PostgresTrialPlanReader resolves trial expiry by joining organizations
// to organization_plans to plans on the public schema.
type PostgresTrialPlanReader struct {
	db    *sql.DB
	nowFn func() time.Time
}

// NewPostgresTrialPlanReader constructs the production reader.
func NewPostgresTrialPlanReader(db *sql.DB) *PostgresTrialPlanReader {
	return &PostgresTrialPlanReader{db: db, nowFn: time.Now}
}

// TrialExpired returns true when the org's active plan is the trial plan and
// trial_ends_at < now() AND converted_at IS NULL.
func (r *PostgresTrialPlanReader) TrialExpired(ctx context.Context, subdomain string) (bool, error) {
	var planCode string
	var trialEndsAt sql.NullTime
	var convertedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT p.code, op.trial_ends_at, op.converted_at
		  FROM public.organizations o
		  JOIN public.organization_plans op
		    ON op.organization_id = o.id
		   AND op.status = 'active'
		   AND op.deleted_at IS NULL
		  JOIN public.plans p ON p.id = op.plan_id
		 WHERE o.subdomain = $1
		   AND o.deleted_at IS NULL
		 ORDER BY op.started_at DESC
		 LIMIT 1
	`, subdomain).Scan(&planCode, &trialEndsAt, &convertedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if planCode != "trial" {
		return false, nil
	}
	if convertedAt.Valid {
		return false, nil
	}
	if !trialEndsAt.Valid {
		return false, nil
	}
	return trialEndsAt.Time.Before(r.nowFn()), nil
}
