package session

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

// ErrSessionInvalid is returned when a token matches no active session.
var ErrSessionInvalid = errors.New("session invalid")

// createParams are the inputs the service passes to Create.
type createParams struct {
	UserID    int64
	TokenHash string
	Label     *string
	IP        *string
	UserAgent *string
	ExpiresAt *time.Time
}

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, p createParams) (Session, error) {
	row, err := r.q.CreateCliSession(ctx, db.CreateCliSessionParams{
		UserID:    p.UserID,
		TokenHash: p.TokenHash,
		Label:     p.Label,
		IpAddress: parseAddr(p.IP),
		UserAgent: p.UserAgent,
		ExpiresAt: tsFromPtr(p.ExpiresAt),
	})
	if err != nil {
		return Session{}, err
	}
	return toSession(row), nil
}

func (r *Repository) GetActiveByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	row, err := r.q.GetActiveSessionByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionInvalid
	}
	if err != nil {
		return Session{}, err
	}
	return toSession(row), nil
}

func (r *Repository) ListActiveByUser(ctx context.Context, userID int64) ([]Session, error) {
	rows, err := r.q.ListActiveSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Session, 0, len(rows))
	for _, row := range rows {
		out = append(out, toSession(row))
	}
	return out, nil
}

func (r *Repository) TouchLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.TouchSessionLastUsed(ctx, id)
}

// Revoke marks one of the user's sessions revoked; returns false if none matched.
func (r *Repository) Revoke(ctx context.Context, id uuid.UUID, userID int64) (bool, error) {
	n, err := r.q.RevokeSession(ctx, db.RevokeSessionParams{ID: id, UserID: userID})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *Repository) RevokeAll(ctx context.Context, userID int64) error {
	return r.q.RevokeAllUserSessions(ctx, userID)
}

func toSession(s db.CliSession) Session {
	return Session{
		ID:         s.ID,
		UserID:     s.UserID,
		Label:      s.Label,
		IPAddress:  addrPtr(s.IpAddress),
		UserAgent:  s.UserAgent,
		LastUsedAt: tsPtr(s.LastUsedAt),
		ExpiresAt:  tsPtr(s.ExpiresAt),
		CreatedAt:  s.CreatedAt,
	}
}

func parseAddr(s *string) *netip.Addr {
	if s == nil || *s == "" {
		return nil
	}
	a, err := netip.ParseAddr(*s)
	if err != nil {
		return nil
	}
	return &a
}

func addrPtr(a *netip.Addr) *string {
	if a == nil {
		return nil
	}
	s := a.String()
	return &s
}

func tsFromPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		v := t.Time
		return &v
	}
	return nil
}
