package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Slurm     SlurmConfig       `yaml:"slurm"`
	Server    ServerConfig      `yaml:"server"`
	Endpoints []EndpointConfig  `yaml:"endpoints"`
	Labels    map[string]string `yaml:"labels"`
	Logging   LoggingConfig     `yaml:"logging"`
}

// SlurmConfig holds the Slurm API connection settings
type SlurmConfig struct {
	URL     string `yaml:"url"`
	Timeout string `yaml:"timeout"`
}

// ServerConfig holds the HTTP server configuration
type ServerConfig struct {
	Port      int             `yaml:"port"`
	BasicAuth BasicAuthConfig `yaml:"basic_auth"`
	SSL       SSLConfig       `yaml:"ssl"`
}

// BasicAuthConfig holds the Basic Authentication settings
type BasicAuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// SSLConfig holds the SSL/TLS settings
type SSLConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// EndpointConfig represents a Slurm endpoint configuration
type EndpointConfig struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Enabled bool   `yaml:"enabled"`
}

// LoggingConfig holds the logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML configuration
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate Slurm configuration
	if c.Slurm.URL == "" {
		return fmt.Errorf("slurm.url is required")
	}

	if c.Slurm.Timeout == "" {
		return fmt.Errorf("slurm.timeout is required")
	}

	// Validate timeout format
	if _, err := time.ParseDuration(c.Slurm.Timeout); err != nil {
		return fmt.Errorf("invalid slurm.timeout format: %w", err)
	}

	// Validate server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}

	// Validate Basic Auth configuration
	if c.Server.BasicAuth.Enabled {
		if c.Server.BasicAuth.Username == "" || c.Server.BasicAuth.Password == "" {
			return fmt.Errorf("basic auth is enabled but username or password is empty")
		}
	}

	// Validate SSL configuration
	if c.Server.SSL.Enabled {
		if c.Server.SSL.CertFile == "" || c.Server.SSL.KeyFile == "" {
			return fmt.Errorf("ssl is enabled but cert_file or key_file is empty")
		}
	}

	// Validate endpoints
	if len(c.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint must be configured")
	}

	for i, endpoint := range c.Endpoints {
		if endpoint.Name == "" {
			return fmt.Errorf("endpoint %d: name is required", i)
		}
		if endpoint.Path == "" {
			return fmt.Errorf("endpoint %d: path is required", i)
		}
	}

	// Validate logging configuration
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	validLevels := []string{"debug", "info", "warn", "error"}
	isValid := false
	for _, level := range validLevels {
		if c.Logging.Level == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}

	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	return nil
}

// GetTimeoutDuration returns the timeout as a time.Duration
func (c *Config) GetTimeoutDuration() (time.Duration, error) {
	return time.ParseDuration(c.Slurm.Timeout)
}

// GetEnabledEndpoints returns only the enabled endpoints
func (c *Config) GetEnabledEndpoints() []EndpointConfig {
	var enabled []EndpointConfig
	for _, endpoint := range c.Endpoints {
		if endpoint.Enabled {
			enabled = append(enabled, endpoint)
		}
	}
	return enabled
}
