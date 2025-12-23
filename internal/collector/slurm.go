package collector

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/config"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/metrics"
)

// Collector is responsible for collecting metrics from Slurm
type Collector struct {
	config   *config.Config
	client   *http.Client
	registry *metrics.Registry
	logger   *slog.Logger
}

// NewCollector creates a new Slurm metrics collector
func NewCollector(cfg *config.Config, registry *metrics.Registry, logger *slog.Logger) (*Collector, error) {
	timeout, err := cfg.GetTimeoutDuration()
	if err != nil {
		return nil, fmt.Errorf("invalid timeout configuration: %w", err)
	}

	// Create HTTP client with optional TLS insecure skip verify
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Configure TLS if needed
	if cfg.Slurm.TLSInsecureVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		logger.Warn("TLS certificate verification is disabled - this is insecure and should only be used for testing")
	}

	return &Collector{
		config:   cfg,
		client:   httpClient,
		registry: registry,
		logger:   logger,
	}, nil
}

// CollectAll collects metrics from all enabled Slurm endpoints
func (c *Collector) CollectAll(ctx context.Context) (map[string]string, error) {
	enabledEndpoints := c.config.GetEnabledEndpoints()
	results := make(map[string]string)

	for _, endpoint := range enabledEndpoints {
		c.logger.Debug("collecting metrics from endpoint",
			"name", endpoint.Name,
			"path", endpoint.Path)

		timer := prometheus.NewTimer(c.registry.ScrapeDuration.WithLabelValues(endpoint.Name))
		metrics, err := c.collectEndpoint(ctx, endpoint)
		timer.ObserveDuration()

		if err != nil {
			c.logger.Error("failed to collect metrics from endpoint",
				"endpoint", endpoint.Name,
				"error", err)
			c.registry.ScrapeSuccess.WithLabelValues(endpoint.Name).Set(0)
			c.registry.ScrapeErrors.WithLabelValues(endpoint.Name).Inc()
			continue
		}

		c.registry.ScrapeSuccess.WithLabelValues(endpoint.Name).Set(1)
		results[endpoint.Name] = metrics
	}

	return results, nil
}

// collectEndpoint collects metrics from a single Slurm endpoint as raw text
func (c *Collector) collectEndpoint(ctx context.Context, endpoint config.EndpointConfig) (string, error) {
	url := c.config.Slurm.URL + endpoint.Path

	c.logger.Debug("fetching metrics from URL", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the metrics as raw text
	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, resp.Body); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Add custom labels to each metric line
	metricsText := c.addCustomLabels(buffer.String())

	return metricsText, nil
}

// addCustomLabels adds configured custom labels to all metric lines
func (c *Collector) addCustomLabels(metricsText string) string {
	if len(c.config.Labels) == 0 {
		return metricsText
	}

	// Build the labels string to append
	var labelsBuilder strings.Builder
	first := true
	for key, value := range c.config.Labels {
		if !first {
			labelsBuilder.WriteString(",")
		}
		labelsBuilder.WriteString(key)
		labelsBuilder.WriteString("=\"")
		labelsBuilder.WriteString(value)
		labelsBuilder.WriteString("\"")
		first = false
	}
	labelsStr := labelsBuilder.String()

	// Parse each line and add labels
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(metricsText))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Parse metric name and value
		// Format: metric_name{labels} value timestamp
		// or: metric_name value
		parts := strings.Fields(line)
		if len(parts) < 2 {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		metricPart := parts[0]
		rest := strings.Join(parts[1:], " ")

		// Check if there are already labels
		if strings.Contains(metricPart, "{") {
			// Add our labels to existing ones
			idx := strings.Index(metricPart, "{")
			name := metricPart[:idx]
			existingLabels := metricPart[idx+1 : len(metricPart)-1]
			result.WriteString(name)
			result.WriteString("{")
			result.WriteString(existingLabels)
			result.WriteString(",")
			result.WriteString(labelsStr)
			result.WriteString("} ")
		} else {
			// Add labels to metric without any
			result.WriteString(metricPart)
			result.WriteString("{")
			result.WriteString(labelsStr)
			result.WriteString("} ")
		}

		result.WriteString(rest)
		result.WriteString("\n")
	}

	return result.String()
}

// WriteMetrics writes all collected metrics in Prometheus format
func (c *Collector) WriteMetrics(w io.Writer, metricsMap map[string]string) error {
	for _, metricsText := range metricsMap {
		if _, err := w.Write([]byte(metricsText)); err != nil {
			return fmt.Errorf("failed to write metrics: %w", err)
		}
	}
	return nil
}

// Health checks if the Slurm API is reachable
func (c *Collector) Health(ctx context.Context) error {
	url := c.config.Slurm.URL

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("slurm API is not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("slurm API returned unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
