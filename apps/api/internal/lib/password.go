package lib

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = 10
)

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a hashed password with a plain text password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
