package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// HashFile calculates SHA-256 hash of a file
func HashFile(r io.Reader) (string, error) {
	hash := sha256.New()

	if _, err := io.Copy(hash, r); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// HashBytes calculates SHA-256 hash of bytes
func HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
