package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword_Success(t *testing.T) {
	password := "securepassword123"

	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestVerifyPassword_Correct(t *testing.T) {
	password := "securepassword123"

	hash, _ := HashPassword(password)
	err := VerifyPassword(hash, password)

	assert.NoError(t, err)
}

func TestVerifyPassword_Incorrect(t *testing.T) {
	password := "securepassword123"
	wrongPassword := "wrongpassword"

	hash, _ := HashPassword(password)
	err := VerifyPassword(hash, wrongPassword)

	assert.Error(t, err)
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "securepassword123"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Same password should produce different hashes (due to salt)
	assert.NotEqual(t, hash1, hash2)

	// But both should verify
	assert.NoError(t, VerifyPassword(hash1, password))
	assert.NoError(t, VerifyPassword(hash2, password))
}
