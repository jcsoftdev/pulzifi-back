// Package testguard protects integration tests from running against shared
// or production databases picked up from ambient environment variables.
package testguard

import (
	"net/url"
	"os"
	"testing"
)

// localHosts are hosts integration tests may write to without an explicit override.
// RequireLocalDB skips the test unless the DSN host is local (localhost,
// 127.0.0.1, ::1, postgres, host.docker.internal) or INTEGRATION_DB_ALLOW_REMOTE=1.
var localHosts = map[string]struct{}{
	"localhost":            {},
	"127.0.0.1":            {},
	"::1":                  {},
	"postgres":             {},
	"host.docker.internal": {},
}

// RequireLocalDB skips the test unless dsn points at a local/dev database or
// the developer explicitly opts in via INTEGRATION_DB_ALLOW_REMOTE=1. Call
// this immediately after resolving the DSN and before any sql.Open/Ping.
func RequireLocalDB(t *testing.T, dsn string) {
	t.Helper()
	if isLocalDSN(dsn) {
		return
	}
	if os.Getenv("INTEGRATION_DB_ALLOW_REMOTE") == "1" {
		return
	}
	t.Skip("refusing to run destructive integration test against non-local database; set INTEGRATION_DB_ALLOW_REMOTE=1 to override")
}

// isLocalDSN reports whether dsn points at a database on a well-known local
// host. It fails closed: unparseable DSNs or empty hostnames are treated as
// non-local.
func isLocalDSN(dsn string) bool {
	u, err := url.Parse(dsn)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "" {
		return false
	}
	_, ok := localHosts[host]
	return ok
}
