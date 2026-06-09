package memory

import (
	"context"
	"sync"
	"time"

	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
)

// CheckQuota is an in-memory implementation of services.CheckQuota.
// It provides atomic consume/compensate semantics for unit tests.
type CheckQuota struct {
	mu        sync.Mutex
	usageByDay map[string]int // key = "YYYY-MM-DD"
}

func NewCheckQuota() *CheckQuota {
	return &CheckQuota{usageByDay: make(map[string]int)}
}

func dayKey(t time.Time) string {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

func (q *CheckQuota) Consume(ctx context.Context, date time.Time, limit int) (services.QuotaResult, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	key := dayKey(date)
	used := q.usageByDay[key]
	// limit 0 = unlimited
	if limit > 0 && used >= limit {
		return services.QuotaResult{Allowed: false, Used: used, Limit: limit}, domainerrors.ErrQuotaExceeded
	}
	q.usageByDay[key] = used + 1
	return services.QuotaResult{Allowed: true, Used: used + 1, Limit: limit}, nil
}

func (q *CheckQuota) Compensate(ctx context.Context, date time.Time) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	key := dayKey(date)
	if q.usageByDay[key] > 0 {
		q.usageByDay[key]--
	}
	return nil
}

// UsedToday returns the current usage count for today — helper for tests.
func (q *CheckQuota) UsedToday() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.usageByDay[dayKey(time.Now())]
}
