package generateinsights

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services/mocks"
)

func TestGenerateInsightsHandler_Handle(t *testing.T) {
	pageID := uuid.New()
	checkID := uuid.New()

	tests := []struct {
		name       string
		req        *Request
		setupMock  func(gen *mocks.MockInsightGenerator)
		wantErr    bool
	}{
		{
			name: "generator error",
			req: &Request{
				PageID:     pageID,
				CheckID:    checkID,
				PageURL:    "https://example.com",
				PrevText:   "old content",
				NewText:    "new content",
				SchemaName: "tenant_test",
			},
			setupMock: func(gen *mocks.MockInsightGenerator) {
				gen.GenerateErr = errors.New("AI service unavailable")
			},
			wantErr: true,
		},
		{
			name: "empty insights returns nil",
			req: &Request{
				PageID:     pageID,
				CheckID:    checkID,
				PageURL:    "https://example.com",
				PrevText:   "same content",
				NewText:    "same content",
				SchemaName: "tenant_test",
			},
			setupMock: func(gen *mocks.MockInsightGenerator) {
				gen.GenerateResult = []*entities.Insight{}
			},
			wantErr: false,
		},
		{
			name: "defaults insight types when empty",
			req: &Request{
				PageID:     pageID,
				CheckID:    checkID,
				PageURL:    "https://example.com",
				PrevText:   "old",
				NewText:    "new",
				SchemaName: "tenant_test",
			},
			setupMock: func(gen *mocks.MockInsightGenerator) {
				gen.GenerateFn = func(_ context.Context, _, _, _ string, enabledTypes []string) ([]*entities.Insight, error) {
					if len(enabledTypes) != 2 {
						return nil, errors.New("expected 2 default types")
					}
					if enabledTypes[0] != "marketing" || enabledTypes[1] != "market_analysis" {
						return nil, errors.New("unexpected default types")
					}
					return []*entities.Insight{}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "uses provided insight types",
			req: &Request{
				PageID:              pageID,
				CheckID:             checkID,
				PageURL:             "https://example.com",
				PrevText:            "old",
				NewText:             "new",
				SchemaName:          "tenant_test",
				EnabledInsightTypes: []string{"seo"},
			},
			setupMock: func(gen *mocks.MockInsightGenerator) {
				gen.GenerateFn = func(_ context.Context, _, _, _ string, enabledTypes []string) ([]*entities.Insight, error) {
					if len(enabledTypes) != 1 || enabledTypes[0] != "seo" {
						return nil, errors.New("expected [seo] types")
					}
					return []*entities.Insight{}, nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &mocks.MockInsightGenerator{}
			if tt.setupMock != nil {
				tt.setupMock(gen)
			}

			// Pass nil for db since we test paths that don't reach the repo
			handler := NewGenerateInsightsHandler(gen, nil)
			err := handler.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
