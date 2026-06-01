package middleware

import (
	"net"
	"net/http/httptest"
	"testing"
)

func mustParseCIDR(s string) net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *ipNet
}

func TestClientIP_NoTrustedProxies_UsesRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	req.Header.Set("X-Forwarded-For", "9.9.9.9")

	ip := ClientIP(req, nil)
	if ip != "1.2.3.4" {
		t.Errorf("no trusted proxies: got %q, want %q", ip, "1.2.3.4")
	}
}

func TestClientIP_TrustedProxy_RightmostUntrusted(t *testing.T) {
	// Proxy at 10.0.0.1 is trusted; real client is 203.0.113.50
	// XFF is built left-to-right: [original-client, ...proxies]
	// rightmost-untrusted walk: 10.0.0.1 is trusted → skip; 203.0.113.50 is not → return it
	trusted := []net.IPNet{mustParseCIDR("10.0.0.0/8")}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 10.0.0.1")

	ip := ClientIP(req, trusted)
	if ip != "203.0.113.50" {
		t.Errorf("rightmost untrusted: got %q, want %q", ip, "203.0.113.50")
	}
}

func TestClientIP_SpoofedLeftmostIgnored(t *testing.T) {
	// Attacker injects a fake IP at leftmost position.
	// XFF: "spoofed-attacker, real-client, trusted-proxy"
	// RemoteAddr is the trusted proxy.
	// Walk right-to-left: trusted-proxy (trusted) → skip; real-client (not trusted) → return.
	trusted := []net.IPNet{mustParseCIDR("10.0.0.0/8")}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.5, 10.0.0.1")

	ip := ClientIP(req, trusted)
	if ip != "198.51.100.5" {
		t.Errorf("spoofed leftmost: got %q, want %q", ip, "198.51.100.5")
	}
}

func TestClientIP_AllTrusted_FallsBackToRemoteAddr(t *testing.T) {
	trusted := []net.IPNet{mustParseCIDR("10.0.0.0/8"), mustParseCIDR("172.16.0.0/12")}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.3, 172.16.0.1")

	ip := ClientIP(req, trusted)
	// All XFF entries trusted → fall back to host of RemoteAddr
	if ip != "10.0.0.2" {
		t.Errorf("all trusted: got %q, want %q", ip, "10.0.0.2")
	}
}

func TestClientIP_NoXFF_TrustedRemoteAddr_UsesRemoteAddr(t *testing.T) {
	trusted := []net.IPNet{mustParseCIDR("10.0.0.0/8")}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	// no X-Forwarded-For

	ip := ClientIP(req, trusted)
	if ip != "10.0.0.1" {
		t.Errorf("no xff: got %q, want %q", ip, "10.0.0.1")
	}
}

func TestClientIP_XRealIP_HonoredWhenRemoteAddrTrusted(t *testing.T) {
	trusted := []net.IPNet{mustParseCIDR("10.0.0.0/8")}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Real-IP", "203.0.113.99")
	// no X-Forwarded-For → falls through to X-Real-IP check

	ip := ClientIP(req, trusted)
	if ip != "203.0.113.99" {
		t.Errorf("X-Real-IP with trusted remote: got %q, want %q", ip, "203.0.113.99")
	}
}

func TestClientIP_XRealIP_IgnoredWhenRemoteAddrUntrusted(t *testing.T) {
	trusted := []net.IPNet{mustParseCIDR("10.0.0.0/8")}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.5.5.5:1234" // NOT trusted
	req.Header.Set("X-Real-IP", "203.0.113.99")

	ip := ClientIP(req, trusted)
	// RemoteAddr not trusted → use it directly, X-Real-IP ignored
	if ip != "5.5.5.5" {
		t.Errorf("X-Real-IP untrusted remote: got %q, want %q", ip, "5.5.5.5")
	}
}

func TestClientIP_EmptyTrustedList_NoXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:9999"

	ip := ClientIP(req, []net.IPNet{})
	if ip != "192.168.1.1" {
		t.Errorf("empty trusted list: got %q, want %q", ip, "192.168.1.1")
	}
}
