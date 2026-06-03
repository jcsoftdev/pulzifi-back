package minio

import (
	"context"
	"testing"
	"time"
)

func TestClient_ObjectNameFromURL(t *testing.T) {
	c := &Client{bucketName: "snapshots"}

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "standard minio URL",
			url:      "http://localhost:4566/snapshots/page-id/123.png",
			expected: "page-id/123.png",
		},
		{
			name:     "deep path",
			url:      "http://minio:9000/snapshots/tenants/abc/checks/456/diff.png",
			expected: "tenants/abc/checks/456/diff.png",
		},
		{
			name:     "tenant-prefixed key",
			url:      "http://localhost:4566/snapshots/tenant1/page-id/123.png",
			expected: "tenant1/page-id/123.png",
		},
		{
			name:     "wrong bucket",
			url:      "http://localhost:4566/other-bucket/page-id/123.png",
			expected: "",
		},
		{
			name:     "no path at all",
			url:      "http://localhost:4566",
			expected: "",
		},
		{
			name:     "empty string",
			url:      "",
			expected: "",
		},
		{
			name:     "bucket present but no object after it",
			url:      "http://localhost:4566/snapshots/",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.ObjectNameFromURL(tt.url)
			if got != tt.expected {
				t.Errorf("ObjectNameFromURL(%q) = %q, want %q", tt.url, got, tt.expected)
			}
		})
	}
}

// TestClient_PresignedGetURL_PassThrough verifies that when the client is
// configured as non-private (bucketPrivate=false) PresignedGetURL returns
// the stored URL unchanged (pass-through behaviour).
func TestClient_PresignedGetURL_PassThrough(t *testing.T) {
	c := &Client{
		bucketName:    "snapshots",
		publicURL:     "http://localhost:4566",
		bucketPrivate: false,
		// presignClient intentionally nil — pass-through must not call it.
	}

	storedURL := "http://localhost:4566/snapshots/page-id/123.png"
	got, err := c.PresignedGetURL(context.Background(), storedURL, 15*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != storedURL {
		t.Errorf("PresignedGetURL pass-through = %q, want %q", got, storedURL)
	}
}

// TestClient_PresignedGetURL_EmptyURL verifies an empty storedURL is returned
// unchanged regardless of privacy mode.
func TestClient_PresignedGetURL_EmptyURL(t *testing.T) {
	for _, priv := range []bool{false, true} {
		c := &Client{bucketName: "snapshots", bucketPrivate: priv}
		got, err := c.PresignedGetURL(context.Background(), "", 15*time.Minute)
		if err != nil {
			t.Fatalf("priv=%v: unexpected error: %v", priv, err)
		}
		if got != "" {
			t.Errorf("priv=%v: PresignedGetURL(\"\") = %q, want \"\"", priv, got)
		}
	}
}
