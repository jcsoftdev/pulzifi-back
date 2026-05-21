// Package services_test validates the frequency resolution logic that underpins
// the Scheduler domain service. The actual frequency math lives in entities.FrequencyIntervals
// and is accessed via entities.ResolveFrequency.
package services_test

import (
	"testing"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

func TestResolveFrequency_CanonicalKeys(t *testing.T) {
	tests := []struct {
		frequency string
		want      time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"10m", 10 * time.Minute},
		{"15m", 15 * time.Minute},
		{"30m", 30 * time.Minute},
		{"1h", 1 * time.Hour},
		{"2h", 2 * time.Hour},
		{"4h", 4 * time.Hour},
		{"6h", 6 * time.Hour},
		{"12h", 12 * time.Hour},
		{"24h", 24 * time.Hour},
		{"168h", 168 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.frequency, func(t *testing.T) {
			got, ok := entities.ResolveFrequency(tt.frequency)
			if !ok {
				t.Fatalf("ResolveFrequency(%q) returned ok=false; expected it to be a known frequency", tt.frequency)
			}
			if got != tt.want {
				t.Errorf("ResolveFrequency(%q): want %v, got %v", tt.frequency, tt.want, got)
			}
		})
	}
}

func TestResolveFrequency_VerboseAliases(t *testing.T) {
	tests := []struct {
		alias string
		want  time.Duration
	}{
		{"Every 5 minutes", 5 * time.Minute},
		{"Every 30 minutes", 30 * time.Minute},
		{"Every hour", 1 * time.Hour},
		{"Every 1 hour", 1 * time.Hour},
		{"Every day", 24 * time.Hour},
		{"Every week", 168 * time.Hour},
		{"1d", 24 * time.Hour},
		{"7d", 168 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.alias, func(t *testing.T) {
			got, ok := entities.ResolveFrequency(tt.alias)
			if !ok {
				t.Fatalf("ResolveFrequency(%q) returned ok=false; expected alias to resolve", tt.alias)
			}
			if got != tt.want {
				t.Errorf("ResolveFrequency(%q): want %v, got %v", tt.alias, tt.want, got)
			}
		})
	}
}

func TestResolveFrequency_UnknownKey_ReturnsFalse(t *testing.T) {
	unknownKeys := []string{"Off", "never", "daily", "weekly", "", "random-string"}
	for _, key := range unknownKeys {
		t.Run(key, func(t *testing.T) {
			got, ok := entities.ResolveFrequency(key)
			if ok {
				t.Errorf("ResolveFrequency(%q): expected ok=false for unknown key, got duration=%v", key, got)
			}
		})
	}
}

func TestFrequencyIntervals_AllValuesPositive(t *testing.T) {
	for key, d := range entities.FrequencyIntervals {
		if d <= 0 {
			t.Errorf("FrequencyIntervals[%q] = %v; expected positive duration", key, d)
		}
	}
}
