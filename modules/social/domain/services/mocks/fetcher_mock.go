package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
)

// MockSocialFetcher is a hand-rolled mock for services.SocialFetcher.
type MockSocialFetcher struct {
	FetchProfileResult *entities.ProfileData
	FetchProfileErr    error

	FetchProfileFn func(ctx context.Context, platform valueobjects.Platform, handle string, postsLimit int) (*entities.ProfileData, error)

	FetchProfileCalls int
}

func (m *MockSocialFetcher) FetchProfile(
	ctx context.Context,
	platform valueobjects.Platform,
	handle string,
	postsLimit int,
) (*entities.ProfileData, error) {
	m.FetchProfileCalls++
	if m.FetchProfileFn != nil {
		return m.FetchProfileFn(ctx, platform, handle, postsLimit)
	}
	return m.FetchProfileResult, m.FetchProfileErr
}
