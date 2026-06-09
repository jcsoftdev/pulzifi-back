package services

import (
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// Diff compares two consecutive ProfileData snapshots and returns the list of
// detected ChangeTypes plus a structured Summary.
//
// This is a pure function: it performs no I/O, no DB calls, no HTTP calls.
// It is safe to call multiple times with the same inputs (deterministic output).
func Diff(prev, next entities.ProfileData) ([]valueobjects.ChangeType, entities.ChangeSummary) {
	var changeTypes []valueobjects.ChangeType
	var summary entities.ChangeSummary

	// --- Bio ---
	if prev.Bio != next.Bio {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeBioChanged)
		summary.Bio = &entities.TextDiff{From: prev.Bio, To: next.Bio}
	}

	// --- Display name ---
	if prev.DisplayName != next.DisplayName {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeDisplayNameChanged)
		summary.DisplayName = &entities.TextDiff{From: prev.DisplayName, To: next.DisplayName}
	}

	// --- Avatar ---
	if prev.AvatarURL != next.AvatarURL {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeAvatarChanged)
		summary.AvatarChanged = true
	}

	// --- Followers ---
	delta := next.FollowersCount - prev.FollowersCount
	if delta != 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeFollowersChanged)
		summary.FollowersDelta = delta
	}

	// --- Posts: index previous posts by ExternalID ---
	prevByID := make(map[string]entities.Post, len(prev.Posts))
	for _, p := range prev.Posts {
		prevByID[p.ExternalID] = p
	}

	nextByID := make(map[string]entities.Post, len(next.Posts))
	for _, p := range next.Posts {
		nextByID[p.ExternalID] = p
	}

	// New posts (in next but not in prev)
	for _, p := range next.Posts {
		if _, existed := prevByID[p.ExternalID]; !existed {
			summary.NewPosts = append(summary.NewPosts, p.ExternalID)
		}
	}
	if len(summary.NewPosts) > 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeNewPost)
	}

	// Removed posts (in prev but not in next)
	for _, p := range prev.Posts {
		if _, stillExists := nextByID[p.ExternalID]; !stillExists {
			summary.RemovedPosts = append(summary.RemovedPosts, p.ExternalID)
		}
	}
	if len(summary.RemovedPosts) > 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeRemovedPost)
	}

	// Edited captions (post exists in both but caption changed)
	for _, np := range next.Posts {
		pp, existed := prevByID[np.ExternalID]
		if existed && pp.Caption != np.Caption {
			summary.EditedCaptions = append(summary.EditedCaptions, entities.CaptionDiff{
				ExternalID: np.ExternalID,
				From:       pp.Caption,
				To:         np.Caption,
			})
		}
	}
	if len(summary.EditedCaptions) > 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeCaptionEdited)
	}

	return changeTypes, summary
}
