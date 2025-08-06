# Configuration Reference

This document provides a comprehensive reference for all configuration options available in the SLURM Prometheus Exporter.

## Table of Contents

- [Configuration Format](#configuration-format)
- [Configuration Loading](#configuration-loading)
- [Environment Variables](#environment-variables)
- [Hot-Reloading](#hot-reloading)
- [Server Configuration](#server-configuration)
- [SLURM Configuration](#slurm-configuration)
- [Collectors Configuration](#collectors-configuration)
- [Logging Configuration](#logging-configuration)
- [Metrics Configuration](#metrics-configuration)
- [Complete Example](#complete-example)

## Configuration Format

The exporter uses YAML format for configuration files. All configuration options can also be overridden using environment variables with the prefix `SLURM_EXPORTER_`.

## Configuration Loading

The configuration is loaded in the following order (later sources override earlier ones):

1. **Default values** (built-in defaults)
2. **Configuration file** (specified with `-config` flag)
3. **Environment variables** (with `SLURM_EXPORTER_` prefix)

### Command Line Usage

```bash
# Use specific configuration file
slurm-exporter -config /path/to/config.yaml

# Use environment variables only (no config file)
SLURM_EXPORTER_SLURM_BASE_URL=https://slurm.example.com:6820 slurm-exporter

# Combine file and environment variables
SLURM_EXPORTER_SERVER_ADDRESS=:9090 slurm-exporter -config production.yaml
```

## Environment Variables

All configuration options can be overridden using environment variables with the pattern:

```
SLURM_EXPORTER_<SECTION>_<FIELD>
```

### Examples

```bash
# Server configuration
SLURM_EXPORTER_SERVER_ADDRESS=":9090"
SLURM_EXPORTER_SERVER_TLS_ENABLED="true"

# SLURM configuration  
SLURM_EXPORTER_SLURM_BASE_URL="https://slurm.example.com:6820"
SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt"
SLURM_EXPORTER_SLURM_AUTH_TOKEN="your-jwt-token"

# Logging configuration
SLURM_EXPORTER_LOGGING_LEVEL="debug"
SLURM_EXPORTER_LOGGING_FORMAT="json"

# Collector configuration
SLURM_EXPORTER_COLLECTORS_JOBS_INTERVAL="15s"
SLURM_EXPORTER_COLLECTORS_NODES_ENABLED="false"
```

## Hot-Reloading

The exporter supports hot-reloading of configuration files. When the configuration file is modified, the exporter will:

1. **Detect** file changes using filesystem notifications
2. **Validate** the new configuration 
3. **Apply** changes without restarting the process
4. **Log** successful reloads or validation errors

Note: Some changes (like server address) may require a restart.

## Server Configuration

Controls the HTTP server that exposes metrics and health endpoints.

```yaml
server:
  address: ":8080"              # Server listen address
  metrics_path: "/metrics"      # Prometheus metrics endpoint
  health_path: "/health"        # Health check endpoint  
  ready_path: "/ready"          # Readiness probe endpoint
  timeout: "60s"                # Server shutdown timeout
  read_timeout: "30s"           # HTTP read timeout
  write_timeout: "30s"          # HTTP write timeout  
  idle_timeout: "60s"           # HTTP idle timeout
  max_request_size: 1048576     # Maximum request size in bytes

  # TLS configuration
  tls:
    enabled: false              # Enable HTTPS
    cert_file: ""               # Path to TLS certificate
    key_file: ""                # Path to TLS private key
    min_version: "1.2"          # Minimum TLS version
    prefer_server_ciphers: true # Prefer server cipher suites

  # Basic authentication for metrics endpoint
  basic_auth:
    enabled: false              # Enable basic auth
    username: ""                # Username for basic auth
    password: ""                # Password for basic auth
    realm: "SLURM Exporter"     # Authentication realm
```

### Server Configuration Details

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `address` | string | `:8080` | Listen address (`:port` or `host:port`) |
| `metrics_path` | string | `/metrics` | Prometheus metrics endpoint path |
| `health_path` | string | `/health` | Health check endpoint path |
| `ready_path` | string | `/ready` | Readiness probe endpoint path |
| `timeout` | duration | `60s` | Server graceful shutdown timeout |
| `read_timeout` | duration | `30s` | HTTP request read timeout |
| `write_timeout` | duration | `30s` | HTTP response write timeout |
| `idle_timeout` | duration | `60s` | HTTP connection idle timeout |
| `max_request_size` | int | `1048576` | Maximum HTTP request body size |

#### TLS Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable HTTPS/TLS |
| `cert_file` | string | `""` | Path to PEM-encoded certificate file |
| `key_file` | string | `""` | Path to PEM-encoded private key file |
| `min_version` | string | `"1.2"` | Minimum TLS version (`1.0`, `1.1`, `1.2`, `1.3`) |
| `prefer_server_ciphers` | bool | `true` | Prefer server cipher suite ordering |

#### Basic Authentication

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable HTTP Basic Authentication |
| `username` | string | `""` | Username for authentication |
| `password` | string | `""` | Password for authentication |
| `realm` | string | `"SLURM Exporter"` | Authentication realm |

## SLURM Configuration

Configures connection to the SLURM REST API.

```yaml
slurm:
  base_url: "https://slurm.example.com:6820"  # SLURM REST API base URL
  api_version: "v0.0.42"                      # SLURM REST API version
  timeout: "30s"                              # Request timeout
  retry_attempts: 3                           # Number of retry attempts
  retry_delay: "5s"                           # Initial retry delay
  insecure_skip_verify: false                 # Skip TLS certificate verification

  # Authentication configuration
  auth:
    type: "jwt"                               # Auth type: none, jwt, basic, apikey
    token: ""                                 # JWT token (for jwt auth)
    token_file: ""                            # Path to JWT token file
    username: ""                              # Username (for basic auth)
    password: ""                              # Password (for basic auth)  
    password_file: ""                         # Path to password file
    api_key: ""                               # API key (for apikey auth)
    api_key_file: ""                          # Path to API key file

  # Rate limiting configuration
  rate_limit:
    requests_per_second: 10.0                 # Sustained request rate
    burst_size: 20                            # Burst request capacity
```

### SLURM Configuration Details

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `base_url` | string | **required** | SLURM REST API base URL |
| `api_version` | string | `v0.0.42` | SLURM REST API version |
| `timeout` | duration | `30s` | HTTP request timeout |
| `retry_attempts` | int | `3` | Number of retry attempts on failure |
| `retry_delay` | duration | `5s` | Initial delay between retries |
| `insecure_skip_verify` | bool | `false` | Skip TLS certificate verification |

#### Supported API Versions

- `v0.0.40` - SLURM 21.08.x
- `v0.0.41` - SLURM 22.05.x  
- `v0.0.42` - SLURM 23.02.x (recommended)
- `v0.0.43` - SLURM 23.11.x

#### Authentication Types

| Type | Description | Required Fields |
|------|-------------|----------------|
| `none` | No authentication | None |
| `jwt` | JWT token authentication | `token` or `token_file` |
| `basic` | HTTP Basic authentication | `username`, `password` or `password_file` |
| `apikey` | API key authentication | `api_key` or `api_key_file` |

#### Rate Limiting

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `requests_per_second` | float | `10.0` | Sustained request rate limit |
| `burst_size` | int | `20` | Maximum burst request capacity |

## Collectors Configuration

Controls which metrics collectors are enabled and their collection intervals.

```yaml
collectors:
  # Global collector settings
  global:
    default_interval: "30s"      # Default collection interval
    default_timeout: "30s"       # Default collection timeout
    max_concurrency: 5           # Maximum concurrent collectors
    error_threshold: 3           # Error threshold before disabling collector
    recovery_delay: "60s"        # Delay before re-enabling failed collector
    graceful_degradation: true   # Enable graceful degradation on errors

  # Individual collector configurations
  cluster:
    enabled: true
    interval: "60s"
    timeout: "30s"
    max_concurrency: 1
    
  nodes:
    enabled: true
    interval: "30s" 
    timeout: "30s"
    max_concurrency: 3
    error_handling:
      max_retries: 3
      retry_delay: "10s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
      
  jobs:
    enabled: true
    interval: "15s"
    timeout: "30s"
    max_concurrency: 5
    
  users:
    enabled: true
    interval: "60s"
    timeout: "30s"
    max_concurrency: 2
    
  partitions:
    enabled: true
    interval: "60s"
    timeout: "30s" 
    max_concurrency: 2
    
  performance:
    enabled: true
    interval: "30s"
    timeout: "30s"
    max_concurrency: 2
    
  system:
    enabled: true
    interval: "30s"
    timeout: "10s"
    max_concurrency: 1
```

### Collector Types

| Collector | Description | Typical Interval |
|-----------|-------------|------------------|
| `cluster` | Cluster overview and capacity metrics | 60s |
| `nodes` | Node status, hardware, and utilization | 30s |
| `jobs` | Job states, queues, and resource usage | 15s |
| `users` | User and account usage statistics | 60s |
| `partitions` | Partition configuration and utilization | 60s |
| `performance` | Performance and efficiency metrics | 30s |
| `system` | Exporter self-monitoring metrics | 30s |

### Global Collector Settings

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `default_interval` | duration | `30s` | Default collection interval |
| `default_timeout` | duration | `30s` | Default collection timeout |
| `max_concurrency` | int | `5` | Maximum concurrent collections |
| `error_threshold` | int | `3` | Errors before disabling collector |
| `recovery_delay` | duration | `60s` | Delay before re-enabling failed collector |
| `graceful_degradation` | bool | `true` | Continue on collector failures |

### Individual Collector Settings

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | bool | Enable this collector |
| `interval` | duration | Collection interval |
| `timeout` | duration | Collection timeout |
| `max_concurrency` | int | Maximum concurrent requests |

### Error Handling Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_retries` | int | `3` | Maximum retry attempts |
| `retry_delay` | duration | `10s` | Initial retry delay |
| `backoff_factor` | float | `2.0` | Backoff multiplier |
| `max_retry_delay` | duration | `60s` | Maximum retry delay |
| `fail_fast` | bool | `false` | Fail fast on errors |

## Logging Configuration

Controls logging output, format, and levels.

```yaml
logging:
  level: "info"                 # Log level: debug, info, warn, error
  format: "json"                # Log format: json, text
  output: "stdout"              # Output: stdout, stderr, file
  file: ""                      # Log file path (when output=file)
  max_size: 100                 # Max log file size in MB
  max_backups: 3                # Number of backup files to keep
  max_age: 28                   # Max age of log files in days
  compress: true                # Compress rotated log files
  suppress_http: false          # Suppress HTTP access logs
```

### Logging Configuration Details

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | `info` | Minimum log level to output |
| `format` | string | `json` | Log message format |
| `output` | string | `stdout` | Log output destination |
| `file` | string | `""` | Log file path (required when output=file) |
| `max_size` | int | `100` | Maximum log file size in MB |
| `max_backups` | int | `3` | Number of backup files to retain |
| `max_age` | int | `28` | Maximum age of log files in days |
| `compress` | bool | `true` | Compress rotated log files |
| `suppress_http` | bool | `false` | Suppress HTTP access logs |

#### Log Levels

| Level | Description |
|-------|-------------|
| `debug` | Verbose debugging information |
| `info` | General information messages |
| `warn` | Warning messages for potential issues |
| `error` | Error messages for failures |

#### Log Formats

| Format | Description |
|--------|-------------|
| `json` | Structured JSON format (recommended for production) |
| `text` | Human-readable text format (good for development) |

## Metrics Configuration

Controls Prometheus metrics collection and cardinality limits.

```yaml
metrics:
  namespace: "slurm"            # Prometheus metrics namespace
  subsystem: "exporter"         # Prometheus metrics subsystem
  const_labels:                 # Constant labels added to all metrics
    cluster: "production"
    datacenter: "us-east-1"
  max_age: "5m"                 # Maximum age of metric samples
  age_buckets: 5                # Number of age buckets for histogram

  # Prometheus registry configuration
  registry:
    enable_go_collector: true          # Enable Go runtime metrics
    enable_process_collector: true     # Enable process metrics
    enable_build_info: true            # Enable build info metric

  # Cardinality management
  cardinality:
    max_series: 10000           # Maximum number of time series
    max_labels: 10              # Maximum labels per metric
    max_label_size: 128         # Maximum label value size in characters
    warn_limit: 8000            # Warning threshold for series count
```

### Metrics Configuration Details

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `namespace` | string | `slurm` | Prometheus metrics namespace prefix |
| `subsystem` | string | `exporter` | Prometheus metrics subsystem |
| `const_labels` | map | `{}` | Constant labels for all metrics |
| `max_age` | duration | `5m` | Maximum age for metric samples |
| `age_buckets` | int | `5` | Number of buckets for age histograms |

#### Registry Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enable_go_collector` | bool | `true` | Include Go runtime metrics |
| `enable_process_collector` | bool | `true` | Include process metrics |
| `enable_build_info` | bool | `true` | Include build information |

#### Cardinality Limits

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_series` | int | `10000` | Maximum time series (prevents memory issues) |
| `max_labels` | int | `10` | Maximum labels per metric |
| `max_label_size` | int | `128` | Maximum characters in label values |
| `warn_limit` | int | `8000` | Warning threshold for series count |

## Complete Example

```yaml
# Complete configuration example with all sections
server:
  address: ":8080"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"
  timeout: "60s"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  max_request_size: 1048576
  
  tls:
    enabled: false
    cert_file: "/etc/ssl/certs/slurm-exporter.pem"
    key_file: "/etc/ssl/private/slurm-exporter.key"
    min_version: "1.2"
    prefer_server_ciphers: true
    
  basic_auth:
    enabled: false
    username: "prometheus"
    password: "secret"
    realm: "SLURM Exporter"

slurm:
  base_url: "https://slurm.example.com:6820"
  api_version: "v0.0.42"
  timeout: "30s"
  retry_attempts: 3
  retry_delay: "5s"
  insecure_skip_verify: false
  
  auth:
    type: "jwt"
    token: ""
    token_file: "/etc/slurm-exporter/jwt-token"
    username: ""
    password: ""
    password_file: ""
    api_key: ""
    api_key_file: ""
    
  rate_limit:
    requests_per_second: 10.0
    burst_size: 20

collectors:
  global:
    default_interval: "30s"
    default_timeout: "30s"
    max_concurrency: 5
    error_threshold: 3
    recovery_delay: "60s"
    graceful_degradation: true
    
  cluster:
    enabled: true
    interval: "60s"
    timeout: "30s"
    max_concurrency: 1
    
  nodes:
    enabled: true
    interval: "30s"
    timeout: "30s"
    max_concurrency: 3
    error_handling:
      max_retries: 3
      retry_delay: "10s"
      backoff_factor: 2.0
      max_retry_delay: "60s"
      fail_fast: false
      
  jobs:
    enabled: true
    interval: "15s"
    timeout: "30s"
    max_concurrency: 5
    
  users:
    enabled: true
    interval: "60s"
    timeout: "30s"
    max_concurrency: 2
    
  partitions:
    enabled: true
    interval: "60s"
    timeout: "30s"
    max_concurrency: 2
    
  performance:
    enabled: true
    interval: "30s"
    timeout: "30s"
    max_concurrency: 2
    
  system:
    enabled: true
    interval: "30s"
    timeout: "10s"
    max_concurrency: 1

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: "/var/log/slurm-exporter.log"
  max_size: 100
  max_backups: 3
  max_age: 28
  compress: true
  suppress_http: false

metrics:
  namespace: "slurm"
  subsystem: "exporter"
  const_labels:
    cluster: "production"
    environment: "prod"
    datacenter: "us-east-1"
  max_age: "5m"
  age_buckets: 5
  
  registry:
    enable_go_collector: true
    enable_process_collector: true
    enable_build_info: true
    
  cardinality:
    max_series: 10000
    max_labels: 10
    max_label_size: 128
    warn_limit: 8000
```

## Configuration Validation

The exporter performs comprehensive validation of all configuration options:

- **Required fields** are checked for presence
- **Numeric values** are validated for proper ranges
- **Duration values** are parsed and validated
- **File paths** are checked for existence (when specified)
- **URLs** are validated for proper format
- **Enum values** are checked against allowed options

Validation errors include:
- Current invalid value
- Expected format or range
- Configuration path for easy identification
- Example correct values
- Environment variable alternatives

## Best Practices

### Performance Tuning

1. **Collection Intervals**: Start with conservative intervals and reduce based on needs
   - Jobs: 15-30s (change frequently)
   - Nodes: 30-60s (moderate changes)
   - Cluster/Partitions: 60s+ (infrequent changes)

2. **Concurrency**: Start low and increase based on SLURM API performance
   - Monitor SLURM API response times
   - Watch for rate limiting errors

3. **Rate Limiting**: Configure based on SLURM API capacity
   - Start with 10 requests/second
   - Increase gradually while monitoring

### Security

1. **TLS**: Always use HTTPS in production
2. **Authentication**: Store tokens/passwords in files, not configuration
3. **File Permissions**: Protect configuration files (600 or 640)
4. **Secrets**: Use environment variables for sensitive values

### Monitoring

1. **Enable system collector** for self-monitoring
2. **Monitor cardinality** to prevent memory issues  
3. **Watch error rates** in logs and metrics
4. **Set up alerts** for collector failures

### Example Environment-Based Configuration

```bash
# Production environment variables
export SLURM_EXPORTER_SERVER_ADDRESS=":9090"
export SLURM_EXPORTER_SERVER_TLS_ENABLED="true"
export SLURM_EXPORTER_SERVER_TLS_CERT_FILE="/etc/ssl/certs/exporter.pem"
export SLURM_EXPORTER_SERVER_TLS_KEY_FILE="/etc/ssl/private/exporter.key"

export SLURM_EXPORTER_SLURM_BASE_URL="https://slurm.example.com:6820"
export SLURM_EXPORTER_SLURM_AUTH_TYPE="jwt"
export SLURM_EXPORTER_SLURM_AUTH_TOKEN_FILE="/etc/slurm-exporter/jwt-token"

export SLURM_EXPORTER_LOGGING_LEVEL="info"
export SLURM_EXPORTER_LOGGING_FORMAT="json"
export SLURM_EXPORTER_LOGGING_OUTPUT="file"
export SLURM_EXPORTER_LOGGING_FILE="/var/log/slurm-exporter.log"

export SLURM_EXPORTER_METRICS_CONST_LABELS_CLUSTER="production"
export SLURM_EXPORTER_METRICS_CONST_LABELS_DATACENTER="us-east-1"
```