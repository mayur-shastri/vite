package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bhavyajaix/BalkanID-filevault/pkg/auth" // Your JWT package
)

// userCtxKey is a custom type to use as a key for the context.
// This prevents collisions with other context keys.
type userCtxKey string

// UserContextKey is the key used to store the user ID in the request context.
const UserContextKey = userCtxKey("userID")

// AuthMiddleware decodes the JWT token and adds the user ID to the context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token provided, proceed without a user in the context
			next.ServeHTTP(w, r)
			return
		}

		// 2. Validate the header format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			// Invalid format, proceed without a user
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]

		// 3. Validate the token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			// Invalid token, proceed without a user
			next.ServeHTTP(w, r)
			return
		}

		// 4. Token is valid, add the user ID to the request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)

		// 5. Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
