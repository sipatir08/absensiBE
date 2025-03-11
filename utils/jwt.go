package utils

import (

	"time"

	"github.com/dgrijalva/jwt-go"
)

// SECRET_KEY default (bisa ubah sesuai kebutuhan)
var SECRET_KEY = []byte("mysecretkey")

// Generate JWT dengan expiry 1 jam
func GenerateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 1).Unix(), // Expire 1 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SECRET_KEY)
}