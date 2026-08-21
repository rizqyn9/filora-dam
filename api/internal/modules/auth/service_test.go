package auth

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	iauth "github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

func TestLogin_WrongPassword_Fails(t *testing.T) {
	// Unit test: bcrypt comparison directly (the service wraps this)
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)

	err := bcrypt.CompareHashAndPassword(hash, []byte("wrong-password"))
	if err == nil {
		t.Fatal("expected bcrypt mismatch error")
	}
}

func TestLogin_CorrectPassword_Succeeds(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("my-secure-pass"), bcrypt.DefaultCost)

	err := bcrypt.CompareHashAndPassword(hash, []byte("my-secure-pass"))
	if err != nil {
		t.Fatalf("expected password to match, got: %v", err)
	}
}

func TestTokenCache_HitAfterSet(t *testing.T) {
	cache := iauth.NewTokenCache()

	cache.Set("hash123", &iauth.CachedSession{
		SessionID: 1,
		UserID:    42,
		ExpiresAt: timeInFuture(),
	})

	got, ok := cache.Get("hash123")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.UserID != 42 {
		t.Fatalf("expected user 42, got %d", got.UserID)
	}
}

func TestTokenCache_MissAfterDelete(t *testing.T) {
	cache := iauth.NewTokenCache()

	cache.Set("hash123", &iauth.CachedSession{
		SessionID: 1,
		UserID:    42,
		ExpiresAt: timeInFuture(),
	})
	cache.Delete("hash123")

	_, ok := cache.Get("hash123")
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

func TestTokenCache_ExpiredEntry_Miss(t *testing.T) {
	cache := iauth.NewTokenCache()

	cache.Set("hash123", &iauth.CachedSession{
		SessionID: 1,
		UserID:    42,
		ExpiresAt: timeInPast(),
	})

	_, ok := cache.Get("hash123")
	if ok {
		t.Fatal("expected cache miss for expired entry")
	}
}

func TestIdleTTL_WebShorterThanCLI(t *testing.T) {
	web := IdleTTL(db.ClientTypeWeb)
	cli := IdleTTL(db.ClientTypeCli)

	if cli <= web {
		t.Fatalf("expected CLI TTL > web TTL, got cli=%v web=%v", cli, web)
	}
}

// --- Helpers ---

func timeInFuture() time.Time {
	return time.Now().Add(24 * time.Hour)
}

func timeInPast() time.Time {
	return time.Now().Add(-24 * time.Hour)
}
