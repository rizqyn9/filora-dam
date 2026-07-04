package gallery

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
	ErrGalleryNotFound    = errors.New("gallery not found")
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

// CreateGalleryWithOwner creates a gallery and its owner membership in one tx.
func (r *Repository) CreateGalleryWithOwner(ctx context.Context, ownerID int64, name string, desc *string, isDefault bool) (Gallery, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Gallery{}, err
	}
	defer tx.Rollback(ctx)

	q := r.q.WithTx(tx)
	g, err := q.CreateGallery(ctx, db.CreateGalleryParams{
		OwnerID:     ownerID,
		Name:        name,
		Description: desc,
		IsDefault:   isDefault,
	})
	if err != nil {
		return Gallery{}, err
	}
	if err := q.UpsertGalleryMember(ctx, db.UpsertGalleryMemberParams{
		GalleryID: g.ID,
		UserID:    ownerID,
		Role:      db.MemberRoleOwner,
	}); err != nil {
		return Gallery{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Gallery{}, err
	}
	return toGallery(g), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Gallery, error) {
	g, err := r.q.GetGalleryByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Gallery{}, ErrGalleryNotFound
	}
	if err != nil {
		return Gallery{}, err
	}
	return toGallery(g), nil
}

func (r *Repository) GetDefault(ctx context.Context, ownerID int64) (Gallery, error) {
	g, err := r.q.GetDefaultGallery(ctx, ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Gallery{}, ErrGalleryNotFound
	}
	if err != nil {
		return Gallery{}, err
	}
	return toGallery(g), nil
}

func (r *Repository) ListForUser(ctx context.Context, userID int64) ([]Gallery, error) {
	rows, err := r.q.ListGalleriesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Gallery, 0, len(rows))
	for _, g := range rows {
		out = append(out, toGallery(g))
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, id int64, name string, desc *string) (Gallery, error) {
	g, err := r.q.UpdateGallery(ctx, db.UpdateGalleryParams{ID: id, Name: name, Description: desc})
	if errors.Is(err, pgx.ErrNoRows) {
		return Gallery{}, ErrGalleryNotFound
	}
	if err != nil {
		return Gallery{}, err
	}
	return toGallery(g), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteGallery(ctx, id)
}

func (r *Repository) AddUsed(ctx context.Context, id, delta int64) error {
	return r.q.AddGalleryUsed(ctx, db.AddGalleryUsedParams{ID: id, StorageUsed: delta})
}

// --- members ---

func (r *Repository) GetMemberRole(ctx context.Context, galleryID, userID int64) (db.MemberRole, error) {
	m, err := r.q.GetGalleryMember(ctx, db.GetGalleryMemberParams{GalleryID: galleryID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotMember
	}
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (r *Repository) ListMembers(ctx context.Context, galleryID int64) ([]Member, error) {
	rows, err := r.q.ListGalleryMembers(ctx, galleryID)
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

func (r *Repository) UpdateMemberRole(ctx context.Context, galleryID, userID int64, role db.MemberRole) (bool, error) {
	n, err := r.q.UpdateGalleryMemberRole(ctx, db.UpdateGalleryMemberRoleParams{GalleryID: galleryID, UserID: userID, Role: role})
	return n > 0, err
}

func (r *Repository) RemoveMember(ctx context.Context, galleryID, userID int64) (bool, error) {
	n, err := r.q.RemoveGalleryMember(ctx, db.RemoveGalleryMemberParams{GalleryID: galleryID, UserID: userID})
	return n > 0, err
}

// --- invitations ---

func (r *Repository) CreateInvitation(ctx context.Context, galleryID int64, email string, role db.MemberRole, token string, invitedBy *int64, expiresAt *time.Time) (Invitation, error) {
	inv, err := r.q.CreateGalleryInvitation(ctx, db.CreateGalleryInvitationParams{
		GalleryID: &galleryID,
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

func (r *Repository) ListInvitations(ctx context.Context, galleryID int64) ([]Invitation, error) {
	rows, err := r.q.ListGalleryInvitations(ctx, &galleryID)
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
func (r *Repository) AcceptInvitationTx(ctx context.Context, invID, galleryID int64, role db.MemberRole, invitedBy *int64, userID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := r.q.WithTx(tx)
	if err := q.UpsertGalleryMember(ctx, db.UpsertGalleryMemberParams{
		GalleryID: galleryID,
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

func (r *Repository) RevokeInvitation(ctx context.Context, invID, galleryID int64) (bool, error) {
	n, err := r.q.RevokeGalleryInvitation(ctx, db.RevokeGalleryInvitationParams{ID: invID, GalleryID: &galleryID})
	return n > 0, err
}

func toGallery(g db.Gallery) Gallery {
	return Gallery{
		ID:           g.ID,
		OwnerID:      g.OwnerID,
		Name:         g.Name,
		Description:  g.Description,
		IsDefault:    g.IsDefault,
		StorageQuota: g.StorageQuota,
		StorageUsed:  g.StorageUsed,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}

func toInvitation(i db.Invitation) Invitation {
	return Invitation{
		ID:        i.ID,
		GalleryID: i.GalleryID,
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
