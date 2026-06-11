package createreport_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	createreport "github.com/jcsoftdev/pulzifi-back/modules/report/application/create_report"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/services"
)

// stubInsightReader is a hand-rolled InsightReader for tests.
type stubInsightReader struct {
	result []services.InsightSummary
	err    error
}

func (s stubInsightReader) ListByPage(_ context.Context, _ uuid.UUID) ([]services.InsightSummary, error) {
	return s.result, s.err
}

// stubSummarizer records its input and returns a canned summary.
type stubSummarizer struct {
	out      string
	err      error
	gotCount int
}

func (s *stubSummarizer) Summarize(_ context.Context, _ string, insights []services.InsightSummary) (string, error) {
	s.gotCount = len(insights)
	return s.out, s.err
}

func TestCreateReportHandler_Handle(t *testing.T) {
	pageID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name    string
		req     *createreport.Request
		setup   func(*mocks.MockReportRepository)
		wantErr bool
	}{
		{
			name: "creates report successfully",
			req: &createreport.Request{
				PageID:     pageID,
				Title:      "Monthly Report",
				ReportDate: time.Now(),
				Content:    entities.Content{"key": "value"},
				CreatedBy:  createdBy,
			},
			setup: func(m *mocks.MockReportRepository) {
				// CreateErr is nil by default — success
			},
		},
		{
			name: "propagates repo error",
			req: &createreport.Request{
				PageID:    pageID,
				Title:     "Monthly Report",
				CreatedBy: createdBy,
			},
			setup: func(m *mocks.MockReportRepository) {
				m.CreateErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockReportRepository{}
			tt.setup(repo)

			h := createreport.NewHandler(repo)
			resp, err := h.Handle(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Report == nil {
				t.Fatal("expected report in response, got nil")
			}
			if resp.Report.Title != tt.req.Title {
				t.Errorf("title: want %q, got %q", tt.req.Title, resp.Report.Title)
			}
			if resp.Report.ID == uuid.Nil {
				t.Error("report ID should be set")
			}
		})
	}
}

func TestCreateReportHandler_WithAI(t *testing.T) {
	pageID := uuid.New()
	baseReq := func() *createreport.Request {
		return &createreport.Request{
			PageID:    pageID,
			Title:     "Monthly Report",
			Content:   entities.Content{},
			CreatedBy: uuid.New(),
		}
	}

	t.Run("summarizes insights into content", func(t *testing.T) {
		repo := &mocks.MockReportRepository{}
		reader := stubInsightReader{result: []services.InsightSummary{
			{Type: "overview", Title: "A", Content: "first"},
			{Type: "pricing", Title: "B", Content: "second"},
		}}
		summarizer := &stubSummarizer{out: "Overview text.\n\nRecommended actions:\n- Do X"}

		h := createreport.NewHandlerWithAI(repo, reader, summarizer)
		resp, err := h.Handle(context.Background(), baseReq())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := resp.Report.Content["summary"]; got != summarizer.out {
			t.Errorf("summary: want %q, got %q", summarizer.out, got)
		}
		if got := resp.Report.Content["insights_count"]; got != 2 {
			t.Errorf("insights_count: want 2, got %v", got)
		}
		if summarizer.gotCount != 2 {
			t.Errorf("summarizer received %d insights, want 2", summarizer.gotCount)
		}
	})

	t.Run("placeholder when no insights, skips LLM", func(t *testing.T) {
		repo := &mocks.MockReportRepository{}
		reader := stubInsightReader{result: nil}
		summarizer := &stubSummarizer{out: "should not be used"}

		h := createreport.NewHandlerWithAI(repo, reader, summarizer)
		resp, err := h.Handle(context.Background(), baseReq())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summarizer.gotCount != 0 {
			t.Error("summarizer should not be called when there are no insights")
		}
		summary, _ := resp.Report.Content["summary"].(string)
		if summary == "" || summary == summarizer.out {
			t.Errorf("expected placeholder summary, got %q", summary)
		}
		if got := resp.Report.Content["insights_count"]; got != 0 {
			t.Errorf("insights_count: want 0, got %v", got)
		}
	})

	t.Run("propagates summarizer error without creating report", func(t *testing.T) {
		repo := &mocks.MockReportRepository{}
		reader := stubInsightReader{result: []services.InsightSummary{{Title: "A"}}}
		summarizer := &stubSummarizer{err: errors.New("llm down")}

		h := createreport.NewHandlerWithAI(repo, reader, summarizer)
		if _, err := h.Handle(context.Background(), baseReq()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("propagates insight reader error", func(t *testing.T) {
		repo := &mocks.MockReportRepository{}
		reader := stubInsightReader{err: errors.New("db down")}
		summarizer := &stubSummarizer{}

		h := createreport.NewHandlerWithAI(repo, reader, summarizer)
		if _, err := h.Handle(context.Background(), baseReq()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
