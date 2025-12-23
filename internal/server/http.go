package server

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/collector"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/config"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/metrics"
)

// Server represents the HTTP server for the exporter
type Server struct {
	config    *config.Config
	collector *collector.Collector
	registry  *metrics.Registry
	logger    *slog.Logger
	server    *http.Server
	version   string
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, coll *collector.Collector, reg *metrics.Registry, logger *slog.Logger, version string) *Server {
	return &Server{
		config:    cfg,
		collector: coll,
		registry:  reg,
		logger:    logger,
		version:   version,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("/", s.handleLandingPage())
	mux.Handle("/metrics", s.instrumentHandler(s.handleMetrics()))

	// Wrap with basic auth if enabled
	var handler http.Handler = mux
	if s.config.Server.BasicAuth.Enabled {
		handler = s.basicAuthMiddleware(mux)
	}

	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("starting HTTP server",
		"address", addr,
		"ssl_enabled", s.config.Server.SSL.Enabled,
		"basic_auth_enabled", s.config.Server.BasicAuth.Enabled)

	// Start with SSL or without
	if s.config.Server.SSL.Enabled {
		return s.server.ListenAndServeTLS(
			s.config.Server.SSL.CertFile,
			s.config.Server.SSL.KeyFile,
		)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping HTTP server")
	return s.server.Shutdown(ctx)
}

// handleLandingPage returns a handler for the landing page
func (s *Server) handleLandingPage() http.HandlerFunc {
	landingPageHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Slurm Exporter</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #007bff;
            padding-bottom: 10px;
        }
        .info {
            margin: 20px 0;
            padding: 15px;
            background-color: #e7f3ff;
            border-left: 4px solid #007bff;
            border-radius: 4px;
        }
        a {
            color: #007bff;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .version {
            color: #666;
            font-size: 0.9em;
            margin-top: 20px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Slurm Exporter</h1>
        <div class="info">
            <p>Welcome to the Slurm Metrics Exporter for Prometheus!</p>
            <p>This exporter collects metrics from Slurm and exposes them in Prometheus format.</p>
        </div>
        <h2>Available Endpoints:</h2>
        <ul>
            <li><a href="/metrics">/metrics</a> - Prometheus metrics endpoint</li>
        </ul>
        <div class="version">
            <strong>Version:</strong> %s
        </div>
    </div>
</body>
</html>`

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, landingPageHTML, s.version)
	}
}

// handleMetrics returns a handler for the metrics endpoint
func (s *Server) handleMetrics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		// Collect metrics from all endpoints
		metricsMap, err := s.collector.CollectAll(ctx)
		if err != nil {
			s.logger.Error("failed to collect metrics", "error", err)
			http.Error(w, "Failed to collect metrics", http.StatusInternalServerError)
			return
		}

		// Write metrics in Prometheus format
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		if err := s.collector.WriteMetrics(w, metricsMap); err != nil {
			s.logger.Error("failed to write metrics", "error", err)
			return
		}

		// Also expose the exporter's own metrics
		promhttp.Handler().ServeHTTP(w, r)
	})
}

// basicAuthMiddleware implements HTTP Basic Authentication
func (s *Server) basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		// Use constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.config.Server.BasicAuth.Username))
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.config.Server.BasicAuth.Password))

		if !ok || usernameMatch != 1 || passwordMatch != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Slurm Exporter"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			s.logger.Warn("unauthorized access attempt",
				"remote_addr", r.RemoteAddr,
				"username", username)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// instrumentHandler wraps a handler to collect metrics about HTTP requests
func (s *Server) instrumentHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture the status code
		wrw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve the request
		handler.ServeHTTP(wrw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		s.registry.HTTPRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", wrw.statusCode),
		).Inc()

		s.registry.HTTPRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	if w.statusCode == 0 {
		w.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write ensures that if WriteHeader wasn't called, we default to 200
func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
		w.ResponseWriter.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
