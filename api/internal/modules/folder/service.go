package folder

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetFolder(ctx context.Context, id uuid.UUID) (*db.Folder, error) {
	f, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("folder not found: %w", err)
		}
		return nil, err
	}
	return f, nil
}

func (s *Service) ListFolders(ctx context.Context, spaceID uuid.UUID, parentID *uuid.UUID) ([]db.Folder, error) {
	if parentID == nil {
		return s.repo.ListRoot(ctx, spaceID)
	}
	return s.repo.ListByParent(ctx, spaceID, *parentID)
}

func (s *Service) CreateFolder(ctx context.Context, req CreateFolderRequest) (*db.Folder, error) {
	return s.repo.Create(ctx, req.SpaceID, req.ParentID, req.Name)
}

func (s *Service) RenameFolder(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.Rename(ctx, id, name)
}

func (s *Service) MoveFolder(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) error {
	if parentID != nil {
		// Validate: target parent must not be the folder itself or one of its descendants
		if *parentID == id {
			return fmt.Errorf("cannot move folder into itself")
		}
		ancestors, err := s.repo.GetAncestors(ctx, *parentID)
		if err == nil {
			for _, a := range ancestors {
				if a.ID == id {
					return fmt.Errorf("cannot move folder into its own descendant")
				}
			}
		}
	}
	return s.repo.Move(ctx, id, parentID)
}

func (s *Service) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) RestoreFolder(ctx context.Context, id uuid.UUID) error {
	return s.repo.Restore(ctx, id)
}

func (s *Service) GetBreadcrumbs(ctx context.Context, id uuid.UUID) ([]BreadcrumbItem, error) {
	rows, err := s.repo.GetAncestors(ctx, id)
	if err != nil {
		return nil, err
	}

	items := make([]BreadcrumbItem, len(rows))
	for i, row := range rows {
		items[i] = BreadcrumbItem{ID: row.ID, Name: row.Name}
	}
	return items, nil
}
