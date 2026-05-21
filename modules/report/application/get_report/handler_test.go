package getreport_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	getreport "github.com/jcsoftdev/pulzifi-back/modules/report/application/get_report"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories/mocks"
)

func TestGetReportHandler_Handle(t *testing.T) {
	reportID := uuid.New()
	existingReport := &entities.Report{ID: reportID, Title: "Q1 Report"}

	tests := []struct {
		name    string
		setup   func(*mocks.MockReportRepository)
		wantNil bool
		wantErr bool
	}{
		{
			name: "returns report when found",
			setup: func(m *mocks.MockReportRepository) {
				m.GetByIDResult = existingReport
			},
			wantNil: false,
		},
		{
			name: "returns nil when not found",
			setup: func(m *mocks.MockReportRepository) {
				m.GetByIDResult = nil
			},
			wantNil: true,
		},
		{
			name: "propagates repo error",
			setup: func(m *mocks.MockReportRepository) {
				m.GetByIDErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockReportRepository{}
			tt.setup(repo)

			h := getreport.NewHandler(repo)
			resp, err := h.Handle(context.Background(), &getreport.Request{ReportID: reportID})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if resp.Report != nil {
					t.Errorf("expected nil report, got %v", resp.Report)
				}
				return
			}
			if resp.Report.ID != reportID {
				t.Errorf("report ID: want %v, got %v", reportID, resp.Report.ID)
			}
		})
	}
}
