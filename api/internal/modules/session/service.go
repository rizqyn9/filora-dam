package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// TokenPrefix distinguishes opaque CLI tokens from Clerk session JWTs.
const TokenPrefix = "flr_"

// IsToken reports whether a bearer token looks like a Filora CLI token.
func IsToken(token string) bool {
	return strings.HasPrefix(token, TokenPrefix)
}

type Service struct {
	repo     *Repository
	ttlHours int
}

func NewService(repo *Repository, ttlHours int) *Service {
	return &Service{repo: repo, ttlHours: ttlHours}
}

// Issue creates a new CLI session and returns the raw token exactly once.
func (s *Service) Issue(ctx context.Context, userID int64, in IssueInput, ip, userAgent *string) (IssueResult, error) {
	if err := lib.Validate(in); err != nil {
		return IssueResult{}, err
	}

	raw, err := generateToken()
	if err != nil {
		return IssueResult{}, lib.ErrInternal("failed to generate token").Wrap(err)
	}

	var expires *time.Time
	if s.ttlHours > 0 {
		t := time.Now().Add(time.Duration(s.ttlHours) * time.Hour)
		expires = &t
	}

	sess, err := s.repo.Create(ctx, createParams{
		UserID:    userID,
		TokenHash: lib.HashBytes([]byte(raw)),
		Label:     in.Label,
		IP:        ip,
		UserAgent: userAgent,
		ExpiresAt: expires,
	})
	if err != nil {
		return IssueResult{}, err
	}
	return IssueResult{Token: raw, Session: sess}, nil
}

// Authenticate resolves a raw CLI token to its active session (and touches it).
func (s *Service) Authenticate(ctx context.Context, rawToken string) (*Session, error) {
	sess, err := s.repo.GetActiveByTokenHash(ctx, lib.HashBytes([]byte(rawToken)))
	if errors.Is(err, ErrSessionInvalid) {
		return nil, lib.ErrUnauthorized("invalid or expired session")
	}
	if err != nil {
		return nil, err
	}
	_ = s.repo.TouchLastUsed(ctx, sess.ID)
	return &sess, nil
}

func (s *Service) List(ctx context.Context, userID int64) ([]Session, error) {
	return s.repo.ListActiveByUser(ctx, userID)
}

func (s *Service) Revoke(ctx context.Context, id uuid.UUID, userID int64) error {
	ok, err := s.repo.Revoke(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("session not found")
	}
	return nil
}

func (s *Service) RevokeAll(ctx context.Context, userID int64) error {
	return s.repo.RevokeAll(ctx, userID)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}
