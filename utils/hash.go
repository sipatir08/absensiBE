package utils

import (
	"golang.org/x/crypto/bcrypt"
    "log"
)

// HashPassword hashes the user's password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares the given password with the stored hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}