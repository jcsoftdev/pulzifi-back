package create_monitoring_config

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestCreateMonitoringConfigHandler_Handle(t *testing.T) {
	pageID := uuid.New()

	tests := []struct {
		name    string
		req     *CreateMonitoringConfigRequest
		repoErr error
		wantErr bool
	}{
		{
			name: "successful creation",
			req: &CreateMonitoringConfigRequest{
				PageID:          pageID,
				CheckFrequency:  "1h",
				ScheduleType:    "all_time",
				Timezone:        "UTC",
				BlockAdsCookies: true,
			},
		},
		{
			name: "repo error",
			req: &CreateMonitoringConfigRequest{
				PageID:         pageID,
				CheckFrequency: "1h",
				ScheduleType:   "all_time",
				Timezone:       "UTC",
			},
			repoErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "frequency Off does not wake scheduler",
			req: &CreateMonitoringConfigRequest{
				PageID:         pageID,
				CheckFrequency: "Off",
				ScheduleType:   "all_time",
				Timezone:       "UTC",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockMonitoringConfigRepository{
				CreateErr: tt.repoErr,
			}

			handler := NewCreateMonitoringConfigHandler(repo, nil)
			resp, err := handler.Handle(context.Background(), tt.req)

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
			if resp.PageID != pageID {
				t.Errorf("page_id: want %v, got %v", pageID, resp.PageID)
			}
			if resp.CheckFrequency != tt.req.CheckFrequency {
				t.Errorf("frequency: want %q, got %q", tt.req.CheckFrequency, resp.CheckFrequency)
			}
			if repo.CreateCalls != 1 {
				t.Errorf("expected 1 Create call, got %d", repo.CreateCalls)
			}
		})
	}
}
