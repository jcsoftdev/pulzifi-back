package ai

import (
	"strings"
	"testing"
)

func TestBuildOrderedTypes(t *testing.T) {
	tests := []struct {
		name     string
		enabled  []string
		wantFirst string
		wantLen  int
	}{
		{
			name:      "overview always first when not in input",
			enabled:   []string{"marketing", "market_analysis"},
			wantFirst: "overview",
			wantLen:   3,
		},
		{
			name:      "overview not duplicated when already in input",
			enabled:   []string{"overview", "marketing"},
			wantFirst: "overview",
			wantLen:   2,
		},
		{
			name:      "single enabled type — overview prepended",
			enabled:   []string{"marketing"},
			wantFirst: "overview",
			wantLen:   2,
		},
		{
			name:      "empty input — only overview",
			enabled:   []string{},
			wantFirst: "overview",
			wantLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildOrderedTypes(tt.enabled)

			if len(result) == 0 {
				t.Fatal("result must not be empty")
			}
			if result[0] != tt.wantFirst {
				t.Errorf("first type: want %q, got %q", tt.wantFirst, result[0])
			}
			if len(result) != tt.wantLen {
				t.Errorf("len: want %d, got %d (got %v)", tt.wantLen, len(result), result)
			}
			// No duplicates
			seen := map[string]bool{}
			for _, r := range result {
				if seen[r] {
					t.Errorf("duplicate type %q in result %v", r, result)
				}
				seen[r] = true
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{
			name:  "short string not truncated",
			input: "hello",
			max:   10,
			want:  "hello",
		},
		{
			name:  "exact max length not truncated",
			input: "hello",
			max:   5,
			want:  "hello",
		},
		{
			name:  "over max — truncated with suffix",
			input: "hello world",
			max:   5,
			want:  "hello... [truncated]",
		},
		{
			name:  "empty string not truncated",
			input: "",
			max:   10,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

func TestParseInsights_ValidJSON(t *testing.T) {
	raw := `[{"insight_type":"marketing","title":"Marketing Lens","content":"Some analysis"}]`

	payloads, err := parseInsights(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(payloads))
	}
	if payloads[0].InsightType != "marketing" {
		t.Errorf("InsightType: want %q, got %q", "marketing", payloads[0].InsightType)
	}
	if payloads[0].Title != "Marketing Lens" {
		t.Errorf("Title: want %q, got %q", "Marketing Lens", payloads[0].Title)
	}
}

func TestParseInsights_JSONWrappedInMarkdown(t *testing.T) {
	// LLM output often wraps JSON in markdown code fences
	raw := "```json\n[{\"insight_type\":\"overview\",\"title\":\"Overview\",\"content\":\"Content here\"}]\n```"

	payloads, err := parseInsights(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloads) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(payloads))
	}
	if payloads[0].InsightType != "overview" {
		t.Errorf("InsightType: want %q, got %q", "overview", payloads[0].InsightType)
	}
}

func TestParseInsights_InvalidJSON(t *testing.T) {
	raw := "not valid json at all"

	_, err := parseInsights(raw)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestInsightTypeDescriptionsComplete(t *testing.T) {
	// Ensure all known types have descriptions and titles
	knownTypes := []string{"overview", "marketing", "market_analysis", "business_opportunities", "job_recommendation"}

	for _, typ := range knownTypes {
		if desc, ok := insightTypeDescriptions[typ]; !ok || strings.TrimSpace(desc) == "" {
			t.Errorf("insightTypeDescriptions missing or empty for type %q", typ)
		}
		if title, ok := insightTitles[typ]; !ok || strings.TrimSpace(title) == "" {
			t.Errorf("insightTitles missing or empty for type %q", typ)
		}
	}
}
