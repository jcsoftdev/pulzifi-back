package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

func TestMockVisionAnalyzer_AnalyzeChange(t *testing.T) {
	tests := []struct {
		name               string
		analyzeResult      *services.VisionChangeResult
		analyzeErr         error
		useFn              bool
		wantErr            bool
		wantHasChange      bool
		wantChangeSummary  string
	}{
		{
			name: "change detected — happy path",
			analyzeResult: &services.VisionChangeResult{
				HasMeaningfulChange: true,
				ChangeSummary:       "Hero banner text changed",
				ChangeDetails:       "The headline changed from X to Y",
			},
			wantErr:           false,
			wantHasChange:     true,
			wantChangeSummary: "Hero banner text changed",
		},
		{
			name: "no meaningful change",
			analyzeResult: &services.VisionChangeResult{
				HasMeaningfulChange: false,
				ChangeSummary:       "",
			},
			wantErr:       false,
			wantHasChange: false,
		},
		{
			name:       "AI error propagated",
			analyzeErr: errors.New("vision API timeout"),
			wantErr:    true,
		},
		{
			name:  "custom function override called",
			useFn: true,
			analyzeResult: &services.VisionChangeResult{
				HasMeaningfulChange: true,
				ChangeSummary:       "From fn",
			},
			wantErr:           false,
			wantHasChange:     true,
			wantChangeSummary: "From fn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mock := &MockVisionAnalyzer{
				AnalyzeChangeResult: tt.analyzeResult,
				AnalyzeChangeErr:    tt.analyzeErr,
			}
			if tt.useFn {
				fnResult := tt.analyzeResult
				mock.AnalyzeChangeFn = func(ctx context.Context, prev, curr, url string) (*services.VisionChangeResult, error) {
					return fnResult, nil
				}
			}

			// Act
			result, err := mock.AnalyzeChange(context.Background(), "prevBase64", "currBase64", "https://example.com")

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
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.HasMeaningfulChange != tt.wantHasChange {
				t.Errorf("HasMeaningfulChange: want %v, got %v", tt.wantHasChange, result.HasMeaningfulChange)
			}
			if tt.wantChangeSummary != "" && result.ChangeSummary != tt.wantChangeSummary {
				t.Errorf("ChangeSummary: want %q, got %q", tt.wantChangeSummary, result.ChangeSummary)
			}
			if mock.AnalyzeChangeCalls != 1 {
				t.Errorf("AnalyzeChangeCalls: want 1, got %d", mock.AnalyzeChangeCalls)
			}
		})
	}
}
