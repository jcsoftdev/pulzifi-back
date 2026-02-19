package repositories

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/dashboard/domain/entities"
)

type DashboardRepository interface {
	GetStats(ctx context.Context) (*entities.DashboardStats, error)
}
