package cookies

import (
	"net/http"
	"time"
)

const (
	SessionIDCookie = "session_id"
)

// SetSessionCookie sets the opaque session identifier as HttpOnly cookie
func SetSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time, domain string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDCookie,
		Value:    sessionID,
		Path:     "/",
		Domain:   domain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie removes the authentication cookie
func ClearSessionCookie(w http.ResponseWriter, domain string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDCookie,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetSessionIDFromCookie extracts the session identifier from the cookie
func GetSessionIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionIDCookie)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
