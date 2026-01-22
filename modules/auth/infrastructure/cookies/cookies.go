package cookies

import (
	"net/http"
	"time"
)

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	// Default refresh token duration for cookie (30 days)
	// This should match or exceed the actual JWT refresh token expiry
	RefreshTokenDuration = 30 * 24 * time.Hour
)

// SetAuthCookies sets the access and refresh tokens as HttpOnly cookies
func SetAuthCookies(w http.ResponseWriter, accessToken string, accessExpiresIn int64, refreshToken string) {
	// Access Token Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(accessExpiresIn) * time.Second),
		MaxAge:   int(accessExpiresIn),
		HttpOnly: true,
		Secure:   true, // Ensure this is true in production
		SameSite: http.SameSiteLaxMode,
	})

	// Refresh Token Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/", // Could be restricted to /auth/refresh if desired, but "/" is more flexible for client-side checks
		Expires:  time.Now().Add(RefreshTokenDuration),
		MaxAge:   int(RefreshTokenDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookies removes the authentication cookies
func ClearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetRefreshTokenFromCookie extracts the refresh token from the cookie
func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetAccessTokenFromCookie extracts the access token from the cookie
func GetAccessTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AccessTokenCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
