package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	iauth "github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

const (
	WebIdleTTL = 7 * 24 * time.Hour  // 7 days
	CLIIdleTTL = 90 * 24 * time.Hour // 90 days
)

// IdleTTL returns the sliding window TTL for the given client type.
func IdleTTL(client db.ClientType) time.Duration {
	if client == db.ClientTypeCli {
		return CLIIdleTTL
	}
	return WebIdleTTL
}

type Service struct {
	queries *db.Queries
	cache   *iauth.TokenCache
}

func NewService(queries *db.Queries, cache *iauth.TokenCache) *Service {
	return &Service{queries: queries, cache: cache}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	token, sess, err := s.createSession(ctx, user.ID, db.ClientType(req.Client), req.Label)
	if err != nil {
		return nil, err
	}

	// Cache the new session
	s.cache.Set(lib.HashToken(token), &iauth.CachedSession{
		SessionID: sess.ID,
		UserID:    user.ID,
		ExpiresAt: sess.ExpiresAt,
	})

	return &AuthResponse{
		Token: token,
		User:  UserData{ID: user.ID, Email: user.Email, Name: user.Name},
	}, nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Verify invitation token
	tokenHash := lib.HashToken(req.InviteToken)
	invite, err := s.queries.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired invitation")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        invite.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Assign default role
	_ = s.queries.AssignRole(ctx, db.AssignRoleParams{UserID: user.ID, RoleName: "member"})

	// Accept invitation
	_ = s.queries.AcceptInvitation(ctx, invite.ID)

	// Add as space member
	_, _ = s.queries.AddSpaceMember(ctx, db.AddSpaceMemberParams{
		SpaceID: invite.SpaceID,
		UserID:  user.ID,
		Role:    invite.Role,
	})

	// Create session (default web for registration)
	token, sess, err := s.createSession(ctx, user.ID, db.ClientTypeWeb, "")
	if err != nil {
		return nil, err
	}

	s.cache.Set(lib.HashToken(token), &iauth.CachedSession{
		SessionID: sess.ID,
		UserID:    user.ID,
		ExpiresAt: sess.ExpiresAt,
	})

	return &AuthResponse{
		Token: token,
		User:  UserData{ID: user.ID, Email: user.Email, Name: user.Name},
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID int64) error {
	return s.queries.RevokeSession(ctx, sessionID)
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, req ChangePasswordRequest) error {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("incorrect current password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.queries.UpdatePassword(ctx, db.UpdatePasswordParams{ID: userID, PasswordHash: string(hash)}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	// Revoke all sessions (force re-login with new password)
	_ = s.queries.RevokeAllUserSessions(ctx, userID)
	s.cache.DeleteByUserID(userID)

	return nil
}

func (s *Service) GetMe(ctx context.Context, userID int64) (*UserData, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &UserData{ID: user.ID, Email: user.Email, Name: user.Name}, nil
}

// GetByTokenHash implements auth.SessionStore for the middleware fallback.
func (s *Service) GetByTokenHash(ctx context.Context, hash string) (*db.Session, error) {
	sess, err := s.queries.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	return &sess, nil
}

// Touch implements auth.SessionStore for sliding window update.
func (s *Service) Touch(ctx context.Context, id int64, newExpiry time.Time) error {
	return s.queries.TouchSession(ctx, db.TouchSessionParams{ID: id, ExpiresAt: newExpiry})
}

func (s *Service) createSession(ctx context.Context, userID int64, client db.ClientType, label string) (string, *db.Session, error) {
	raw, err := lib.GenerateToken()
	if err != nil {
		return "", nil, err
	}

	ttl := IdleTTL(client)
	expiresAt := time.Now().Add(ttl)

	sess, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		TokenHash: lib.HashToken(raw),
		Client:    client,
		Label:     label,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", nil, fmt.Errorf("create session: %w", err)
	}

	return raw, &sess, nil
}
