package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SpaceMemberChecker is implemented by the space service.
type SpaceMemberChecker interface {
	HasMember(ctx context.Context, spaceID uuid.UUID, userID int64) (bool, error)
}

// RequireSpaceAccess checks that the authenticated user has access to the given space.
// Returns nil if allowed, error if not.
func RequireSpaceAccess(ctx context.Context, checker SpaceMemberChecker, spaceID uuid.UUID, userID int64) error {
	ok, err := checker.HasMember(ctx, spaceID, userID)
	if err != nil {
		return fmt.Errorf("check space access: %w", err)
	}
	if !ok {
		return fmt.Errorf("user %d has no access to space %s", userID, spaceID)
	}
	return nil
}
