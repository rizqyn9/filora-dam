package session

import (
	"strings"
	"testing"
)

func TestGenerateToken(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tok, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken: %v", err)
		}
		if !strings.HasPrefix(tok, TokenPrefix) {
			t.Fatalf("token %q missing prefix %q", tok, TokenPrefix)
		}
		if !IsToken(tok) {
			t.Fatalf("IsToken false for %q", tok)
		}
		if seen[tok] {
			t.Fatalf("duplicate token generated: %q", tok)
		}
		seen[tok] = true
	}
}

func TestIsToken(t *testing.T) {
	if IsToken("eyJhbGciOiJ...") {
		t.Fatal("clerk-style JWT should not be a CLI token")
	}
	if !IsToken(TokenPrefix + "abc") {
		t.Fatal("prefixed token should be a CLI token")
	}
}
