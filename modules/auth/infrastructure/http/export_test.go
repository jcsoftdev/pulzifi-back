// Package http export_test.go exposes unexported symbols for testing purposes.
// This file is compiled only during test builds and is not part of the
// production binary.
package http

// BuildTenantRedirectURLForTest exposes the unexported buildTenantRedirectURL
// function so that package http_test can verify its behaviour without needing
// a full HTTP round-trip through handleOAuthCallback.
func BuildTenantRedirectURLForTest(frontendURL, cookieDomain, subdomain string) (string, bool) {
	return buildTenantRedirectURL(frontendURL, cookieDomain, subdomain)
}
