package adapters

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ImageKitAdapter implements StorageAdapter for ImageKit
type ImageKitAdapter struct {
	publicKey   string
	privateKey  string
	urlEndpoint string
}

// NewImageKitAdapter creates a new ImageKit adapter
func NewImageKitAdapter(config *AdapterConfig) (*ImageKitAdapter, error) {
	publicKey, ok := config.Credentials["public_key"].(string)
	if !ok || publicKey == "" {
		return nil, ErrInvalidConfig
	}

	privateKey, ok := config.Credentials["private_key"].(string)
	if !ok || privateKey == "" {
		return nil, ErrInvalidConfig
	}

	urlEndpoint, ok := config.Credentials["url_endpoint"].(string)
	if !ok || urlEndpoint == "" {
		return nil, ErrInvalidConfig
	}

	return &ImageKitAdapter{
		publicKey:   publicKey,
		privateKey:  privateKey,
		urlEndpoint: urlEndpoint,
	}, nil
}

type imageKitUploadResponse struct {
	FileID   string `json:"fileId"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FilePath string `json:"filePath"`
}

// Upload uploads a file to ImageKit
func (a *ImageKitAdapter) Upload(ctx context.Context, input *UploadInput) (*UploadResult, error) {
	fileBytes, err := io.ReadAll(input.File)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	// ImageKit expects base64 encoded file
	encoded := base64.StdEncoding.EncodeToString(fileBytes)

	// Build form data
	form := fmt.Sprintf(
		"file=%s&fileName=%s&isPrivateFile=false",
		encoded,
		input.Filename,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		"https://upload.imagekit.io/api/v1/files/upload",
		strings.NewReader(form),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.privateKey+":")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: status %d - %s", ErrUploadFailed, resp.StatusCode, string(body))
	}

	var uploadResp imageKitUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	metadata := map[string]interface{}{
		"file_id":  uploadResp.FileID,
		"width":    uploadResp.Width,
		"height":   uploadResp.Height,
		"file_path": uploadResp.FilePath,
	}

	return &UploadResult{
		Key:      uploadResp.FileID,
		URL:      uploadResp.URL,
		Size:     uploadResp.Size,
		Metadata: metadata,
	}, nil
}

// Download downloads a file from ImageKit
func (a *ImageKitAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	url, err := a.GetURL(ctx, key)
	if err != nil {
		return nil, err
	}

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

type imageKitDeleteResponse struct {
	Succeed bool `json:"succeed"`
}

// Delete deletes a file from ImageKit
func (a *ImageKitAdapter) Delete(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(
		ctx,
		"DELETE",
		fmt.Sprintf("https://api.imagekit.io/v1/files/%s", key),
		nil,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.privateKey+":")))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d - %s", ErrDeleteFailed, resp.StatusCode, string(body))
	}

	return nil
}

// Exists checks if a file exists in ImageKit
func (a *ImageKitAdapter) Exists(ctx context.Context, key string) (bool, error) {
	url, err := a.GetURL(ctx, key)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// GetURL returns the public URL for an ImageKit file
func (a *ImageKitAdapter) GetURL(ctx context.Context, key string) (string, error) {
	// ImageKit URL format: https://{urlEndpoint}/ik-{fileId}
	url := fmt.Sprintf("https://%s/ik-%s", a.urlEndpoint, key)
	return url, nil
}
