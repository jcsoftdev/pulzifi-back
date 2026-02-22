package get_monitoring_config

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories/mocks"
)

func TestGetMonitoringConfigHandler_Handle(t *testing.T) {
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
		name     string
		result   *entities.MonitoringConfig
		repoErr  error
		wantNil  bool
		wantErr  bool
	}{
		{
			name:   "config found",
			result: existingConfig,
		},
		{
			name:    "config not found",
			result:  nil,
			wantNil: true,
		},
		{
			name:    "repo error",
			repoErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockMonitoringConfigRepository{
				GetByPageIDResult: tt.result,
				GetByPageIDErr:    tt.repoErr,
			}

			handler := NewGetMonitoringConfigHandler(repo)
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

			if tt.wantNil {
				if resp != nil {
					t.Fatal("expected nil response")
				}
				return
			}

			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			if resp.ID != configID {
				t.Errorf("ID: want %v, got %v", configID, resp.ID)
			}
			if resp.CheckFrequency != "1h" {
				t.Errorf("frequency: want %q, got %q", "1h", resp.CheckFrequency)
			}
		})
	}
}
