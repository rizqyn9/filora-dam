package lib

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// HashBytes returns the hex-encoded SHA-256 of b.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// HashReader returns the hex-encoded SHA-256 of everything read from r.
func HashReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("hash reader: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
