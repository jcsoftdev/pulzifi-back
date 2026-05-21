package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

func TestMockAlertCreator_Create(t *testing.T) {
	workspaceID := uuid.New()
	pageID := uuid.New()
	checkID := uuid.New()

	validInput := services.AlertInput{
		SchemaName:    "acme",
		WorkspaceID:   workspaceID,
		PageID:        pageID,
		CheckID:       checkID,
		AlertType:     "change_detected",
		Title:         "Content changed",
		Description:   "Significant changes detected",
		ChangeSummary: "Hero text updated",
	}

	tests := []struct {
		name        string
		input       services.AlertInput
		createErr   error
		useFn       bool
		wantErr     bool
		wantCalls   int
	}{
		{
			name:      "change detected — alert created",
			input:     validInput,
			wantErr:   false,
			wantCalls: 1,
		},
		{
			name:      "repo error propagated",
			input:     validInput,
			createErr: errors.New("db error"),
			wantErr:   true,
			wantCalls: 1,
		},
		{
			name:      "custom function called",
			input:     validInput,
			useFn:     true,
			wantErr:   false,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mock := &MockAlertCreator{
				CreateErr: tt.createErr,
			}
			if tt.useFn {
				mock.CreateFn = func(ctx context.Context, input services.AlertInput) error {
					return nil
				}
			}

			// Act
			err := mock.Create(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mock.CreateCalls != tt.wantCalls {
				t.Errorf("CreateCalls: want %d, got %d", tt.wantCalls, mock.CreateCalls)
			}
		})
	}
}

func TestMockAlertCreator_NoCreateWhenNotCalled(t *testing.T) {
	// Assert CreateCalls == 0 when no change detected (Create never called)
	mock := &MockAlertCreator{}

	// Simulate: no change detected, Create is NOT called
	if mock.CreateCalls != 0 {
		t.Errorf("expected 0 CreateCalls before any call, got %d", mock.CreateCalls)
	}
}
