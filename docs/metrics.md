# SLURM Exporter Metrics Reference

This document provides a comprehensive reference for all metrics exposed by the SLURM Prometheus Exporter, including interpretation guides and use cases.

## Table of Contents

- [Metric Naming Convention](#metric-naming-convention)
- [Cluster Metrics](#cluster-metrics)
- [Node Metrics](#node-metrics)
- [Job Metrics](#job-metrics)
- [User and Account Metrics](#user-and-account-metrics)
- [Partition Metrics](#partition-metrics)
- [Performance Metrics](#performance-metrics)
- [Exporter Metrics](#exporter-metrics)
- [Use Cases and Examples](#use-cases-and-examples)
- [Alerting Guidelines](#alerting-guidelines)
- [Dashboard Templates](#dashboard-templates)

## Metric Naming Convention

All SLURM metrics follow the Prometheus naming conventions:

- **Prefix**: `slurm_` for all SLURM-related metrics
- **Subsystem**: Component name (e.g., `cluster_`, `node_`, `job_`)
- **Unit**: Included in metric name (e.g., `_bytes`, `_seconds`, `_total`)
- **Type suffix**: `_total` for counters, `_info` for information metrics

Example: `slurm_node_memory_total_bytes`

## Cluster Metrics

### slurm_cluster_info

**Type**: Gauge  
**Description**: SLURM cluster information and configuration  
**Labels**:
- `cluster_name`: Name of the SLURM cluster
- `version`: SLURM version
- `control_host`: Primary controller hostname
- `backup_host`: Backup controller hostname

**Example**:
```
slurm_cluster_info{cluster_name="hpc-prod",version="23.02.5",control_host="ctrl1",backup_host="ctrl2"} 1
```

**Use Cases**:
- Inventory tracking of SLURM clusters
- Version compatibility monitoring
- Controller failover detection

### slurm_cluster_nodes_total

**Type**: Gauge  
**Description**: Total number of nodes in the cluster by state  
**Labels**:
- `state`: Node state (idle, allocated, mixed, down, drain, etc.)

**Example**:
```
slurm_cluster_nodes_total{state="idle"} 150
slurm_cluster_nodes_total{state="allocated"} 300
slurm_cluster_nodes_total{state="down"} 5
```

**Use Cases**:
- Cluster health monitoring
- Capacity planning
- Alert on high percentage of down nodes

**Interpretation**:
- **idle**: Nodes available for new jobs
- **allocated**: Nodes fully utilized
- **mixed**: Nodes partially allocated
- **down**: Failed nodes requiring attention
- **drain**: Nodes being prepared for maintenance

### slurm_cluster_cpus_total

**Type**: Gauge  
**Description**: Total number of CPUs in the cluster by state  
**Labels**:
- `state`: CPU state (idle, allocated, other, total)

**Example**:
```
slurm_cluster_cpus_total{state="idle"} 2400
slurm_cluster_cpus_total{state="allocated"} 4800
slurm_cluster_cpus_total{state="total"} 7200
```

**Use Cases**:
- CPU utilization tracking
- Efficiency metrics calculation
- Capacity forecasting

**Queries**:
```promql
# CPU utilization percentage
100 * slurm_cluster_cpus_total{state="allocated"} / slurm_cluster_cpus_total{state="total"}

# Available CPU capacity
slurm_cluster_cpus_total{state="idle"} + (slurm_cluster_cpus_total{state="total"} - slurm_cluster_cpus_total{state="allocated"} - slurm_cluster_cpus_total{state="idle"})
```

### slurm_cluster_memory_total_bytes

**Type**: Gauge  
**Description**: Total memory in the cluster by state  
**Labels**:
- `state`: Memory state (allocated, total)

**Example**:
```
slurm_cluster_memory_total_bytes{state="total"} 1.099511627776e+13  # 10TB
slurm_cluster_memory_total_bytes{state="allocated"} 8.796093022208e+12  # 8TB
```

**Use Cases**:
- Memory pressure monitoring
- Memory utilization trends
- Capacity planning for memory-intensive workloads

### slurm_cluster_jobs_total

**Type**: Gauge  
**Description**: Total number of jobs by state  
**Labels**:
- `state`: Job state (pending, running, suspended, completed, failed)

**Example**:
```
slurm_cluster_jobs_total{state="running"} 1250
slurm_cluster_jobs_total{state="pending"} 3400
slurm_cluster_jobs_total{state="suspended"} 45
```

**Use Cases**:
- Queue depth monitoring
- Job throughput tracking
- Backlog detection

**Alerts**:
```yaml
- alert: HighJobQueueDepth
  expr: slurm_cluster_jobs_total{state="pending"} > 5000
  for: 30m
  annotations:
    summary: "High number of pending jobs ({{ $value }})"
```

## Node Metrics

### slurm_node_info

**Type**: Gauge  
**Description**: Node information and specifications  
**Labels**:
- `node`: Node name
- `partition`: Associated partition(s)
- `arch`: CPU architecture
- `os`: Operating system
- `features`: Node features (comma-separated)

**Example**:
```
slurm_node_info{node="compute001",partition="gpu,general",arch="x86_64",os="Linux",features="gpu,infiniband"} 1
```

**Use Cases**:
- Node inventory management
- Feature-based job routing verification
- Heterogeneous cluster monitoring

### slurm_node_state

**Type**: Gauge  
**Description**: Current state of the node (1 if in state, 0 otherwise)  
**Labels**:
- `node`: Node name
- `state`: Node state

**Example**:
```
slurm_node_state{node="compute001",state="idle"} 1
slurm_node_state{node="compute001",state="allocated"} 0
slurm_node_state{node="compute002",state="down"} 1
```

**Use Cases**:
- Node availability tracking
- Maintenance window monitoring
- Failure detection

**Queries**:
```promql
# Nodes currently down
slurm_node_state{state="down"} == 1

# Count of nodes in each state
sum by (state) (slurm_node_state == 1)
```

### slurm_node_cpus_total

**Type**: Gauge  
**Description**: Total number of CPUs on the node  
**Labels**:
- `node`: Node name

**Example**:
```
slurm_node_cpus_total{node="compute001"} 64
slurm_node_cpus_total{node="gpu001"} 48
```

### slurm_node_cpus_allocated

**Type**: Gauge  
**Description**: Number of allocated CPUs on the node  
**Labels**:
- `node`: Node name

**Example**:
```
slurm_node_cpus_allocated{node="compute001"} 32
```

**Use Cases**:
- Node-level CPU utilization
- Load balancing verification
- Hot spot detection

**Queries**:
```promql
# Node CPU utilization percentage
100 * slurm_node_cpus_allocated / slurm_node_cpus_total

# Nodes with high CPU utilization
(slurm_node_cpus_allocated / slurm_node_cpus_total) > 0.9
```

### slurm_node_memory_total_bytes

**Type**: Gauge  
**Description**: Total memory on the node in bytes  
**Labels**:
- `node`: Node name

**Example**:
```
slurm_node_memory_total_bytes{node="compute001"} 2.74877906944e+11  # 256GB
```

### slurm_node_memory_allocated_bytes

**Type**: Gauge  
**Description**: Allocated memory on the node in bytes  
**Labels**:
- `node`: Node name

**Example**:
```
slurm_node_memory_allocated_bytes{node="compute001"} 2.06158430208e+11  # 192GB
```

**Use Cases**:
- Memory pressure identification
- Memory leak detection
- Allocation efficiency

### slurm_node_gpus_total

**Type**: Gauge  
**Description**: Total number of GPUs on the node  
**Labels**:
- `node`: Node name
- `gpu_type`: GPU model/type

**Example**:
```
slurm_node_gpus_total{node="gpu001",gpu_type="nvidia_a100"} 8
slurm_node_gpus_total{node="gpu002",gpu_type="nvidia_v100"} 4
```

### slurm_node_gpus_allocated

**Type**: Gauge  
**Description**: Number of allocated GPUs on the node  
**Labels**:
- `node`: Node name
- `gpu_type`: GPU model/type

**Example**:
```
slurm_node_gpus_allocated{node="gpu001",gpu_type="nvidia_a100"} 6
```

**Use Cases**:
- GPU utilization tracking
- GPU resource availability
- Cost optimization for GPU resources

### slurm_node_power_usage_watts

**Type**: Gauge  
**Description**: Current power consumption of the node in watts (if available)  
**Labels**:
- `node`: Node name

**Example**:
```
slurm_node_power_usage_watts{node="compute001"} 450
```

**Use Cases**:
- Power efficiency monitoring
- Data center capacity planning
- Environmental impact tracking

### slurm_node_temperature_celsius

**Type**: Gauge  
**Description**: Node temperature in Celsius (if available)  
**Labels**:
- `node`: Node name
- `sensor`: Temperature sensor location

**Example**:
```
slurm_node_temperature_celsius{node="compute001",sensor="cpu0"} 65
slurm_node_temperature_celsius{node="compute001",sensor="inlet"} 25
```

**Use Cases**:
- Thermal monitoring
- Cooling efficiency
- Predictive maintenance

## Job Metrics

### slurm_job_info

**Type**: Gauge  
**Description**: Job information (value is always 1)  
**Labels**:
- `job_id`: Job ID
- `job_name`: Job name
- `user`: Username
- `account`: Account name
- `partition`: Partition name
- `qos`: Quality of Service

**Example**:
```
slurm_job_info{job_id="12345",job_name="simulation_run_1",user="jdoe",account="physics",partition="general",qos="normal"} 1
```

**Use Cases**:
- Job tracking and auditing
- User activity monitoring
- Account resource usage

### slurm_job_state

**Type**: Gauge  
**Description**: Current job state (1 if in state, 0 otherwise)  
**Labels**:
- `job_id`: Job ID
- `state`: Job state

**Example**:
```
slurm_job_state{job_id="12345",state="running"} 1
slurm_job_state{job_id="12346",state="pending"} 1
```

**States**:
- **pending**: Waiting for resources
- **running**: Currently executing
- **suspended**: Temporarily paused
- **completing**: In cleanup phase
- **completed**: Successfully finished
- **failed**: Terminated with error
- **timeout**: Exceeded time limit
- **cancelled**: Manually terminated

### slurm_job_cpus_requested

**Type**: Gauge  
**Description**: Number of CPUs requested by the job  
**Labels**:
- `job_id`: Job ID
- `user`: Username
- `account`: Account name

**Example**:
```
slurm_job_cpus_requested{job_id="12345",user="jdoe",account="physics"} 128
```

### slurm_job_memory_requested_bytes

**Type**: Gauge  
**Description**: Memory requested by the job in bytes  
**Labels**:
- `job_id`: Job ID
- `user`: Username
- `account`: Account name

**Example**:
```
slurm_job_memory_requested_bytes{job_id="12345",user="jdoe",account="physics"} 5.49755813888e+11  # 512GB
```

### slurm_job_gpus_requested

**Type**: Gauge  
**Description**: Number of GPUs requested by the job  
**Labels**:
- `job_id`: Job ID
- `user`: Username
- `account`: Account name
- `gpu_type`: Requested GPU type

**Example**:
```
slurm_job_gpus_requested{job_id="12345",user="jdoe",account="ml_research",gpu_type="nvidia_a100"} 4
```

### slurm_job_wait_time_seconds

**Type**: Gauge  
**Description**: Time job waited in queue before starting (seconds)  
**Labels**:
- `job_id`: Job ID
- `user`: Username
- `partition`: Partition name

**Example**:
```
slurm_job_wait_time_seconds{job_id="12345",user="jdoe",partition="general"} 3600  # 1 hour
```

**Use Cases**:
- Queue performance analysis
- Fair-share policy effectiveness
- User experience monitoring

**Queries**:
```promql
# Average wait time by partition
avg by (partition) (slurm_job_wait_time_seconds)

# Jobs waiting more than 24 hours
slurm_job_wait_time_seconds > 86400
```

### slurm_job_runtime_seconds

**Type**: Gauge  
**Description**: Current runtime of the job in seconds  
**Labels**:
- `job_id`: Job ID
- `user`: Username

**Example**:
```
slurm_job_runtime_seconds{job_id="12345",user="jdoe"} 7200  # 2 hours
```

### slurm_job_time_limit_seconds

**Type**: Gauge  
**Description**: Job time limit in seconds  
**Labels**:
- `job_id`: Job ID

**Example**:
```
slurm_job_time_limit_seconds{job_id="12345"} 86400  # 24 hours
```

**Use Cases**:
- Jobs approaching time limit
- Time limit policy compliance
- Walltime accuracy improvement

**Queries**:
```promql
# Jobs using >90% of time limit
slurm_job_runtime_seconds / slurm_job_time_limit_seconds > 0.9
```

### slurm_job_nodes_allocated

**Type**: Gauge  
**Description**: Number of nodes allocated to the job  
**Labels**:
- `job_id`: Job ID
- `user`: Username

**Example**:
```
slurm_job_nodes_allocated{job_id="12345",user="jdoe"} 16
```

### slurm_job_priority

**Type**: Gauge  
**Description**: Job priority score  
**Labels**:
- `job_id`: Job ID
- `user`: Username
- `partition`: Partition name

**Example**:
```
slurm_job_priority{job_id="12345",user="jdoe",partition="general"} 1000000
```

**Use Cases**:
- Priority queue analysis
- Fair-share monitoring
- QoS effectiveness

## User and Account Metrics

### slurm_user_jobs_total

**Type**: Gauge  
**Description**: Total number of jobs per user by state  
**Labels**:
- `user`: Username
- `account`: Account name
- `state`: Job state

**Example**:
```
slurm_user_jobs_total{user="jdoe",account="physics",state="running"} 5
slurm_user_jobs_total{user="jdoe",account="physics",state="pending"} 12
```

**Use Cases**:
- User activity monitoring
- Resource hogging detection
- Fair-share verification

**Queries**:
```promql
# Users with most running jobs
topk(10, sum by (user) (slurm_user_jobs_total{state="running"}))

# Users with high pending/running ratio
sum by (user) (slurm_user_jobs_total{state="pending"}) 
/ 
sum by (user) (slurm_user_jobs_total{state="running"}) > 10
```

### slurm_user_cpus_allocated

**Type**: Gauge  
**Description**: Total CPUs currently allocated to user  
**Labels**:
- `user`: Username
- `account`: Account name

**Example**:
```
slurm_user_cpus_allocated{user="jdoe",account="physics"} 640
```

### slurm_user_memory_allocated_bytes

**Type**: Gauge  
**Description**: Total memory currently allocated to user  
**Labels**:
- `user`: Username
- `account`: Account name

**Example**:
```
slurm_user_memory_allocated_bytes{user="jdoe",account="physics"} 2.748779069440e+12  # 2.5TB
```

### slurm_user_gpus_allocated

**Type**: Gauge  
**Description**: Total GPUs currently allocated to user  
**Labels**:
- `user`: Username
- `account`: Account name
- `gpu_type`: GPU type

**Example**:
```
slurm_user_gpus_allocated{user="jdoe",account="ml_research",gpu_type="nvidia_a100"} 16
```

### slurm_account_jobs_total

**Type**: Gauge  
**Description**: Total number of jobs per account by state  
**Labels**:
- `account`: Account name
- `state`: Job state

**Example**:
```
slurm_account_jobs_total{account="physics",state="running"} 45
slurm_account_jobs_total{account="chemistry",state="pending"} 120
```

### slurm_account_cpus_allocated

**Type**: Gauge  
**Description**: Total CPUs allocated to account  
**Labels**:
- `account`: Account name

**Example**:
```
slurm_account_cpus_allocated{account="physics"} 2048
```

### slurm_account_fairshare

**Type**: Gauge  
**Description**: Account fair-share factor  
**Labels**:
- `account`: Account name

**Example**:
```
slurm_account_fairshare{account="physics"} 0.95
slurm_account_fairshare{account="chemistry"} 1.05
```

**Use Cases**:
- Fair-share policy monitoring
- Resource allocation balance
- Account priority adjustments

**Interpretation**:
- Values < 1.0: Account is over-utilizing resources
- Values > 1.0: Account is under-utilizing resources
- Value = 1.0: Account usage matches allocation

### slurm_account_quota_cpu_hours

**Type**: Gauge  
**Description**: CPU hour quota for account (if configured)  
**Labels**:
- `account`: Account name
- `period`: Quota period (monthly, yearly)

**Example**:
```
slurm_account_quota_cpu_hours{account="physics",period="monthly"} 1000000
```

### slurm_account_usage_cpu_hours

**Type**: Counter  
**Description**: CPU hours used by account  
**Labels**:
- `account`: Account name
- `period`: Usage period

**Example**:
```
slurm_account_usage_cpu_hours{account="physics",period="monthly"} 750000
```

**Use Cases**:
- Quota enforcement monitoring
- Billing and chargeback
- Capacity planning

**Queries**:
```promql
# Quota utilization percentage
100 * slurm_account_usage_cpu_hours / slurm_account_quota_cpu_hours

# Accounts near quota limit (>90%)
(slurm_account_usage_cpu_hours / slurm_account_quota_cpu_hours) > 0.9
```

## Partition Metrics

### slurm_partition_info

**Type**: Gauge  
**Description**: Partition configuration information  
**Labels**:
- `partition`: Partition name
- `state`: Partition state (up, down, drain, inactive)
- `default`: Whether this is the default partition
- `allow_groups`: Allowed groups (comma-separated)
- `allow_accounts`: Allowed accounts (comma-separated)

**Example**:
```
slurm_partition_info{partition="general",state="up",default="yes",allow_groups="all",allow_accounts="all"} 1
slurm_partition_info{partition="gpu",state="up",default="no",allow_groups="gpu_users",allow_accounts="ml_research,physics"} 1
```

### slurm_partition_nodes_total

**Type**: Gauge  
**Description**: Total nodes in partition by state  
**Labels**:
- `partition`: Partition name
- `state`: Node state

**Example**:
```
slurm_partition_nodes_total{partition="general",state="idle"} 50
slurm_partition_nodes_total{partition="general",state="allocated"} 100
slurm_partition_nodes_total{partition="general",state="total"} 150
```

### slurm_partition_cpus_total

**Type**: Gauge  
**Description**: Total CPUs in partition by state  
**Labels**:
- `partition`: Partition name
- `state`: CPU state

**Example**:
```
slurm_partition_cpus_total{partition="general",state="idle"} 800
slurm_partition_cpus_total{partition="general",state="allocated"} 3200
slurm_partition_cpus_total{partition="general",state="total"} 4000
```

**Use Cases**:
- Partition utilization monitoring
- Load balancing between partitions
- Capacity planning per partition

### slurm_partition_jobs_total

**Type**: Gauge  
**Description**: Number of jobs in partition by state  
**Labels**:
- `partition`: Partition name
- `state`: Job state

**Example**:
```
slurm_partition_jobs_total{partition="general",state="running"} 250
slurm_partition_jobs_total{partition="general",state="pending"} 500
```

### slurm_partition_max_time_seconds

**Type**: Gauge  
**Description**: Maximum job time limit for partition  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_partition_max_time_seconds{partition="general"} 259200     # 3 days
slurm_partition_max_time_seconds{partition="long"} 2592000       # 30 days
slurm_partition_max_time_seconds{partition="infinite"} -1        # No limit
```

### slurm_partition_default_time_seconds

**Type**: Gauge  
**Description**: Default job time limit for partition  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_partition_default_time_seconds{partition="general"} 3600   # 1 hour
```

### slurm_partition_priority_tier

**Type**: Gauge  
**Description**: Priority tier of the partition  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_partition_priority_tier{partition="high_priority"} 1000
slurm_partition_priority_tier{partition="general"} 100
slurm_partition_priority_tier{partition="low_priority"} 10
```

## Performance Metrics

### slurm_scheduler_cycle_duration_seconds

**Type**: Histogram  
**Description**: Time taken for scheduler cycle  
**Labels**: None

**Example**:
```
slurm_scheduler_cycle_duration_seconds_bucket{le="0.1"} 950
slurm_scheduler_cycle_duration_seconds_bucket{le="0.5"} 990
slurm_scheduler_cycle_duration_seconds_bucket{le="1"} 999
slurm_scheduler_cycle_duration_seconds_count 1000
slurm_scheduler_cycle_duration_seconds_sum 125.5
```

**Use Cases**:
- Scheduler performance monitoring
- Bottleneck identification
- Configuration tuning

**Queries**:
```promql
# 95th percentile scheduler cycle time
histogram_quantile(0.95, rate(slurm_scheduler_cycle_duration_seconds_bucket[5m]))

# Scheduler cycles taking >1 second
rate(slurm_scheduler_cycle_duration_seconds_bucket{le="1"}[5m])
```

### slurm_job_throughput_per_minute

**Type**: Gauge  
**Description**: Number of jobs completed per minute  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_job_throughput_per_minute{partition="general"} 45
slurm_job_throughput_per_minute{partition="gpu"} 12
```

### slurm_job_start_rate_per_minute

**Type**: Gauge  
**Description**: Number of jobs started per minute  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_job_start_rate_per_minute{partition="general"} 50
```

### slurm_backfill_cycle_duration_seconds

**Type**: Histogram  
**Description**: Duration of backfill scheduler cycles  
**Labels**: None

**Example**:
```
slurm_backfill_cycle_duration_seconds_bucket{le="1"} 800
slurm_backfill_cycle_duration_seconds_bucket{le="5"} 950
slurm_backfill_cycle_duration_seconds_bucket{le="10"} 990
```

**Use Cases**:
- Backfill efficiency monitoring
- Scheduler tuning
- Performance optimization

### slurm_backfill_jobs_total

**Type**: Counter  
**Description**: Total number of jobs scheduled via backfill  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_backfill_jobs_total{partition="general"} 15420
```

### slurm_cluster_efficiency_percentage

**Type**: Gauge  
**Description**: Overall cluster efficiency (allocated/total resources)  
**Labels**:
- `resource`: Resource type (cpu, memory, gpu)

**Example**:
```
slurm_cluster_efficiency_percentage{resource="cpu"} 85.5
slurm_cluster_efficiency_percentage{resource="memory"} 72.3
slurm_cluster_efficiency_percentage{resource="gpu"} 95.2
```

**Use Cases**:
- Resource utilization optimization
- Capacity planning
- Cost efficiency tracking

### slurm_job_efficiency_percentage

**Type**: Histogram  
**Description**: Job efficiency (actual usage vs allocated)  
**Labels**:
- `partition`: Partition name

**Example**:
```
slurm_job_efficiency_percentage_bucket{partition="general",le="25"} 100
slurm_job_efficiency_percentage_bucket{partition="general",le="50"} 250
slurm_job_efficiency_percentage_bucket{partition="general",le="75"} 700
slurm_job_efficiency_percentage_bucket{partition="general",le="100"} 1000
```

**Use Cases**:
- Resource waste identification
- User education targets
- Policy enforcement

**Queries**:
```promql
# Median job efficiency
histogram_quantile(0.5, rate(slurm_job_efficiency_percentage_bucket[1h]))

# Percentage of jobs with <50% efficiency
sum(rate(slurm_job_efficiency_percentage_bucket{le="50"}[1h])) 
/ 
sum(rate(slurm_job_efficiency_percentage_count[1h])) * 100
```

## Exporter Metrics

### slurm_exporter_up

**Type**: Gauge  
**Description**: Whether the SLURM exporter is up (1) or down (0)  
**Labels**: None

**Example**:
```
slurm_exporter_up 1
```

### slurm_exporter_collection_duration_seconds

**Type**: Histogram  
**Description**: Time taken to collect metrics from SLURM  
**Labels**:
- `collector`: Collector name (nodes, jobs, users, etc.)

**Example**:
```
slurm_exporter_collection_duration_seconds_bucket{collector="nodes",le="0.5"} 950
slurm_exporter_collection_duration_seconds_bucket{collector="nodes",le="1"} 990
slurm_exporter_collection_duration_seconds_count{collector="nodes"} 1000
```

**Use Cases**:
- Exporter performance monitoring
- Collection timeout tuning
- Bottleneck identification

### slurm_exporter_collection_errors_total

**Type**: Counter  
**Description**: Total number of collection errors  
**Labels**:
- `collector`: Collector name
- `error_type`: Type of error

**Example**:
```
slurm_exporter_collection_errors_total{collector="nodes",error_type="timeout"} 5
slurm_exporter_collection_errors_total{collector="jobs",error_type="api_error"} 2
```

**Use Cases**:
- Error rate monitoring
- Reliability tracking
- Troubleshooting

**Alerts**:
```yaml
- alert: HighExporterErrorRate
  expr: rate(slurm_exporter_collection_errors_total[5m]) > 0.1
  for: 10m
  annotations:
    summary: "High error rate in SLURM exporter {{ $labels.collector }}"
```

### slurm_exporter_last_collection_timestamp

**Type**: Gauge  
**Description**: Unix timestamp of last successful collection  
**Labels**:
- `collector`: Collector name

**Example**:
```
slurm_exporter_last_collection_timestamp{collector="nodes"} 1704067200
```

**Use Cases**:
- Freshness monitoring
- Stale data detection
- Collection interval verification

**Queries**:
```promql
# Time since last collection
time() - slurm_exporter_last_collection_timestamp

# Collectors not updated in last 5 minutes
(time() - slurm_exporter_last_collection_timestamp) > 300
```

### slurm_exporter_api_requests_total

**Type**: Counter  
**Description**: Total API requests made to SLURM  
**Labels**:
- `endpoint`: API endpoint
- `status`: HTTP status code

**Example**:
```
slurm_exporter_api_requests_total{endpoint="/nodes",status="200"} 10000
slurm_exporter_api_requests_total{endpoint="/jobs",status="200"} 15000
slurm_exporter_api_requests_total{endpoint="/jobs",status="500"} 10
```

### slurm_exporter_api_request_duration_seconds

**Type**: Histogram  
**Description**: Duration of API requests to SLURM  
**Labels**:
- `endpoint`: API endpoint

**Example**:
```
slurm_exporter_api_request_duration_seconds_bucket{endpoint="/nodes",le="0.1"} 8000
slurm_exporter_api_request_duration_seconds_bucket{endpoint="/nodes",le="0.5"} 9800
```

## Use Cases and Examples

### Capacity Planning

**Objective**: Predict when additional resources are needed

```promql
# CPU utilization trend over 7 days
avg_over_time(
  (slurm_cluster_cpus_total{state="allocated"} / slurm_cluster_cpus_total{state="total"})
  [7d:1h]
)

# Memory pressure indicator
(slurm_cluster_memory_total_bytes{state="allocated"} / slurm_cluster_memory_total_bytes{state="total"}) > 0.9

# Queue depth trend
deriv(slurm_cluster_jobs_total{state="pending"}[1h]) > 0
```

### User Experience Monitoring

**Objective**: Ensure good user experience and fair access

```promql
# Average wait time by partition (last hour)
avg by (partition) (slurm_job_wait_time_seconds)

# Users with excessive wait times
slurm_job_wait_time_seconds > 14400  # 4 hours

# Fair-share balance
stddev by () (slurm_account_fairshare) > 0.2
```

### Resource Efficiency

**Objective**: Identify and reduce resource waste

```promql
# Jobs with low CPU efficiency
histogram_quantile(0.5, slurm_job_efficiency_percentage_bucket) < 50

# Memory over-allocation
avg(
  slurm_node_memory_allocated_bytes / slurm_node_memory_total_bytes
) by (node) > 0.95

# Idle GPU resources
sum(slurm_node_gpus_total - slurm_node_gpus_allocated) > 10
```

### Reliability and Health

**Objective**: Maintain cluster health and availability

```promql
# Node failure rate
rate(slurm_node_state{state="down"}[1h]) > 0

# Scheduler performance degradation
histogram_quantile(0.95, slurm_scheduler_cycle_duration_seconds_bucket) > 2

# API errors
rate(slurm_exporter_api_requests_total{status!="200"}[5m]) > 0.01
```

### Cost Optimization

**Objective**: Optimize resource usage for cost efficiency

```promql
# GPU utilization for cost tracking
avg(slurm_node_gpus_allocated / slurm_node_gpus_total) by (gpu_type)

# Power consumption trends
sum(slurm_node_power_usage_watts) by (partition)

# Underutilized expensive resources
(slurm_node_gpus_allocated{gpu_type="nvidia_a100"} / slurm_node_gpus_total{gpu_type="nvidia_a100"}) < 0.7
```

## Alerting Guidelines

### Critical Alerts

```yaml
# Cluster availability
- alert: SlurmControllerDown
  expr: slurm_cluster_info == 0
  for: 5m
  severity: critical
  annotations:
    summary: "SLURM controller is not responding"

# Mass node failure
- alert: HighNodeFailureRate
  expr: |
    (sum(slurm_node_state{state="down"} == 1) / sum(slurm_node_state{state=~".*"} == 1)) > 0.1
  for: 10m
  severity: critical
  annotations:
    summary: "More than 10% of nodes are down"

# Scheduler failure
- alert: SchedulerStalled
  expr: |
    time() - slurm_exporter_last_collection_timestamp{collector="scheduler"} > 600
  severity: critical
  annotations:
    summary: "SLURM scheduler metrics not updated for 10 minutes"
```

### Warning Alerts

```yaml
# High resource utilization
- alert: HighCPUUtilization
  expr: |
    (slurm_cluster_cpus_total{state="allocated"} / slurm_cluster_cpus_total{state="total"}) > 0.95
  for: 30m
  severity: warning
  annotations:
    summary: "CPU utilization above 95% for 30 minutes"

# Queue buildup
- alert: ExcessivePendingJobs
  expr: slurm_cluster_jobs_total{state="pending"} > 1000
  for: 1h
  severity: warning
  annotations:
    summary: "More than 1000 jobs pending for 1 hour"

# Efficiency concerns
- alert: LowJobEfficiency
  expr: |
    histogram_quantile(0.5, slurm_job_efficiency_percentage_bucket) < 40
  for: 2h
  severity: warning
  annotations:
    summary: "Median job efficiency below 40%"
```

### Informational Alerts

```yaml
# Maintenance reminders
- alert: NodeInDrainState
  expr: slurm_node_state{state="drain"} == 1
  for: 24h
  severity: info
  annotations:
    summary: "Node {{ $labels.node }} in drain state for 24 hours"

# Quota warnings
- alert: ApproachingQuotaLimit
  expr: |
    (slurm_account_usage_cpu_hours / slurm_account_quota_cpu_hours) > 0.8
  severity: info
  annotations:
    summary: "Account {{ $labels.account }} at 80% of CPU hour quota"
```

## Dashboard Templates

### Executive Dashboard

Key metrics for management overview:

1. **Cluster Utilization Gauge**
   ```promql
   100 * slurm_cluster_cpus_total{state="allocated"} / slurm_cluster_cpus_total{state="total"}
   ```

2. **Job Throughput Timeline**
   ```promql
   sum(slurm_job_throughput_per_minute)
   ```

3. **Resource Efficiency Score**
   ```promql
   avg(slurm_cluster_efficiency_percentage)
   ```

4. **Cost per CPU Hour**
   ```promql
   sum(slurm_node_power_usage_watts) * 0.10 / 1000 / sum(slurm_cluster_cpus_total{state="allocated"})
   ```

### Operations Dashboard

Detailed view for system administrators:

1. **Node State Distribution**
   ```promql
   sum by (state) (slurm_node_state == 1)
   ```

2. **API Performance**
   ```promql
   histogram_quantile(0.95, rate(slurm_exporter_api_request_duration_seconds_bucket[5m]))
   ```

3. **Error Rate**
   ```promql
   sum(rate(slurm_exporter_collection_errors_total[5m])) by (collector)
   ```

4. **Scheduler Performance**
   ```promql
   histogram_quantile(0.99, slurm_scheduler_cycle_duration_seconds_bucket)
   ```

### User Dashboard

Self-service view for end users:

1. **My Jobs**
   ```promql
   slurm_user_jobs_total{user="$username"}
   ```

2. **Wait Time Comparison**
   ```promql
   avg(slurm_job_wait_time_seconds{user="$username"}) 
   vs 
   avg(slurm_job_wait_time_seconds)
   ```

3. **Resource Usage**
   ```promql
   slurm_user_cpus_allocated{user="$username"}
   slurm_user_memory_allocated_bytes{user="$username"}
   slurm_user_gpus_allocated{user="$username"}
   ```

4. **Job Efficiency**
   ```promql
   histogram_quantile(0.5, slurm_job_efficiency_percentage_bucket{user="$username"})
   ```

### Capacity Planning Dashboard

Forward-looking metrics for planning:

1. **Utilization Trends**
   ```promql
   predict_linear(slurm_cluster_cpus_total{state="allocated"}[30d], 86400 * 30)
   ```

2. **Queue Growth Rate**
   ```promql
   deriv(slurm_cluster_jobs_total{state="pending"}[7d])
   ```

3. **Peak Usage Patterns**
   ```promql
   max_over_time(slurm_cluster_cpus_total{state="allocated"}[1d])
   ```

4. **Resource Saturation Forecast**
   ```promql
   predict_linear(
     (slurm_cluster_cpus_total{state="allocated"} / slurm_cluster_cpus_total{state="total"})
     [30d], 
     86400 * 90
   ) > 0.95
   ```

## Best Practices

### Metric Selection

1. **Start with key metrics**:
   - Cluster utilization
   - Job wait times
   - Node availability
   - Queue depth

2. **Add detail gradually**:
   - User-level metrics for chargeback
   - Efficiency metrics for optimization
   - Performance metrics for tuning

3. **Avoid metric explosion**:
   - Use label limits for high-cardinality data
   - Aggregate where appropriate
   - Set retention policies

### Query Optimization

1. **Use recording rules for complex queries**:
   ```yaml
   - record: cluster:cpu_utilization:rate5m
     expr: |
       slurm_cluster_cpus_total{state="allocated"} 
       / 
       slurm_cluster_cpus_total{state="total"}
   ```

2. **Aggregate before graphing**:
   ```promql
   # Good: Pre-aggregate
   sum by (partition) (slurm_node_cpus_allocated)
   
   # Bad: Graph then aggregate
   slurm_node_cpus_allocated
   ```

3. **Use appropriate time ranges**:
   - Real-time: 5m-15m ranges
   - Trends: 1h-6h ranges
   - Historical: 1d-7d ranges

### Alert Design

1. **Layer alerts by severity**:
   - Critical: Immediate action required
   - Warning: Investigation needed
   - Info: Awareness only

2. **Include context in alerts**:
   - Current value
   - Threshold crossed
   - Duration of condition
   - Suggested actions

3. **Avoid alert fatigue**:
   - Set appropriate thresholds
   - Use time-based conditions
   - Group related alerts
   - Implement alert suppression

This comprehensive metrics reference provides the foundation for effective SLURM cluster monitoring, optimization, and capacity planning.