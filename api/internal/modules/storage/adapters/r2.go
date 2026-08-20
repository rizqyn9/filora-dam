package adapters

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Credentials are the fields expected in the encrypted credentials JSON.
type R2Credentials struct {
	AccountID   string `json:"account_id"`
	AccessKeyID string `json:"access_key_id"`
	SecretKey   string `json:"secret_key"`
	Bucket      string `json:"bucket"`
	PublicURL   string `json:"public_url"` // optional, for serving layer
}

// R2Adapter implements Adapter for Cloudflare R2 (S3-compatible).
type R2Adapter struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Adapter(creds R2Credentials) *R2Adapter {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", creds.AccountID)

	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: &endpoint,
		Credentials:  credentials.NewStaticCredentialsProvider(creds.AccessKeyID, creds.SecretKey, ""),
	})

	return &R2Adapter{
		client:    client,
		bucket:    creds.Bucket,
		publicURL: creds.PublicURL,
	}
}

func (a *R2Adapter) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	_, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &input.Key,
		Body:        input.Body,
		ContentType: &input.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("r2 upload: %w", err)
	}

	var remoteURL string
	if a.publicURL != "" {
		remoteURL = fmt.Sprintf("%s/%s", a.publicURL, input.Key)
	}

	return &UploadResult{
		RemotePath: input.Key,
		RemoteURL:  remoteURL,
	}, nil
}

func (a *R2Adapter) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &a.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("r2 delete: %w", err)
	}
	return nil
}

func (a *R2Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &a.bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("r2 download: %w", err)
	}
	return out.Body, nil
}

// Ensure R2Adapter implements the Adapter interface.
var _ Adapter = (*R2Adapter)(nil)
