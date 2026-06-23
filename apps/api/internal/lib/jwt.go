package lib

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
	}
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(userID, email string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims, nil
}

// ExtractUserID extracts user ID from token
func (m *JWTManager) ExtractUserID(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
