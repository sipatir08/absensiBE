package middleware

import (
	"absensi/utils"
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Middleware untuk verifikasi token dan ekstraksi user_id
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil token dari header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Cek format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]

		// Verifikasi token
		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Tambahkan user_id ke context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


func JWTMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Authorization token is missing", http.StatusUnauthorized)
            return
        }

        // Parsing token dan memverifikasi
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // Verifikasi JWT
            return []byte("your-secret-key"), nil
        })
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        // Extract user_id dari claims
        userID := claims["user_id"].(string)
        if userID == "" {
            http.Error(w, "User ID missing in token", http.StatusUnauthorized)
            return
        }

        // Set user_id ke context request
        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
