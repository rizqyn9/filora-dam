package account

import (
	"context"
	"fmt"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ResolveUserID implements auth.UserResolver.
func (s *Service) ResolveUserID(ctx context.Context, clerkID string) (int64, error) {
	u, err := s.repo.GetByClerkID(ctx, clerkID)
	if err != nil {
		return 0, fmt.Errorf("resolve user: %w", err)
	}
	return u.ID, nil
}

// SyncUser creates or updates a local user mirror from Clerk webhook data.
func (s *Service) SyncUser(ctx context.Context, clerkID, email, name string, avatarURL *string) (*db.User, error) {
	existing, err := s.repo.GetByClerkID(ctx, clerkID)
	if err == nil && existing != nil {
		return s.repo.Update(ctx, db.UpdateUserParams{
			ClerkID:   clerkID,
			Email:     email,
			Name:      name,
			AvatarUrl: avatarURL,
		})
	}

	return s.repo.Create(ctx, db.CreateUserParams{
		ClerkID:   clerkID,
		Email:     email,
		Name:      name,
		AvatarUrl: avatarURL,
	})
}

func (s *Service) DeleteUser(ctx context.Context, clerkID string) error {
	return s.repo.Delete(ctx, clerkID)
}
