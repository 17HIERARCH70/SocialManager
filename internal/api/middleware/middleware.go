package middleware

import (
	"context"
	"net/http"
	"strings"

	jwtService "github.com/17HIERARCH70/SocialManager/internal/lib/jwt"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware is a middleware for JWT authentication
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		token, err := jwtService.VerifyAccessToken(tokenString)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid Token Claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid User ID in Token Claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user", uint64(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
