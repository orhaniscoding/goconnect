package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/config"
	"gopkg.in/yaml.v3"
)

type setupTestResponse struct {
	Status         string   `json:"status"`
	ConfigPresent  bool     `json:"config_present"`
	ConfigValid    bool     `json:"config_valid"`
	ValidationErr  string   `json:"validation_error"`
	NextStep       string   `json:"next_step"`
	CompletedSteps []string `json:"completed_steps"`
}

func TestSetupStatusWithoutConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "goconnect.yaml")

	r := gin.New()
	registerSetupRoutes(r, "sqlite", "data/goconnect.db", configPath, nil)

	req := httptest.NewRequest(http.MethodGet, "/setup/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp setupTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.ConfigPresent {
		t.Fatalf("expected config_present=false")
	}
	if resp.NextStep != "mode" {
		t.Fatalf("expected next_step=mode, got %s", resp.NextStep)
	}
}

func TestSetupStatusWithValidConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "goconnect.yaml")

	cfg := validSQLiteConfig()
	if err := config.SaveToFile(&cfg, configPath); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	r := gin.New()
	registerSetupRoutes(r, "sqlite", "data/goconnect.db", configPath, nil)

	req := httptest.NewRequest(http.MethodGet, "/setup/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp setupTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.ConfigPresent {
		t.Fatalf("expected config_present=true")
	}
	if !resp.ConfigValid {
		t.Fatalf("expected config_valid=true, got false (validation=%s)", resp.ValidationErr)
	}
	if resp.NextStep != setupCompletedStep {
		t.Fatalf("expected next_step=%s, got %s", setupCompletedStep, resp.NextStep)
	}
}

func TestSetupStatusWithPartialConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "goconnect.yaml")

	cfg := config.Config{
		Server: config.ServerConfig{Host: "0.0.0.0", Port: "8080"},
		Database: config.DatabaseConfig{
			Backend:    "sqlite",
			SQLitePath: "data/goconnect.db",
		},
		// Missing JWT/WireGuard values on purpose to keep it incomplete.
	}
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, raw, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	r := gin.New()
	registerSetupRoutes(r, "sqlite", "data/goconnect.db", configPath, nil)

	req := httptest.NewRequest(http.MethodGet, "/setup/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp setupTestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.NextStep != "admin" {
		t.Fatalf("expected next_step=admin, got %s", resp.NextStep)
	}
	if len(resp.CompletedSteps) != 1 || resp.CompletedSteps[0] != "mode" {
		t.Fatalf("expected completed_steps to contain only mode, got %+v", resp.CompletedSteps)
	}
	if resp.ConfigValid {
		t.Fatalf("expected config_valid=false for incomplete config")
	}
}

func TestSetupPersistAndRestartFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "goconnect.yaml")

	r := gin.New()
	stopCh := make(chan struct{}, 1)
	stop := func() {
		stopCh <- struct{}{}
	}
	registerSetupRoutes(r, "sqlite", "data/goconnect.db", configPath, stop)

	payload := setupRequest{
		Config:  validSQLiteConfig(),
		Restart: true,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	if !fileExists(configPath) {
		t.Fatalf("expected config file to be written")
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if restartRequired, ok := resp["restart_required"].(bool); !ok || !restartRequired {
		t.Fatalf("expected restart_required=true, got %v", resp["restart_required"])
	}

	select {
	case <-stopCh:
		// restart triggered
	case <-time.After(time.Second):
		t.Fatalf("expected restart callback to be invoked")
	}
}

func validSQLiteConfig() config.Config {
	return config.Config{
		Server: config.ServerConfig{Host: "0.0.0.0", Port: "8080"},
		Database: config.DatabaseConfig{
			Backend:    "sqlite",
			SQLitePath: "data/goconnect.db",
		},
		JWT: config.JWTConfig{
			Secret:          strings.Repeat("s", 32),
			AccessTokenTTL:  time.Minute,
			RefreshTokenTTL: time.Hour,
		},
		WireGuard: config.WireGuardConfig{
			ServerEndpoint: "example.com:51820",
			ServerPubKey:   strings.Repeat("A", 44),
		},
	}
}
