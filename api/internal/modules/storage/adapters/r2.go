package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

// r2Adapter stores objects in a Cloudflare R2 (or any S3-compatible) bucket.
// Suitable for either layer; use PublicBaseURL for a servable public URL.
type r2Adapter struct {
	creds  R2Credentials
	client *s3.Client
	bucket string
}

func newR2Adapter(c R2Credentials) *r2Adapter {
	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(c.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(c.AccessKeyID, c.SecretAccessKey, ""),
		UsePathStyle: true,
	})
	return &r2Adapter{creds: c, client: client, bucket: c.BucketName}
}

func (a *r2Adapter) Upload(ctx context.Context, in UploadInput) (*UploadResult, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(in.Key),
		Body:   in.Reader,
	}
	if in.ContentType != "" {
		input.ContentType = aws.String(in.ContentType)
	}
	if in.Size > 0 {
		input.ContentLength = aws.Int64(in.Size)
	}
	if _, err := a.client.PutObject(ctx, input); err != nil {
		return nil, fmt.Errorf("r2 put object: %w", err)
	}
	return &UploadResult{Key: in.Key, URL: a.publicURL(in.Key)}, nil
}

func (a *r2Adapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := a.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("r2 get object: %w", err)
	}
	return out.Body, nil
}

func (a *r2Adapter) Delete(ctx context.Context, key string) error {
	if _, err := a.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("r2 delete object: %w", err)
	}
	return nil
}

func (a *r2Adapter) Exists(ctx context.Context, key string) (bool, error) {
	_, err := a.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NotFound", "NoSuchKey":
			return false, nil
		}
	}
	return false, fmt.Errorf("r2 head object: %w", err)
}

func (a *r2Adapter) publicURL(key string) string {
	if a.creds.PublicBaseURL == "" {
		return ""
	}
	return strings.TrimRight(a.creds.PublicBaseURL, "/") + "/" + key
}
