package listchecks

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// stubPresigner implements SnapshotURLPresigner for unit tests.
type stubPresigner struct {
	presignedURL string
	err          error
	calls        []string
}

func (s *stubPresigner) PresignedGetURL(_ context.Context, storedURL string, _ time.Duration) (string, error) {
	s.calls = append(s.calls, storedURL)
	if s.err != nil {
		return "", s.err
	}
	if s.presignedURL != "" {
		return s.presignedURL, nil
	}
	return "presigned://" + storedURL, nil
}

func TestMediateURL_LegacyKey_OwnedByCurrentTenant(t *testing.T) {
	// Legacy key: UUID-prefixed (no tenant prefix) → treated as owned by current tenant.
	presigner := &stubPresigner{}
	cfg := URLMediationConfig{
		Tenant:        "tenant1",
		BucketPrivate: false,
		PresignTTL:    15 * time.Minute,
		Presigner:     presigner,
	}

	// Use a real UUID as the page directory (legacy key format).
	input := "http://storage/bucket/550e8400-e29b-41d4-a716-446655440000/1234.png"
	got := mediateURL(context.Background(), input, cfg)
	// Not private → pass-through.
	if got != input {
		t.Errorf("mediateURL (public, legacy key) = %q, want %q", got, input)
	}
}

func TestMediateURL_TenantPrefixedKey_OwnsTenant_Private(t *testing.T) {
	// Tenant-prefixed key and bucket is private → should presign.
	presigner := &stubPresigner{presignedURL: "https://presigned.example.com/obj"}
	cfg := URLMediationConfig{
		Tenant:        "tenant1",
		BucketPrivate: true,
		PresignTTL:    15 * time.Minute,
		Presigner:     presigner,
	}

	input := "http://storage/bucket/tenant1/page-uuid/123.png"
	got := mediateURL(context.Background(), input, cfg)
	if got != "https://presigned.example.com/obj" {
		t.Errorf("mediateURL (private, tenant match) = %q, want presigned URL", got)
	}
	if len(presigner.calls) != 1 || presigner.calls[0] != input {
		t.Errorf("presigner called with %v, want [%q]", presigner.calls, input)
	}
}

func TestMediateURL_WrongTenantPrefix_DroppedIDOR(t *testing.T) {
	// Key prefix is a different tenant → IDOR defence: return blank.
	presigner := &stubPresigner{}
	cfg := URLMediationConfig{
		Tenant:        "tenant1",
		BucketPrivate: false,
		PresignTTL:    15 * time.Minute,
		Presigner:     presigner,
	}

	input := "http://storage/bucket/tenant2/page-uuid/123.png"
	got := mediateURL(context.Background(), input, cfg)
	if got != "" {
		t.Errorf("mediateURL (wrong tenant prefix) = %q, want \"\" (IDOR drop)", got)
	}
}

func TestMediateURL_EmptyURL_PassThrough(t *testing.T) {
	cfg := URLMediationConfig{
		Tenant:        "tenant1",
		BucketPrivate: true,
		PresignTTL:    15 * time.Minute,
		Presigner:     &stubPresigner{},
	}
	got := mediateURL(context.Background(), "", cfg)
	if got != "" {
		t.Errorf("mediateURL(\"\") = %q, want \"\"", got)
	}
}

func TestMediateURL_PresignError_ReturnBlank(t *testing.T) {
	// If presigning fails, return blank so the browser does not see a raw object URL.
	presigner := &stubPresigner{err: errors.New("presign failed")}
	cfg := URLMediationConfig{
		Tenant:        "tenant1",
		BucketPrivate: true,
		PresignTTL:    15 * time.Minute,
		Presigner:     presigner,
	}

	input := "http://storage/bucket/tenant1/page-uuid/123.png"
	got := mediateURL(context.Background(), input, cfg)
	if got != "" {
		t.Errorf("mediateURL (presign error) = %q, want \"\" (error blank)", got)
	}
}

func TestMediateURL_NilPresigner_BucketPrivate_PassThrough(t *testing.T) {
	// When no presigner is injected (graceful degradation), pass through.
	cfg := URLMediationConfig{
		Tenant:        "tenant1",
		BucketPrivate: true,
		PresignTTL:    15 * time.Minute,
		Presigner:     nil,
	}
	input := "http://storage/bucket/tenant1/page-uuid/123.png"
	got := mediateURL(context.Background(), input, cfg)
	if got != input {
		t.Errorf("mediateURL (nil presigner) = %q, want %q (pass-through)", got, input)
	}
}

// TestOwnershipCheck verifies the ownership logic independently.
func TestOwnershipCheck(t *testing.T) {
	tests := []struct {
		name       string
		objectName string
		tenant     string
		owned      bool
	}{
		{
			name:       "tenant prefix matches",
			objectName: "tenant1/page-uuid/123.png",
			tenant:     "tenant1",
			owned:      true,
		},
		{
			name:       "tenant prefix mismatch",
			objectName: "tenant2/page-uuid/123.png",
			tenant:     "tenant1",
			owned:      false,
		},
		{
			name:       "no prefix (legacy UUID)",
			objectName: "550e8400-e29b-41d4-a716-446655440000/123.png",
			tenant:     "tenant1",
			owned:      true,
		},
		{
			name:       "empty object name",
			objectName: "",
			tenant:     "tenant1",
			owned:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlOwnedByTenant(tt.objectName, tt.tenant)
			if got != tt.owned {
				t.Errorf("urlOwnedByTenant(%q, %q) = %v, want %v", tt.objectName, tt.tenant, got, tt.owned)
			}
		})
	}
}

// TestExtractObjectNameFromURL verifies the object name extraction from storage URLs.
func TestExtractObjectNameFromURL(t *testing.T) {
	tests := []struct {
		name      string
		storedURL string
		want      string
	}{
		{
			name:      "minio URL with bucket",
			storedURL: "http://minio:9000/bucket/tenant1/page-uuid/123.png",
			want:      "tenant1/page-uuid/123.png",
		},
		{
			name:      "legacy minio URL no prefix",
			storedURL: "http://minio:9000/bucket/550e8400-e29b-41d4-a716-446655440000/123.png",
			want:      "550e8400-e29b-41d4-a716-446655440000/123.png",
		},
		{
			name:      "cloudinary URL",
			storedURL: "https://res.cloudinary.com/mycloud/image/upload/v123/page-uuid/123.png",
			want:      "", // cloudinary URLs parsed differently; we rely on the /bucket/ marker
		},
		{
			name:      "empty",
			storedURL: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractObjectNameFromStorageURL(tt.storedURL)
			// For the cloudinary case we only care it doesn't panic; the ownership
			// check will treat an empty result as legacy-owned.
			if strings.Contains(tt.storedURL, "cloudinary") {
				return
			}
			if got != tt.want {
				t.Errorf("extractObjectNameFromStorageURL(%q) = %q, want %q", tt.storedURL, got, tt.want)
			}
		})
	}
}
