package getsnapshot

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

type Response struct {
	ID             uuid.UUID             `json:"id"`
	ProfileID      uuid.UUID             `json:"profile_id"`
	CapturedAt     time.Time             `json:"captured_at"`
	Status         string                `json:"status"`
	Error          string                `json:"error,omitempty"`
	FollowersCount int                   `json:"followers_count"`
	PostsCount     int                   `json:"posts_count"`
	Data           *entities.ProfileData `json:"data,omitempty"`
}

func snapshotToResponse(s *entities.SocialSnapshot) *Response {
	return &Response{
		ID:             s.ID,
		ProfileID:      s.ProfileID,
		CapturedAt:     s.CapturedAt,
		Status:         string(s.Status),
		Error:          s.Error,
		FollowersCount: s.FollowersCount,
		PostsCount:     s.PostsCount,
		Data:           s.Data,
	}
}
