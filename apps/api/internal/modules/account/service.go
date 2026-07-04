package account

import (
	"context"
	"errors"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Provisioner sets up per-user resources on first sync (e.g. default gallery).
// Implemented by the gallery service; injected via SetProvisioner.
type Provisioner interface {
	EnsureDefaultGallery(ctx context.Context, userID int64) error
}

// Service holds user business logic.
type Service struct {
	repo        *Repository
	provisioner Provisioner
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SetProvisioner wires the per-user provisioning hook (optional).
func (s *Service) SetProvisioner(p Provisioner) {
	s.provisioner = p
}

// GetByID returns a user or a not-found AppError.
func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, ErrUserNotFound) {
		return nil, lib.ErrNotFound("user not found")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetByClerkID returns a user by Clerk id, or ErrUserNotFound.
func (s *Service) GetByClerkID(ctx context.Context, clerkID string) (*User, error) {
	return s.repo.GetByClerkID(ctx, clerkID)
}

// SyncFromClerk creates or refreshes the local user from a Clerk identity.
func (s *Service) SyncFromClerk(ctx context.Context, id auth.ClerkIdentity) (*User, error) {
	if id.ClerkUserID == "" || id.Email == "" {
		return nil, lib.ErrBadRequest("clerk identity requires id and email")
	}
	if id.Name == "" {
		id.Name = id.Email
	}
	u, err := s.repo.UpsertByClerkID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.provisioner != nil {
		// Best-effort: ensure the user has a default gallery. Idempotent.
		_ = s.provisioner.EnsureDefaultGallery(ctx, u.ID)
	}
	return u, nil
}

// UpdateProfile updates the current user's editable fields.
func (s *Service) UpdateProfile(ctx context.Context, id int64, in UpdateProfileInput) (*User, error) {
	if err := lib.Validate(in); err != nil {
		return nil, err
	}
	u, err := s.repo.UpdateProfile(ctx, id, in.Name, in.AvatarURL)
	if errors.Is(err, ErrUserNotFound) {
		return nil, lib.ErrNotFound("user not found")
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// DeactivateByClerkID marks a user inactive (Clerk user.deleted).
func (s *Service) DeactivateByClerkID(ctx context.Context, clerkID string) error {
	return s.repo.DeactivateByClerkID(ctx, clerkID)
}

// TouchLastSeen updates the last-seen timestamp (best effort).
func (s *Service) TouchLastSeen(ctx context.Context, id int64) error {
	return s.repo.TouchLastSeen(ctx, id)
}
