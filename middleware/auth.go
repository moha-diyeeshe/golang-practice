package middleware

import (
	"context"
	"net/http"
	"strings"

	"rest_api_go/auth"
)

type contextKey string

// UserIDKey is the request context key for the authenticated user id.
const UserIDKey contextKey = "user_id"

// SessionIDKey is the Redis session id (JWT jti) for the current access token.
const SessionIDKey contextKey = "session_id"

// UserIDFromContext returns the user id set by RequireAuth, or false if missing.
func UserIDFromContext(ctx context.Context) (int, bool) {
	v := ctx.Value(UserIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

// SessionIDFromContext returns the Redis session id set by RequireAuth.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(SessionIDKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// RequireAuth validates Bearer JWT + Redis session, then sets user id on the context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if h == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(h, prefix) {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(h, prefix))
		if raw == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		claims, err := auth.ParseToken(raw)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if claims.Type != "access" {
			http.Error(w, "use access token for this endpoint", http.StatusUnauthorized)
			return
		}
		if claims.ID == "" {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}
		if err := auth.ValidateSession(claims.ID, claims.UserID); err != nil {
			http.Error(w, "session expired or logged out", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, SessionIDKey, claims.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
