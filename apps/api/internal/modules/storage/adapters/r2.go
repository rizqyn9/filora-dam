package adapters

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
)

// R2Adapter implements StorageAdapter for Cloudflare R2
type R2Adapter struct {
	client     *s3.Client
	bucketName string
	endpoint   string
	accountID  string
}

// NewR2Adapter creates a new Cloudflare R2 adapter
func NewR2Adapter(config *AdapterConfig) (*R2Adapter, error) {
	accountID, ok := config.Credentials["account_id"].(string)
	if !ok || accountID == "" {
		return nil, ErrInvalidConfig
	}

	accessKeyID, ok := config.Credentials["access_key_id"].(string)
	if !ok || accessKeyID == "" {
		return nil, ErrInvalidConfig
	}

	secretAccessKey, ok := config.Credentials["secret_access_key"].(string)
	if !ok || secretAccessKey == "" {
		return nil, ErrInvalidConfig
	}

	bucketName, ok := config.Credentials["bucket_name"].(string)
	if !ok || bucketName == "" {
		return nil, ErrInvalidConfig
	}

	endpoint, ok := config.Credentials["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, ErrInvalidConfig
	}

	// Create AWS credentials
	creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")

	// Create S3 client configured for R2
	client := s3.NewFromConfig(aws.Config{
		Region:      "auto",
		Credentials: creds,
		BaseEndpoint: aws.String(endpoint),
		ClientLogMode: aws.LogRetries | aws.LogRequestWithBody,
		Logger: logging.NewStandardLogger(nil),
	})

	return &R2Adapter{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		accountID:  accountID,
	}, nil
}

// Upload uploads a file to R2
func (a *R2Adapter) Upload(ctx context.Context, input *UploadInput) (*UploadResult, error) {
	result, err := a.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(a.bucketName),
		Key:         aws.String(input.Filename),
		Body:        input.File,
		ContentType: aws.String(input.MimeType),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	metadata := map[string]interface{}{
		"etag":      result.ETag,
		"version":   result.VersionId,
	}

	url, _ := a.GetURL(ctx, input.Filename)

	return &UploadResult{
		Key:      input.Filename,
		URL:      url,
		Size:     input.Size,
		Metadata: metadata,
	}, nil
}

// Download downloads a file from R2
func (a *R2Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	return result.Body, nil
}

// Delete deletes a file from R2
func (a *R2Adapter) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	return nil
}

// Exists checks if a file exists in R2
func (a *R2Adapter) Exists(ctx context.Context, key string) (bool, error) {
	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}

	return true, nil
}

// GetURL returns the public URL for an R2 file
func (a *R2Adapter) GetURL(ctx context.Context, key string) (string, error) {
	// R2 public URL format: https://{bucket}.{account-id}.r2.cloudflarestorage.com/{key}
	// Or if custom domain: https://{custom-domain}/{key}
	url := fmt.Sprintf("https://%s.%s.r2.cloudflarestorage.com/%s", a.bucketName, a.accountID, key)
	return url, nil
}
