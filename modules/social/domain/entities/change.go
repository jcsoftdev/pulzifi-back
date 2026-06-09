package entities

import (
	"time"

	"github.com/google/uuid"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// ChangeSummary is a structured, JSONB-compatible representation of what changed
// between two consecutive snapshots. It is human-readable and machine-parseable
// so the frontend can render typed cards without re-diffing.
type ChangeSummary struct {
	// FollowersDelta is set when ChangeTypeFollowersChanged is present.
	// Positive = gained followers, negative = lost followers.
	FollowersDelta int `json:"followers_delta,omitempty"`

	// Bio holds before/after text when ChangeTypeBioChanged is present.
	Bio *TextDiff `json:"bio,omitempty"`

	// DisplayName holds before/after text when ChangeTypeDisplayNameChanged is present.
	DisplayName *TextDiff `json:"display_name,omitempty"`

	// NewPosts contains the ExternalIDs of posts added since the previous snapshot.
	NewPosts []string `json:"new_posts,omitempty"`

	// RemovedPosts contains the ExternalIDs of posts removed since the previous snapshot.
	RemovedPosts []string `json:"removed_posts,omitempty"`

	// EditedCaptions maps ExternalID → TextDiff for posts whose caption changed.
	EditedCaptions []CaptionDiff `json:"edited_captions,omitempty"`

	// AvatarChanged is true when the stored avatar URL differs.
	AvatarChanged bool `json:"avatar_changed,omitempty"`
}

// TextDiff holds the before and after text for a text-based change.
type TextDiff struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// CaptionDiff records a caption change for a specific post.
type CaptionDiff struct {
	ExternalID string `json:"external_id"`
	From       string `json:"from"`
	To         string `json:"to"`
}

// SocialChange records detected differences between two consecutive snapshots.
type SocialChange struct {
	ID             uuid.UUID
	ProfileID      uuid.UUID
	FromSnapshotID uuid.UUID
	ToSnapshotID   uuid.UUID

	// ChangeTypes lists every category of change detected by the differ.
	ChangeTypes []valueobjects.ChangeType

	// Summary is the structured breakdown used by the frontend for typed rendering.
	Summary ChangeSummary

	CreatedAt time.Time
}

// NewSocialChange constructs a SocialChange.
func NewSocialChange(
	profileID, fromSnapshotID, toSnapshotID uuid.UUID,
	changeTypes []valueobjects.ChangeType,
	summary ChangeSummary,
) *SocialChange {
	return &SocialChange{
		ID:             uuid.New(),
		ProfileID:      profileID,
		FromSnapshotID: fromSnapshotID,
		ToSnapshotID:   toSnapshotID,
		ChangeTypes:    changeTypes,
		Summary:        summary,
		CreatedAt:      time.Now().UTC(),
	}
}
