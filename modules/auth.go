// File: modules/auth.go
package modules

import (
    "context"
    "log"
    "net/http"
)

type contextKey string

const usernameKey contextKey = "username"

// GetUsernameFromContext retrieves the username from the context
func GetUsernameFromContext(ctx context.Context) (string, bool) {
    username, ok := ctx.Value(usernameKey).(string)
    return username, ok
}

// AuthMiddleware is a middleware that checks for a valid JWT token
func (app *AppContext) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            log.Println("No token cookie found:", err)
            sendErrorResponse(w, "Unauthorized: No token provided", http.StatusUnauthorized)
            return
        }

        username, err := app.UserManager.ValidateToken(cookie.Value)
        if err != nil {
            log.Println("Invalid token:", err)
            sendErrorResponse(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
            return
        }

        // Add username to request context
        ctx := context.WithValue(r.Context(), usernameKey, username)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
