package entities

import "time"

// ProfileData holds provider-agnostic structured data captured from a social
// profile at a single point in time. It is embedded in SocialSnapshot.Data
// (serialised as JSONB) and is the input to differ.Diff.
type ProfileData struct {
	// Bio is the profile biography / description text.
	Bio string `json:"bio"`

	// DisplayName is the human-readable name shown on the profile.
	DisplayName string `json:"display_name"`

	// AvatarURL is the stored (MinIO/Cloudinary) URL of the profile picture.
	// CDN URLs expire; the infrastructure layer downloads and re-uploads them.
	AvatarURL string `json:"avatar_url"`

	// FollowersCount is the number of followers at capture time.
	FollowersCount int `json:"followers_count"`

	// FollowingCount is the number of accounts the profile follows.
	FollowingCount int `json:"following_count"`

	// PostsCount is the total number of posts reported by the platform.
	PostsCount int `json:"posts_count"`

	// Posts contains the most recent posts (capped by SOCIAL_POSTS_PER_CHECK).
	Posts []Post `json:"posts"`
}

// Post represents a single post captured from a social profile.
type Post struct {
	// ExternalID is the platform-native post identifier (used for deduplication).
	ExternalID string `json:"external_id"`

	// Caption is the post text / description.
	Caption string `json:"caption"`

	// MediaURL is the CDN URL of the post thumbnail (expires; do not persist).
	MediaURL string `json:"media_url"`

	// StoredMediaURL is the object-storage URL (MinIO/Cloudinary) written by
	// MediaStore.Store. Persisted alongside the snapshot.
	StoredMediaURL string `json:"stored_media_url"`

	// PostedAt is the time the post was published on the platform.
	PostedAt time.Time `json:"posted_at"`

	// LikesCount is the number of likes at capture time.
	LikesCount int `json:"likes_count"`

	// CommentsCount is the number of comments at capture time.
	CommentsCount int `json:"comments_count"`
}
