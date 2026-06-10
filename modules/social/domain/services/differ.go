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

	changeTypes = diffScalarFields(prev, next, changeTypes, &summary)
	changeTypes = diffPosts(prev.Posts, next.Posts, changeTypes, &summary)

	return changeTypes, summary
}

// diffScalarFields checks bio, display name, avatar, and follower count changes.
func diffScalarFields(
	prev, next entities.ProfileData,
	changeTypes []valueobjects.ChangeType,
	summary *entities.ChangeSummary,
) []valueobjects.ChangeType {
	if prev.Bio != next.Bio {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeBioChanged)
		summary.Bio = &entities.TextDiff{From: prev.Bio, To: next.Bio}
	}
	if prev.DisplayName != next.DisplayName {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeDisplayNameChanged)
		summary.DisplayName = &entities.TextDiff{From: prev.DisplayName, To: next.DisplayName}
	}
	// Avatar URL comparison is intentionally skipped: Instagram CDN URLs contain
	// expiring tokens that change on every Apify call regardless of whether the
	// image actually changed. Reliable avatar-change detection requires downloading
	// and hashing the image, which is not yet implemented.
	_ = prev.AvatarURL
	_ = next.AvatarURL
	if delta := next.FollowersCount - prev.FollowersCount; delta != 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeFollowersChanged)
		summary.FollowersDelta = delta
	}
	return changeTypes
}

// diffPosts detects new, removed, and caption-edited posts.
func diffPosts(
	prevPosts, nextPosts []entities.Post,
	changeTypes []valueobjects.ChangeType,
	summary *entities.ChangeSummary,
) []valueobjects.ChangeType {
	prevByID := indexPostsByID(prevPosts)
	nextByID := indexPostsByID(nextPosts)

	// New posts (in next but not in prev)
	for _, p := range nextPosts {
		if _, existed := prevByID[p.ExternalID]; !existed {
			summary.NewPosts = append(summary.NewPosts, p.ExternalID)
		}
	}
	if len(summary.NewPosts) > 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeNewPost)
	}

	// Removed posts (in prev but not in next)
	for _, p := range prevPosts {
		if _, stillExists := nextByID[p.ExternalID]; !stillExists {
			summary.RemovedPosts = append(summary.RemovedPosts, p.ExternalID)
		}
	}
	if len(summary.RemovedPosts) > 0 {
		changeTypes = append(changeTypes, valueobjects.ChangeTypeRemovedPost)
	}

	// Edited captions (post exists in both but caption changed)
	for _, np := range nextPosts {
		if pp, existed := prevByID[np.ExternalID]; existed && pp.Caption != np.Caption {
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

	return changeTypes
}

// indexPostsByID builds a lookup map from ExternalID → Post.
func indexPostsByID(posts []entities.Post) map[string]entities.Post {
	m := make(map[string]entities.Post, len(posts))
	for _, p := range posts {
		m[p.ExternalID] = p
	}
	return m
}
