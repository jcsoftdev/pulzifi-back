package listchecks

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestListChecksHandler_Handle(t *testing.T) {
	pageID := uuid.New()

	checks := []*entities.Check{
		{ID: uuid.New(), PageID: pageID, Status: "success", CheckedAt: time.Now()},
		{ID: uuid.New(), PageID: pageID, Status: "error", ErrorMessage: "timeout", CheckedAt: time.Now()},
	}

	tests := []struct {
		name      string
		result    []*entities.Check
		repoErr   error
		wantErr   bool
		wantCount int
	}{
		{
			name:      "returns checks",
			result:    checks,
			wantCount: 2,
		},
		{
			name:      "empty result",
			result:    nil,
			wantCount: 0,
		},
		{
			name:    "repo error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockCheckRepository{
				ListByPageResult: tt.result,
				ListByPageErr:    tt.repoErr,
			}

			handler := NewListChecksHandler(repo)
			resp, err := handler.Handle(context.Background(), pageID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if len(resp.Checks) != tt.wantCount {
				t.Errorf("checks count: want %d, got %d", tt.wantCount, len(resp.Checks))
			}
		})
	}
}

// TestToCheckResponse_ContentDiffIsJSONObject is the regression for the Text
// Changes tab showing "No text changes detected": the stored content-diff jsonb
// must be serialized as a JSON OBJECT, not a quoted string, so the frontend can
// read content_diff.has_changes / .diffs directly.
func TestToCheckResponse_ContentDiffIsJSONObject(t *testing.T) {
	diffJSON := `{"diffs":[{"op":"changed","block":{"text":"new"},"old_block":{"text":"old"}}],"has_changes":true,"total_changes":1}`
	check := &entities.Check{
		ID:              uuid.New(),
		PageID:          uuid.New(),
		Status:          "success",
		ContentDiffJSON: diffJSON,
		CheckedAt:       time.Now(),
	}

	resp := toCheckResponse(context.Background(), check, URLMediationConfig{})

	out, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Must embed as object: "content_diff":{...}, NOT a quoted string "content_diff":"{...}".
	if !strings.Contains(string(out), `"content_diff":{`) {
		t.Errorf("content_diff not embedded as JSON object; got: %s", out)
	}
	if strings.Contains(string(out), `"content_diff":"`) {
		t.Errorf("content_diff serialized as quoted string (frontend can't parse it): %s", out)
	}

	// Round-trips back into an object exposing has_changes.
	var parsed struct {
		ContentDiff struct {
			HasChanges bool `json:"has_changes"`
		} `json:"content_diff"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("re-unmarshal failed: %v", err)
	}
	if !parsed.ContentDiff.HasChanges {
		t.Error("content_diff.has_changes = false, want true")
	}
}

// TestRawContentDiff_EmptyOrInvalid verifies empty/invalid input is omitted.
func TestRawContentDiff_EmptyOrInvalid(t *testing.T) {
	if got := rawContentDiff(""); got != nil {
		t.Errorf("empty string: want nil, got %q", got)
	}
	if got := rawContentDiff("not json"); got != nil {
		t.Errorf("invalid json: want nil, got %q", got)
	}
	if got := rawContentDiff(`{"has_changes":true}`); got == nil {
		t.Error("valid json: want non-nil, got nil")
	}
}
