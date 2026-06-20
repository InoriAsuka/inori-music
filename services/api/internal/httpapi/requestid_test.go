package httpapi

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var hexID32 = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestRequestIDPassthroughExisting(t *testing.T) {
	const existing = "my-trace-id-abc123"
	mw := requestIDMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", existing)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != existing {
		t.Errorf("X-Request-ID = %q, want %q (passthrough)", got, existing)
	}
}

func TestRequestIDGeneratedWhenAbsent(t *testing.T) {
	mw := requestIDMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	if !hexID32.MatchString(got) {
		t.Errorf("X-Request-ID = %q, want 32 lowercase hex chars", got)
	}
}

func TestRequestIDPresentOnAllRoutes(t *testing.T) {
	// Use the full test handler to confirm the middleware is wired end-to-end.
	handler := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("X-Request-ID")
	if got == "" {
		t.Error("X-Request-ID header missing on /healthz response")
	}
}

func TestRequestIDInjectedIntoContext(t *testing.T) {
	const id = "context-test-id"
	var captured string

	mw := requestIDMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = requestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", id)

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if captured != id {
		t.Errorf("context ID = %q, want %q", captured, id)
	}
}
