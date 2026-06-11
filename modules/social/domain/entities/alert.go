package entities

import (
	"time"

	"github.com/google/uuid"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// Alert type discriminators stored in social_alerts.alert_type.
const (
	AlertTypeSocialChange     = "social_change"
	AlertTypeSocialSuspension = "social_suspension"
)

// SocialAlert is an in-app notification for a detected social change or a
// profile suspension. Persisted in the tenant-scoped social_alerts table.
type SocialAlert struct {
	ID          uuid.UUID
	ProfileID   uuid.UUID
	WorkspaceID uuid.UUID
	ChangeID    *uuid.UUID // nil for suspension alerts
	AlertType   string
	Title       string
	Description string
	ChangeTypes []valueobjects.ChangeType
	IsRead      bool
	CreatedAt   time.Time
}

// NewSocialChangeAlert builds an unread change alert linked to a SocialChange.
func NewSocialChangeAlert(
	profileID, workspaceID, changeID uuid.UUID,
	title, description string,
	changeTypes []valueobjects.ChangeType,
) *SocialAlert {
	cid := changeID
	return &SocialAlert{
		ID:          uuid.New(),
		ProfileID:   profileID,
		WorkspaceID: workspaceID,
		ChangeID:    &cid,
		AlertType:   AlertTypeSocialChange,
		Title:       title,
		Description: description,
		ChangeTypes: changeTypes,
		IsRead:      false,
		CreatedAt:   time.Now().UTC(),
	}
}

// NewSocialSuspensionAlert builds an unread suspension alert (no change link).
func NewSocialSuspensionAlert(
	profileID, workspaceID uuid.UUID,
	title, description string,
) *SocialAlert {
	return &SocialAlert{
		ID:          uuid.New(),
		ProfileID:   profileID,
		WorkspaceID: workspaceID,
		ChangeID:    nil,
		AlertType:   AlertTypeSocialSuspension,
		Title:       title,
		Description: description,
		ChangeTypes: nil,
		IsRead:      false,
		CreatedAt:   time.Now().UTC(),
	}
}
