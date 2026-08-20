package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

const (
	tokenBytes = 32
	sessionTTL = 90 * 24 * time.Hour // 90 days
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSession(ctx context.Context, userID int64, label string) (string, *db.CliSession, error) {
	raw, err := generateToken()
	if err != nil {
		return "", nil, err
	}

	hash := hashToken(raw)
	expiresAt := time.Now().Add(sessionTTL)

	sess, err := s.repo.Create(ctx, userID, hash, label, expiresAt)
	if err != nil {
		return "", nil, err
	}

	return raw, sess, nil
}

func (s *Service) ListSessions(ctx context.Context, userID int64) ([]db.CliSession, error) {
	return s.repo.ListActive(ctx, userID)
}

func (s *Service) RevokeSession(ctx context.Context, id int64) error {
	return s.repo.Revoke(ctx, id)
}

func (s *Service) RevokeAll(ctx context.Context, userID int64) error {
	return s.repo.RevokeAll(ctx, userID)
}

// ResolveToken implements auth.CLITokenResolver.
func (s *Service) ResolveToken(ctx context.Context, rawToken string) (int64, error) {
	hash := hashToken(rawToken)
	sess, err := s.repo.GetByTokenHash(ctx, hash)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}

	// Touch last_used
	_ = s.repo.Touch(ctx, sess.ID)

	return sess.UserID, nil
}

func generateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
