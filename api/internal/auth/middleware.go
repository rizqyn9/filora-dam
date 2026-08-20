package auth

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

// SessionStore looks up sessions from the database (fallback when cache misses).
type SessionStore interface {
	GetByTokenHash(ctx context.Context, hash string) (*db.Session, error)
	Touch(ctx context.Context, id int64, newExpiry time.Time) error
}

// Middleware returns a Fiber middleware that authenticates requests via opaque tokens.
// Flow: extract token → hash → check cache → fallback to DB → set user context.
func Middleware(cache *TokenCache, store SessionStore, idleTTL func(db.ClientType) time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		raw := extractBearerToken(c)
		if raw == "" {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing authorization token")
		}

		tokenHash := lib.HashToken(raw)

		// Check cache first (no DB call)
		if cached, ok := cache.Get(tokenHash); ok {
			SetUser(c, &AuthUser{UserID: cached.UserID, SessionID: cached.SessionID})
			return c.Next()
		}

		// Cache miss → DB lookup
		sess, err := store.GetByTokenHash(c.Context(), tokenHash)
		if err != nil {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
		}

		// Sliding window: extend expiry
		ttl := idleTTL(sess.Client)
		newExpiry := time.Now().Add(ttl)
		_ = store.Touch(c.Context(), sess.ID, newExpiry)

		// Populate cache
		cache.Set(tokenHash, &CachedSession{
			SessionID: sess.ID,
			UserID:    sess.UserID,
			ExpiresAt: newExpiry,
		})

		SetUser(c, &AuthUser{UserID: sess.UserID, SessionID: sess.ID})
		return c.Next()
	}
}

func extractBearerToken(c fiber.Ctx) string {
	h := c.Get("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return h[7:]
	}
	return ""
}
