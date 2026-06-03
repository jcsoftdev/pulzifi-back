package listchecks

import (
	"context"
	"strings"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// SnapshotURLPresigner is the port for generating time-limited presigned URLs.
// Satisfied by both MinIO and Cloudinary clients, and by nil (graceful degradation).
type SnapshotURLPresigner interface {
	PresignedGetURL(ctx context.Context, storedURL string, ttl time.Duration) (string, error)
}

// URLMediationConfig carries the context needed for URL ownership checking and
// optional presigning. All zero values yield a safe pass-through behaviour.
type URLMediationConfig struct {
	// Tenant is the current request's tenant schema name (e.g. "acme").
	Tenant string
	// BucketPrivate mirrors SNAPSHOT_BUCKET_PRIVATE: when true reads are routed
	// through presigned URLs instead of direct object URLs.
	BucketPrivate bool
	// PresignTTL is the validity window for presigned URLs (default 15 min).
	PresignTTL time.Duration
	// Presigner is the ObjectStorage implementation that mints presigned URLs.
	// When nil, the function falls back to pass-through (graceful degradation).
	Presigner SnapshotURLPresigner
}

// mediateURL applies the read-mediation logic to a single stored URL:
//  1. Empty URL → return ""  (no-op)
//  2. Extract object name from stored URL to determine prefix.
//  3. Ownership check: if the object has a tenant prefix that does NOT match
//     cfg.Tenant, return "" and log (IDOR defense). No-prefix keys are treated
//     as legacy and owned by the current tenant.
//  4. If BucketPrivate is true and a Presigner is available, replace the URL
//     with a presigned URL. On presign error, return "" (fail-closed).
//  5. Otherwise return the stored URL unchanged.
func mediateURL(ctx context.Context, storedURL string, cfg URLMediationConfig) string {
	if storedURL == "" {
		return ""
	}

	// Extract the object name from the stored URL to check ownership.
	objectName := extractObjectNameFromStorageURL(storedURL)

	// Ownership check — Cloudinary URLs may not yield an object name via the
	// /bucket/ marker; treat those as legacy (pass-through). This is safe because
	// Cloudinary public IDs are non-guessable and ACL is managed at the Cloudinary
	// level (authenticated type) rather than by key prefix.
	if objectName != "" && !urlOwnedByTenant(objectName, cfg.Tenant) {
		logger.Warn("URL ownership mismatch — dropping URL from response (IDOR defense)",
			zap.String("tenant", cfg.Tenant),
			zap.String("url_prefix", extractTenantPrefix(objectName)))
		return ""
	}

	// Presigning: only when the bucket is private.
	if !cfg.BucketPrivate || cfg.Presigner == nil {
		return storedURL
	}

	ttl := cfg.PresignTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	presigned, err := cfg.Presigner.PresignedGetURL(ctx, storedURL, ttl)
	if err != nil {
		logger.Error("Failed to presign URL — dropping from response",
			zap.String("url", storedURL), zap.Error(err))
		return ""
	}
	return presigned
}

// MediateURLExported is the exported version of mediateURL for use by the HTTP
// layer within the same module (infrastructure/http). This keeps the core logic
// in the application package while allowing the HTTP handlers to reuse it.
func MediateURLExported(ctx context.Context, storedURL string, cfg URLMediationConfig) string {
	return mediateURL(ctx, storedURL, cfg)
}

// urlOwnedByTenant returns true when objectName is either:
//   - empty (legacy / parse failure → assume current tenant owns it)
//   - has no "/" (flat key, no prefix → legacy-owned)
//   - the first path segment matches tenant exactly
func urlOwnedByTenant(objectName, tenant string) bool {
	if objectName == "" {
		return true
	}
	idx := strings.Index(objectName, "/")
	if idx < 0 {
		// Flat key — no prefix → legacy-owned.
		return true
	}
	firstSegment := objectName[:idx]
	// If the first segment looks like a UUID, it is a legacy key (no tenant prefix).
	if isUUIDShape(firstSegment) {
		return true
	}
	return firstSegment == tenant
}

// extractObjectNameFromStorageURL extracts the storage object name (key) from a
// stored URL by looking for a common /bucket-name/ marker pattern. Only works
// for MinIO / S3-style URLs; returns "" for Cloudinary URLs.
// This is intentionally a best-effort parser — it need not handle all URL shapes.
func extractObjectNameFromStorageURL(storedURL string) string {
	if storedURL == "" {
		return ""
	}
	// MinIO/S3 URLs have the form: http[s]://<host>/<bucket>/<key...>
	// We look for any path segment followed by the key by splitting on the third "/".
	// E.g. "http://minio:9000/snapshots/tenant1/page-uuid/123.png" → "tenant1/page-uuid/123.png"
	stripped := strings.TrimPrefix(storedURL, "https://")
	stripped = strings.TrimPrefix(stripped, "http://")

	// Find the first "/" (end of host), then the second "/" (end of bucket-name).
	hostEnd := strings.Index(stripped, "/")
	if hostEnd < 0 {
		return ""
	}

	// Cloudinary URLs (res.cloudinary.com/<cloud>/<resource_type>/<delivery>/...)
	// do not follow the host/bucket/key shape: parsing them naively yields a bogus
	// key prefixed with "image" or "raw", which would trip the tenant-ownership
	// check and drop legitimately-owned URLs. Cloudinary ACL is enforced at the
	// Cloudinary level (authenticated type, non-guessable public IDs), so we
	// return "" here and let the caller treat it as legacy pass-through.
	if host := stripped[:hostEnd]; strings.Contains(host, "cloudinary.com") {
		return ""
	}
	rest := stripped[hostEnd+1:] // "snapshots/tenant1/page-uuid/123.png"

	bucketEnd := strings.Index(rest, "/")
	if bucketEnd < 0 {
		return ""
	}
	key := rest[bucketEnd+1:] // "tenant1/page-uuid/123.png"

	// Strip query string.
	if qIdx := strings.Index(key, "?"); qIdx >= 0 {
		key = key[:qIdx]
	}

	return key
}

// extractTenantPrefix returns the first path segment of an object name.
func extractTenantPrefix(objectName string) string {
	idx := strings.Index(objectName, "/")
	if idx < 0 {
		return objectName
	}
	return objectName[:idx]
}

// isUUIDShape reports whether s looks like a UUID (contains hyphens and is ~36 chars).
// Used to distinguish legacy UUID-prefixed keys from tenant-name-prefixed keys.
func isUUIDShape(s string) bool {
	if len(s) < 32 {
		return false
	}
	return strings.Count(s, "-") >= 4
}
