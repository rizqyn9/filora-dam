// Package adapters defines the storage provider abstraction. Business logic
// depends only on StorageAdapter, never on a provider SDK.
//
// Implemented: r2 (Cloudflare R2 / any S3-compatible endpoint) — usable for
// both the serving and archive layers. Still stubbed (ErrNotImplemented):
// cloudinary, imagekit, gcs — added when their SDKs + credentials are wired.
package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ErrNotImplemented is returned by adapter operations not yet wired to a SDK.
var ErrNotImplemented = errors.New("storage adapter operation not implemented")

// UploadInput is the payload for storing an object.
type UploadInput struct {
	Key         string    // logical object key/path
	ContentType string    // MIME type
	Size        int64     // bytes (best effort)
	Reader      io.Reader // object content
}

// UploadResult describes where an object landed.
type UploadResult struct {
	Key      string         // provider object key
	URL      string         // access/public URL (may be empty for archive)
	Metadata map[string]any // provider-specific metadata
}

// StorageAdapter is implemented by every provider.
type StorageAdapter interface {
	Upload(ctx context.Context, in UploadInput) (*UploadResult, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// --- credential shapes ---

type CloudinaryCredentials struct {
	CloudName string `json:"cloud_name"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type ImageKitCredentials struct {
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
	URLEndpoint string `json:"url_endpoint"`
}

type R2Credentials struct {
	AccountID       string `json:"account_id"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BucketName      string `json:"bucket_name"`
	Endpoint        string `json:"endpoint"`
	// PublicBaseURL is optional: when set, uploaded objects get a public URL of
	// "<public_base_url>/<key>" (e.g. an R2 public bucket / custom domain).
	PublicBaseURL string `json:"public_base_url"`
}

type GCSCredentials struct {
	BucketName        string `json:"bucket_name"`
	ServiceAccountKey string `json:"service_account_key"` // JSON key, base64 or raw
}

// ValidateCredentials checks that credentials for a provider type contain the
// required fields.
func ValidateCredentials(providerType string, raw []byte) error {
	switch providerType {
	case "cloudinary":
		var c CloudinaryCredentials
		if err := json.Unmarshal(raw, &c); err != nil {
			return fmt.Errorf("invalid cloudinary credentials: %w", err)
		}
		return requireFields(map[string]string{"cloud_name": c.CloudName, "api_key": c.APIKey, "api_secret": c.APISecret})
	case "imagekit":
		var c ImageKitCredentials
		if err := json.Unmarshal(raw, &c); err != nil {
			return fmt.Errorf("invalid imagekit credentials: %w", err)
		}
		return requireFields(map[string]string{"public_key": c.PublicKey, "private_key": c.PrivateKey, "url_endpoint": c.URLEndpoint})
	case "r2":
		var c R2Credentials
		if err := json.Unmarshal(raw, &c); err != nil {
			return fmt.Errorf("invalid r2 credentials: %w", err)
		}
		return requireFields(map[string]string{"access_key_id": c.AccessKeyID, "secret_access_key": c.SecretAccessKey, "bucket_name": c.BucketName, "endpoint": c.Endpoint})
	case "gcs":
		var c GCSCredentials
		if err := json.Unmarshal(raw, &c); err != nil {
			return fmt.Errorf("invalid gcs credentials: %w", err)
		}
		return requireFields(map[string]string{"bucket_name": c.BucketName, "service_account_key": c.ServiceAccountKey})
	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// NewAdapter builds the adapter for a provider type from its raw credentials.
func NewAdapter(providerType string, raw []byte) (StorageAdapter, error) {
	if err := ValidateCredentials(providerType, raw); err != nil {
		return nil, err
	}
	switch providerType {
	case "cloudinary":
		var c CloudinaryCredentials
		_ = json.Unmarshal(raw, &c)
		return &cloudinaryAdapter{creds: c}, nil
	case "imagekit":
		var c ImageKitCredentials
		_ = json.Unmarshal(raw, &c)
		return &imagekitAdapter{creds: c}, nil
	case "r2":
		var c R2Credentials
		_ = json.Unmarshal(raw, &c)
		return newR2Adapter(c), nil
	case "gcs":
		var c GCSCredentials
		_ = json.Unmarshal(raw, &c)
		return &gcsAdapter{creds: c}, nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

func requireFields(fields map[string]string) error {
	for name, val := range fields {
		if val == "" {
			return fmt.Errorf("missing credential field: %s", name)
		}
	}
	return nil
}
