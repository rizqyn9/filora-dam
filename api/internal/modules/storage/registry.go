package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/rizqyn9/filora-dam/api/internal/database/db"
	"github.com/rizqyn9/filora-dam/api/internal/modules/storage/adapters"
)

// Registry creates and caches Adapter instances per storage account.
type Registry struct {
	repo      *Repository
	decryptFn func(ciphertext []byte) ([]byte, error)

	mu    sync.RWMutex
	cache map[int64]Adapter
}

func NewRegistry(repo *Repository, decryptFn func([]byte) ([]byte, error)) *Registry {
	return &Registry{
		repo:      repo,
		decryptFn: decryptFn,
		cache:     make(map[int64]Adapter),
	}
}

// Get returns an Adapter for the given account ID, creating it if needed.
func (r *Registry) Get(ctx context.Context, accountID int64) (Adapter, error) {
	r.mu.RLock()
	if a, ok := r.cache[accountID]; ok {
		r.mu.RUnlock()
		return a, nil
	}
	r.mu.RUnlock()

	// Build adapter
	account, err := r.repo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	adapter, err := r.buildAdapter(account)
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	r.cache[accountID] = adapter
	r.mu.Unlock()

	return adapter, nil
}

// Invalidate removes a cached adapter (e.g. when credentials change).
func (r *Registry) Invalidate(accountID int64) {
	r.mu.Lock()
	delete(r.cache, accountID)
	r.mu.Unlock()
}

func (r *Registry) buildAdapter(account *db.StorageAccount) (Adapter, error) {
	plaintext, err := r.decryptFn(account.CredentialsEncrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt credentials for account %d: %w", account.ID, err)
	}

	switch account.Provider {
	case db.StorageProviderR2:
		var creds adapters.R2Credentials
		if err := json.Unmarshal(plaintext, &creds); err != nil {
			return nil, fmt.Errorf("unmarshal r2 credentials: %w", err)
		}
		return adapters.NewR2Adapter(creds), nil

	case db.StorageProviderCloudinary:
		var creds adapters.CloudinaryCredentials
		if err := json.Unmarshal(plaintext, &creds); err != nil {
			return nil, fmt.Errorf("unmarshal cloudinary credentials: %w", err)
		}
		return adapters.NewCloudinaryAdapter(creds), nil

	case db.StorageProviderImagekit:
		var creds adapters.ImageKitCredentials
		if err := json.Unmarshal(plaintext, &creds); err != nil {
			return nil, fmt.Errorf("unmarshal imagekit credentials: %w", err)
		}
		return adapters.NewImageKitAdapter(creds), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", account.Provider)
	}
}
