package services

import (
	"context"
	"strings"
	"testing"
)

func TestValidateEmailContent_AdditionalCases(t *testing.T) {
	svc := NewEmailService()
	ctx := context.Background()

	tests := []struct {
		name    string
		subject string
		body    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid subject and body",
			subject: "Welcome to Pulzifi",
			body:    "Hello, your account is ready.",
			wantErr: false,
		},
		{
			name:    "empty subject",
			subject: "",
			body:    "Some body content",
			wantErr: true,
			errMsg:  "subject cannot be empty",
		},
		{
			name:    "empty body",
			subject: "A valid subject",
			body:    "",
			wantErr: true,
			errMsg:  "body cannot be empty",
		},
		{
			name:    "subject exactly 255 chars — valid",
			subject: strings.Repeat("a", 255),
			body:    "body content",
			wantErr: false,
		},
		{
			name:    "subject 256 chars — too long",
			subject: strings.Repeat("a", 256),
			body:    "body content",
			wantErr: true,
			errMsg:  "subject too long",
		},
		{
			name:    "body exactly 10000 chars — valid",
			subject: "subject",
			body:    strings.Repeat("x", 10000),
			wantErr: false,
		},
		{
			name:    "body 10001 chars — too long",
			subject: "subject",
			body:    strings.Repeat("x", 10001),
			wantErr: true,
			errMsg:  "body too long",
		},
		{
			name:    "both empty",
			subject: "",
			body:    "",
			wantErr: true,
			errMsg:  "subject cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := svc.ValidateEmailContent(ctx, tt.subject, tt.body)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
