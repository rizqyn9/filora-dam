package space

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

var (
	ErrSpaceNotFound      = errors.New("space not found")
	ErrNotMember          = errors.New("not a member")
	ErrInvitationNotFound = errors.New("invitation not found")
)

type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

// CreateSpaceWithOwner creates a space and its owner membership in one tx.
func (r *Repository) CreateSpaceWithOwner(ctx context.Context, ownerID int64, name string, desc *string, isDefault bool) (Space, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Space{}, err
	}
	defer tx.Rollback(ctx)

	q := r.q.WithTx(tx)
	s, err := q.CreateSpace(ctx, db.CreateSpaceParams{
		OwnerID:     ownerID,
		Name:        name,
		Description: desc,
		IsDefault:   isDefault,
	})
	if err != nil {
		return Space{}, err
	}
	if err := q.UpsertSpaceMember(ctx, db.UpsertSpaceMemberParams{
		SpaceID: s.ID,
		UserID:  ownerID,
		Role:    db.MemberRoleOwner,
	}); err != nil {
		return Space{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Space{}, err
	}
	return toSpace(s), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Space, error) {
	s, err := r.q.GetSpaceByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Space{}, ErrSpaceNotFound
	}
	if err != nil {
		return Space{}, err
	}
	return toSpace(s), nil
}

func (r *Repository) GetDefault(ctx context.Context, ownerID int64) (Space, error) {
	s, err := r.q.GetDefaultSpace(ctx, ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Space{}, ErrSpaceNotFound
	}
	if err != nil {
		return Space{}, err
	}
	return toSpace(s), nil
}

func (r *Repository) ListForUser(ctx context.Context, userID int64) ([]Space, error) {
	rows, err := r.q.ListSpacesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Space, 0, len(rows))
	for _, s := range rows {
		out = append(out, toSpace(s))
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, id int64, name string, desc *string) (Space, error) {
	s, err := r.q.UpdateSpace(ctx, db.UpdateSpaceParams{ID: id, Name: name, Description: desc})
	if errors.Is(err, pgx.ErrNoRows) {
		return Space{}, ErrSpaceNotFound
	}
	if err != nil {
		return Space{}, err
	}
	return toSpace(s), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteSpace(ctx, id)
}

func (r *Repository) AddUsed(ctx context.Context, id, delta int64) error {
	return r.q.AddSpaceUsed(ctx, db.AddSpaceUsedParams{ID: id, StorageUsed: delta})
}

// --- members ---

func (r *Repository) GetMemberRole(ctx context.Context, spaceID, userID int64) (db.MemberRole, error) {
	m, err := r.q.GetSpaceMember(ctx, db.GetSpaceMemberParams{SpaceID: spaceID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotMember
	}
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (r *Repository) ListMembers(ctx context.Context, spaceID int64) ([]Member, error) {
	rows, err := r.q.ListSpaceMembers(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	out := make([]Member, 0, len(rows))
	for _, m := range rows {
		out = append(out, Member{
			UserID:    m.UserID,
			Role:      string(m.Role),
			Email:     m.Email,
			Name:      m.Name,
			AvatarURL: m.AvatarUrl,
			CreatedAt: m.CreatedAt,
		})
	}
	return out, nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, spaceID, userID int64, role db.MemberRole) (bool, error) {
	n, err := r.q.UpdateSpaceMemberRole(ctx, db.UpdateSpaceMemberRoleParams{SpaceID: spaceID, UserID: userID, Role: role})
	return n > 0, err
}

func (r *Repository) RemoveMember(ctx context.Context, spaceID, userID int64) (bool, error) {
	n, err := r.q.RemoveSpaceMember(ctx, db.RemoveSpaceMemberParams{SpaceID: spaceID, UserID: userID})
	return n > 0, err
}

// --- invitations ---

func (r *Repository) CreateInvitation(ctx context.Context, spaceID int64, email string, role db.MemberRole, token string, invitedBy *int64, expiresAt *time.Time) (Invitation, error) {
	inv, err := r.q.CreateSpaceInvitation(ctx, db.CreateSpaceInvitationParams{
		SpaceID:   spaceID,
		Email:     email,
		Role:      role,
		Token:     token,
		InvitedBy: invitedBy,
		ExpiresAt: tsFromPtr(expiresAt),
	})
	if err != nil {
		return Invitation{}, err
	}
	return toInvitation(inv), nil
}

func (r *Repository) GetInvitationByToken(ctx context.Context, token string) (Invitation, error) {
	inv, err := r.q.GetInvitationByToken(ctx, token)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invitation{}, ErrInvitationNotFound
	}
	if err != nil {
		return Invitation{}, err
	}
	return toInvitation(inv), nil
}

func (r *Repository) ListInvitations(ctx context.Context, spaceID int64) ([]Invitation, error) {
	rows, err := r.q.ListSpaceInvitations(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	out := make([]Invitation, 0, len(rows))
	for _, inv := range rows {
		out = append(out, toInvitation(inv))
	}
	return out, nil
}

// AcceptInvitationTx adds the member and marks the invitation accepted atomically.
func (r *Repository) AcceptInvitationTx(ctx context.Context, invID, spaceID int64, role db.MemberRole, invitedBy *int64, userID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := r.q.WithTx(tx)
	if err := q.UpsertSpaceMember(ctx, db.UpsertSpaceMemberParams{
		SpaceID:   spaceID,
		UserID:    userID,
		Role:      role,
		InvitedBy: invitedBy,
	}); err != nil {
		return err
	}
	if err := q.MarkInvitationAccepted(ctx, db.MarkInvitationAcceptedParams{ID: invID, AcceptedUserID: &userID}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) RevokeInvitation(ctx context.Context, invID, spaceID int64) (bool, error) {
	n, err := r.q.RevokeSpaceInvitation(ctx, db.RevokeSpaceInvitationParams{ID: invID, SpaceID: spaceID})
	return n > 0, err
}

func toSpace(s db.Space) Space {
	return Space{
		ID:           s.ID,
		OwnerID:      s.OwnerID,
		Name:         s.Name,
		Description:  s.Description,
		IsDefault:    s.IsDefault,
		StorageQuota: s.StorageQuota,
		StorageUsed:  s.StorageUsed,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

func toInvitation(i db.Invitation) Invitation {
	return Invitation{
		ID:        i.ID,
		SpaceID:   i.SpaceID,
		Email:     i.Email,
		Role:      string(i.Role),
		Token:     i.Token,
		Status:    string(i.Status),
		ExpiresAt: tsPtr(i.ExpiresAt),
		CreatedAt: i.CreatedAt,
	}
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
