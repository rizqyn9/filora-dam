package storage

import (
	"context"
	"errors"

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
