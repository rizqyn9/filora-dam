package lib

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken_Success(t *testing.T) {
	manager := NewJWTManager("test-secret-key-123")

	token, err := manager.GenerateToken("user-123", "user@example.com")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken_Success(t *testing.T) {
	manager := NewJWTManager("test-secret-key-123")

	token, _ := manager.GenerateToken("user-123", "user@example.com")
	claims, err := manager.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-123")

	claims, err := manager.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret-1")
	manager2 := NewJWTManager("secret-2")

	token, _ := manager1.GenerateToken("user-123", "user@example.com")
	claims, err := manager2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestExtractUserID_Success(t *testing.T) {
	manager := NewJWTManager("test-secret-key-123")

	token, _ := manager.GenerateToken("user-456", "user@example.com")
	userID, err := manager.ExtractUserID(token)

	assert.NoError(t, err)
	assert.Equal(t, "user-456", userID)
}

func TestExtractUserID_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-123")

	userID, err := manager.ExtractUserID("invalid-token")

	assert.Error(t, err)
	assert.Empty(t, userID)
}

func TestToken_Expiration(t *testing.T) {
	manager := NewJWTManager("test-secret-key-123")

	token, _ := manager.GenerateToken("user-123", "user@example.com")
	claims, err := manager.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Token should expire in 24 hours
	expiresAt := claims.ExpiresAt.Time
	now := time.Now()
	diff := expiresAt.Sub(now)

	// Should be approximately 24 hours (allow some tolerance)
	assert.Greater(t, diff, 23*time.Hour)
	assert.Less(t, diff, 25*time.Hour)
}
