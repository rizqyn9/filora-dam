package storage

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"

	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/storage/adapters"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, in CreateProviderInput, createdBy *int64) (Provider, error) {
	if err := lib.Validate(in); err != nil {
		return Provider{}, err
	}
	// Validate credentials shape for the provider type early.
	creds, err := marshalCredentials(in.Credentials)
	if err != nil {
		return Provider{}, lib.ErrBadRequest("invalid credentials")
	}
	if err := adapters.ValidateCredentials(in.Type, creds); err != nil {
		return Provider{}, lib.ErrBadRequest(err.Error())
	}
	return s.repo.Create(ctx, in.Layer, in.Name, in.Type, creds, in.Quota, createdBy)
}

func (s *Service) Get(ctx context.Context, id int64) (Provider, error) {
	p, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, ErrProviderNotFound) {
		return Provider{}, lib.ErrNotFound("storage provider not found")
	}
	return p, err
}

func (s *Service) List(ctx context.Context) ([]Provider, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, id int64, in UpdateProviderInput) (Provider, error) {
	if err := lib.Validate(in); err != nil {
		return Provider{}, err
	}
	current, err := s.repo.GetRaw(ctx, id)
	if errors.Is(err, ErrProviderNotFound) {
		return Provider{}, lib.ErrNotFound("storage provider not found")
	}
	if err != nil {
		return Provider{}, err
	}

	creds := current.Credentials
	if in.Credentials != nil {
		b, err := marshalCredentials(in.Credentials)
		if err != nil {
			return Provider{}, lib.ErrBadRequest("invalid credentials")
		}
		if err := adapters.ValidateCredentials(current.Type, b); err != nil {
			return Provider{}, lib.ErrBadRequest(err.Error())
		}
		creds = b
	}

	isActive := current.IsActive
	if in.IsActive != nil {
		isActive = *in.IsActive
	}
	quota := current.Quota
	if in.Quota != nil {
		quota = in.Quota
	}

	return s.repo.Update(ctx, id, in.Name, creds, quota, isActive)
}

func (s *Service) Deactivate(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	return s.repo.Deactivate(ctx, id)
}

func (s *Service) Usage(ctx context.Context) ([]AccountUsage, error) {
	return s.repo.Usage(ctx)
}

// --- orchestration (used by the asset module via a consumer-defined interface) ---

// StoreServing uploads content to an active serving-layer account and records a
// storage_location. Returns the chosen provider id, stored key, and access URL.
func (s *Service) StoreServing(ctx context.Context, assetID uuid.UUID, key, contentType string, size int64, r io.Reader) (int64, string, string, error) {
	accounts, err := s.repo.ListActiveByLayer(ctx, "serving")
	if err != nil {
		return 0, "", "", err
	}
	if len(accounts) == 0 {
		return 0, "", "", lib.NewAppError(503, "NO_SERVING_STORAGE", "no active serving storage account configured")
	}
	// Account election is a backlog concern; use the first active account.
	acct := accounts[0]

	adapter, err := adapters.NewAdapter(acct.Type, acct.Credentials)
	if err != nil {
		return 0, "", "", lib.ErrInternal("storage adapter error").Wrap(err)
	}
	res, err := adapter.Upload(ctx, adapters.UploadInput{
		Key:         key,
		ContentType: contentType,
		Size:        size,
		Reader:      r,
	})
	if err != nil {
		return 0, "", "", err
	}

	var urlPtr *string
	if res.URL != "" {
		urlPtr = &res.URL
	}
	if err := s.repo.CreateLocation(ctx, assetID, acct.ID, "serving", res.Key, urlPtr, "stored"); err != nil {
		return 0, "", "", err
	}
	_ = s.repo.AddUsed(ctx, acct.ID, size)

	return acct.ID, res.Key, res.URL, nil
}

// ServingURL returns the public URL of an asset's serving copy.
func (s *Service) ServingURL(ctx context.Context, assetID uuid.UUID) (string, error) {
	url, ok, err := s.repo.GetServingURL(ctx, assetID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", lib.ErrNotFound("no serving copy for this asset")
	}
	return url, nil
}

// EnqueueArchive schedules replication of an asset to the archive layer.
func (s *Service) EnqueueArchive(ctx context.Context, assetID uuid.UUID) error {
	return s.repo.EnqueueArchive(ctx, assetID)
}
