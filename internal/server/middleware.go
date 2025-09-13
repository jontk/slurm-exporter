package server

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggingMiddleware provides structured HTTP request logging
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap the response writer to capture status and size
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the request
		duration := time.Since(start)
		
		logEntry := s.logger.WithFields(logrus.Fields{
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"remote_addr":   r.RemoteAddr,
			"user_agent":    r.Header.Get("User-Agent"),
			"referer":       r.Header.Get("Referer"),
			"status":        rw.statusCode,
			"response_size": rw.written,
			"duration":      duration,
			"duration_ms":   float64(duration.Nanoseconds()) / 1000000,
		})

		// Add request ID if present
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			logEntry = logEntry.WithField("request_id", requestID)
		}

		// Log at different levels based on status code
		message := "HTTP request"
		switch {
		case rw.statusCode >= 500:
			logEntry.Error(message)
		case rw.statusCode >= 400:
			logEntry.Warn(message)
		default:
			logEntry.Info(message)
		}
	})
}

// MetricsMiddleware adds metrics collection for HTTP requests
func (s *Server) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap the response writer
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics (if we had HTTP metrics - placeholder for future)
		duration := time.Since(start)
		
		s.logger.WithFields(logrus.Fields{
			"component": "http_metrics",
			"method":    r.Method,
			"path":      r.URL.Path,
			"status":    rw.statusCode,
			"duration":  duration,
		}).Debug("HTTP metrics recorded")
	})
}

// HeadersMiddleware adds standard security and info headers
func (s *Server) HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		// Add server info
		w.Header().Set("Server", "slurm-exporter")
		
		// Add cache control for metrics endpoint
		if r.URL.Path == s.config.Server.MetricsPath {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics and logs them
func (s *Server) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithFields(logrus.Fields{
					"component":   "recovery",
					"method":      r.Method,
					"path":        r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"panic":       err,
				}).Error("HTTP handler panic recovered")

				// Return 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CombinedMiddleware applies all middleware in the correct order
func (s *Server) CombinedMiddleware(next http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied = first executed)
	handler := next
	handler = s.MetricsMiddleware(handler)
	handler = s.LoggingMiddleware(handler)
	handler = s.HeadersMiddleware(handler)
	handler = s.RecoveryMiddleware(handler)
	
	return handler
}