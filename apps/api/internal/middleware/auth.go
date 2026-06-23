package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(jwtManager *lib.JWTManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return lib.Unauthorized(c, "Missing authorization header")
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return lib.Unauthorized(c, "Invalid authorization format")
		}

		token := parts[1]

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			if err == lib.ErrExpiredToken {
				return lib.Unauthorized(c, "Token has expired")
			}
			return lib.Unauthorized(c, "Invalid token")
		}

		// Store user info in context
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)

		return c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c fiber.Ctx) string {
	userID, ok := c.Locals("userID").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserEmail extracts user email from context
func GetUserEmail(c fiber.Ctx) string {
	email, ok := c.Locals("userEmail").(string)
	if !ok {
		return ""
	}
	return email
}
