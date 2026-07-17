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

// fakeScheduler records dispatch calls so tests can assert whether changing a
// frequency triggered an immediate check.
type fakeScheduler struct {
	wakeUpCalls  int
	triggerCalls int
	triggerErr   error
}

func (f *fakeScheduler) WakeUp() { f.wakeUpCalls++ }

func (f *fakeScheduler) TriggerPageCheck(_ context.Context, _ string, _ uuid.UUID) error {
	f.triggerCalls++
	return f.triggerErr
}

func (f *fakeScheduler) Start(_ context.Context) {}

func TestUpdateMonitoringConfigHandler_ImmediateDispatch(t *testing.T) {
	pageID := uuid.New()
	now := time.Now()
	recent := now.Add(-1 * time.Minute)
	old := now.Add(-2 * time.Hour)

	cfg := func(freq string) *entities.MonitoringConfig {
		return &entities.MonitoringConfig{
			ID:             uuid.New(),
			PageID:         pageID,
			CheckFrequency: freq,
			ScheduleType:   "all_time",
			Timezone:       "UTC",
		}
	}

	tests := []struct {
		name        string
		existingCfg *entities.MonitoringConfig
		lastChecked *time.Time
		newFreq     string
		wantTrigger bool
	}{
		{
			name:        "re-enable from Off triggers immediately even when recently checked",
			existingCfg: cfg("Off"),
			lastChecked: &recent,
			newFreq:     "1h",
			wantTrigger: true,
		},
		{
			name:        "active to active recently checked does not trigger",
			existingCfg: cfg("1h"),
			lastChecked: &recent,
			newFreq:     "30m",
			wantTrigger: false,
		},
		{
			name:        "active to active overdue still triggers",
			existingCfg: cfg("24h"),
			lastChecked: &old,
			newFreq:     "1h",
			wantTrigger: true,
		},
		{
			name:        "switching active to Off never triggers",
			existingCfg: cfg("1h"),
			lastChecked: &old,
			newFreq:     "Off",
			wantTrigger: false,
		},
		{
			name:        "brand new active config triggers immediately",
			existingCfg: nil,
			lastChecked: nil,
			newFreq:     "1h",
			wantTrigger: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockMonitoringConfigRepository{
				GetByPageIDResult:      tt.existingCfg,
				GetLastCheckedAtResult: tt.lastChecked,
			}
			sched := &fakeScheduler{}

			handler := NewUpdateMonitoringConfigHandler(repo, nil, "test_tenant", sched)
			_, err := handler.Handle(context.Background(), pageID, &UpdateMonitoringConfigRequest{
				CheckFrequency: strPtr(tt.newFreq),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := sched.triggerCalls > 0
			if got != tt.wantTrigger {
				t.Errorf("immediate trigger: want %v, got %v (triggerCalls=%d)", tt.wantTrigger, got, sched.triggerCalls)
			}
		})
	}
}

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
