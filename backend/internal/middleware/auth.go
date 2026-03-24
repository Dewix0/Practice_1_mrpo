package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a custom type used for context keys to avoid collisions.
type contextKey string

const (
	contextKeyUserID contextKey = "userId"
	contextKeyRole   contextKey = "role"
)

// JWTAuth extracts a JWT from the httpOnly cookie named "token" and enriches
// the request context with userId and role claims. It does NOT block requests
// without a valid token — it simply passes them through unenriched.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("token")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				next.ServeHTTP(w, r)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			userIDFloat, ok := claims["userId"].(float64)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			role, _ := claims["role"].(string)

			ctx := context.WithValue(r.Context(), contextKeyUserID, int64(userIDFloat))
			ctx = context.WithValue(ctx, contextKeyRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth rejects requests that do not have a valid JWT in context with 401.
func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !IsAuthenticated(r.Context()) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks that the authenticated user's role is one of the allowed roles.
// Returns 403 if the role is not allowed.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetRole(r.Context())
			for _, allowed := range roles {
				if userRole == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "forbidden"})
		})
	}
}

// GetUserID returns the user ID from the context, or 0 if not present.
func GetUserID(ctx context.Context) int64 {
	id, _ := ctx.Value(contextKeyUserID).(int64)
	return id
}

// GetRole returns the role from the context, or empty string if not present.
func GetRole(ctx context.Context) string {
	role, _ := ctx.Value(contextKeyRole).(string)
	return role
}

// IsAuthenticated returns true if the context contains a valid user ID.
func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != 0
}
