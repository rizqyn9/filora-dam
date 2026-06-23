package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrInvalidPassword   = errors.New("invalid password")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *Service) GetQuota(ctx context.Context, userID string) (*QuotaInfo, error) {
	quota, err := s.repo.GetQuota(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}
	return quota, nil
}

func (s *Service) UpdateStorageUsed(ctx context.Context, userID string, used int64) error {
	if err := s.repo.UpdateStorageUsed(ctx, userID, used); err != nil {
		return fmt.Errorf("failed to update storage used: %w", err)
	}
	return nil
}

func (s *Service) CheckQuota(ctx context.Context, userID string, additionalSize int64) error {
	quota, err := s.GetQuota(ctx, userID)
	if err != nil {
		return err
	}

	if quota.Used+additionalSize > quota.Quota {
		return errors.New("storage quota exceeded")
	}

	return nil
}
