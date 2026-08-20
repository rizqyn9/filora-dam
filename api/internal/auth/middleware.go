package auth

import (
	"context"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gofiber/fiber/v3"

	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

// UserResolver looks up the local user ID by Clerk ID.
// Implemented by the account/user repository at the compose root.
type UserResolver interface {
	ResolveUserID(ctx context.Context, clerkID string) (int64, error)
}

// Middleware returns a Fiber middleware that verifies Clerk JWTs.
func Middleware(resolver UserResolver) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := extractBearerToken(c)
		if token == "" {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing authorization token")
		}

		claims, err := jwt.Verify(c.Context(), &jwt.VerifyParams{Token: token})
		if err != nil {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
		}

		clerkID := claims.Subject
		if clerkID == "" {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid token subject")
		}

		userID, err := resolver.ResolveUserID(c.Context(), clerkID)
		if err != nil {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "user not found")
		}

		SetUser(c, &AuthUser{
			ClerkID: clerkID,
			UserID:  userID,
		})

		return c.Next()
	}
}

// CLITokenMiddleware verifies opaque CLI tokens.
// tokenResolver returns (userID, error) given a raw token string.
type CLITokenResolver interface {
	ResolveToken(ctx context.Context, rawToken string) (int64, error)
}

func CLITokenMiddleware(resolver CLITokenResolver) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := extractBearerToken(c)
		if token == "" {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing authorization token")
		}

		userID, err := resolver.ResolveToken(c.Context(), token)
		if err != nil {
			return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
		}

		SetUser(c, &AuthUser{UserID: userID})
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
