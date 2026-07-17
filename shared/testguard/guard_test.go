package testguard

import (
	"os"
	"testing"
)

func TestIsLocalDSN(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
		want bool
	}{
		{
			name: "localhost",
			dsn:  "postgres://user:pass@localhost:5432/db?sslmode=disable",
			want: true,
		},
		{
			name: "loopback ipv4",
			dsn:  "postgres://user:pass@127.0.0.1:5432/db",
			want: true,
		},
		{
			name: "docker host alias",
			dsn:  "postgres://user:pass@host.docker.internal:5432/db",
			want: true,
		},
		{
			name: "compose service name postgres",
			dsn:  "postgres://user:pass@postgres:5432/db",
			want: true,
		},
		{
			name: "remote host",
			dsn:  "postgres://user:pass@prod-db.example.com:5432/db",
			want: false,
		},
		{
			name: "unparseable dsn",
			dsn:  "postgres://user:pass@[::1/db", // malformed bracket, fails url.Parse
			want: false,
		},
		{
			name: "empty string",
			dsn:  "",
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isLocalDSN(tc.dsn)
			if got != tc.want {
				t.Errorf("isLocalDSN(%q) = %v, want %v", tc.dsn, got, tc.want)
			}
		})
	}
}

func TestIsLocalDSN_LoopbackIPv6(t *testing.T) {
	// ::1 must be bracketed in a URL authority to parse correctly.
	got := isLocalDSN("postgres://user:pass@[::1]:5432/db")
	if !got {
		t.Errorf("isLocalDSN(bracketed ::1) = false, want true")
	}
}

func TestRequireLocalDB_SkipsForRemoteDSNWithoutOverride(t *testing.T) {
	os.Unsetenv("INTEGRATION_DB_ALLOW_REMOTE")

	var subtestRef *testing.T
	t.Run("remote", func(st *testing.T) {
		subtestRef = st
		RequireLocalDB(st, "postgres://user:pass@prod-db.example.com:5432/db")
		st.Error("expected RequireLocalDB to skip before reaching this line")
	})
	if subtestRef == nil || !subtestRef.Skipped() {
		t.Error("expected subtest to be marked skipped by RequireLocalDB")
	}
}

func TestRequireLocalDB_AllowsRemoteWithOverride(t *testing.T) {
	t.Setenv("INTEGRATION_DB_ALLOW_REMOTE", "1")

	ran := false
	t.Run("remote-with-override", func(t *testing.T) {
		RequireLocalDB(t, "postgres://user:pass@prod-db.example.com:5432/db")
		ran = true
	})
	if !ran {
		t.Error("expected RequireLocalDB to allow test to proceed when override is set")
	}
}
