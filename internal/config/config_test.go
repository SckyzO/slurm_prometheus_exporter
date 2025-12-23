package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	content := `
slurm:
  url: "http://localhost:6817"
  timeout: "10s"

server:
  port: 8080
  basic_auth:
    enabled: true
    username: "admin"
    password: "password"
  ssl:
    enabled: false
    cert_file: ""
    key_file: ""

endpoints:
  - name: "jobs"
    path: "/metrics/jobs"
    enabled: true

labels:
  cluster: "test"

logging:
  level: "info"
  output: "stdout"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading the config
	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.Slurm.URL != "http://localhost:6817" {
		t.Errorf("Expected slurm.url to be 'http://localhost:6817', got '%s'", cfg.Slurm.URL)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server.port to be 8080, got %d", cfg.Server.Port)
	}

	if !cfg.Server.BasicAuth.Enabled {
		t.Error("Expected basic_auth to be enabled")
	}

	if len(cfg.Endpoints) != 1 {
		t.Errorf("Expected 1 endpoint, got %d", len(cfg.Endpoints))
	}

	if cfg.Labels["cluster"] != "test" {
		t.Errorf("Expected label 'cluster' to be 'test', got '%s'", cfg.Labels["cluster"])
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		shouldErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Slurm: SlurmConfig{
					URL:     "http://localhost:6817",
					Timeout: "10s",
				},
				Server: ServerConfig{
					Port: 8080,
					BasicAuth: BasicAuthConfig{
						Enabled:  true,
						Username: "admin",
						Password: "password",
					},
				},
				Endpoints: []EndpointConfig{
					{Name: "jobs", Path: "/metrics/jobs", Enabled: true},
				},
				Logging: LoggingConfig{
					Level:  "info",
					Output: "stdout",
				},
			},
			shouldErr: false,
		},
		{
			name: "missing slurm url",
			config: Config{
				Slurm: SlurmConfig{
					URL:     "",
					Timeout: "10s",
				},
				Server: ServerConfig{Port: 8080},
				Endpoints: []EndpointConfig{
					{Name: "jobs", Path: "/metrics/jobs", Enabled: true},
				},
			},
			shouldErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Slurm: SlurmConfig{
					URL:     "http://localhost:6817",
					Timeout: "10s",
				},
				Server: ServerConfig{Port: 0},
				Endpoints: []EndpointConfig{
					{Name: "jobs", Path: "/metrics/jobs", Enabled: true},
				},
			},
			shouldErr: true,
		},
		{
			name: "no endpoints",
			config: Config{
				Slurm: SlurmConfig{
					URL:     "http://localhost:6817",
					Timeout: "10s",
				},
				Server:    ServerConfig{Port: 8080},
				Endpoints: []EndpointConfig{},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr && err == nil {
				t.Error("Expected an error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestGetEnabledEndpoints(t *testing.T) {
	cfg := Config{
		Endpoints: []EndpointConfig{
			{Name: "jobs", Path: "/metrics/jobs", Enabled: true},
			{Name: "nodes", Path: "/metrics/nodes", Enabled: false},
			{Name: "partitions", Path: "/metrics/partitions", Enabled: true},
		},
	}

	enabled := cfg.GetEnabledEndpoints()
	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled endpoints, got %d", len(enabled))
	}

	for _, ep := range enabled {
		if !ep.Enabled {
			t.Errorf("Endpoint '%s' should be enabled", ep.Name)
		}
	}
}
