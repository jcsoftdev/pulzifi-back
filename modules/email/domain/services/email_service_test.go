package services

import (
	"context"
	"strings"
	"testing"

	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/email/domain/errors"
)

func TestValidateEmail(t *testing.T) {
	svc := NewEmailService()
	ctx := context.Background()

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid simple email", "user@example.com", false},
		{"valid with dots", "first.last@example.com", false},
		{"valid with plus", "user+tag@example.com", false},
		{"valid with subdomain", "user@mail.example.com", false},
		{"valid with percent", "user%name@example.com", false},
		{"empty string", "", true},
		{"missing at sign", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing local part", "@example.com", true},
		{"missing tld", "user@example", true},
		{"double at sign", "user@@example.com", true},
		{"spaces in email", "user @example.com", true},
		{"single char tld", "user@example.c", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateEmail(ctx, tt.email)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateEmail(%q) expected error, got nil", tt.email)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateEmail(%q) unexpected error: %v", tt.email, err)
			}
			if err != nil {
				var invalidErr *domainerrors.InvalidEmailError
				if ok := isInvalidEmailError(err, &invalidErr); !ok {
					t.Errorf("ValidateEmail(%q) error type = %T, want *InvalidEmailError", tt.email, err)
				}
			}
		})
	}
}

func TestValidateEmailContent(t *testing.T) {
	svc := NewEmailService()
	ctx := context.Background()

	longSubject := strings.Repeat("a", 256)
	longBody := strings.Repeat("b", 10001)
	validSubject := "Test Subject"
	validBody := "Test Body"

	tests := []struct {
		name       string
		subject    string
		body       string
		wantErr    bool
		wantErrMsg string
	}{
		{"valid content", validSubject, validBody, false, ""},
		{"empty subject", "", validBody, true, "subject cannot be empty"},
		{"empty body", validSubject, "", true, "body cannot be empty"},
		{"both empty", "", "", true, "subject cannot be empty"},
		{"subject too long", longSubject, validBody, true, "subject too long"},
		{"body too long", validSubject, longBody, true, "body too long"},
		{"subject at max length", strings.Repeat("a", 255), validBody, false, ""},
		{"body at max length", validSubject, strings.Repeat("b", 10000), false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateEmailContent(ctx, tt.subject, tt.body)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateEmailContent(%q, %q) expected error, got nil", tt.subject, truncate(tt.body))
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateEmailContent(%q, %q) unexpected error: %v", tt.subject, truncate(tt.body), err)
			}
			if tt.wantErr && err != nil && tt.wantErrMsg != "" {
				var invalidErr *domainerrors.InvalidEmailError
				if ok := isInvalidEmailError(err, &invalidErr); !ok {
					t.Errorf("error type = %T, want *InvalidEmailError", err)
				} else if !strings.Contains(invalidErr.Message, tt.wantErrMsg) {
					t.Errorf("error message = %q, want it to contain %q", invalidErr.Message, tt.wantErrMsg)
				}
			}
		})
	}
}

// isInvalidEmailError checks if the error is an *InvalidEmailError and assigns it.
func isInvalidEmailError(err error, target **domainerrors.InvalidEmailError) bool {
	e, ok := err.(*domainerrors.InvalidEmailError)
	if ok {
		*target = e
	}
	return ok
}

func truncate(s string) string {
	if len(s) > 20 {
		return s[:20] + "..."
	}
	return s
}
