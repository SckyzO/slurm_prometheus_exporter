package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry holds all the Prometheus metrics for the exporter
type Registry struct {
	// Build information
	BuildInfo *prometheus.GaugeVec

	// Scrape metrics
	ScrapeDuration *prometheus.HistogramVec
	ScrapeSuccess  *prometheus.GaugeVec
	ScrapeErrors   *prometheus.CounterVec

	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec

	// Custom registry for Slurm metrics
	customRegistry *prometheus.Registry
}

// NewRegistry creates and registers all metrics for the exporter
func NewRegistry(version, gitCommit, buildTime string) *Registry {
	reg := &Registry{
		customRegistry: prometheus.NewRegistry(),
	}

	// Build information metric
	reg.BuildInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slurm_exporter_build_info",
			Help: "A metric with a constant '1' value labeled by version, git_commit, and build_time",
		},
		[]string{"version", "git_commit", "build_time"},
	)
	reg.BuildInfo.WithLabelValues(version, gitCommit, buildTime).Set(1)

	// Scrape duration histogram
	reg.ScrapeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slurm_exporter_scrape_duration_seconds",
			Help:    "Duration of scrapes by the exporter",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)

	// Scrape success gauge
	reg.ScrapeSuccess = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slurm_exporter_scrape_success",
			Help: "Whether the last scrape was successful (1 = success, 0 = failure)",
		},
		[]string{"endpoint"},
	)

	// Scrape errors counter
	reg.ScrapeErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slurm_exporter_scrape_errors_total",
			Help: "Total number of scrape errors by endpoint",
		},
		[]string{"endpoint"},
	)

	// HTTP requests total counter
	reg.HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slurm_exporter_http_requests_total",
			Help: "Total number of HTTP requests received by the exporter",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP request duration histogram
	reg.HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slurm_exporter_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	return reg
}

// GetRegistry returns the custom Prometheus registry
func (r *Registry) GetRegistry() *prometheus.Registry {
	return r.customRegistry
}
