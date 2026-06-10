package services

import "context"

// MediaStore is the port for persisting social media assets (avatar images,
// post thumbnails) to object storage (MinIO or Cloudinary).
//
// Instagram CDN URLs expire quickly. The infrastructure layer downloads the
// source URL and re-uploads it under a stable key in object storage, returning
// a durable URL that is stored in ProfileData.Post.StoredMediaURL.
type MediaStore interface {
	// Store downloads the asset at sourceURL and uploads it to object storage
	// under the given key (e.g. "social/{profileID}/posts/{externalID}").
	//
	// Returns the stored (durable) URL on success.
	// Returns an error when the download or upload fails.
	Store(ctx context.Context, sourceURL, key string) (storedURL string, err error)
}
