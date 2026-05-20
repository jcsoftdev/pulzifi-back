package middleware

import "net/http"

// AuthVerifier is the minimum interface that modules need from the auth middleware.
// It is satisfied by *modules/auth/infrastructure/middleware.AuthMiddleware,
// which is wired at startup via SetAuthMiddleware in cmd/server/modules.go.
type AuthVerifier interface {
	Authenticate(http.Handler) http.Handler
	RequirePermission(resource, action string) func(http.Handler) http.Handler
}

// authMiddlewareProxy is a concrete struct that wraps an AuthVerifier.
// Using a concrete struct (rather than a bare interface variable) means that
// taking a method value such as middleware.AuthMiddleware.Authenticate does not
// panic at the point of registration when the inner verifier is nil — the panic
// is deferred until Chi actually calls the middleware, which matches the
// original behavior of a typed nil concrete pointer.
type authMiddlewareProxy struct {
	inner AuthVerifier
}

// Authenticate wraps the inner verifier's Authenticate. The lookup of p.inner
// is deferred into the returned handler so that Chi can register the route
// before SetAuthMiddleware is called (e.g. in tests that only exercise
// unauthenticated routes).
func (p *authMiddlewareProxy) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.inner.Authenticate(next).ServeHTTP(w, r)
	})
}

// RequirePermission wraps the inner verifier's RequirePermission with the same
// deferred-lookup pattern.
func (p *authMiddlewareProxy) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p.inner.RequirePermission(resource, action)(next).ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware is the global auth middleware singleton.
// Consumers call middleware.AuthMiddleware.Authenticate and
// middleware.AuthMiddleware.RequirePermission directly (unchanged call sites).
var AuthMiddleware = &authMiddlewareProxy{}

var OrgMiddleware *OrganizationMiddleware

// SetAuthMiddleware wires the concrete auth middleware implementation.
// Called once at server startup from cmd/server/modules.go.
func SetAuthMiddleware(m AuthVerifier) {
	AuthMiddleware.inner = m
}

func SetOrganizationMiddleware(m *OrganizationMiddleware) {
	OrgMiddleware = m
}
