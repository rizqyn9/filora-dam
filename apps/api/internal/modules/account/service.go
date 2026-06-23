package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyTaken = errors.New("email already taken")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type Service struct {
	repo       *Repository
	jwtManager *lib.JWTManager
}

func NewService(repo *Repository, jwtManager *lib.JWTManager) *Service {
	return &Service{
		repo:       repo,
		jwtManager: jwtManager,
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

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*LoginResponse, error) {
	// Check if email already exists
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailAlreadyTaken
	}

	// Hash password
	hashedPassword, err := lib.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.repo.Create(ctx, req.Email, req.Name, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Get user by email
	dbUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify password
	if err := lib.VerifyPassword(dbUser.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Convert to User model
	user := &User{
		ID:           pgUUIDToString(dbUser.ID),
		Email:        dbUser.Email,
		Name:         dbUser.Name,
		StorageQuota: dbUser.StorageQuota,
		StorageUsed:  dbUser.StorageUsed,
		CreatedAt:    dbUser.CreatedAt.Time,
		UpdatedAt:    dbUser.UpdatedAt.Time,
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		User:  user,
		Token: token,
	}, nil
}
