package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad_ValidConfig(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "configs", "config.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Address != ":8080" {
		t.Fatalf("expected server address :8080, got %q", cfg.Server.Address)
	}
	if cfg.HashRing.Replicas != 100 {
		t.Fatalf("expected 100 replicas, got %d", cfg.HashRing.Replicas)
	}
	if cfg.Health.Interval != 5*time.Second {
		t.Fatalf("expected 5s interval, got %v", cfg.Health.Interval)
	}
	if cfg.Health.Timeout != 2*time.Second {
		t.Fatalf("expected 2s timeout, got %v", cfg.Health.Timeout)
	}
	if cfg.Health.FailureThreshold != 3 {
		t.Fatalf("expected failure threshold 3, got %d", cfg.Health.FailureThreshold)
	}
	if cfg.Health.SuccessThreshold != 2 {
		t.Fatalf("expected success threshold 2, got %d", cfg.Health.SuccessThreshold)
	}
	if len(cfg.Backends) != 3 {
		t.Fatalf("expected 3 backends, got %d", len(cfg.Backends))
	}
	if cfg.Backends[0].ID != "server-8081" {
		t.Fatalf("expected first backend server-8081, got %q", cfg.Backends[0].ID)
	}
	if cfg.Backends[0].Address != "http://localhost:8081" {
		t.Fatalf("expected first backend address http://localhost:8081, got %q", cfg.Backends[0].Address)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("does-not-exist.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTempConfig(t, "not: [valid")
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid yaml")
	}
}

func TestLoad_ValidationFailure(t *testing.T) {
	path := writeTempConfig(t, `
server:
  address: ":8080"
hashRing:
  replicas: 0
backends:
  - id: server-1
    address: http://localhost:8081
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "hashRing.replicas") {
		t.Fatalf("expected replicas validation error, got %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid",
			cfg: Config{
				Server:   ServerConfig{Address: ":8080"},
				HashRing: HashRingConfig{Replicas: 10},
				Backends: []BackendConfig{{ID: "a", Address: "http://localhost:8081"}},
			},
		},
		{
			name: "missing server address",
			cfg: Config{
				Server:   ServerConfig{Address: ""},
				HashRing: HashRingConfig{Replicas: 10},
				Backends: []BackendConfig{{ID: "a", Address: "http://localhost:8081"}},
			},
			wantErr: "server.address is required",
		},
		{
			name: "zero replicas",
			cfg: Config{
				Server:   ServerConfig{Address: ":8080"},
				HashRing: HashRingConfig{Replicas: 0},
				Backends: []BackendConfig{{ID: "a", Address: "http://localhost:8081"}},
			},
			wantErr: "hashRing.replicas must be greater than zero",
		},
		{
			name: "no backends",
			cfg: Config{
				Server:   ServerConfig{Address: ":8080"},
				HashRing: HashRingConfig{Replicas: 10},
			},
			wantErr: "at least one backend is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
