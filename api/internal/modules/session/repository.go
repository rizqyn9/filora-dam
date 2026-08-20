package session

import (
	"context"
	"fmt"
	"time"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, userID int64, tokenHash, label string, expiresAt time.Time) (*db.CliSession, error) {
	s, err := r.q.CreateCLISession(ctx, db.CreateCLISessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		Label:     label,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &s, nil
}

func (r *Repository) GetByTokenHash(ctx context.Context, hash string) (*db.CliSession, error) {
	s, err := r.q.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get session by token: %w", err)
	}
	return &s, nil
}

func (r *Repository) ListActive(ctx context.Context, userID int64) ([]db.CliSession, error) {
	sessions, err := r.q.ListActiveSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	return sessions, nil
}

func (r *Repository) Touch(ctx context.Context, id int64) error {
	return r.q.TouchSession(ctx, id)
}

func (r *Repository) Revoke(ctx context.Context, id int64) error {
	return r.q.RevokeSession(ctx, id)
}

func (r *Repository) RevokeAll(ctx context.Context, userID int64) error {
	return r.q.RevokeAllUserSessions(ctx, userID)
}
