package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/account"
)

// ClerkVerifier verifies a Clerk session token into an identity.
// (Consumer-defined interface; implemented by internal/clerk.Verifier.)
type ClerkVerifier interface {
	Verify(ctx context.Context, token string) (auth.ClerkIdentity, error)
}

// AuthDeps are the dependencies of the auth middleware.
type AuthDeps struct {
	Clerk    ClerkVerifier
	Accounts *account.Service
}

// RequireAuth authenticates the request (currently: Clerk web sessions) and
// attaches the Principal. CLI-token support is added in a later phase.
func RequireAuth(deps AuthDeps) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := bearerToken(c)
		if token == "" {
			return lib.ErrUnauthorized("missing bearer token")
		}
		if deps.Clerk == nil {
			return lib.ErrUnauthorized("authentication is not configured")
		}

		ident, err := deps.Clerk.Verify(c.Context(), token)
		if err != nil {
			return lib.ErrUnauthorized("invalid or expired token").Wrap(err)
		}

		ctx := c.Context()
		user, err := deps.Accounts.GetByClerkID(ctx, ident.ClerkUserID)
		if errors.Is(err, account.ErrUserNotFound) {
			user, err = deps.Accounts.SyncFromClerk(ctx, ident) // JIT create
		}
		if err != nil {
			return err
		}
		if !user.IsActive {
			return lib.ErrForbidden("account is inactive")
		}

		auth.SetPrincipal(c, &auth.Principal{
			UserID:      user.ID,
			ClerkUserID: user.ClerkUserID,
			Email:       user.Email,
			Name:        user.Name,
			IsActive:    user.IsActive,
		})
		_ = deps.Accounts.TouchLastSeen(ctx, user.ID)

		return c.Next()
	}
}

func bearerToken(c fiber.Ctx) string {
	h := c.Get(fiber.HeaderAuthorization)
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
