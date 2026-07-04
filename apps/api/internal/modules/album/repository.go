package album

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

var (
	ErrAlbumNotFound = errors.New("album not found")
	ErrNotMember     = errors.New("not a member")
)

type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

func (r *Repository) CreateWithOwner(ctx context.Context, galleryID, ownerID int64, name string, desc *string) (Album, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Album{}, err
	}
	defer tx.Rollback(ctx)

	q := r.q.WithTx(tx)
	a, err := q.CreateAlbum(ctx, db.CreateAlbumParams{
		GalleryID:   galleryID,
		OwnerID:     ownerID,
		Name:        name,
		Description: desc,
	})
	if err != nil {
		return Album{}, err
	}
	if err := q.UpsertAlbumMember(ctx, db.UpsertAlbumMemberParams{
		AlbumID: a.ID,
		UserID:  ownerID,
		Role:    db.MemberRoleOwner,
	}); err != nil {
		return Album{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Album{}, err
	}
	return toAlbum(a), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Album, error) {
	a, err := r.q.GetAlbumByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Album{}, ErrAlbumNotFound
	}
	if err != nil {
		return Album{}, err
	}
	return toAlbum(a), nil
}

func (r *Repository) ListByGallery(ctx context.Context, galleryID int64) ([]Album, error) {
	rows, err := r.q.ListAlbumsByGallery(ctx, galleryID)
	if err != nil {
		return nil, err
	}
	out := make([]Album, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAlbum(a))
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, id int64, name string, desc, cover *string) (Album, error) {
	coverUUID, err := pgUUIDFromPtr(cover)
	if err != nil {
		return Album{}, err
	}
	a, err := r.q.UpdateAlbum(ctx, db.UpdateAlbumParams{
		ID:           id,
		Name:         name,
		Description:  desc,
		CoverAssetID: coverUUID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Album{}, ErrAlbumNotFound
	}
	if err != nil {
		return Album{}, err
	}
	return toAlbum(a), nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteAlbum(ctx, id)
}

func (r *Repository) GetMemberRole(ctx context.Context, albumID, userID int64) (db.MemberRole, error) {
	m, err := r.q.GetAlbumMember(ctx, db.GetAlbumMemberParams{AlbumID: albumID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotMember
	}
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func (r *Repository) UpsertMember(ctx context.Context, albumID, userID int64, role db.MemberRole, invitedBy *int64) error {
	return r.q.UpsertAlbumMember(ctx, db.UpsertAlbumMemberParams{
		AlbumID: albumID, UserID: userID, Role: role, InvitedBy: invitedBy,
	})
}

func (r *Repository) ListMembers(ctx context.Context, albumID int64) ([]Member, error) {
	rows, err := r.q.ListAlbumMembers(ctx, albumID)
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

func (r *Repository) RemoveMember(ctx context.Context, albumID, userID int64) (bool, error) {
	n, err := r.q.RemoveAlbumMember(ctx, db.RemoveAlbumMemberParams{AlbumID: albumID, UserID: userID})
	return n > 0, err
}

func (r *Repository) AddAsset(ctx context.Context, albumID int64, assetID uuid.UUID, addedBy *int64, sort int32) error {
	return r.q.AddAssetToAlbum(ctx, db.AddAssetToAlbumParams{
		AlbumID: albumID, AssetID: assetID, AddedBy: addedBy, SortOrder: sort,
	})
}

func (r *Repository) RemoveAsset(ctx context.Context, albumID int64, assetID uuid.UUID) (bool, error) {
	n, err := r.q.RemoveAssetFromAlbum(ctx, db.RemoveAssetFromAlbumParams{AlbumID: albumID, AssetID: assetID})
	return n > 0, err
}

func (r *Repository) ListAssetIDs(ctx context.Context, albumID int64) ([]uuid.UUID, error) {
	return r.q.ListAlbumAssetIDs(ctx, albumID)
}

func toAlbum(a db.Album) Album {
	return Album{
		ID:           a.ID,
		GalleryID:    a.GalleryID,
		OwnerID:      a.OwnerID,
		Name:         a.Name,
		Description:  a.Description,
		CoverAssetID: uuidPtr(a.CoverAssetID),
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}
}

func pgUUIDFromPtr(s *string) (pgtype.UUID, error) {
	if s == nil || *s == "" {
		return pgtype.UUID{}, nil
	}
	u, err := uuid.Parse(*s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: u, Valid: true}, nil
}

func uuidPtr(p pgtype.UUID) *string {
	if !p.Valid {
		return nil
	}
	s := uuid.UUID(p.Bytes).String()
	return &s
}
