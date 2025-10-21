# Troubleshooting Guide

This guide covers common issues encountered when deploying and operating the SLURM exporter, along with step-by-step solutions.

## Table of Contents

- [Connection Issues](#connection-issues)
- [Authentication Problems](#authentication-problems)
- [Performance Issues](#performance-issues)
- [Metric Collection Problems](#metric-collection-problems)
- [Kubernetes Deployment Issues](#kubernetes-deployment-issues)
- [Configuration Problems](#configuration-problems)
- [Resource Constraints](#resource-constraints)
- [Monitoring and Alerting Issues](#monitoring-and-alerting-issues)
- [Debugging Tools](#debugging-tools)

## Connection Issues

### SLURM REST API Connection Failed

**Symptoms:**
- Log messages like `Failed to connect to SLURM REST API`
- Health check endpoint returns unhealthy
- No metrics being collected

**Common Causes & Solutions:**

1. **Incorrect API URL**
   ```bash
   # Check if SLURM REST API is accessible
   curl -k https://your-slurm-server:6820/slurm/v0.0.40/diag
   
   # Verify configuration
   grep -i url /etc/slurm-exporter/config.yaml
   ```

2. **Network connectivity issues**
   ```bash
   # Test network connectivity
   telnet your-slurm-server 6820
   
   # Check DNS resolution
   nslookup your-slurm-server
   
   # Test with curl
   curl -v https://your-slurm-server:6820/slurm/v0.0.40/ping
   ```

3. **TLS/SSL certificate problems**
   ```bash
   # Test with TLS verification disabled
   curl -k https://your-slurm-server:6820/slurm/v0.0.40/ping
   
   # Check certificate
   openssl s_client -connect your-slurm-server:6820 -showcerts
   ```

**Resolution:**
```yaml
# config.yaml - For self-signed certificates
slurm:
  url: "https://your-slurm-server:6820"
  tls_skip_verify: true  # Only for testing
  
# For proper certificate validation
slurm:
  url: "https://your-slurm-server:6820"
  ca_cert_file: "/path/to/ca-bundle.pem"
```

### Connection Timeouts

**Symptoms:**
- Intermittent connection failures
- Slow metric collection
- Timeout errors in logs

**Solutions:**

1. **Increase timeout values**
   ```yaml
   slurm:
     timeout: 30s      # Increase from default 10s
     retry_attempts: 3
     retry_delay: 5s
   ```

2. **Check SLURM API performance**
   ```bash
   # Time API responses
   time curl -k https://your-slurm-server:6820/slurm/v0.0.40/nodes
   
   # Check SLURM controller load
   scontrol show config | grep -i max
   ```

## Authentication Problems

### JWT Token Authentication Failed

**Symptoms:**
- HTTP 401 Unauthorized errors
- Log messages about invalid tokens
- Authentication header issues

**Solutions:**

1. **Verify token format and expiration**
   ```bash
   # Decode JWT token (without verification)
   echo "your-jwt-token" | base64 -d
   
   # Check token expiration
   python3 -c "
   import jwt
   import datetime
   token = 'your-jwt-token'
   decoded = jwt.decode(token, options={'verify_signature': False})
   print('Expires:', datetime.datetime.fromtimestamp(decoded['exp']))
   "
   ```

2. **Generate new token**
   ```bash
   # Using scontrol (SLURM 21.08+)
   scontrol token
   
   # Or use JWT generation tool
   jwt encode --secret="your-secret" --exp="+1 day" user=slurm-exporter
   ```

3. **Configure token in exporter**
   ```yaml
   slurm:
     auth:
       type: "jwt"
       token: "your-jwt-token"
   ```

### API Key Authentication Issues

**Symptoms:**
- API key rejected by SLURM
- Inconsistent authentication behavior

**Solutions:**

1. **Verify API key format**
   ```bash
   # Test API key directly
   curl -H "X-SLURM-USER-NAME: slurm-exporter" \
        -H "X-SLURM-USER-TOKEN: your-api-key" \
        https://your-slurm-server:6820/slurm/v0.0.40/ping
   ```

2. **Check SLURM user permissions**
   ```bash
   # Verify user exists in SLURM
   sacctmgr show user slurm-exporter
   
   # Check associations
   sacctmgr show association user=slurm-exporter
   ```

## Performance Issues

### Slow Metric Collection

**Symptoms:**
- Long collection durations
- High CPU/memory usage
- Collection timeouts

**Solutions:**

1. **Optimize collection intervals**
   ```yaml
   collection:
     intervals:
       cluster: 30s     # Reduce frequency for expensive calls
       jobs: 60s        # Jobs change less frequently
       nodes: 15s       # Nodes need frequent monitoring
   ```

2. **Enable selective collection**
   ```yaml
   collectors:
     cluster:
       enabled: true
     jobs:
       enabled: true
       max_jobs: 1000   # Limit job collection
     nodes:
       enabled: true
   ```

3. **Use connection pooling**
   ```yaml
   slurm:
     max_connections: 10
     keep_alive: true
     connection_timeout: 10s
   ```

### High Memory Usage

**Symptoms:**
- OOMKilled containers
- Memory usage constantly increasing
- Slow garbage collection

**Solutions:**

1. **Limit metric cardinality**
   ```yaml
   metrics:
     max_series: 50000
     label_limits:
       user: 1000      # Limit user metrics to top 1000
       partition: 50   # Limit partitions
   ```

2. **Adjust Go runtime settings**
   ```bash
   # Set environment variables
   export GOGC=100           # Default garbage collection
   export GOMEMLIMIT=500MiB  # Set memory limit
   ```

3. **Configure resource limits**
   ```yaml
   # Kubernetes
   resources:
     limits:
       memory: "1Gi"
     requests:
       memory: "256Mi"
   ```

## Metric Collection Problems

### Missing Metrics

**Symptoms:**
- Expected metrics not appearing in Prometheus
- Gaps in metric series
- Partial data collection

**Solutions:**

1. **Check collector status**
   ```bash
   # Query exporter health endpoint
   curl http://slurm-exporter:8080/health
   
   # Check specific collector metrics
   curl http://slurm-exporter:8080/metrics | grep slurm_exporter_collection
   ```

2. **Verify SLURM API responses**
   ```bash
   # Test individual endpoints
   curl -H "Authorization: Bearer your-token" \
        https://slurm-server:6820/slurm/v0.0.40/nodes
   
   curl -H "Authorization: Bearer your-token" \
        https://slurm-server:6820/slurm/v0.0.40/jobs
   ```

3. **Enable debug logging**
   ```yaml
   logging:
     level: debug
     format: json
   ```

### Incorrect Metric Values

**Symptoms:**
- Metrics show unexpected values
- Inconsistent data between SLURM commands and metrics
- Mathematical errors in calculations

**Solutions:**

1. **Compare with SLURM commands**
   ```bash
   # Cross-check node information
   sinfo -Nel
   curl http://slurm-exporter:8080/metrics | grep slurm_node_cpus_total
   
   # Verify job counts
   squeue | wc -l
   curl http://slurm-exporter:8080/metrics | grep slurm_cluster_jobs_total
   ```

2. **Check metric labeling**
   ```bash
   # Verify label consistency
   curl http://slurm-exporter:8080/metrics | grep slurm_node_state
   ```

3. **Review metric definitions**
   ```bash
   # Check help text
   curl http://slurm-exporter:8080/metrics | grep "# HELP slurm_"
   ```

## Kubernetes Deployment Issues

### Pod Startup Failures

**Symptoms:**
- Pods stuck in CrashLoopBackOff
- Init container failures
- Image pull errors

**Solutions:**

1. **Check pod logs**
   ```bash
   kubectl logs -l app=slurm-exporter
   kubectl describe pod -l app=slurm-exporter
   ```

2. **Verify image availability**
   ```bash
   # Test image pull
   docker pull your-registry/slurm-exporter:latest
   
   # Check registry access
   kubectl get secrets -o yaml | grep dockerconfig
   ```

3. **Check resource constraints**
   ```bash
   kubectl top pods
   kubectl describe nodes
   ```

### Service Discovery Issues

**Symptoms:**
- Prometheus not scraping metrics
- Service endpoints not found
- Network policy blocking access

**Solutions:**

1. **Verify service configuration**
   ```bash
   kubectl get svc slurm-exporter -o yaml
   kubectl get endpoints slurm-exporter
   ```

2. **Test service connectivity**
   ```bash
   # From within cluster
   kubectl run debug --image=busybox --rm -it -- sh
   wget -O- http://slurm-exporter:8080/metrics
   ```

3. **Check Prometheus configuration**
   ```bash
   kubectl get servicemonitor slurm-exporter -o yaml
   
   # Verify Prometheus targets
   curl http://prometheus:9090/api/v1/targets
   ```

### RBAC Permission Issues

**Symptoms:**
- Permission denied errors
- ServiceAccount authentication failures
- API access denied

**Solutions:**

1. **Verify RBAC configuration**
   ```bash
   kubectl get serviceaccount slurm-exporter
   kubectl get rolebinding,clusterrolebinding | grep slurm-exporter
   ```

2. **Test permissions**
   ```bash
   kubectl auth can-i get pods --as=system:serviceaccount:monitoring:slurm-exporter
   ```

3. **Check security contexts**
   ```bash
   kubectl get pod -l app=slurm-exporter -o yaml | grep -A10 securityContext
   ```

## Configuration Problems

### Invalid Configuration File

**Symptoms:**
- Application fails to start
- YAML parsing errors
- Configuration validation failures

**Solutions:**

1. **Validate YAML syntax**
   ```bash
   # Check YAML syntax
   yamllint config.yaml
   
   # Use Python to validate
   python3 -c "import yaml; yaml.safe_load(open('config.yaml'))"
   ```

2. **Test configuration**
   ```bash
   # Dry-run configuration
   ./slurm-exporter --config=config.yaml --dry-run
   
   # Validate against schema
   jsonschema -i config.yaml schema.json
   ```

### Environment Variable Override Issues

**Symptoms:**
- Environment variables not taking effect
- Unexpected configuration values
- Precedence problems

**Solutions:**

1. **Check environment variable format**
   ```bash
   # List all SLURM_EXPORTER_* variables
   env | grep SLURM_EXPORTER
   
   # Verify naming convention
   export SLURM_EXPORTER_SLURM_URL="https://slurm-server:6820"
   ```

2. **Debug configuration loading**
   ```yaml
   logging:
     level: debug
   ```

## Resource Constraints

### CPU Throttling

**Symptoms:**
- High CPU wait times
- Slow metric collection
- Increased response times

**Solutions:**

1. **Monitor CPU usage**
   ```bash
   # Check container CPU metrics
   kubectl top pods slurm-exporter
   
   # Monitor CPU throttling
   kubectl exec slurm-exporter -- cat /sys/fs/cgroup/cpu/cpu.stat
   ```

2. **Adjust resource limits**
   ```yaml
   resources:
     requests:
       cpu: "100m"
     limits:
       cpu: "500m"
   ```

### Storage Issues

**Symptoms:**
- Log file write failures
- Disk space errors
- Configuration file read errors

**Solutions:**

1. **Check disk usage**
   ```bash
   kubectl exec slurm-exporter -- df -h
   kubectl describe pv,pvc
   ```

2. **Configure log rotation**
   ```yaml
   logging:
     file: "/var/log/slurm-exporter.log"
     max_size: "100MB"
     max_backups: 3
     max_age: 7
   ```

## Monitoring and Alerting Issues

### Alert Fatigue

**Symptoms:**
- Too many false positive alerts
- Important alerts being ignored
- Alert storms during maintenance

**Solutions:**

1. **Tune alert thresholds**
   ```yaml
   # Adjust sensitivity
   - alert: HighCPUUtilization
     expr: cpu_usage > 0.85  # Increase from 0.80
     for: 15m                # Increase duration
   ```

2. **Implement alert suppression**
   ```yaml
   # Add maintenance window routing
   route:
     routes:
     - match:
         severity: info
       receiver: low-priority
   ```

3. **Use alert dependencies**
   ```yaml
   - alert: NodeDown
     expr: slurm_node_state{state="down"} == 1
   - alert: PartitionDegraded
     expr: slurm_partition_nodes_down > 0
     # Only fire if nodes are actually down
     for: 5m
   ```

### Grafana Dashboard Issues

**Symptoms:**
- Panels showing no data
- Query timeouts
- Incorrect visualizations

**Solutions:**

1. **Verify data source**
   ```bash
   # Test Prometheus connectivity
   curl http://prometheus:9090/api/v1/query?query=up
   ```

2. **Debug queries**
   ```promql
   # Test individual queries in Prometheus UI
   sum(slurm_node_cpus_total)
   rate(slurm_job_submissions_total[5m])
   ```

3. **Check time ranges**
   ```bash
   # Verify metric timestamps
   curl 'http://prometheus:9090/api/v1/query?query=slurm_cluster_info'
   ```

## Debugging Tools

### Log Analysis

```bash
# View recent logs
kubectl logs -l app=slurm-exporter --tail=100

# Follow logs in real-time
kubectl logs -l app=slurm-exporter -f

# Search for specific errors
kubectl logs -l app=slurm-exporter | grep -i "error\|failed\|timeout"

# Export logs for analysis
kubectl logs -l app=slurm-exporter --since=1h > slurm-exporter.log
```

### Metric Debugging

```bash
# Check exporter metrics
curl http://slurm-exporter:8080/metrics | grep slurm_exporter_

# Monitor collection performance
curl http://slurm-exporter:8080/metrics | grep collection_duration

# Check for collection errors
curl http://slurm-exporter:8080/metrics | grep collection_errors_total
```

### Network Debugging

```bash
# Test connectivity from pod
kubectl exec -it slurm-exporter -- nslookup slurm-controller
kubectl exec -it slurm-exporter -- telnet slurm-controller 6820

# Check network policies
kubectl get networkpolicy
kubectl describe networkpolicy slurm-exporter

# Monitor network traffic
kubectl exec -it slurm-exporter -- tcpdump -i eth0 port 6820
```

### Configuration Debugging

```bash
# Dump effective configuration
kubectl exec slurm-exporter -- cat /etc/slurm-exporter/config.yaml

# Check environment variables
kubectl exec slurm-exporter -- env | grep SLURM_EXPORTER

# Validate configuration
kubectl exec slurm-exporter -- /slurm-exporter --config=/etc/slurm-exporter/config.yaml --validate
```

## Getting Help

### Enable Debug Mode

```yaml
# Temporary debug configuration
logging:
  level: debug
  format: json

# Enable request tracing
slurm:
  debug: true
  trace_requests: true
```

### Collect Diagnostic Information

```bash
#!/bin/bash
# diagnostic-collect.sh

echo "=== SLURM Exporter Diagnostics ==="
echo "Date: $(date)"
echo "Version: $(kubectl exec slurm-exporter -- /slurm-exporter --version)"

echo -e "\n=== Configuration ==="
kubectl get configmap slurm-exporter-config -o yaml

echo -e "\n=== Pod Status ==="
kubectl describe pod -l app=slurm-exporter

echo -e "\n=== Recent Logs ==="
kubectl logs -l app=slurm-exporter --tail=50

echo -e "\n=== Metrics Sample ==="
curl -s http://slurm-exporter:8080/metrics | head -20

echo -e "\n=== Health Check ==="
curl -s http://slurm-exporter:8080/health
```

### Contact Support

When reporting issues, please include:

1. SLURM exporter version
2. SLURM version and configuration
3. Kubernetes version (if applicable)
4. Complete error messages and logs
5. Configuration files (sanitized)
6. Steps to reproduce the issue
7. Expected vs. actual behavior

Submit issues at: https://github.com/jontk/slurm-exporter/issues

## Common Error Patterns

### Connection Refused
```
Error: dial tcp 10.0.0.1:6820: connect: connection refused
```
**Solution:** Check if SLURM REST API daemon (slurmrestd) is running

### Permission Denied
```
Error: HTTP 403 Forbidden - insufficient privileges
```
**Solution:** Verify user permissions in SLURM accounting database

### Context Deadline Exceeded
```
Error: context deadline exceeded
```
**Solution:** Increase timeout values in configuration

### Certificate Verification Failed
```
Error: x509: certificate signed by unknown authority
```
**Solution:** Add CA certificate or set `tls_skip_verify: true` for testing

### Memory Limit Exceeded
```
Error: signal: killed (OOMKilled)
```
**Solution:** Increase memory limits or reduce metric cardinality