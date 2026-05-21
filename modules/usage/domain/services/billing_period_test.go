package services_test

import (
	"testing"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
)

func TestBillingPeriodCalculator_For(t *testing.T) {
	calc := services.New()

	tests := []struct {
		name      string
		today     time.Time
		anchorDay int
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "today is on anchor day — period starts today",
			today:     date(2025, 3, 15),
			anchorDay: 15,
			wantStart: date(2025, 3, 15),
			wantEnd:   date(2025, 4, 14),
		},
		{
			name:      "today is after anchor — period started this month",
			today:     date(2025, 3, 20),
			anchorDay: 15,
			wantStart: date(2025, 3, 15),
			wantEnd:   date(2025, 4, 14),
		},
		{
			name:      "today is before anchor — period started last month",
			today:     date(2025, 3, 10),
			anchorDay: 15,
			wantStart: date(2025, 2, 15),
			wantEnd:   date(2025, 3, 14),
		},
		{
			name:      "anchor day = 31 in February — clamps to 28",
			today:     date(2025, 2, 15),
			anchorDay: 31,
			wantStart: date(2025, 1, 31),
			wantEnd:   date(2025, 2, 27),
		},
		{
			name:      "anchor day = 31, today before clamped anchor in March",
			today:     date(2025, 3, 5),
			anchorDay: 31,
			wantStart: date(2025, 2, 28),
			wantEnd:   date(2025, 3, 30),
		},
		{
			name:      "leap year — February 29 anchor clamped from 31",
			today:     date(2024, 2, 20),
			anchorDay: 31,
			wantStart: date(2024, 1, 31),
			wantEnd:   date(2024, 2, 28),
		},
		{
			name:      "year boundary — December to January",
			today:     date(2025, 1, 5),
			anchorDay: 10,
			wantStart: date(2024, 12, 10),
			wantEnd:   date(2025, 1, 9),
		},
		{
			name:      "anchor day 1 — calendar-month aligned",
			today:     date(2025, 6, 30),
			anchorDay: 1,
			wantStart: date(2025, 6, 1),
			wantEnd:   date(2025, 6, 30),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := calc.For(tt.today, tt.anchorDay)
			if !start.Equal(tt.wantStart) {
				t.Errorf("start: want %v, got %v", tt.wantStart, start)
			}
			if !end.Equal(tt.wantEnd) {
				t.Errorf("end: want %v, got %v", tt.wantEnd, end)
			}
		})
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
