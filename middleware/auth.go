package middleware

import (
	"net/http"
	"strings"
	"trello-lite/utils"

	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get token from Authorization header (Format: "Bearer <token>")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.SendError(w, http.StatusUnauthorized, "Missing token")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &utils.Claims{}

		// 2. Parse and Verify the token
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			// Reference the key from the utils package
			return utils.JwtKey, nil
		})

		if err != nil || !token.Valid {
			utils.SendError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// 3. Inject verified data into headers so handlers can still use them
		r.Header.Set("User-ID", claims.UserID)
		r.Header.Set("Role", claims.Role)

		next.ServeHTTP(w, r)
	}
}
