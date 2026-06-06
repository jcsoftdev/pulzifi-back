package generateinsights

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services/mocks"
)

// fakeInsightRepo captures Create calls so quota tests can count persisted
// insights without touching a real database.
type fakeInsightRepo struct {
	mu      sync.Mutex
	created []*entities.Insight
	err     error
}

func (r *fakeInsightRepo) Create(_ context.Context, insight *entities.Insight) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	cp := *insight
	r.created = append(r.created, &cp)
	return nil
}

func (r *fakeInsightRepo) ListByPageID(_ context.Context, _ uuid.UUID) ([]*entities.Insight, error) {
	return nil, nil
}

func (r *fakeInsightRepo) ListByCheckID(_ context.Context, _ uuid.UUID) ([]*entities.Insight, error) {
	return nil, nil
}

func (r *fakeInsightRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Insight, error) {
	return nil, nil
}

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

			// Pass nil for db since we test paths that don't reach the repo.
			// quotaReader nil → guard disabled (legacy / generator-only tests).
			handler := NewGenerateInsightsHandler(gen, nil, nil)
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

// ── Quota guard tests ────────────────────────────────────────────────────────

func TestHandle_QuotaExhausted_ShortCircuitsBeforeLLM(t *testing.T) {
	gen := &mocks.MockInsightGenerator{
		GenerateResult: []*entities.Insight{{InsightType: "marketing"}},
	}
	repo := &fakeInsightRepo{}
	quota := &mocks.MockInsightQuotaReader{
		ReadResult: services.InsightQuotaSnapshot{Used: 5, Allowed: 5, Unlimited: false},
	}

	h := NewGenerateInsightsHandler(gen, repo, quota)
	err := h.Handle(context.Background(), &Request{
		PageID:     uuid.New(),
		CheckID:    uuid.New(),
		PageURL:    "https://example.com",
		PrevText:   "old",
		NewText:    "new",
		SchemaName: "tenant_quota",
	})

	if !errors.Is(err, ErrInsightQuotaExhausted) {
		t.Fatalf("want ErrInsightQuotaExhausted, got %v", err)
	}
	if gen.GenerateCalls != 0 {
		t.Errorf("LLM must NOT be called when quota exhausted, got %d calls", gen.GenerateCalls)
	}
	if len(repo.created) != 0 {
		t.Errorf("must NOT persist insights when quota exhausted, got %d", len(repo.created))
	}
	if quota.IncrementCalls != 0 {
		t.Errorf("Increment must NOT run when guard short-circuits, got %d calls", quota.IncrementCalls)
	}
}

func TestHandle_QuotaAllowedZero_FailsOpen(t *testing.T) {
	// A stale usage_tracking row (created before the ai_insights_allowed
	// backfill) has Allowed=0. No real plan grants 0 insights, so this must be
	// treated as "no cap" and NOT block generation.
	gen := &mocks.MockInsightGenerator{
		GenerateResult: []*entities.Insight{{InsightType: "marketing"}},
	}
	repo := &fakeInsightRepo{}
	quota := &mocks.MockInsightQuotaReader{
		ReadResult:      services.InsightQuotaSnapshot{Used: 0, Allowed: 0, Unlimited: false},
		IncrementResult: services.InsightQuotaSnapshot{Used: 1, Allowed: 0, Unlimited: false},
	}

	h := NewGenerateInsightsHandler(gen, repo, quota)
	err := h.Handle(context.Background(), &Request{
		PageID:     uuid.New(),
		CheckID:    uuid.New(),
		PageURL:    "https://example.com",
		PrevText:   "old",
		NewText:    "new",
		SchemaName: "tenant_stale_zero",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gen.GenerateCalls == 0 {
		t.Error("LLM must be called when Allowed=0 (fail-open), got 0 calls")
	}
	if len(repo.created) != 1 {
		t.Errorf("want 1 insight persisted, got %d", len(repo.created))
	}
}

func TestHandle_QuotaPartial_StopsBatchOnOverflow(t *testing.T) {
	insights := []*entities.Insight{
		{InsightType: "marketing"},
		{InsightType: "market_analysis"},
		{InsightType: "seo"},
	}
	gen := &mocks.MockInsightGenerator{GenerateResult: insights}
	repo := &fakeInsightRepo{}

	// Read says 48/50 — pre-flight passes. Each Increment returns post-value;
	// after 2 inserts we hit 50/50 → loop breaks; insight #3 dropped.
	post := []services.InsightQuotaSnapshot{
		{Used: 49, Allowed: 50},
		{Used: 50, Allowed: 50},
	}
	var incCalls int
	quota := &mocks.MockInsightQuotaReader{
		ReadResult: services.InsightQuotaSnapshot{Used: 48, Allowed: 50},
		IncrementFn: func(_ context.Context, _ string) (services.InsightQuotaSnapshot, error) {
			snap := post[incCalls]
			incCalls++
			return snap, nil
		},
	}

	h := NewGenerateInsightsHandler(gen, repo, quota)
	err := h.Handle(context.Background(), &Request{
		PageID:     uuid.New(),
		CheckID:    uuid.New(),
		PageURL:    "https://example.com",
		PrevText:   "old",
		NewText:    "new",
		SchemaName: "tenant_partial",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.created) != 2 {
		t.Errorf("want 2 insights persisted (cap hit mid-batch), got %d", len(repo.created))
	}
	if incCalls != 2 {
		t.Errorf("want 2 Increment calls, got %d", incCalls)
	}
}

func TestHandle_QuotaUnlimited_NoGuardShortCircuit(t *testing.T) {
	insights := []*entities.Insight{
		{InsightType: "marketing"},
		{InsightType: "market_analysis"},
	}
	gen := &mocks.MockInsightGenerator{GenerateResult: insights}
	repo := &fakeInsightRepo{}
	quota := &mocks.MockInsightQuotaReader{
		ReadResult:      services.InsightQuotaSnapshot{Used: 1_000_000, Allowed: 2_147_483_647, Unlimited: true},
		IncrementResult: services.InsightQuotaSnapshot{Used: 1_000_001, Allowed: 2_147_483_647, Unlimited: true},
	}

	h := NewGenerateInsightsHandler(gen, repo, quota)
	err := h.Handle(context.Background(), &Request{
		PageID:     uuid.New(),
		CheckID:    uuid.New(),
		PageURL:    "https://example.com",
		PrevText:   "old",
		NewText:    "new",
		SchemaName: "tenant_enterprise",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.created) != 2 {
		t.Errorf("unlimited plan must persist all insights, got %d", len(repo.created))
	}
	if quota.IncrementCalls != 2 {
		t.Errorf("Increment must still run for visibility under unlimited, got %d", quota.IncrementCalls)
	}
}

func TestHandle_QuotaReadError_FailsOpen(t *testing.T) {
	gen := &mocks.MockInsightGenerator{
		GenerateResult: []*entities.Insight{{InsightType: "marketing"}},
	}
	repo := &fakeInsightRepo{}
	quota := &mocks.MockInsightQuotaReader{ReadErr: errors.New("transient db error")}

	h := NewGenerateInsightsHandler(gen, repo, quota)
	err := h.Handle(context.Background(), &Request{
		PageID:     uuid.New(),
		CheckID:    uuid.New(),
		PageURL:    "https://example.com",
		PrevText:   "old",
		NewText:    "new",
		SchemaName: "tenant_failopen",
	})
	if err != nil {
		t.Fatalf("quota read error must NOT block generation, got %v", err)
	}
	if len(repo.created) != 1 {
		t.Errorf("want 1 insight persisted (fail-open), got %d", len(repo.created))
	}
}

func TestHandle_NoSchemaName_SkipsQuotaEntirely(t *testing.T) {
	gen := &mocks.MockInsightGenerator{
		GenerateResult: []*entities.Insight{{InsightType: "marketing"}},
	}
	repo := &fakeInsightRepo{}
	quota := &mocks.MockInsightQuotaReader{
		ReadResult: services.InsightQuotaSnapshot{Used: 1000, Allowed: 5}, // would exhaust if checked
	}

	h := NewGenerateInsightsHandler(gen, repo, quota)
	err := h.Handle(context.Background(), &Request{
		PageID:   uuid.New(),
		CheckID:  uuid.New(),
		PageURL:  "https://example.com",
		PrevText: "old",
		NewText:  "new",
		// SchemaName intentionally empty
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if quota.ReadCalls != 0 {
		t.Errorf("quota.Read must NOT run without SchemaName, got %d calls", quota.ReadCalls)
	}
}
