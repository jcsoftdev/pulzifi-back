package listreports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	listreports "github.com/jcsoftdev/pulzifi-back/modules/report/application/list_reports"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories/mocks"
)

func TestListReportsHandler_Handle(t *testing.T) {
	pageID := uuid.New()
	report1 := &entities.Report{ID: uuid.New(), Title: "Report 1"}
	report2 := &entities.Report{ID: uuid.New(), Title: "Report 2"}

	tests := []struct {
		name    string
		req     *listreports.Request
		setup   func(*mocks.MockReportRepository)
		wantLen int
		wantErr bool
	}{
		{
			name: "lists all reports",
			req:  &listreports.Request{},
			setup: func(m *mocks.MockReportRepository) {
				m.ListResult = []*entities.Report{report1, report2}
			},
			wantLen: 2,
		},
		{
			name: "lists reports filtered by page",
			req:  &listreports.Request{PageID: &pageID},
			setup: func(m *mocks.MockReportRepository) {
				m.ListByPageResult = []*entities.Report{report1}
			},
			wantLen: 1,
		},
		{
			name: "returns empty list when none exist",
			req:  &listreports.Request{},
			setup: func(m *mocks.MockReportRepository) {
				m.ListResult = nil
			},
			wantLen: 0,
		},
		{
			name: "propagates list error",
			req:  &listreports.Request{},
			setup: func(m *mocks.MockReportRepository) {
				m.ListErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockReportRepository{}
			tt.setup(repo)

			h := listreports.NewHandler(repo)
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
			if len(resp.Reports) != tt.wantLen {
				t.Errorf("count: want %d, got %d", tt.wantLen, len(resp.Reports))
			}
			if resp.Count != tt.wantLen {
				t.Errorf("Count field: want %d, got %d", tt.wantLen, resp.Count)
			}
		})
	}
}
