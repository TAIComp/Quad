// File: modules/auth.go
package modules

import (
    "log"
    "net/http"
)

// AuthMiddleware is a middleware that checks for a valid JWT token
func AuthMiddleware(next func(http.ResponseWriter, *http.Request, string), userManager *UserManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("token")
        if err != nil {
            log.Println("No token cookie found:", err)
            sendErrorResponse(w, "Unauthorized: No token provided", http.StatusUnauthorized)
            return
        }

        log.Println("Token cookie found:", cookie.Value)

        username, err := userManager.ValidateToken(cookie.Value)
        if err != nil {
            log.Println("Invalid token:", err)
            sendErrorResponse(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
            return
        }

        next(w, r, username)
    }
}
