package auth

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type contextKey string

const userKey contextKey = "auth_user"

// AuthUser represents the authenticated user.
type AuthUser struct {
	UserID    int64
	SessionID int64
}

// SetUser stores the authenticated user in the request context.
func SetUser(c fiber.Ctx, user *AuthUser) {
	ctx := context.WithValue(c.Context(), userKey, user)
	c.SetContext(ctx)
}

// GetUser retrieves the authenticated user from the request context.
func GetUser(c fiber.Ctx) *AuthUser {
	u, _ := c.Context().Value(userKey).(*AuthUser)
	return u
}
