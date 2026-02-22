package updatemonitoringconfig

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func strPtr(s string) *string { return &s }

func TestUpdateMonitoringConfigHandler_Handle(t *testing.T) {
	pageID := uuid.New()
	configID := uuid.New()

	existingConfig := &entities.MonitoringConfig{
		ID:                     configID,
		PageID:                 pageID,
		CheckFrequency:         "1h",
		ScheduleType:           "all_time",
		Timezone:               "UTC",
		BlockAdsCookies:        true,
		EnabledInsightTypes:    []string{"marketing"},
		EnabledAlertConditions: []string{"any_changes"},
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	tests := []struct {
		name          string
		req           *UpdateMonitoringConfigRequest
		existingCfg   *entities.MonitoringConfig
		repoErr       error
		updateErr     error
		wantErr       bool
		wantFrequency string
	}{
		{
			name:          "update existing config frequency",
			req:           &UpdateMonitoringConfigRequest{CheckFrequency: strPtr("30m")},
			existingCfg:   existingConfig,
			wantErr:       false,
			wantFrequency: "30m",
		},
		{
			name:          "create new config when not found",
			req:           &UpdateMonitoringConfigRequest{CheckFrequency: strPtr("2h")},
			existingCfg:   nil,
			wantErr:       false,
			wantFrequency: "2h",
		},
		{
			name:        "repo get error",
			req:         &UpdateMonitoringConfigRequest{CheckFrequency: strPtr("1h")},
			repoErr:     errors.New("db error"),
			wantErr:     true,
		},
		{
			name:        "repo update error",
			req:         &UpdateMonitoringConfigRequest{CheckFrequency: strPtr("1h")},
			existingCfg: existingConfig,
			updateErr:   errors.New("update failed"),
			wantErr:     true,
		},
		{
			name:          "normalize verbose frequency",
			req:           &UpdateMonitoringConfigRequest{CheckFrequency: strPtr("every 30 minutes")},
			existingCfg:   existingConfig,
			wantErr:       false,
			wantFrequency: "30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockMonitoringConfigRepository{
				GetByPageIDResult: tt.existingCfg,
				GetByPageIDErr:    tt.repoErr,
				UpdateErr:         tt.updateErr,
			}

			handler := NewUpdateMonitoringConfigHandler(repo, nil, "test_tenant", nil)
			resp, err := handler.Handle(context.Background(), pageID, tt.req)

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

			if tt.wantFrequency != "" && resp.CheckFrequency != tt.wantFrequency {
				t.Errorf("frequency: want %q, got %q", tt.wantFrequency, resp.CheckFrequency)
			}
		})
	}
}
