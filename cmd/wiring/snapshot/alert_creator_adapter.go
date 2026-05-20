package snapshotwiring

import (
	"context"
	"database/sql"

	alertentities "github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
	alertPersistence "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/persistence"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// alertCreatorAdapter implements snapshot's AlertCreator port by wrapping
// alert/infrastructure/persistence.
type alertCreatorAdapter struct {
	db *sql.DB
}

// NewAlertCreator builds an AlertCreator adapter backed by the shared DB pool.
func NewAlertCreator(db *sql.DB) snapServices.AlertCreator {
	return &alertCreatorAdapter{db: db}
}

func (a *alertCreatorAdapter) Create(ctx context.Context, input snapServices.AlertInput) error {
	alert := alertentities.NewAlert(
		input.WorkspaceID,
		input.PageID,
		input.CheckID,
		input.AlertType,
		input.Title,
		input.Description,
	)
	alert.ChangeSummary = input.ChangeSummary

	repo := alertPersistence.NewAlertPostgresRepository(a.db, input.SchemaName)
	return repo.Create(ctx, alert)
}
