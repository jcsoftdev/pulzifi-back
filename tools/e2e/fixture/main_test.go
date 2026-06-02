package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageServesV1ThenV2AfterSwitch(t *testing.T) {
	srv := newServer()

	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/page", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), "Version ONE") {
		t.Fatalf("expected v1 body, got code=%d body=%q", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/switch", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("switch failed: code=%d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/page", nil))
	if !strings.Contains(rr.Body.String(), "Version TWO") {
		t.Fatalf("expected v2 body, got %q", rr.Body.String())
	}
}

func TestHealth(t *testing.T) {
	srv := newServer()
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("health code=%d", rr.Code)
	}
}
