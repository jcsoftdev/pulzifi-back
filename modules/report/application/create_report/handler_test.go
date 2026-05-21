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
)

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
