// File: modules/middleware.go
package modules

import (
    "context"
    "net/http"
    "strings"
)

type contextKey string

const usernameKey contextKey = "username"

// AuthMiddleware verifies JWT tokens and protects routes
func (app *AppContext) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Retrieve the token from the Authorization header (Bearer token)
        authHeader := r.Header.Get("Authorization")
        var tokenString string

        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString = strings.TrimPrefix(authHeader, "Bearer ")
        } else {
            // Alternatively, retrieve the token from cookies
            cookie, err := r.Cookie("token")
            if err != nil {
                http.Error(w, "User not authenticated", http.StatusUnauthorized)
                return
            }
            tokenString = cookie.Value
        }

        username, err := app.UserManager.ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add the username to the request context
        ctx := context.WithValue(r.Context(), usernameKey, username)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetUsernameFromContext retrieves the username from the context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
    username, ok := ctx.Value(usernameKey).(string)
    return username, ok
}
