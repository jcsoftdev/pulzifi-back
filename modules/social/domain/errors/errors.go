package errors

import "errors"

// Domain errors for the social module.
// HTTP adapters map these to the appropriate status codes.

// ErrProfileNotFound is returned when a social profile lookup produces no result.
var ErrProfileNotFound = errors.New("social profile not found")

// ErrQuotaExceeded is returned when the daily check quota for the organisation
// has been consumed. run_check must NOT call the external fetcher in this case.
var ErrQuotaExceeded = errors.New("daily social check quota exceeded")

// ErrPlatformNotSupported is returned by the SocialFetcher when it has no
// adapter for the requested platform. Phase 1: only Instagram is supported
// end-to-end; tiktok/facebook pass schema validation but fail here.
var ErrPlatformNotSupported = errors.New("social platform not supported")

// ErrProfileLimitReached is returned when the organisation's plan cap on
// social_profiles_limit would be exceeded by creating a new profile.
var ErrProfileLimitReached = errors.New("social profile limit for current plan reached")

// ErrProfileAlreadyExists is returned when (workspace_id, platform, handle)
// is not unique within the tenant schema.
var ErrProfileAlreadyExists = errors.New("social profile already exists for this platform and handle")

// ErrFetchFailed is returned when the external provider (Apify) returns an
// error, is unreachable, or times out. The caller should persist a failed
// snapshot and apply exponential backoff.
var ErrFetchFailed = errors.New("social profile fetch failed")
