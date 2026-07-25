// Package client is a thin HTTP client for the Filora API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client talks to the Filora API using an optional bearer token.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http: &http.Client{
			Timeout: 60 * time.Second,
			// Do not auto-follow redirects (download returns a storage URL we
			// fetch separately, without our Authorization header).
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// envelope mirrors the API response wrapper.
type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Do performs a JSON request and decodes data into out (may be nil).
func (c *Client) Do(method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	c.auth(req)

	return c.send(req, out)
}

// Upload posts a multipart file to the given path.
func (c *Client) Upload(path, filePath string, out any) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, f); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	c.auth(req)

	return c.send(req, out)
}

// Download resolves the redirect URL for path and streams the object to w.
func (c *Client) Download(path string, w io.Writer) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.auth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		return c.decodeError(resp)
	}
	loc := resp.Header.Get("Location")
	if loc == "" {
		return fmt.Errorf("no download location returned")
	}

	// Fetch the storage URL without our API credentials.
	obj, err := http.Get(loc)
	if err != nil {
		return fmt.Errorf("fetch object: %w", err)
	}
	defer obj.Body.Close()
	if obj.StatusCode >= 300 {
		return fmt.Errorf("object fetch failed: %s", obj.Status)
	}
	_, err = io.Copy(w, obj.Body)
	return err
}

func (c *Client) auth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")
}

func (c *Client) send(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("unexpected response (%s): %s", resp.Status, strings.TrimSpace(string(data)))
	}
	if !env.Success {
		if env.Error != nil {
			return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
		}
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	if out != nil && len(env.Data) > 0 {
		return json.Unmarshal(env.Data, out)
	}
	return nil
}

func (c *Client) decodeError(resp *http.Response) error {
	data, _ := io.ReadAll(resp.Body)
	var env envelope
	if json.Unmarshal(data, &env) == nil && env.Error != nil {
		return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
	}
	return fmt.Errorf("request failed: %s", resp.Status)
}
