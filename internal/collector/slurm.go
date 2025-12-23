package collector

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
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

	return &Collector{
		config: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
		registry: registry,
		logger:   logger,
	}, nil
}

// CollectAll collects metrics from all enabled Slurm endpoints
func (c *Collector) CollectAll(ctx context.Context) (map[string][]*dto.MetricFamily, error) {
	enabledEndpoints := c.config.GetEnabledEndpoints()
	results := make(map[string][]*dto.MetricFamily)

	for _, endpoint := range enabledEndpoints {
		c.logger.Debug("collecting metrics from endpoint",
			"name", endpoint.Name,
			"path", endpoint.Path)

		timer := prometheus.NewTimer(c.registry.ScrapeDuration.WithLabelValues(endpoint.Name))
		metricFamilies, err := c.collectEndpoint(ctx, endpoint)
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
		results[endpoint.Name] = metricFamilies
	}

	return results, nil
}

// collectEndpoint collects metrics from a single Slurm endpoint
func (c *Collector) collectEndpoint(ctx context.Context, endpoint config.EndpointConfig) ([]*dto.MetricFamily, error) {
	url := c.config.Slurm.URL + endpoint.Path

	c.logger.Debug("fetching metrics from URL", "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the OpenMetrics/Prometheus format
	metricFamilies, err := c.parseMetrics(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	// Add custom labels to all metrics
	metricFamilies = c.addCustomLabels(metricFamilies)

	return metricFamilies, nil
}

// parseMetrics parses OpenMetrics/Prometheus format into MetricFamily structs
func (c *Collector) parseMetrics(reader io.Reader) ([]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", err)
	}

	// Convert map to slice
	families := make([]*dto.MetricFamily, 0, len(metricFamilies))
	for _, mf := range metricFamilies {
		families = append(families, mf)
	}

	return families, nil
}

// addCustomLabels adds configured custom labels to all metrics
func (c *Collector) addCustomLabels(families []*dto.MetricFamily) []*dto.MetricFamily {
	if len(c.config.Labels) == 0 {
		return families
	}

	// Create new label pairs from config
	customLabels := make([]*dto.LabelPair, 0, len(c.config.Labels))
	for key, value := range c.config.Labels {
		k := key
		v := value
		customLabels = append(customLabels, &dto.LabelPair{
			Name:  &k,
			Value: &v,
		})
	}

	// Add custom labels to each metric
	for _, family := range families {
		for _, metric := range family.Metric {
			metric.Label = append(metric.Label, customLabels...)
		}
	}

	return families
}

// WriteMetrics writes all collected metrics in Prometheus format
func (c *Collector) WriteMetrics(w io.Writer, families map[string][]*dto.MetricFamily) error {
	encoder := expfmt.NewEncoder(w, expfmt.FmtText)

	for _, metricFamilies := range families {
		for _, family := range metricFamilies {
			if err := encoder.Encode(family); err != nil {
				return fmt.Errorf("failed to encode metric family: %w", err)
			}
		}
	}

	return nil
}

// CollectFromFile collects metrics from a test file (for testing purposes)
func (c *Collector) CollectFromFile(filePath string) ([]*dto.MetricFamily, error) {
	file, err := http.Get(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}
	defer file.Body.Close()

	return c.parseMetrics(file.Body)
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

// GetMetricsAsText returns metrics as text for a single endpoint (for debugging)
func (c *Collector) GetMetricsAsText(ctx context.Context, endpointName string) (string, error) {
	var endpoint *config.EndpointConfig
	for _, ep := range c.config.Endpoints {
		if ep.Name == endpointName && ep.Enabled {
			endpoint = &ep
			break
		}
	}

	if endpoint == nil {
		return "", fmt.Errorf("endpoint '%s' not found or not enabled", endpointName)
	}

	url := c.config.Slurm.URL + endpoint.Path

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

	// Read response as text
	var builder strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		builder.WriteString(scanner.Text())
		builder.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return builder.String(), nil
}
