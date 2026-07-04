package account

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

// ErrUserNotFound is returned when a user lookup finds nothing.
var ErrUserNotFound = errors.New("user not found")

// Repository provides persistence for users and Clerk webhook events.
type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return toUser(row), nil
}

func (r *Repository) GetByClerkID(ctx context.Context, clerkID string) (*User, error) {
	row, err := r.q.GetUserByClerkID(ctx, clerkID)
	if err != nil {
		return nil, mapErr(err)
	}
	return toUser(row), nil
}

// UpsertByClerkID creates or refreshes a user from a Clerk identity.
func (r *Repository) UpsertByClerkID(ctx context.Context, id auth.ClerkIdentity) (*User, error) {
	row, err := r.q.UpsertUserByClerkID(ctx, db.UpsertUserByClerkIDParams{
		ClerkUserID: id.ClerkUserID,
		Email:       id.Email,
		Name:        id.Name,
		AvatarUrl:   id.AvatarURL,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toUser(row), nil
}

func (r *Repository) UpdateProfile(ctx context.Context, id int64, name string, avatar *string) (*User, error) {
	row, err := r.q.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		ID:        id,
		Name:      name,
		AvatarUrl: avatar,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return toUser(row), nil
}

func (r *Repository) DeactivateByClerkID(ctx context.Context, clerkID string) error {
	return r.q.DeactivateUserByClerkID(ctx, clerkID)
}

func (r *Repository) TouchLastSeen(ctx context.Context, id int64) error {
	return r.q.TouchUserLastSeen(ctx, id)
}

// --- Clerk webhook events ---

// InsertWebhookEvent records a delivery; returns false if the event was already
// seen (idempotent no-op).
func (r *Repository) InsertWebhookEvent(ctx context.Context, eventID, eventType string, payload []byte) (int64, bool, error) {
	id, err := r.q.InsertWebhookEvent(ctx, db.InsertWebhookEventParams{
		EventID:   eventID,
		EventType: eventType,
		Payload:   payload,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil // duplicate delivery
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func (r *Repository) MarkWebhookProcessed(ctx context.Context, id int64) error {
	return r.q.MarkWebhookProcessed(ctx, id)
}

func (r *Repository) MarkWebhookFailed(ctx context.Context, id int64, msg string) error {
	return r.q.MarkWebhookFailed(ctx, db.MarkWebhookFailedParams{ID: id, Error: &msg})
}

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}
	return err
}

func toUser(u db.User) *User {
	return &User{
		ID:          u.ID,
		ClerkUserID: u.ClerkUserID,
		Email:       u.Email,
		Name:        u.Name,
		AvatarURL:   u.AvatarUrl,
		IsActive:    u.IsActive,
		LastSeenAt:  tsPtr(u.LastSeenAt),
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

func tsPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		v := t.Time
		return &v
	}
	return nil
}
