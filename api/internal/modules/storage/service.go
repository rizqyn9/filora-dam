package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
)

type Service struct {
	repo *Repository
	// encryptFn encrypts credentials JSON before storage.
	encryptFn func(plaintext []byte) ([]byte, error)
}

func NewService(repo *Repository, encryptFn func([]byte) ([]byte, error)) *Service {
	return &Service{repo: repo, encryptFn: encryptFn}
}

func (s *Service) ListAccounts(ctx context.Context) ([]db.StorageAccount, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) GetAccount(ctx context.Context, id int64) (*db.StorageAccount, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*db.StorageAccount, error) {
	credsJSON, err := json.Marshal(req.Credentials)
	if err != nil {
		return nil, fmt.Errorf("marshal credentials: %w", err)
	}

	encrypted, err := s.encryptFn(credsJSON)
	if err != nil {
		return nil, fmt.Errorf("encrypt credentials: %w", err)
	}

	return s.repo.Create(ctx, db.CreateStorageAccountParams{
		Provider:             db.StorageProvider(req.Provider),
		Label:                req.Label,
		Layer:                db.StorageLayer(req.Layer),
		CredentialsEncrypted: encrypted,
		QuotaBytes:           req.QuotaBytes,
	})
}

func (s *Service) UpdateAccount(ctx context.Context, id int64, req UpdateAccountRequest) error {
	return s.repo.Update(ctx, db.UpdateStorageAccountParams{
		ID:         id,
		Label:      req.Label,
		IsActive:   req.IsActive,
		QuotaBytes: req.QuotaBytes,
	})
}

func (s *Service) DeactivateAccount(ctx context.Context, id int64) error {
	return s.repo.Deactivate(ctx, id)
}

// ElectAccount picks the first active account in the given layer with remaining capacity.
// ponytail: naive first-fit election. Upgrade path: least-used or weighted strategy.
func (s *Service) ElectAccount(ctx context.Context, layer db.StorageLayer, sizeBytes int64) (*db.StorageAccount, error) {
	accounts, err := s.repo.ListActiveByLayer(ctx, layer)
	if err != nil {
		return nil, err
	}

	for i := range accounts {
		a := &accounts[i]
		if a.QuotaBytes == 0 || a.UsedBytes+sizeBytes <= a.QuotaBytes {
			return a, nil
		}
	}

	return nil, fmt.Errorf("no available storage account for layer %s", string(layer))
}
