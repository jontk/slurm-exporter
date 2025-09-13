package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
)

// RegistryInterface defines the methods needed by the server from the registry
type RegistryInterface interface {
	GetStats() map[string]collector.CollectorState
	CollectAll(ctx context.Context) error
}

// Server represents the HTTP server.
type Server struct {
	config   *config.Config
	logger   *logrus.Logger
	server   *http.Server
	registry RegistryInterface
	promRegistry *prometheus.Registry
}

// New creates a new server instance.
func New(cfg *config.Config, logger *logrus.Logger, registry RegistryInterface) (*Server, error) {
	// Create Prometheus registry if not provided
	promRegistry := prometheus.NewRegistry()
	
	s := &Server{
		config:       cfg,
		logger:       logger,
		registry:     registry,
		promRegistry: promRegistry,
	}
	
	// Create HTTP mux and setup routes
	mux := s.setupRoutes()
	
	// Configure HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
	
	s.server = server
	return s, nil
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)
	
	// Readiness check endpoint
	mux.HandleFunc("/ready", s.handleReady)
	
	// Metrics endpoint
	mux.Handle(s.config.Server.MetricsPath, s.createMetricsHandler())
	
	// Root endpoint with basic info
	mux.HandleFunc("/", s.handleRoot)
	
	return mux
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.logger.WithField("address", s.config.Server.Address).Info("Starting HTTP server")
	
	go func() {
		<-ctx.Done()
		s.logger.Info("Context cancelled, shutting down server")
		s.server.Shutdown(context.Background())
	}()
	
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Health check requested")
	
	// Simple health check - always returns OK
	// Could be extended to check dependencies
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReady handles the readiness check endpoint
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Readiness check requested")
	
	// Check if collectors are ready
	if s.registry != nil {
		stats := s.registry.GetStats()
		
		// Consider ready if at least one collector is enabled
		ready := false
		for _, stat := range stats {
			if stat.Enabled {
				ready = true
				break
			}
		}
		
		if !ready {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("No collectors enabled"))
			return
		}
	}
	
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ready"))
}

// handleRoot handles the root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.logger.Debug("Root endpoint requested")
	
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	
	html := `<!DOCTYPE html>
<html>
<head>
    <title>SLURM Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 10px 0; }
        .endpoint a { text-decoration: none; color: #0066cc; }
        .endpoint a:hover { text-decoration: underline; }
        .stats { background-color: #f5f5f5; padding: 15px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>SLURM Prometheus Exporter</h1>
    <p>This is a Prometheus exporter for SLURM workload manager metrics.</p>
    
    <h2>Available Endpoints</h2>
    <div class="endpoint">📊 <a href="%s">Metrics</a> - Prometheus metrics endpoint</div>
    <div class="endpoint">❤️ <a href="/health">Health</a> - Health check endpoint</div>
    <div class="endpoint">⚡ <a href="/ready">Ready</a> - Readiness check endpoint</div>
    
    <div class="stats">
        <h3>Collector Status</h3>
        %s
    </div>
</body>
</html>`
	
	// Generate collector status
	collectorStatus := "No collectors configured"
	if s.registry != nil {
		stats := s.registry.GetStats()
		if len(stats) > 0 {
			collectorStatus = "<ul>"
			for name, stat := range stats {
				status := "✅ Enabled"
				if !stat.Enabled {
					status = "❌ Disabled"
				}
				collectorStatus += fmt.Sprintf("<li><strong>%s:</strong> %s</li>", name, status)
			}
			collectorStatus += "</ul>"
		}
	}
	
	content := fmt.Sprintf(html, s.config.Server.MetricsPath, collectorStatus)
	w.Write([]byte(content))
}

// createMetricsHandler creates the Prometheus metrics handler
func (s *Server) createMetricsHandler() http.Handler {
	// Create a custom gatherer that collects from our registry
	gatherer := prometheus.Gatherers{
		s.promRegistry,
		prometheus.DefaultGatherer, // Include Go runtime metrics
	}
	
	// Create promhttp handler with custom configuration
	handler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		ErrorLog:      s.logger,
		ErrorHandling: promhttp.ContinueOnError,
		Timeout:       30 * time.Second,
	})
	
	// Wrap with request logging and collection triggering
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_agent": r.Header.Get("User-Agent"),
		}).Debug("Metrics request received")
		
		// Trigger collection from all collectors if registry is available
		if s.registry != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
			defer cancel()
			
			if err := s.registry.CollectAll(ctx); err != nil {
				s.logger.WithError(err).Warn("Failed to collect all metrics")
				// Continue serving cached/existing metrics
			}
		}
		
		// Serve the metrics
		handler.ServeHTTP(w, r)
		
		duration := time.Since(start)
		s.logger.WithFields(logrus.Fields{
			"duration": duration,
			"status":   "completed",
		}).Debug("Metrics request completed")
	})
}

// GetPrometheusRegistry returns the Prometheus registry
func (s *Server) GetPrometheusRegistry() *prometheus.Registry {
	return s.promRegistry
}

// SetPrometheusRegistry sets the Prometheus registry
func (s *Server) SetPrometheusRegistry(registry *prometheus.Registry) {
	s.promRegistry = registry
}

// RegisterCollector registers a Prometheus collector
func (s *Server) RegisterCollector(collector prometheus.Collector) error {
	return s.promRegistry.Register(collector)
}

// UnregisterCollector unregisters a Prometheus collector
func (s *Server) UnregisterCollector(collector prometheus.Collector) bool {
	return s.promRegistry.Unregister(collector)
}

// GetMetricsPath returns the configured metrics path
func (s *Server) GetMetricsPath() string {
	return s.config.Server.MetricsPath
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return s.config.Server.Address
}