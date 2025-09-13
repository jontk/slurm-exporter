package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/logging"
	"github.com/jontk/slurm-exporter/internal/server"
	"github.com/jontk/slurm-exporter/pkg/version"
)

var (
	configFile = flag.String("config", "configs/config.yaml", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showVersion = flag.Bool("version", false, "Show version information and exit")
	addr        = flag.String("addr", ":8080", "Address to listen on")
	metricsPath = flag.String("metrics-path", "/metrics", "Path for metrics endpoint")
)

func main() {
	flag.Parse()

	// Show version information if requested
	if *showVersion {
		versionInfo := version.Get()
		fmt.Println(versionInfo.String())
		os.Exit(0)
	}

	// Load configuration first
	cfg, err := config.Load(*configFile)
	if err != nil {
		// Use basic logrus for configuration errors
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Override log level from command line if provided
	if *logLevel != "info" {
		cfg.Logging.Level = *logLevel
	}

	// Set up structured logging
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize logger")
	}
	defer logger.Close()

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Override config with command line flags if provided
	if *addr != ":8080" {
		cfg.Server.Address = *addr
	}
	if *metricsPath != "/metrics" {
		cfg.Server.MetricsPath = *metricsPath
	}

	logger.WithComponent("main").WithFields(logrus.Fields{
		"version":      version.Get().Short(),
		"config_file":  *configFile,
		"address":      cfg.Server.Address,
		"metrics_path": cfg.Server.MetricsPath,
		"log_level":    cfg.Logging.Level,
		"log_format":   cfg.Logging.Format,
	}).Info("Starting SLURM Prometheus Exporter")

	// Create Prometheus registry for collectors
	promRegistry := prometheus.NewRegistry()
	
	// Create collector registry
	registry, err := collector.NewRegistry(&cfg.Collectors, promRegistry)
	if err != nil {
		logger.WithComponent("main").WithError(err).Fatal("Failed to create collector registry")
	}

	// Create and start the server
	srv, err := server.New(cfg, logger.Logger, registry)
	if err != nil {
		logger.WithComponent("main").WithError(err).Fatal("Failed to create server")
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		logger.WithField("signal", sig).Info("Received shutdown signal")
	case err := <-errChan:
		logger.WithError(err).Error("Server error")
	}

	// Graceful shutdown
	logger.Info("Initiating graceful shutdown...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.WithError(err).Error("Server shutdown error")
		os.Exit(1)
	}

	logger.Info("Server shut down gracefully")
}