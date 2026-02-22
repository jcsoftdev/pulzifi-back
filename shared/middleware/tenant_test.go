package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		host      string
		xTenant   string
		xFwdHost  string
		want      string
	}{
		{"X-Tenant header", "localhost:8080", "acme", "", "acme"},
		{"X-Forwarded-Host with subdomain", "localhost:8080", "", "acme.example.com", "acme"},
		{"Host with subdomain", "acme.example.com:443", "", "", "acme"},
		{"Host with two parts extracts first", "example.com", "", "", "example"},
		{"localhost no subdomain", "localhost:3000", "", "", ""},
		{"empty host", "", "", "", ""},
		{"X-Tenant takes priority", "other.example.com:80", "priority", "", "priority"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.Host = tt.host
			if tt.xTenant != "" {
				r.Header.Set("X-Tenant", tt.xTenant)
			}
			if tt.xFwdHost != "" {
				r.Header.Set("X-Forwarded-Host", tt.xFwdHost)
			}
			got := extractSubdomain(r)
			if got != tt.want {
				t.Errorf("extractSubdomain() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain string
		want      bool
	}{
		{"valid subdomain", "acme", true},
		{"empty string", "", false},
		{"localhost is invalid", "localhost", false},
		{"app is valid", "app", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSubdomain(tt.subdomain)
			if got != tt.want {
				t.Errorf("isValidSubdomain(%q) = %v, want %v", tt.subdomain, got, tt.want)
			}
		})
	}
}

func TestIsGenericDomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain string
		want      bool
	}{
		{"app", "app", true},
		{"localhost", "localhost", true},
		{"127.0.0.1", "127.0.0.1", true},
		{"acme", "acme", false},
		{"empty", "", false},
		{"custom", "myorg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isGenericDomain(tt.subdomain)
			if got != tt.want {
				t.Errorf("isGenericDomain(%q) = %v, want %v", tt.subdomain, got, tt.want)
			}
		})
	}
}

func TestRequireTenant(t *testing.T) {
	handler := RequireTenant(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("tenant present", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := context.WithValue(r.Context(), TenantContextKey, "my_schema")
		r = r.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("status: want %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("tenant missing", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status: want %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestBuildContextWithTenant(t *testing.T) {
	ctx := buildContextWithTenant(context.Background(), "acme", "acme_schema")

	subdomain := GetSubdomainFromContext(ctx)
	if subdomain != "acme" {
		t.Errorf("subdomain: want %q, got %q", "acme", subdomain)
	}

	tenant := GetTenantFromContext(ctx)
	if tenant != "acme_schema" {
		t.Errorf("tenant: want %q, got %q", "acme_schema", tenant)
	}
}
