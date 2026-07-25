package asset

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Spaces exposes parent-space membership and quota (implemented by space.Service).
type Spaces interface {
	RoleOf(ctx context.Context, spaceID, userID int64) (string, bool, error)
	QuotaInfo(ctx context.Context, spaceID int64) (used, quota int64, err error)
	AddUsed(ctx context.Context, spaceID, delta int64) error
}

// StorageService orchestrates physical storage (implemented by storage.Service).
type StorageService interface {
	StoreServing(ctx context.Context, assetID uuid.UUID, key, contentType string, size int64, r io.Reader) (providerID int64, storedKey, url string, err error)
	ServingURL(ctx context.Context, assetID uuid.UUID) (string, error)
	EnqueueArchive(ctx context.Context, assetID uuid.UUID) error
}

type Service struct {
	repo    *Repository
	authz   *auth.Authorizer
	spaces  Spaces
	storage StorageService
}

func NewService(repo *Repository, authz *auth.Authorizer, spaces Spaces, storage StorageService) *Service {
	return &Service{repo: repo, authz: authz, spaces: spaces, storage: storage}
}

// Upload stores a new asset on the serving layer (with per-space dedup and
// quota), then schedules archive replication. Duplicate content returns the
// existing asset.
func (s *Service) Upload(ctx context.Context, userID int64, in UploadInput) (Asset, error) {
	if err := s.access(ctx, userID, in.SpaceID, "create", rankEditor); err != nil {
		return Asset{}, err
	}

	// Dedup within the space.
	if existing, err := s.repo.GetActiveByHash(ctx, in.SpaceID, in.Hash); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrAssetNotFound) {
		return Asset{}, err
	}

	// Quota check.
	used, quota, err := s.spaces.QuotaInfo(ctx, in.SpaceID)
	if err != nil {
		return Asset{}, err
	}
	if quota > 0 && used+in.Size > quota {
		return Asset{}, lib.NewAppError(507, "INSUFFICIENT_STORAGE", "space storage quota exceeded")
	}

	a, err := s.repo.Create(ctx, in.SpaceID, in.FolderID, &userID, in.Name, in.Type, in.MimeType, in.Size, in.Hash)
	if err != nil {
		return Asset{}, err
	}

	key := buildKey(in.SpaceID, a.ID, in.Name)
	if _, _, _, err := s.storage.StoreServing(ctx, a.ID, key, in.MimeType, in.Size, in.Reader); err != nil {
		// Roll back the orphaned asset row on upload failure.
		_ = s.repo.HardDelete(ctx, a.ID)
		return Asset{}, err
	}

	_ = s.spaces.AddUsed(ctx, in.SpaceID, in.Size)
	_ = s.storage.EnqueueArchive(ctx, a.ID)

	return a, nil
}

func (s *Service) List(ctx context.Context, userID, spaceID int64, folderID *int64, page lib.Page) (ListResult, error) {
	if err := s.access(ctx, userID, spaceID, "read", rankViewer); err != nil {
		return ListResult{}, err
	}

	var assets []Asset
	var total int64
	var err error

	if folderID != nil {
		assets, err = s.repo.ListByFolder(ctx, spaceID, folderID, int32(page.Limit), int32(page.Offset))
		if err != nil {
			return ListResult{}, err
		}
		total, err = s.repo.CountByFolder(ctx, spaceID, folderID)
	} else {
		assets, err = s.repo.ListRootAssets(ctx, spaceID, int32(page.Limit), int32(page.Offset))
		if err != nil {
			return ListResult{}, err
		}
		total, err = s.repo.CountRoot(ctx, spaceID)
	}
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Assets: assets, Total: total, Limit: page.Limit, Offset: page.Offset}, nil
}

func (s *Service) ListAll(ctx context.Context, userID, spaceID int64, page lib.Page) (ListResult, error) {
	if err := s.access(ctx, userID, spaceID, "read", rankViewer); err != nil {
		return ListResult{}, err
	}
	assets, err := s.repo.ListActive(ctx, spaceID, int32(page.Limit), int32(page.Offset))
	if err != nil {
		return ListResult{}, err
	}
	total, err := s.repo.CountActive(ctx, spaceID)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Assets: assets, Total: total, Limit: page.Limit, Offset: page.Offset}, nil
}

func (s *Service) Search(ctx context.Context, userID, spaceID int64, query string, page lib.Page) (ListResult, error) {
	if err := s.access(ctx, userID, spaceID, "read", rankViewer); err != nil {
		return ListResult{}, err
	}
	assets, err := s.repo.Search(ctx, spaceID, "%"+query+"%", int32(page.Limit), int32(page.Offset))
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Assets: assets, Total: int64(len(assets)), Limit: page.Limit, Offset: page.Offset}, nil
}

func (s *Service) FilterByType(ctx context.Context, userID, spaceID int64, typ string, page lib.Page) (ListResult, error) {
	if err := s.access(ctx, userID, spaceID, "read", rankViewer); err != nil {
		return ListResult{}, err
	}
	assets, err := s.repo.FilterByType(ctx, spaceID, typ, int32(page.Limit), int32(page.Offset))
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Assets: assets, Total: int64(len(assets)), Limit: page.Limit, Offset: page.Offset}, nil
}

func (s *Service) Trash(ctx context.Context, userID, spaceID int64, page lib.Page) (ListResult, error) {
	if err := s.access(ctx, userID, spaceID, "read", rankEditor); err != nil {
		return ListResult{}, err
	}
	assets, err := s.repo.ListTrashed(ctx, spaceID, int32(page.Limit), int32(page.Offset))
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Assets: assets, Total: int64(len(assets)), Limit: page.Limit, Offset: page.Offset}, nil
}

func (s *Service) Get(ctx context.Context, userID int64, id uuid.UUID) (Asset, error) {
	a, err := s.load(ctx, id)
	if err != nil {
		return Asset{}, err
	}
	if err := s.access(ctx, userID, a.SpaceID, "read", rankViewer); err != nil {
		return Asset{}, err
	}
	return a, nil
}

// DownloadURL returns the serving URL for an asset the user may read.
func (s *Service) DownloadURL(ctx context.Context, userID int64, id uuid.UUID) (string, error) {
	a, err := s.load(ctx, id)
	if err != nil {
		return "", err
	}
	if err := s.access(ctx, userID, a.SpaceID, "download", rankViewer); err != nil {
		return "", err
	}
	return s.storage.ServingURL(ctx, id)
}

func (s *Service) UpdateName(ctx context.Context, userID int64, id uuid.UUID, in UpdateAssetInput) (Asset, error) {
	if err := lib.Validate(in); err != nil {
		return Asset{}, err
	}
	a, err := s.load(ctx, id)
	if err != nil {
		return Asset{}, err
	}
	if err := s.access(ctx, userID, a.SpaceID, "update", rankEditor); err != nil {
		return Asset{}, err
	}
	return s.repo.UpdateName(ctx, id, in.Name)
}

func (s *Service) Move(ctx context.Context, userID int64, id uuid.UUID, in MoveAssetInput) (Asset, error) {
	a, err := s.load(ctx, id)
	if err != nil {
		return Asset{}, err
	}
	if err := s.access(ctx, userID, a.SpaceID, "update", rankEditor); err != nil {
		return Asset{}, err
	}
	return s.repo.MoveToFolder(ctx, id, in.FolderID)
}

func (s *Service) Delete(ctx context.Context, userID int64, id uuid.UUID) error {
	a, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if err := s.access(ctx, userID, a.SpaceID, "delete", rankEditor); err != nil {
		return err
	}
	ok, err := s.repo.SoftDelete(ctx, id, &userID)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("asset not found")
	}
	return nil
}

func (s *Service) Restore(ctx context.Context, userID int64, id uuid.UUID) error {
	a, err := s.load(ctx, id)
	if err != nil {
		return err
	}
	if err := s.access(ctx, userID, a.SpaceID, "update", rankEditor); err != nil {
		return err
	}
	ok, err := s.repo.Restore(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return lib.ErrNotFound("asset not in trash")
	}
	return nil
}

// --- helpers ---

func (s *Service) load(ctx context.Context, id uuid.UUID) (Asset, error) {
	a, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, ErrAssetNotFound) {
		return Asset{}, lib.ErrNotFound("asset not found")
	}
	return a, err
}

func (s *Service) access(ctx context.Context, userID, spaceID int64, action string, minRank int) error {
	dec, err := s.authz.Authorize(ctx, userID, "asset", action)
	if err != nil {
		return err
	}
	if !dec.Allowed {
		return lib.ErrForbidden("insufficient permission: asset:" + action)
	}
	if dec.Scope == auth.ScopeAll {
		return nil
	}
	role, ok, err := s.spaces.RoleOf(ctx, spaceID, userID)
	if err != nil {
		return err
	}
	if !ok || rank(role) < minRank {
		return lib.ErrForbidden("insufficient access to this space")
	}
	return nil
}

func buildKey(spaceID int64, assetID uuid.UUID, name string) string {
	ext := filepath.Ext(name)
	return "spaces/" + strconv.FormatInt(spaceID, 10) + "/" + assetID.String() + ext
}

const (
	rankViewer = 1
	rankEditor = 2
	rankOwner  = 3
)

func rank(role string) int {
	switch role {
	case "owner":
		return rankOwner
	case "editor":
		return rankEditor
	case "viewer":
		return rankViewer
	}
	return 0
}
