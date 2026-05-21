package trialstatus

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeReader is a deterministic test double for OrgPlanReader.
type fakeReader struct {
	snap *PlanSnapshot
	err  error
}

func (f *fakeReader) ActivePlanBySubdomain(_ context.Context, _ string) (*PlanSnapshot, error) {
	return f.snap, f.err
}

func ptr[T any](v T) *T { return &v }

func TestHandler_Handle(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	in7d := now.Add(7 * 24 * time.Hour)
	in1h := now.Add(1 * time.Hour)
	pastYesterday := now.Add(-24 * time.Hour)
	convertedAt := now.Add(-30 * 24 * time.Hour)

	tests := []struct {
		name    string
		snap    *PlanSnapshot
		err     error
		want    Response
		wantErr bool
	}{
		{
			name: "no active plan returns zero-value Response",
			snap: nil,
			want: Response{},
		},
		{
			name: "non-trial plan returns zero-value Response",
			snap: &PlanSnapshot{PlanCode: "starter"},
			want: Response{},
		},
		{
			name: "converted trial reports Converted=true and IsTrial=false",
			snap: &PlanSnapshot{
				PlanCode:    "trial",
				TrialEndsAt: &pastYesterday,
				ConvertedAt: &convertedAt,
			},
			want: Response{Converted: true},
		},
		{
			name: "active trial with 7 days remaining",
			snap: &PlanSnapshot{
				PlanCode:    "trial",
				TrialEndsAt: &in7d,
			},
			want: Response{
				IsTrial:       true,
				DaysRemaining: 7,
				TrialEndsAt:   &in7d,
			},
		},
		{
			name: "active trial with <1 day rounds up to 1",
			snap: &PlanSnapshot{
				PlanCode:    "trial",
				TrialEndsAt: &in1h,
			},
			want: Response{
				IsTrial:       true,
				DaysRemaining: 1,
				TrialEndsAt:   &in1h,
			},
		},
		{
			name: "expired and not converted reports IsExpired + NeedsUpgrade",
			snap: &PlanSnapshot{
				PlanCode:    "trial",
				TrialEndsAt: &pastYesterday,
			},
			want: Response{
				IsTrial:       true,
				IsExpired:     true,
				NeedsUpgrade:  true,
				DaysRemaining: 0,
				TrialEndsAt:   &pastYesterday,
			},
		},
		{
			name:    "reader error propagates",
			err:     errors.New("db down"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&fakeReader{snap: tt.snap, err: tt.err})
			h.nowFn = func() time.Time { return now }
			got, err := h.Handle(context.Background(), &Request{Subdomain: "acme"})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.IsTrial != tt.want.IsTrial {
				t.Errorf("IsTrial = %v, want %v", got.IsTrial, tt.want.IsTrial)
			}
			if got.IsExpired != tt.want.IsExpired {
				t.Errorf("IsExpired = %v, want %v", got.IsExpired, tt.want.IsExpired)
			}
			if got.NeedsUpgrade != tt.want.NeedsUpgrade {
				t.Errorf("NeedsUpgrade = %v, want %v", got.NeedsUpgrade, tt.want.NeedsUpgrade)
			}
			if got.Converted != tt.want.Converted {
				t.Errorf("Converted = %v, want %v", got.Converted, tt.want.Converted)
			}
			if got.DaysRemaining != tt.want.DaysRemaining {
				t.Errorf("DaysRemaining = %d, want %d", got.DaysRemaining, tt.want.DaysRemaining)
			}
		})
	}
}
