package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// sentinel handler that always writes 200 OK with a short body.
var corsTestNext = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func TestCORSPreflightReturns204(t *testing.T) {
	mw := corsMiddleware(nil)(corsTestNext)
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/catalog/tracks", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", rec.Code)
	}
}

func TestCORSPreflightHeadersPresent(t *testing.T) {
	mw := corsMiddleware(nil)(corsTestNext)
	req := httptest.NewRequest(http.MethodOptions, "/healthz", nil)
	req.Header.Set("Origin", "https://inori.example.com")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	for _, header := range []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Credentials",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Max-Age",
	} {
		if rec.Header().Get(header) == "" {
			t.Errorf("preflight response missing header %q", header)
		}
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q, want \"true\"", got)
	}
}

func TestCORSAllowedOriginReflected(t *testing.T) {
	allowed := "https://inori.example.com"
	mw := corsMiddleware([]string{allowed})(corsTestNext)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", allowed)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowed {
		t.Errorf("Allow-Origin = %q, want %q", got, allowed)
	}
}

func TestCORSDisallowedOriginOmitted(t *testing.T) {
	mw := corsMiddleware([]string{"https://inori.example.com"})(corsTestNext)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://evil.example.com")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin = %q, want empty for disallowed origin", got)
	}
}

func TestCORSPermissiveModeReflectsAnyOrigin(t *testing.T) {
	// Empty origins slice → permissive mode.
	mw := corsMiddleware([]string{})(corsTestNext)
	origin := "https://random-dev.local:3000"
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", origin)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("Allow-Origin = %q, want %q in permissive mode", got, origin)
	}
}

func TestCORSNonPreflightPassesThrough(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	mw := corsMiddleware(nil)(next)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called for non-preflight GET request")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
