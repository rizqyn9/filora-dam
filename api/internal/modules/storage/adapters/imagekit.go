package adapters

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

)

type ImageKitCredentials struct {
	PublicKey   string `json:"public_key"`
	PrivateKey  string `json:"private_key"`
	URLEndpoint string `json:"url_endpoint"`
}

type ImageKitAdapter struct {
	creds ImageKitCredentials
}

func NewImageKitAdapter(creds ImageKitCredentials) *ImageKitAdapter {
	return &ImageKitAdapter{creds: creds}
}

func (a *ImageKitAdapter) Upload(ctx context.Context, input UploadInput) (*UploadResult, error) {
	const url = "https://upload.imagekit.io/api/v1/files/upload"

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("fileName", input.Key)
	_ = w.WriteField("useUniqueFileName", "false")

	part, err := w.CreateFormFile("file", input.Key)
	if err != nil {
		return nil, fmt.Errorf("imagekit create form: %w", err)
	}
	if _, err := io.Copy(part, input.Body); err != nil {
		return nil, fmt.Errorf("imagekit copy body: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("imagekit create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	// ImageKit uses private key as basic auth username, empty password
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.creds.PrivateKey+":")))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("imagekit upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("imagekit upload failed (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		FileID string `json:"fileId"`
		URL    string `json:"url"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("imagekit decode response: %w", err)
	}

	return &UploadResult{
		RemotePath: result.FileID,
		RemoteURL:  result.URL,
	}, nil
}

func (a *ImageKitAdapter) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// ImageKit files are accessed via URL endpoint + path
	url := fmt.Sprintf("%s/%s", a.creds.URLEndpoint, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("imagekit download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("imagekit download: %w", err)
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, fmt.Errorf("imagekit download failed: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (a *ImageKitAdapter) Delete(ctx context.Context, key string) error {
	url := fmt.Sprintf("https://api.imagekit.io/v1/files/%s", key)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("imagekit delete request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.creds.PrivateKey+":")))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("imagekit delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("imagekit delete failed: %d", resp.StatusCode)
	}
	return nil
}

var _ Adapter = (*ImageKitAdapter)(nil)
