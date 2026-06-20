package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type requestIDKeyType struct{}

var requestIDKey requestIDKeyType

// requestIDMiddleware reads X-Request-ID from the incoming request (or generates
// a new random 32-character lowercase hex ID) and echoes it on the response.
// The ID is also available to downstream handlers via requestIDFromContext.
func requestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = generateRequestID()
			}
			w.Header().Set("X-Request-ID", id)
			ctx := context.WithValue(r.Context(), requestIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requestIDFromContext retrieves the request ID injected by requestIDMiddleware.
// Returns an empty string when no ID is present in the context.
func requestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// generateRequestID returns a 32-character lowercase hexadecimal string sourced
// from 16 bytes of cryptographically secure random data.
func generateRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Extremely unlikely; fall back to a fixed sentinel rather than panic.
		return "0000000000000000000000000000000f"
	}
	return hex.EncodeToString(b[:])
}
