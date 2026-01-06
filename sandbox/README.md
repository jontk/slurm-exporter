# SLURM Exporter Sandbox Environment

A complete containerized environment for learning and experimenting with SLURM Exporter.

## 🚀 Quick Start

### Option 1: One-Line Start
```bash
curl -fsSL https://raw.githubusercontent.com/jontk/slurm-exporter/main/sandbox/start.sh | bash
```

### Option 2: Manual Setup
```bash
# Clone the repository
git clone https://github.com/jontk/slurm-exporter.git
cd slurm-exporter/sandbox

# Start the sandbox
docker-compose up -d

# Wait for services to be ready (2-3 minutes)
./wait-for-ready.sh

# Open your browser
open http://localhost:8888  # Tutorial homepage
```

## 📊 Services & URLs

| Service | URL | Credentials | Purpose |
|---------|-----|-------------|---------|
| **Tutorial Homepage** | http://localhost:8888 | - | Interactive learning portal |
| **Grafana** | http://localhost:3000 | admin/admin | Dashboards and visualization |
| **Prometheus** | http://localhost:9090 | - | Metrics storage and queries |
| **SLURM Exporter** | http://localhost:8080 | - | Main exporter with debug endpoints |
| **Alertmanager** | http://localhost:9093 | - | Alert routing and management |
| **SLURM API** | http://localhost:6820 | - | SLURM REST API (direct access) |

## 🎯 What's Included

### SLURM Cluster Simulation
- **Controller**: Full SLURM controller with REST API
- **Compute Nodes**: Simulated worker nodes
- **Job Generator**: Automatically creates realistic workload
- **Sample Data**: Jobs, users, partitions, and resource usage

### Monitoring Stack  
- **SLURM Exporter**: Fully configured with all collectors enabled
- **Prometheus**: Pre-configured to scrape all services
- **Grafana**: All 5 dashboards pre-imported
- **Alertmanager**: Tutorial-focused alert routing

### Tutorial Environment
- **Interactive Tutorials**: Step-by-step guided learning
- **Sandbox Configuration**: Optimized for learning
- **Sample Scenarios**: Real-world troubleshooting cases
- **Performance Testing**: Load testing and optimization tools

## 📚 Tutorial Workflow

### 1. Start with the Homepage
Visit http://localhost:8888 for the interactive tutorial portal.

### 2. Follow Learning Paths
- **Beginner**: Basic setup and configuration
- **Operator**: Monitoring and troubleshooting  
- **Expert**: Performance optimization and advanced features

### 3. Hands-On Practice
Each tutorial includes:
- Step-by-step instructions
- Copy-paste commands
- Expected outcomes
- Troubleshooting exercises

### 4. Experiment Freely
The sandbox resets automatically, so feel free to:
- Break things on purpose
- Test edge cases
- Try different configurations
- Explore failure scenarios

## 🔧 Customization

### Modify SLURM Configuration
```bash
# Edit SLURM cluster configuration
vim slurm-configs/slurm.conf

# Add more nodes or partitions
vim slurm-configs/slurm.conf

# Restart SLURM services
docker-compose restart slurm-controller slurm-node
```

### Adjust Workload Patterns
```bash
# Modify job generation patterns
vim sample-jobs/workload.sh

# Create custom job scenarios
vim sample-jobs/tutorial-jobs.sh

# Restart job generator
docker-compose restart job-generator
```

### Configure Monitoring
```bash
# Adjust scrape intervals
vim prometheus.yml

# Modify alert thresholds
vim alert-rules.yml

# Customize dashboards
# Import custom dashboards via Grafana UI
```

## 🐛 Troubleshooting

### Services Not Starting
```bash
# Check service status
docker-compose ps

# View service logs
docker-compose logs slurm-exporter
docker-compose logs slurm-controller

# Restart specific service
docker-compose restart <service-name>
```

### No Metrics Data
```bash
# Check SLURM API connectivity
curl http://localhost:6820/slurm/v0.0.40/diag

# Verify exporter health
curl http://localhost:8080/health

# Debug collection issues
curl http://localhost:8080/debug/collectors
```

### Performance Issues
```bash
# Check resource usage
docker stats

# Increase resources in docker-compose.yml
# Add under services:
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 2G
```

## 🧪 Advanced Usage

### Load Testing
```bash
# Run performance benchmarks
./scripts/sandbox-loadtest.sh

# Monitor performance impact
./scripts/monitor-performance.sh
```

### Custom Scenarios
```bash
# Simulate node failures
docker-compose exec slurm-controller scontrol update NodeName=node001 State=DOWN

# Create resource contention
docker-compose exec job-generator ./create-heavy-workload.sh

# Test circuit breaker
docker-compose stop slurm-controller
# Watch circuit breaker open in http://localhost:8080/debug/health
```

### Integration Testing
```bash
# Test with external Prometheus
vim docker-compose.override.yml

# Connect to external Grafana
# Export dashboards: ./export-dashboards.sh

# Test Kubernetes deployment
./scripts/test-k8s-migration.sh
```

## 📦 Environment Details

### Generated Workload
- **Job Types**: CPU, GPU, memory-intensive, short/long running
- **Users**: 10 simulated users with different usage patterns  
- **Partitions**: debug, compute, gpu, highmem partitions
- **Failure Rate**: ~5% job failure rate for realistic testing

### Metrics Coverage
- **150+ Metrics**: All SLURM Exporter metrics available
- **High Cardinality**: Realistic label combinations
- **Time Series**: 15-second collection interval
- **Retention**: 7 days of metric history

### Alert Scenarios
- **Tutorial Alerts**: Learning-focused notifications
- **Production Alerts**: Real-world alert examples
- **Severity Levels**: Critical, warning, and info alerts
- **Custom Routing**: Different handling per tutorial

## 🔄 Reset & Cleanup

### Soft Reset (Keep Data)
```bash
# Restart services
docker-compose restart

# Clear Prometheus data
docker volume rm sandbox_prometheus-data
docker-compose up prometheus -d
```

### Full Reset
```bash
# Stop and remove everything
docker-compose down -v

# Remove all data
docker volume prune -f

# Start fresh
docker-compose up -d
```

### Selective Reset
```bash
# Reset only SLURM data
docker-compose restart slurm-controller slurm-node job-generator

# Reset only monitoring data  
docker-compose down prometheus grafana alertmanager
docker volume rm sandbox_prometheus-data sandbox_grafana-data
docker-compose up -d prometheus grafana alertmanager
```

## 💡 Tips & Best Practices

### Learning Efficiently
1. **Start Simple**: Begin with Tutorial 1, don't skip ahead
2. **Hands-On**: Actually run the commands, don't just read
3. **Break Things**: The sandbox is safe for experimentation
4. **Ask Questions**: Use GitHub Discussions for help

### Troubleshooting Practice
1. **Use Decision Trees**: Follow the troubleshooting guide systematically
2. **Check Logs**: Always look at service logs first
3. **Verify Connectivity**: Test each component individually
4. **Monitor Resources**: Watch CPU/memory usage

### Performance Testing
1. **Baseline First**: Measure performance before optimizations
2. **Change One Thing**: Don't modify multiple settings at once
3. **Document Results**: Keep notes on what works
4. **Share Findings**: Help others learn from your experiments

## 🤝 Contributing

### Improving the Sandbox
- **Add Scenarios**: Create new tutorial scenarios
- **Fix Issues**: Report and fix bugs you encounter  
- **Update Configs**: Improve default configurations
- **Share Dashboards**: Contribute useful visualizations

### Documentation
- **Tutorial Feedback**: Suggest improvements
- **Missing Content**: Identify gaps in coverage
- **Clarifications**: Help make instructions clearer
- **Translations**: Help make content accessible

## 📞 Support

### Getting Help
1. **Tutorial Issues**: Check the troubleshooting section
2. **Technical Problems**: Open GitHub issues
3. **Questions**: Use GitHub Discussions
4. **Community**: Join our Discord/Slack

### Providing Feedback
- **What worked well**: Share success stories
- **What was confusing**: Help us improve
- **Missing features**: Suggest enhancements
- **Bug reports**: Include full reproduction steps

---

**Ready to learn?** Start with `docker-compose up -d` and visit http://localhost:8888!

**Need help?** Check our [troubleshooting guide](../docs/troubleshooting.md) or ask in [GitHub Discussions](https://github.com/jontk/slurm-exporter/discussions).