package entities

import (
	"time"

	"github.com/google/uuid"
)

// SnapshotStatus represents the outcome of a profile check.
type SnapshotStatus string

const (
	// SnapshotStatusSuccess indicates the fetch succeeded and data was captured.
	SnapshotStatusSuccess SnapshotStatus = "success"

	// SnapshotStatusFailed indicates the fetch failed (provider error, timeout,
	// profile not found). The Error field contains the reason.
	SnapshotStatusFailed SnapshotStatus = "failed"
)

// SocialSnapshot records the state of a social profile at a single point in time.
type SocialSnapshot struct {
	ID        uuid.UUID
	ProfileID uuid.UUID

	// CapturedAt is when this snapshot was taken.
	CapturedAt time.Time

	// Status is the outcome of the check that produced this snapshot.
	Status SnapshotStatus

	// Error holds the error message when Status is SnapshotStatusFailed.
	Error string

	// FollowersCount is a denormalised copy of ProfileData.FollowersCount for
	// fast queries (avoids deserialising the full JSONB data column).
	FollowersCount int

	// PostsCount is a denormalised copy of ProfileData.PostsCount.
	PostsCount int

	// Data is the full captured profile state (stored as JSONB in PostgreSQL).
	// nil when Status is SnapshotStatusFailed.
	Data *ProfileData
}

// NewSuccessSnapshot constructs a snapshot with status=success.
func NewSuccessSnapshot(profileID uuid.UUID, data *ProfileData) *SocialSnapshot {
	return &SocialSnapshot{
		ID:             uuid.New(),
		ProfileID:      profileID,
		CapturedAt:     time.Now().UTC(),
		Status:         SnapshotStatusSuccess,
		FollowersCount: data.FollowersCount,
		PostsCount:     data.PostsCount,
		Data:           data,
	}
}

// NewFailedSnapshot constructs a snapshot with status=failed.
func NewFailedSnapshot(profileID uuid.UUID, errMsg string) *SocialSnapshot {
	return &SocialSnapshot{
		ID:         uuid.New(),
		ProfileID:  profileID,
		CapturedAt: time.Now().UTC(),
		Status:     SnapshotStatusFailed,
		Error:      errMsg,
	}
}
