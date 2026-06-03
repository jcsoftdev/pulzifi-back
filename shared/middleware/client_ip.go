package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP extracts the real client IP from r, respecting a list of trusted
// proxy CIDRs.
//
// Algorithm:
//  1. If trusted is empty (nil or zero-length), return the host of r.RemoteAddr
//     directly — no headers are trusted.
//  2. Otherwise, if r.RemoteAddr's IP is itself inside a trusted CIDR, honour
//     X-Real-IP if present (and no X-Forwarded-For header exists).
//  3. Parse X-Forwarded-For as a comma-separated list and walk it
//     RIGHT-TO-LEFT: skip addresses that fall inside any trusted CIDR; return
//     the first (rightmost) address that is NOT trusted.
//  4. If the header is absent, or all entries are trusted, fall back to the
//     host of r.RemoteAddr.
func ClientIP(r *http.Request, trusted []net.IPNet) string {
	remoteHost := remoteAddrHost(r.RemoteAddr)

	// No trusted proxies configured — use RemoteAddr directly.
	if len(trusted) == 0 {
		return remoteHost
	}

	remoteIP := net.ParseIP(remoteHost)
	remoteIsTrusted := remoteIP != nil && inCIDRList(remoteIP, trusted)

	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		// No XFF header. If remote is trusted, honour X-Real-IP.
		if remoteIsTrusted {
			if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
				return xri
			}
		}
		return remoteHost
	}

	// Build a slice of all hops: the XFF entries + RemoteAddr as the last hop.
	// We walk right-to-left through XFF only (RemoteAddr is the immediate peer).
	parts := strings.Split(xff, ",")
	// Walk right-to-left through the XFF list.
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		ip := net.ParseIP(candidate)
		if ip == nil {
			// Unparseable entry — treat as untrusted and return it.
			return candidate
		}
		if !inCIDRList(ip, trusted) {
			return ip.String()
		}
	}

	// All XFF entries were trusted. Fall back to RemoteAddr.
	return remoteHost
}

// remoteAddrHost strips the port from a "host:port" RemoteAddr string.
// Returns the original string if there is no port.
func remoteAddrHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// inCIDRList returns true if ip is contained by any CIDR in the list.
func inCIDRList(ip net.IP, cidrs []net.IPNet) bool {
	for i := range cidrs {
		if cidrs[i].Contains(ip) {
			return true
		}
	}
	return false
}
