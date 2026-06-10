package services

import (
	"context"

	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// SocialFetcher is the port that retrieves live profile data from a social
// platform provider (Phase 1: Apify Instagram actor).
//
// The concrete adapter lives in infrastructure/apify/client.go. Swapping to a
// different provider (HikerAPI, official OAuth) only requires a new adapter.
type SocialFetcher interface {
	// FetchProfile retrieves the current state of a social profile.
	//
	// platform   — the social network (validated before this call).
	// handle     — the account username / @handle.
	// postsLimit — maximum number of recent posts to include (from SOCIAL_POSTS_PER_CHECK).
	//
	// Returns ErrPlatformNotSupported when no adapter exists for platform.
	// Returns ErrFetchFailed on provider error, timeout, or profile not found.
	FetchProfile(ctx context.Context, platform valueobjects.Platform, handle string, postsLimit int) (*entities.ProfileData, error)
}
