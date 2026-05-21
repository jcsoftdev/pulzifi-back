package services

import "time"

// BillingPeriodCalculator computes billing period boundaries.
// It is a pure-time service with no I/O.
type BillingPeriodCalculator interface {
	// For returns the period_start and period_end for today,
	// anchored to anchorDay (day-of-month when the plan started).
	// Handles months with fewer days by clamping to end-of-month.
	For(today time.Time, anchorDay int) (start, end time.Time)
}

// New returns the default BillingPeriodCalculator implementation.
func New() BillingPeriodCalculator { return calculator{} }

type calculator struct{}

// For returns the billing period that contains today.
// E.g. anchorDay=15, today=Mar 20 → period Mar 15 – Apr 14.
//
//	anchorDay=15, today=Mar 10 → period Feb 15 – Mar 14.
func (calculator) For(today time.Time, anchorDay int) (start, end time.Time) {
	year, month, day := today.Date()

	// Clamp anchor to last day of current month
	lastDay := daysInMonth(year, month)
	clampedAnchor := anchorDay
	if clampedAnchor > lastDay {
		clampedAnchor = lastDay
	}

	if day >= clampedAnchor {
		// We're in the period that started this month
		start = time.Date(year, month, clampedAnchor, 0, 0, 0, 0, time.UTC)
	} else {
		// We're in the period that started last month
		prevMonth := month - 1
		prevYear := year
		if prevMonth < 1 {
			prevMonth = 12
			prevYear--
		}
		prevLastDay := daysInMonth(prevYear, prevMonth)
		prevAnchor := anchorDay
		if prevAnchor > prevLastDay {
			prevAnchor = prevLastDay
		}
		start = time.Date(prevYear, prevMonth, prevAnchor, 0, 0, 0, 0, time.UTC)
	}

	// End is one day before the next period starts
	nextMonth := start.Month() + 1
	nextYear := start.Year()
	if nextMonth > 12 {
		nextMonth = 1
		nextYear++
	}
	nextLastDay := daysInMonth(nextYear, nextMonth)
	nextAnchor := anchorDay
	if nextAnchor > nextLastDay {
		nextAnchor = nextLastDay
	}
	end = time.Date(nextYear, nextMonth, nextAnchor, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)

	return start, end
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
