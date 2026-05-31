package cloudinary

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestClient_ObjectNameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		folder   string
		url      string
		expected string
	}{
		{
			name:     "image upload with folder and version",
			folder:   "pulzifi",
			url:      "https://res.cloudinary.com/mycloud/image/upload/v1234567890/pulzifi/page-id/123.png",
			expected: "page-id/123.png",
		},
		{
			name:     "raw upload with folder no version",
			folder:   "pulzifi",
			url:      "https://res.cloudinary.com/mycloud/raw/upload/pulzifi/page-id/snapshot.html",
			expected: "page-id/snapshot.html",
		},
		{
			name:     "no folder configured",
			folder:   "",
			url:      "https://res.cloudinary.com/mycloud/image/upload/v9999/page-id/photo.jpg",
			expected: "page-id/photo.jpg",
		},
		{
			name:     "no upload marker in URL",
			folder:   "pulzifi",
			url:      "https://res.cloudinary.com/mycloud/image/fetch/page-id/123.png",
			expected: "",
		},
		{
			name:     "empty URL",
			folder:   "pulzifi",
			url:      "",
			expected: "",
		},
		{
			name:     "non-cloudinary URL returns empty",
			folder:   "pulzifi",
			url:      "http://localhost:9000/bucket/page-id/123.png",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{folder: tt.folder}
			got := c.ObjectNameFromURL(tt.url)
			if got != tt.expected {
				t.Errorf("ObjectNameFromURL(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

// TestClient_PresignedGetURL_PublicPassThrough verifies that a non-authenticated
// Cloudinary URL (no "/authenticated/" segment) is returned unchanged.
func TestClient_PresignedGetURL_PublicPassThrough(t *testing.T) {
	c := &Client{
		cloudName: "mycloud",
		apiKey:    "key",
		apiSecret: "secret",
		folder:    "pulzifi",
	}

	storedURL := "https://res.cloudinary.com/mycloud/image/upload/v1234/pulzifi/page-id/123.png"
	got, err := c.PresignedGetURL(context.Background(), storedURL, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != storedURL {
		t.Errorf("public pass-through = %q, want %q", got, storedURL)
	}
}

// TestClient_PresignedGetURL_EmptyURL verifies an empty URL is returned as-is.
func TestClient_PresignedGetURL_EmptyURL(t *testing.T) {
	c := &Client{cloudName: "mycloud", apiSecret: "secret"}
	got, err := c.PresignedGetURL(context.Background(), "", 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("empty URL should return empty, got %q", got)
	}
}

// TestClient_PresignedGetURL_AuthenticatedURLShape verifies that an authenticated
// Cloudinary URL (contains "/authenticated/") produces a signed URL with the
// expected shape: contains the cloud name, a timestamp, and a signature.
func TestClient_PresignedGetURL_AuthenticatedURLShape(t *testing.T) {
	c := &Client{
		cloudName: "mycloud",
		apiKey:    "apikey123",
		apiSecret: "apisecret456",
		folder:    "pulzifi",
	}

	// Simulate an authenticated delivery URL (private asset).
	authenticatedURL := "https://res.cloudinary.com/mycloud/image/authenticated/v1234/pulzifi/page-id/123.png"
	got, err := c.PresignedGetURL(context.Background(), authenticatedURL, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected non-empty presigned URL")
	}
	// The signed URL must contain the base delivery endpoint.
	if !strings.Contains(got, "mycloud") {
		t.Errorf("signed URL does not contain cloud name: %q", got)
	}
	// Must contain a timestamp query param.
	if !strings.Contains(got, "timestamp=") {
		t.Errorf("signed URL missing timestamp param: %q", got)
	}
	// Must contain a signature query param.
	if !strings.Contains(got, "signature=") {
		t.Errorf("signed URL missing signature param: %q", got)
	}
}

// TestClient_sign verifies the SHA-1 HMAC signature helper produces a
// deterministic, non-empty hex string for a fixed input.
func TestClient_sign(t *testing.T) {
	c := &Client{apiSecret: "mysecret"}
	params := map[string]string{
		"public_id": "test/page-id/123",
		"timestamp": "1700000000",
	}
	got := c.sign(params)
	if got == "" {
		t.Error("sign returned empty string")
	}
	// Deterministic — calling twice gives same result.
	if got2 := c.sign(params); got != got2 {
		t.Errorf("sign non-deterministic: %q vs %q", got, got2)
	}
}
