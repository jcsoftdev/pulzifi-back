package services

import (
	"strings"
	"testing"

	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
)

func TestValidateOrganizationName(t *testing.T) {
	svc := NewOrganizationService()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid name", "Acme Corp", false, ""},
		{"valid short name 2 chars", "AB", false, ""},
		{"valid long name", strings.Repeat("a", 255), false, ""},
		{"empty name", "", true, "name cannot be empty"},
		{"whitespace only", "   ", true, "name cannot be empty"},
		{"too short 1 char", "A", true, "name must be at least 2 characters"},
		{"too long 256 chars", strings.Repeat("a", 256), true, "name must be at most 255 characters"},
		{"name with spaces trimmed to valid", "  AB  ", false, ""},
		{"name with spaces trimmed to too short", "  A  ", true, "name must be at least 2 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateOrganizationName(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateOrganizationName(%q) expected error, got nil", tt.input)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateOrganizationName(%q) unexpected error: %v", tt.input, err)
				return
			}
			if tt.wantErr && err != nil {
				var invalidErr *domainerrors.InvalidOrganizationDataError
				if e, ok := err.(*domainerrors.InvalidOrganizationDataError); !ok {
					t.Errorf("error type = %T, want *InvalidOrganizationDataError", err)
				} else {
					invalidErr = e
					if !strings.Contains(invalidErr.Message, tt.errMsg) {
						t.Errorf("error message = %q, want it to contain %q", invalidErr.Message, tt.errMsg)
					}
				}
			}
		})
	}
}

func TestValidateSubdomain(t *testing.T) {
	svc := NewOrganizationService()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{"valid subdomain", "acme", false, ""},
		{"valid with hyphens", "my-company", false, ""},
		{"valid with numbers", "company123", false, ""},
		{"valid 3 chars min", "abc", false, ""},
		{"valid 63 chars max", strings.Repeat("a", 63), false, ""},
		{"empty subdomain", "", true, "subdomain cannot be empty"},
		{"whitespace only", "   ", true, "subdomain cannot be empty"},
		{"too short 1 char", "a", true, "subdomain must be at least 3 characters"},
		{"too short 2 chars", "ab", true, "subdomain must be at least 3 characters"},
		{"too long 64 chars", strings.Repeat("a", 64), true, "subdomain must be at most 63 characters"},
		{"starts with hyphen", "-acme", true, "cannot start or end with a hyphen"},
		{"ends with hyphen", "acme-", true, "cannot start or end with a hyphen"},
		{"starts and ends with hyphen", "-acme-", true, "cannot start or end with a hyphen"},
		{"special characters", "acme!corp", true, "must contain only alphanumeric"},
		{"spaces", "acme corp", true, "must contain only alphanumeric"},
		{"uppercase converted", "ACME", false, ""},
		{"underscores not allowed", "my_company", true, "must contain only alphanumeric"},
		{"dots not allowed", "my.company", true, "must contain only alphanumeric"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateSubdomain(tt.input)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateSubdomain(%q) expected error, got nil", tt.input)
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateSubdomain(%q) unexpected error: %v", tt.input, err)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				var invalidErr *domainerrors.InvalidOrganizationDataError
				if e, ok := err.(*domainerrors.InvalidOrganizationDataError); !ok {
					t.Errorf("error type = %T, want *InvalidOrganizationDataError", err)
				} else {
					invalidErr = e
					if !strings.Contains(invalidErr.Message, tt.errMsg) {
						t.Errorf("error message = %q, want it to contain %q", invalidErr.Message, tt.errMsg)
					}
				}
			}
		})
	}
}

func TestGenerateSchemaName(t *testing.T) {
	svc := NewOrganizationService()

	tests := []struct {
		name      string
		subdomain string
		want      string
	}{
		{"simple subdomain", "acme", "acme"},
		{"subdomain with hyphens", "my-company", "my_company"},
		{"subdomain with multiple hyphens", "my-big-company", "my_big_company"},
		{"uppercase converted to lower", "ACME", "acme"},
		{"starts with digit gets prefix", "123corp", "s_123corp"},
		{"digit only", "999", "s_999"},
		{"mixed case with hyphens", "My-Company", "my_company"},
		{"no hyphens no change", "company", "company"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.GenerateSchemaName(tt.subdomain)
			if got != tt.want {
				t.Errorf("GenerateSchemaName(%q) = %q, want %q", tt.subdomain, got, tt.want)
			}
		})
	}
}
