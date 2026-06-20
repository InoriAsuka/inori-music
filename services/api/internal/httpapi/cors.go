package httpapi

import (
	"net/http"
	"strings"
)

const (
	corsAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsAllowHeaders = "Authorization, Content-Type, X-Request-ID"
	corsMaxAge       = "86400"
)

// corsMiddleware returns an HTTP middleware that adds CORS headers to every
// response and handles OPTIONS preflight requests.
//
// When origins is non-empty, only requests whose Origin header exactly matches
// one of the listed values receive the Access-Control-Allow-Origin header.
// When origins is empty, the request Origin is reflected back (permissive mode
// suitable for local development).
//
// A wildcard "*" is never used because Access-Control-Allow-Credentials: true
// requires an explicit origin.
func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	// Build a lookup set for O(1) checks.
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		if o != "" {
			allowed[strings.TrimRight(o, "/")] = struct{}{}
		}
	}
	permissive := len(allowed) == 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				// Determine the reflected value.
				reflected := ""
				if permissive {
					reflected = origin
				} else if _, ok := allowed[strings.TrimRight(origin, "/")]; ok {
					reflected = origin
				}

				if reflected != "" {
					w.Header().Set("Access-Control-Allow-Origin", reflected)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
					w.Header().Set("Access-Control-Allow-Headers", corsAllowHeaders)
					w.Header().Set("Access-Control-Max-Age", corsMaxAge)
					// Vary on Origin so caches do not serve one origin's response to another.
					w.Header().Add("Vary", "Origin")
				}
			}

			// Handle preflight without forwarding to the mux.
			if r.Method == http.MethodOptions && origin != "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
