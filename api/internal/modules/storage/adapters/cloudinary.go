package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type CloudinaryCredentials struct {
	CloudName string `json:"cloud_name"`
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type CloudinaryAdapter struct {
	creds CloudinaryCredentials
}

func NewCloudinaryAdapter(creds CloudinaryCredentials) *CloudinaryAdapter {
	return &CloudinaryAdapter{creds: creds}
}

func (a *CloudinaryAdapter) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/raw/upload", a.creds.CloudName)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("public_id", input.Key)
	_ = w.WriteField("resource_type", "raw")

	part, err := w.CreateFormFile("file", input.Key)
	if err != nil {
		return nil, fmt.Errorf("cloudinary create form: %w", err)
	}
	if _, err := io.Copy(part, input.Body); err != nil {
		return nil, fmt.Errorf("cloudinary copy body: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("cloudinary close form: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("cloudinary create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth(a.creds.APIKey, a.creds.APISecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudinary upload failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		PublicID  string `json:"public_id"`
		SecureURL string `json:"secure_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("cloudinary decode response: %w", err)
	}

	return &UploadResult{
		RemotePath: result.PublicID,
		RemoteURL:  result.SecureURL,
	}, nil
}

func (a *CloudinaryAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	url := fmt.Sprintf("https://res.cloudinary.com/%s/raw/upload/%s", a.creds.CloudName, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cloudinary download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudinary download: %w", err)
	}
	if resp.StatusCode >= 400 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("cloudinary download failed: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (a *CloudinaryAdapter) Delete(ctx context.Context, key string) error {
	url := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/resources/raw/upload/%s", a.creds.CloudName, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("cloudinary delete request: %w", err)
	}
	req.SetBasicAuth(a.creds.APIKey, a.creds.APISecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudinary delete: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("cloudinary delete failed: %d", resp.StatusCode)
	}
	return nil
}

var _ Adapter = (*CloudinaryAdapter)(nil)
