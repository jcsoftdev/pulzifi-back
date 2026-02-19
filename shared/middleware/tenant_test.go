package middleware

import (
	"context"
	"strings"
	"testing"
)

func TestGetSetSearchPathSQL(t *testing.T) {
	tests := []struct {
		name       string
		tenant     string
		wantSafe   bool   // true if we expect a real SET search_path, false if we expect the no-op
		wantResult string // exact expected result (optional, only when wantSafe is true)
	}{
		{"valid simple schema", "tenant_one", true, `SET search_path TO "tenant_one", public`},
		{"valid underscore prefix", "_private", true, `SET search_path TO "_private", public`},
		{"valid alphanumeric", "org123", true, `SET search_path TO "org123", public`},
		{"valid single letter", "a", true, `SET search_path TO "a", public`},
		{"sql injection with semicolon", "tenant; DROP TABLE users;--", false, ""},
		{"sql injection with quote", `tenant"; DROP TABLE users;--`, false, ""},
		{"sql injection with single quote", "tenant' OR '1'='1", false, ""},
		{"hyphen in name", "my-tenant", false, ""},
		{"space in name", "my tenant", false, ""},
		{"dot in name", "my.tenant", false, ""},
		{"empty string", "", false, ""},
		{"starts with digit", "1tenant", false, ""},
		{"special characters", "tenant!@#$", false, ""},
		{"newline injection", "tenant\nDROP TABLE users", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSetSearchPathSQL(tt.tenant)
			if tt.wantSafe {
				if got != tt.wantResult {
					t.Errorf("GetSetSearchPathSQL(%q) = %q, want %q", tt.tenant, got, tt.wantResult)
				}
				// Verify it contains SET search_path
				if !strings.HasPrefix(got, "SET search_path TO") {
					t.Errorf("GetSetSearchPathSQL(%q) should start with SET search_path TO, got %q", tt.tenant, got)
				}
			} else {
				// Should return the safe no-op
				if got != "SELECT 1" {
					t.Errorf("GetSetSearchPathSQL(%q) = %q, want %q (safe no-op)", tt.tenant, got, "SELECT 1")
				}
			}
		})
	}
}

func TestIsPublicPath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		isPublic bool
	}{
		{"swagger path", "/api/v1/swagger", true},
		{"swagger subpath", "/api/v1/swagger/index.html", true},
		{"health check", "/api/v1/health", true},
		{"auth login", "/api/v1/auth/login", true},
		{"auth register", "/api/v1/auth/register", true},
		{"auth me", "/api/v1/auth/me", true},
		{"auth refresh", "/api/v1/auth/refresh", true},
		{"auth providers", "/api/v1/auth/providers", true},
		{"auth csrf", "/api/v1/auth/csrf", true},
		{"api docs", "/api/docs", true},
		{"root swagger", "/swagger", true},
		{"root health", "/health", true},
		{"root docs", "/docs", true},
		{"monitoring endpoint", "/api/v1/monitoring/checks", false},
		{"organizations endpoint", "/api/v1/organizations", false},
		{"random path", "/some/random/path", false},
		{"root path", "/", false},
		{"empty path", "", false},
		{"partial match", "/api/v1/heal", false},
		{"auth but different", "/api/v1/authorize", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPublicPath(tt.path)
			if got != tt.isPublic {
				t.Errorf("isPublicPath(%q) = %v, want %v", tt.path, got, tt.isPublic)
			}
		})
	}
}

func TestGetTenantFromContext(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() context.Context
		want   string
	}{
		{
			name: "tenant present in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TenantContextKey, "my_schema")
			},
			want: "my_schema",
		},
		{
			name: "tenant not in context",
			setup: func() context.Context {
				return context.Background()
			},
			want: "",
		},
		{
			name: "tenant is empty string in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TenantContextKey, "")
			},
			want: "",
		},
		{
			name: "wrong type in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TenantContextKey, 12345)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetTenantFromContext(ctx)
			if got != tt.want {
				t.Errorf("GetTenantFromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetSubdomainFromContext(t *testing.T) {
	tests := []struct {
		name  string
		setup func() context.Context
		want  string
	}{
		{
			name: "subdomain present in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), SubdomainContextKey, "acme")
			},
			want: "acme",
		},
		{
			name: "subdomain not in context",
			setup: func() context.Context {
				return context.Background()
			},
			want: "",
		},
		{
			name: "subdomain is empty string",
			setup: func() context.Context {
				return context.WithValue(context.Background(), SubdomainContextKey, "")
			},
			want: "",
		},
		{
			name: "wrong type in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), SubdomainContextKey, 999)
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got := GetSubdomainFromContext(ctx)
			if got != tt.want {
				t.Errorf("GetSubdomainFromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetTenantFromContextOrError(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() context.Context
		want    string
		wantErr bool
	}{
		{
			name: "tenant present",
			setup: func() context.Context {
				return context.WithValue(context.Background(), TenantContextKey, "my_schema")
			},
			want:    "my_schema",
			wantErr: false,
		},
		{
			name: "tenant missing",
			setup: func() context.Context {
				return context.Background()
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			got, err := GetTenantFromContextOrError(ctx)
			if tt.wantErr && err == nil {
				t.Error("GetTenantFromContextOrError() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("GetTenantFromContextOrError() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("GetTenantFromContextOrError() = %q, want %q", got, tt.want)
			}
		})
	}
}
