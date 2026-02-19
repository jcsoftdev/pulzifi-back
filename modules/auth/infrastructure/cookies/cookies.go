package cookies

import (
	"net"
	"net/http"
	"time"
)

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

// resolveDomain returns the cookie domain to use. When a static domain is
// configured it is returned as-is. Otherwise the hostname is extracted from
// the request's Host header. For IP addresses and localhost the empty string
// is returned so the browser scopes the cookie to the exact origin.
func resolveDomain(r *http.Request, staticDomain string) string {
	if staticDomain != "" {
		return staticDomain
	}

	host := r.Host
	// Strip port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	// Never set a Domain attribute for IP addresses or localhost — the
	// browser will scope the cookie to the exact origin automatically.
	if host == "localhost" || net.ParseIP(host) != nil {
		return ""
	}

	return host
}

func setCookie(w http.ResponseWriter, name, value string, expires time.Time, domain string, secure bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
	if domain != "" {
		cookie.Domain = domain
	}
	http.SetCookie(w, cookie)
}

func SetAccessTokenCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time, staticDomain string, secure bool) {
	domain := resolveDomain(r, staticDomain)
	setCookie(w, AccessTokenCookie, token, expires, domain, secure)
}

func SetRefreshTokenCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time, staticDomain string, secure bool) {
	domain := resolveDomain(r, staticDomain)
	setCookie(w, RefreshTokenCookie, token, expires, domain, secure)
}

func ClearAuthCookies(w http.ResponseWriter, r *http.Request, staticDomain string, secure bool) {
	domain := resolveDomain(r, staticDomain)
	expired := time.Unix(0, 0)
	for _, name := range []string{AccessTokenCookie, RefreshTokenCookie} {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			Expires:  expired,
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		}
		if domain != "" {
			cookie.Domain = domain
		}
		http.SetCookie(w, cookie)
	}
}

func GetTokenFromCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
