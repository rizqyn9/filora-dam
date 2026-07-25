// Package auth holds neutral identity types shared across the HTTP edge and the
// modules: the authenticated Principal and the Clerk identity used to sync users.
// (The RBAC Authorizer is added in a later phase.)
package auth

import "github.com/gofiber/fiber/v3"

// ClerkIdentity is the subset of a Clerk user we mirror locally, produced by the
// token verifier or a webhook.
type ClerkIdentity struct {
	ClerkUserID string
	Email       string
	Name        string
	AvatarURL   *string
}

// Principal is the authenticated user attached to a request.
type Principal struct {
	UserID      int64
	ClerkUserID string
	Email       string
	Name        string
	IsActive    bool
}

const principalKey = "principal"

// SetPrincipal stores the principal on the request (Fiber Locals).
func SetPrincipal(c fiber.Ctx, p *Principal) {
	c.Locals(principalKey, p)
}

// GetPrincipal returns the request's principal, if any.
func GetPrincipal(c fiber.Ctx) (*Principal, bool) {
	p, ok := c.Locals(principalKey).(*Principal)
	return p, ok
}

// MustPrincipal returns the principal or nil. Use only on routes guarded by the
// auth middleware.
func MustPrincipal(c fiber.Ctx) *Principal {
	p, _ := GetPrincipal(c)
	return p
}
