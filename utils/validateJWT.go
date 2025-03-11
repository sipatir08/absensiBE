package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// var secretKey = []byte("your_secret_key") // Gantilah dengan secret key yang digunakan untuk encoding JWT

// Validasi JWT, return user_id atau error
func ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Pastikan metode signing sesuai
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return SECRET_KEY, nil
	})

	if err != nil {
		return "", errors.New("invalid token")
	}

	// Ambil claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if exp, ok := claims["exp"].(float64); ok {
			// Cek apakah token expired
			if time.Now().Unix() > int64(exp) {
				return "", errors.New("token expired")
			}
		}
		userID, _ := claims["user_id"].(string)
		return userID, nil
	}

	return "", errors.New("invalid token claims")
}