package auth

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type contextKey string

const userKey contextKey = "auth_user"

// AuthUser represents the authenticated user extracted from the JWT.
type AuthUser struct {
	ClerkID string
	UserID  int64 // local DB user ID, resolved after JWT verification
}

// SetUser stores the authenticated user in the request context.
func SetUser(c fiber.Ctx, user *AuthUser) {
	ctx := context.WithValue(c.Context(), userKey, user)
	c.SetContext(ctx)
}

// GetUser retrieves the authenticated user from the request context.
// Returns nil if no user is set (should not happen behind auth middleware).
func GetUser(c fiber.Ctx) *AuthUser {
	u, _ := c.Context().Value(userKey).(*AuthUser)
	return u
}
