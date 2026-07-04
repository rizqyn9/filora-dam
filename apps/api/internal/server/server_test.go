package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rizqynugroho9/filora-dam/api/internal/config"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

func TestHealthEndpoint(t *testing.T) {
	app := New(Deps{Config: &config.Config{Env: "test", Port: "3000"}})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var env lib.Envelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (body=%s)", err, body)
	}
	if !env.Success {
		t.Fatalf("expected success=true, got %s", body)
	}
}

func TestRootEndpoint(t *testing.T) {
	app := New(Deps{Config: &config.Config{Env: "test", Port: "3000"}})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
