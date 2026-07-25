package folder

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rizqynugroho9/filora-dam/api/internal/database/db"
)

var ErrFolderNotFound = errors.New("folder not found")

type Repository struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, spaceID int64, parentID *int64, ownerID int64, name, path string) (Folder, error) {
	f, err := r.q.CreateFolder(ctx, db.CreateFolderParams{
		SpaceID:  spaceID,
		ParentID: parentID,
		OwnerID:  ownerID,
		Name:     name,
		Path:     path,
	})
	if err != nil {
		return Folder{}, err
	}
	return toFolder(f), nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Folder, error) {
	f, err := r.q.GetFolderByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Folder{}, ErrFolderNotFound
	}
	if err != nil {
		return Folder{}, err
	}
	return toFolder(f), nil
}

func (r *Repository) ListChildren(ctx context.Context, spaceID int64, parentID *int64) ([]Folder, error) {
	if parentID == nil {
		rows, err := r.q.ListRootFolders(ctx, spaceID)
		if err != nil {
			return nil, err
		}
		return toFolders(rows), nil
	}
	rows, err := r.q.ListFoldersByParent(ctx, db.ListFoldersByParentParams{
		SpaceID:  spaceID,
		ParentID: parentID,
	})
	if err != nil {
		return nil, err
	}
	return toFolders(rows), nil
}

func (r *Repository) Rename(ctx context.Context, id int64, name string) (Folder, error) {
	f, err := r.q.UpdateFolderName(ctx, db.UpdateFolderNameParams{ID: id, Name: name})
	if errors.Is(err, pgx.ErrNoRows) {
		return Folder{}, ErrFolderNotFound
	}
	if err != nil {
		return Folder{}, err
	}
	return toFolder(f), nil
}

func (r *Repository) Move(ctx context.Context, id int64, newParentID *int64, newPath string) (Folder, error) {
	f, err := r.q.UpdateFolderParent(ctx, db.UpdateFolderParentParams{
		ID:       id,
		ParentID: newParentID,
		Path:     newPath,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Folder{}, ErrFolderNotFound
	}
	if err != nil {
		return Folder{}, err
	}
	return toFolder(f), nil
}

// UpdateSubtreePaths updates the path for all descendants of a folder after it
// has been moved. oldPathPrefix is the old path + id + "/", newPathPrefix is the
// replacement.
func (r *Repository) UpdateSubtreePaths(ctx context.Context, spaceID int64, oldPathPrefix, newPathPrefix string) error {
	// Get all descendants matching old path prefix.
	descendants, err := r.q.ListFolderSubtree(ctx, db.ListFolderSubtreeParams{
		SpaceID: spaceID,
		Path:    oldPathPrefix + "%",
	})
	if err != nil {
		return err
	}
	for _, d := range descendants {
		newPath := strings.Replace(d.Path, oldPathPrefix, newPathPrefix, 1)
		if err := r.q.UpdateFolderPath(ctx, db.UpdateFolderPathParams{ID: d.ID, Path: newPath}); err != nil {
			return fmt.Errorf("update path for folder %d: %w", d.ID, err)
		}
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	return r.q.DeleteFolder(ctx, id)
}

// GetBreadcrumb returns the ancestor chain for a folder based on its path field.
func (r *Repository) GetBreadcrumb(ctx context.Context, path string) ([]BreadcrumbItem, error) {
	ids := parsePathIDs(path)
	if len(ids) == 0 {
		return []BreadcrumbItem{}, nil
	}
	rows, err := r.q.GetFolderBreadcrumb(ctx, ids)
	if err != nil {
		return nil, err
	}
	items := make([]BreadcrumbItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, BreadcrumbItem{ID: row.ID, Name: row.Name})
	}
	return items, nil
}

// --- helpers ---

func toFolder(f db.Folder) Folder {
	return Folder{
		ID:        f.ID,
		SpaceID:   f.SpaceID,
		ParentID:  f.ParentID,
		OwnerID:   f.OwnerID,
		Name:      f.Name,
		Path:      f.Path,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func toFolders(rows []db.Folder) []Folder {
	out := make([]Folder, 0, len(rows))
	for _, f := range rows {
		out = append(out, toFolder(f))
	}
	return out
}

// parsePathIDs extracts ancestor folder IDs from a materialized path.
// E.g. "/5/12/" → [5, 12].
func parsePathIDs(path string) []int64 {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}
