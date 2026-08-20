package folder

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*db.Folder, error) {
	f, err := r.q.GetFolderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get folder by id: %w", err)
	}
	return &f, nil
}

func (r *Repository) ListByParent(ctx context.Context, spaceID, parentID uuid.UUID) ([]db.Folder, error) {
	folders, err := r.q.ListFoldersByParent(ctx, db.ListFoldersByParentParams{
		SpaceID:  spaceID,
		ParentID: uuidToPgtype(parentID),
	})
	if err != nil {
		return nil, fmt.Errorf("list folders by parent: %w", err)
	}
	return folders, nil
}

func (r *Repository) ListRoot(ctx context.Context, spaceID uuid.UUID) ([]db.Folder, error) {
	folders, err := r.q.ListRootFolders(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("list root folders: %w", err)
	}
	return folders, nil
}

func (r *Repository) Create(ctx context.Context, spaceID uuid.UUID, parentID *uuid.UUID, name string) (*db.Folder, error) {
	f, err := r.q.CreateFolder(ctx, db.CreateFolderParams{
		SpaceID:  spaceID,
		ParentID: uuidPtrToPgtype(parentID),
		Name:     name,
	})
	if err != nil {
		return nil, fmt.Errorf("create folder: %w", err)
	}
	return &f, nil
}

func (r *Repository) Rename(ctx context.Context, id uuid.UUID, name string) error {
	if err := r.q.RenameFolder(ctx, db.RenameFolderParams{ID: id, Name: name}); err != nil {
		return fmt.Errorf("rename folder: %w", err)
	}
	return nil
}

func (r *Repository) Move(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) error {
	if err := r.q.MoveFolder(ctx, db.MoveFolderParams{ID: id, ParentID: uuidPtrToPgtype(parentID)}); err != nil {
		return fmt.Errorf("move folder: %w", err)
	}
	return nil
}

func (r *Repository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.SoftDeleteFolder(ctx, id); err != nil {
		return fmt.Errorf("soft delete folder: %w", err)
	}
	return nil
}

func (r *Repository) Restore(ctx context.Context, id uuid.UUID) error {
	if err := r.q.RestoreFolder(ctx, id); err != nil {
		return fmt.Errorf("restore folder: %w", err)
	}
	return nil
}

func (r *Repository) GetAncestors(ctx context.Context, id uuid.UUID) ([]db.GetFolderAncestorsRow, error) {
	rows, err := r.q.GetFolderAncestors(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get folder ancestors: %w", err)
	}
	return rows, nil
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidPtrToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}
