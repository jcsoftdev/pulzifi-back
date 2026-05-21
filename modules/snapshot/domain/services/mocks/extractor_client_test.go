package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

func TestMockExtractorClient_Extract(t *testing.T) {
	tests := []struct {
		name          string
		extractResult *services.ExtractorResult
		extractErr    error
		useFn         bool
		wantErr       bool
		wantCalls     int
	}{
		{
			name: "returns configured result",
			extractResult: &services.ExtractorResult{
				Title:            "Page Title",
				HTML:             "<html></html>",
				Text:             "Page content",
				ScreenshotBase64: "base64data",
				SelectorMatched:  true,
			},
			wantErr:   false,
			wantCalls: 1,
		},
		{
			name:      "propagates configured error",
			extractErr: errors.New("extractor timeout"),
			wantErr:   true,
			wantCalls: 1,
		},
		{
			name:  "uses custom function when set",
			useFn: true,
			extractResult: &services.ExtractorResult{
				Title: "From Fn",
			},
			wantErr:   false,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mock := &MockExtractorClient{
				ExtractResult: tt.extractResult,
				ExtractErr:    tt.extractErr,
			}
			if tt.useFn {
				fnResult := tt.extractResult
				mock.ExtractFn = func(ctx context.Context, url string, opts services.ExtractOptions) (*services.ExtractorResult, error) {
					return fnResult, nil
				}
			}

			// Act
			result, err := mock.Extract(context.Background(), "https://example.com", services.ExtractOptions{})

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
			if mock.ExtractCalls != tt.wantCalls {
				t.Errorf("ExtractCalls: want %d, got %d", tt.wantCalls, mock.ExtractCalls)
			}
		})
	}
}

func TestMockExtractorClient_CallCountIncrementsPerCall(t *testing.T) {
	mock := &MockExtractorClient{
		ExtractResult: &services.ExtractorResult{Title: "T"},
	}

	for i := 1; i <= 3; i++ {
		_, _ = mock.Extract(context.Background(), "url", services.ExtractOptions{})
		if mock.ExtractCalls != i {
			t.Errorf("after call %d: ExtractCalls = %d", i, mock.ExtractCalls)
		}
	}
}
