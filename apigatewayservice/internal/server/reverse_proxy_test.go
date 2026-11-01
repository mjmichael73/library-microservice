package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// TestReverseProxy_ForwardsToTarget verifies the proxy handler forwards the
// request to the upstream target and returns its response. It also guards
// against the removed per-request debug Sentry call (R-008): this test must
// pass without any Sentry initialization.
func TestReverseProxy_ForwardsToTarget(t *testing.T) {
	var gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	}))
	defer upstream.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := reverseProxy(upstream.URL)(c); err != nil {
		t.Fatalf("reverseProxy returned error: %v", err)
	}

	if gotPath != "/anything" {
		t.Errorf("upstream path = %q, want %q", gotPath, "/anything")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "pong" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "pong")
	}
}

// TestReverseProxy_InvalidTargetReturnsError verifies a malformed target URL is
// reported as an error rather than panicking.
func TestReverseProxy_InvalidTargetReturnsError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := reverseProxy("://not-a-url")(c); err == nil {
		t.Fatal("expected error for invalid target URL, got nil")
	}
}
