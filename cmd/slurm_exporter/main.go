package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/collector"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/config"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/metrics"
	"github.com/sckyzo/slurm_prometheus_exporter/internal/server"
)

var (
	// Version information, set during build time via -ldflags
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"

	// Command-line flags
	configFile = kingpin.Flag("config.file", "Path to configuration file").
			Default("config.yaml").
			String()

	webListenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry").
				Default(":8080").
				String()

	// webTelemetryPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics").
	// 				Default("/metrics").
	// 				String()

	// webConfigFile = kingpin.Flag("web.config.file", "Path to configuration file for TLS and/or basic authentication (optional)").
	// 		String()

	logLevel = kingpin.Flag("log.level", "Log level (debug, info, warn, error)").
			Default("info").
			String()

	logFormat = kingpin.Flag("log.format", "Log format (text, json)").
			Default("text").
			String()

	showVersion = kingpin.Flag("version", "Show version information").
			Short('v').
			Bool()
)

func main() {
	// Parse command-line arguments
	kingpin.Parse()

	// Show version information if requested
	if *showVersion {
		fmt.Printf("Slurm Prometheus Exporter\n")
		fmt.Printf("Version:    %s\n", Version)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		fmt.Printf("Build Time: %s\n", BuildTime)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override config with CLI flags if provided
	if *webListenAddress != ":8080" {
		// Parse port from listen address
		if _, err := fmt.Sscanf(*webListenAddress, ":%d", &cfg.Server.Port); err == nil {
			// Successfully parsed port
		} else if _, err := fmt.Sscanf(*webListenAddress, "%*[^:]:%d", &cfg.Server.Port); err == nil {
			// Successfully parsed with host
			fmt.Printf("Parsed listen address with host: %s\n", *webListenAddress)
		}
	}

	// Override logging configuration with CLI flags
	if *logLevel != "info" {
		cfg.Logging.Level = *logLevel
	}
	if *logFormat != "text" {
		if *logFormat == "json" {
			cfg.Logging.Output = "json"
		}
	}

	// Setup logging
	logger := setupLogger(cfg.Logging)
	logger.Info("starting slurm exporter",
		"version", Version,
		"git_commit", GitCommit,
		"build_time", BuildTime)

	// Create metrics registry
	metricsRegistry := metrics.NewRegistry(Version, GitCommit, BuildTime)

	// Create collector
	coll, err := collector.NewCollector(cfg, metricsRegistry, logger)
	if err != nil {
		logger.Error("failed to create collector", "error", err)
		os.Exit(1)
	}

	// Check Slurm API health
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := coll.Health(ctx); err != nil {
		logger.Warn("slurm API health check failed, but continuing anyway",
			"error", err,
			"url", cfg.Slurm.URL)
	} else {
		logger.Info("slurm API health check passed", "url", cfg.Slurm.URL)
	}

	// Create HTTP server
	srv := server.NewServer(cfg, coll, metricsRegistry, logger, Version)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("exporter is ready",
		"port", cfg.Server.Port,
		"endpoints", len(cfg.GetEnabledEndpoints()))

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down exporter...")

	// Give the server 10 seconds to finish ongoing requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("exporter stopped successfully")
}

// setupLogger configures the structured logger based on the configuration
func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Output == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
