package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateEnhanced_Valid(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Address:        ":8080",
			MetricsPath:    "/metrics",
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxRequestSize: 10 * 1024 * 1024,
		},
		SLURM: SLURMConfig{
			BaseURL: "http://localhost:6820",
			Timeout: 20 * time.Second,
			Auth: AuthConfig{
				Type: "jwt",
			},
		},
		Collectors: CollectorsConfig{
			Jobs: CollectorConfig{
				Enabled: true,
				Timeout: 15 * time.Second,
			},
		},
	}

	err := cfg.ValidateEnhanced()
	assert.NoError(t, err)
}

func TestConfig_ValidateEnhanced_ServerErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		errMsg string
	}{
		{
			name: "invalid address",
			config: &Config{
				Server: ServerConfig{
					Address: "invalid",
				},
			},
			errMsg: "invalid server address",
		},
		{
			name: "invalid metrics path",
			config: &Config{
				Server: ServerConfig{
					Address:     ":8080",
					MetricsPath: "metrics", // missing leading slash
				},
			},
			errMsg: "metrics path must start with '/'",
		},
		{
			name: "timeouts too short",
			config: &Config{
				Server: ServerConfig{
					Address:      ":8080",
					MetricsPath:  "/metrics",
					ReadTimeout:  100 * time.Millisecond,
					WriteTimeout: 100 * time.Millisecond,
				},
			},
			errMsg: "read timeout too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set other required fields
			tt.config.SLURM = SLURMConfig{
				BaseURL: "http://localhost:6820",
				Timeout: 20 * time.Second,
			}

			err := tt.config.ValidateEnhanced()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestConfig_ValidateEnhanced_SLURMErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		errMsg string
	}{
		{
			name: "missing base URL",
			config: &Config{
				SLURM: SLURMConfig{
					BaseURL: "",
					Timeout: 20 * time.Second,
				},
			},
			errMsg: "SLURM base URL is required",
		},
		{
			name: "invalid base URL",
			config: &Config{
				SLURM: SLURMConfig{
					BaseURL: "not-a-url",
					Timeout: 20 * time.Second,
				},
			},
			errMsg: "invalid SLURM base URL",
		},
		{
			name: "invalid auth type",
			config: &Config{
				SLURM: SLURMConfig{
					BaseURL: "http://localhost:6820",
					Timeout: 20 * time.Second,
					Auth: AuthConfig{
						Type: "invalid",
					},
				},
			},
			errMsg: "invalid auth type",
		},
		{
			name: "missing JWT path",
			config: &Config{
				SLURM: SLURMConfig{
					BaseURL: "http://localhost:6820",
					Timeout: 20 * time.Second,
					Auth: AuthConfig{
						Type:    "jwt",
						JWTPath: "",
					},
				},
			},
			errMsg: "JWT path required for JWT auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set other required fields
			tt.config.Server = ServerConfig{
				Address:     ":8080",
				MetricsPath: "/metrics",
			}

			err := tt.config.ValidateEnhanced()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestConfig_ValidateEnhanced_CollectorErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		errMsg string
	}{
		{
			name: "collector timeout too short",
			config: &Config{
				Collectors: CollectorsConfig{
					Jobs: CollectorConfig{
						Enabled: true,
						Timeout: 100 * time.Millisecond,
					},
				},
			},
			errMsg: "jobs collector timeout too short",
		},
		{
			name: "collector timeout exceeds SLURM timeout",
			config: &Config{
				SLURM: SLURMConfig{
					Timeout: 10 * time.Second,
				},
				Collectors: CollectorsConfig{
					Jobs: CollectorConfig{
						Enabled: true,
						Timeout: 15 * time.Second,
					},
				},
			},
			errMsg: "jobs collector timeout exceeds SLURM timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set other required fields
			tt.config.Server = ServerConfig{
				Address:     ":8080",
				MetricsPath: "/metrics",
			}
			if tt.config.SLURM.BaseURL == "" {
				tt.config.SLURM.BaseURL = "http://localhost:6820"
			}

			err := tt.config.ValidateEnhanced()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestConfig_ValidateEnhanced_FilterErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		errMsg string
	}{
		{
			name: "conflicting filter flags",
			config: &Config{
				Collectors: CollectorsConfig{
					Jobs: CollectorConfig{
						Enabled: true,
						Filters: FilterConfig{
							MetricFilter: MetricFilterConfig{
								OnlyCounters:   true,
								OnlyGauges:     true,
							},
						},
					},
				},
			},
			errMsg: "conflicting metric type filters",
		},
		{
			name: "invalid wildcard pattern",
			config: &Config{
				Collectors: CollectorsConfig{
					Jobs: CollectorConfig{
						Enabled: true,
						Filters: FilterConfig{
							MetricFilter: MetricFilterConfig{
								IncludeMetrics: []string{"slurm_job_[*"},
							},
						},
					},
				},
			},
			errMsg: "invalid wildcard pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set other required fields
			tt.config.Server = ServerConfig{
				Address:     ":8080",
				MetricsPath: "/metrics",
			}
			tt.config.SLURM = SLURMConfig{
				BaseURL: "http://localhost:6820",
				Timeout: 20 * time.Second,
			}

			err := tt.config.ValidateEnhanced()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestConfig_ValidateEnhanced_CrossFieldValidation(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Address:      ":8080",
			MetricsPath:  "/metrics",
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 30 * time.Second,
			TLS: TLSConfig{
				Enabled:  true,
				CertFile: "", // missing cert
			},
		},
		SLURM: SLURMConfig{
			BaseURL: "http://localhost:6820",
			Timeout: 20 * time.Second,
		},
	}

	err := cfg.ValidateEnhanced()
	assert.Error(t, err)

	// Should have multiple errors
	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.True(t, len(validationErrs) > 1)

	// Check for specific errors
	errMessages := []string{}
	for _, e := range validationErrs {
		errMessages = append(errMessages, e.Error())
	}

	assert.Contains(t, errMessages, "TLS cert file required when TLS is enabled")
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		ValidationError{Field: "server.address", Message: "invalid address"},
		ValidationError{Field: "slurm.timeout", Message: "timeout too short"},
	}

	errStr := errs.Error()
	assert.Contains(t, errStr, "validation errors")
	assert.Contains(t, errStr, "server.address")
	assert.Contains(t, errStr, "slurm.timeout")
}