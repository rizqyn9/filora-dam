package adapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryAdapter implements StorageAdapter for Cloudinary
type CloudinaryAdapter struct {
	cld       *cloudinary.Cloudinary
	cloudName string
}

// NewCloudinaryAdapter creates a new Cloudinary adapter
func NewCloudinaryAdapter(config *AdapterConfig) (*CloudinaryAdapter, error) {
	cloudName, ok := config.Credentials["cloud_name"].(string)
	if !ok || cloudName == "" {
		return nil, ErrInvalidConfig
	}

	apiKey, ok := config.Credentials["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, ErrInvalidConfig
	}

	apiSecret, ok := config.Credentials["api_secret"].(string)
	if !ok || apiSecret == "" {
		return nil, ErrInvalidConfig
	}

	// Initialize Cloudinary
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &CloudinaryAdapter{
		cld:       cld,
		cloudName: cloudName,
	}, nil
}

// Upload uploads a file to Cloudinary
func (a *CloudinaryAdapter) Upload(ctx context.Context, input *UploadInput) (*UploadResult, error) {
	// Read file into memory (Cloudinary SDK requires it)
	fileBytes, err := io.ReadAll(input.File)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	// Upload to Cloudinary
	uploadParams := uploader.UploadParams{
		PublicID:     input.Filename, // Use filename as public ID
		ResourceType: "auto",         // Auto-detect resource type
		Context: map[string]string{
			"filename":  input.Filename,
			"mime_type": input.MimeType,
		},
	}

	result, err := a.cld.Upload.Upload(ctx, bytes.NewReader(fileBytes), uploadParams)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	metadata := map[string]interface{}{
		"public_id":     result.PublicID,
		"version":       result.Version,
		"format":        result.Format,
		"resource_type": result.ResourceType,
		"width":         result.Width,
		"height":        result.Height,
	}

	return &UploadResult{
		Key:      result.PublicID,
		URL:      result.SecureURL,
		Size:     int64(result.Bytes),
		Metadata: metadata,
	}, nil
}

// Download downloads a file from Cloudinary
func (a *CloudinaryAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// Get URL for the file
	url, err := a.GetURL(ctx, key)
	if err != nil {
		return nil, err
	}

	// Download from URL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%w: status %d", ErrDownloadFailed, resp.StatusCode)
	}

	return resp.Body, nil
}

// Delete deletes a file from Cloudinary
func (a *CloudinaryAdapter) Delete(ctx context.Context, key string) error {
	deleteParams := uploader.DestroyParams{
		PublicID:     key,
		ResourceType: "image", // Default to image, can be enhanced
	}

	_, err := a.cld.Upload.Destroy(ctx, deleteParams)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	return nil
}

// Exists checks if a file exists in Cloudinary
func (a *CloudinaryAdapter) Exists(ctx context.Context, key string) (bool, error) {
	// Try to get the URL, if it succeeds, file exists
	url, err := a.GetURL(ctx, key)
	if err != nil {
		return false, err
	}

	// Check if URL is accessible
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil // Assume doesn't exist on error
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetURL returns the public URL for a Cloudinary file
func (a *CloudinaryAdapter) GetURL(ctx context.Context, key string) (string, error) {
	// Cloudinary URL format: https://res.cloudinary.com/{cloud_name}/image/upload/{public_id}
	url := fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", a.cloudName, key)
	return url, nil
}
