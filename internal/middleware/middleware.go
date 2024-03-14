package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func RoleCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const BearerSchema = "Bearer "
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No token provided", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, BearerSchema)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("your_secret_key"), nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			role := claims["role"].(float64)

			if role == 1 {
				next.ServeHTTP(w, r)
			} else if role == 2 && r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Not authorized for this action", http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		}
	})
}
