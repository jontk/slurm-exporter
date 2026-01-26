# SLURM Exporter Complete Metrics Specification

This document defines the complete set of metrics designed to be collected by the slurm-exporter, organized by collector and metric type. This is the authoritative reference for understanding what metrics *should* be available based on the codebase implementation.

**Document Version**: 1.0
**Last Updated**: 2026-01-26
**Total Metrics Designed**: ~107 metrics across 12 collectors

## Table of Contents

- [Overview](#overview)
- [Collector Reference](#collector-reference)
  - [Cluster Collector](#1-cluster-collector)
  - [Nodes Collector](#2-nodes-collector)
  - [Jobs Collector](#3-jobs-collector)
  - [Accounts Collector](#4-accounts-collector)
  - [Partitions Collector](#5-partitions-collector)
  - [Users Collector](#6-users-collector)
  - [QoS Collector](#7-qos-collector)
  - [TRES Collector](#8-tres-collector)
  - [Licenses Collector](#9-licenses-collector)
  - [Shares/Fairshare Collector](#10-sharesfairshare-collector)
  - [System Collector](#11-system-collector)
  - [Scheduler Collector](#12-scheduler-collector)
- [Metric Type Summary](#metric-type-summary)
- [API Version Dependencies](#api-version-dependencies)
- [Data Conversions](#data-conversions)
- [Design Patterns](#design-patterns)

---

## Overview

The slurm-exporter collects comprehensive SLURM cluster monitoring metrics across 12 specialized collectors:

| Collector | File | Metrics | Status |
|-----------|------|---------|--------|
| Cluster | `cluster_simple.go` | 6 | Production |
| Nodes | `nodes_simple.go` | 6 | Production |
| Jobs | `jobs_simple.go` | 7 | Production |
| Accounts | `accounts_simple.go` | 6 | Production |
| Partitions | `partitions_simple.go` | 11 | Production |
| Users | `users_simple.go` | 6 | Production |
| QoS | `qos.go` | 13 | Production |
| TRES | `tres_simple.go` | 5 | Partial (node-level TODO) |
| Licenses | `licenses_simple.go` | 5 | Conditional (v0.0.42+) |
| Shares | `shares_simple.go` | 7 | Conditional (v0.0.42+) |
| System | `system_simple.go` | 14 | Production |
| Scheduler | `scheduler.go` | 3 | Production |
| **TOTAL** | | **~107** | |

---

## Collector Reference

### 1. Cluster Collector

**File**: `internal/collector/cluster_simple.go`
**Purpose**: High-level cluster status and configuration
**Metrics**: 6

#### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_cluster_info` | Gauge | Cluster metadata (info metric always 1.0) | cluster_name, control_host, control_port, rpc_version, plugin_version |
| `slurm_cluster_nodes_total` | Gauge | Total number of nodes | cluster_name |
| `slurm_cluster_cpus_total` | Gauge | Total CPUs in cluster | cluster_name |
| `slurm_cluster_jobs_total` | Gauge | Total jobs in cluster | cluster_name |
| `slurm_cluster_version_info` | Gauge | SLURM version information | cluster_name, major, minor, patch |
| `slurm_cluster_controllers_total` | Gauge | Number of active controllers | cluster_name |

**Label Breakdown**:
- `cluster_name`: From `ClusterName` field (default: "default_cluster")
- `control_host`: Primary controller hostname
- `control_port`: Controller port
- `rpc_version`: RPC API version string
- `plugin_version`: SLURM plugin version

**Data Source**: Cluster info manager's `Info()` and `Stats()` API calls

**Example Output**:
```
slurm_cluster_info{cluster_name="hpc-prod",control_host="ctrlnode1",control_port="6817",rpc_version="35.01",plugin_version="35.01.00"} 1
slurm_cluster_nodes_total{cluster_name="hpc-prod"} 128
slurm_cluster_cpus_total{cluster_name="hpc-prod"} 2048
slurm_cluster_jobs_total{cluster_name="hpc-prod"} 245
slurm_cluster_version_info{cluster_name="hpc-prod",major="23",minor="02",patch="5"} 1
slurm_cluster_controllers_total{cluster_name="hpc-prod"} 1
```

---

### 2. Nodes Collector

**File**: `internal/collector/nodes_simple.go`
**Purpose**: Per-node resource allocation and health
**Metrics**: 6

#### Metrics

| Metric | Type | Description | Labels | Notes |
|--------|------|-------------|--------|-------|
| `slurm_node_state` | Gauge | Node operational state | node, partition, state, reason, arch, os | 1.0=UP, 0.0=DOWN/DRAIN/other |
| `slurm_node_cpus_total` | Gauge | Total CPUs on node | node, partition | |
| `slurm_node_cpus_allocated` | Gauge | Estimated allocated CPUs | node, partition | Proportional to node state |
| `slurm_node_memory_total_bytes` | Gauge | Total memory in bytes | node, partition | Converted from MB |
| `slurm_node_memory_allocated_bytes` | Gauge | Estimated allocated memory | node, partition | Proportional to CPU allocation |
| `slurm_node_info` | Gauge | Node metadata | node, partition, state, reason, arch, os | Always 1.0 |

**Label Breakdown**:
- `node`: Node name/hostname
- `partition`: First partition node is assigned to
- `state`: Node state (UP, DOWN, DRAIN, MIXED, etc.)
- `reason`: Reason for state (e.g., "None", "Powering up")
- `arch`: CPU architecture (default: "x86_64")
- `os`: Operating system (default: "linux")

**CPU/Memory Allocation Estimation**:
- `ALLOCATED` or `MIXED`: 50% of total resources
- `COMPLETING`: 75% of total resources
- `IDLE`, `DRAIN`, `DOWN`: 0% (no allocation estimate)

**Sanity Checks**:
- Memory validation: < 1TB (skip if exceeds)
- CPU validation: Uses actual values

**Data Source**: Nodes manager API

**Example Output**:
```
slurm_node_state{node="node001",partition="compute",state="UP",reason="None",arch="x86_64",os="linux"} 1
slurm_node_cpus_total{node="node001",partition="compute"} 32
slurm_node_cpus_allocated{node="node001",partition="compute"} 16
slurm_node_memory_total_bytes{node="node001",partition="compute"} 274877906944
slurm_node_memory_allocated_bytes{node="node001",partition="compute"} 137438953472
slurm_node_info{node="node001",partition="compute",state="UP",reason="None",arch="x86_64",os="linux"} 1
```

---

### 3. Jobs Collector

**File**: `internal/collector/jobs_simple.go`
**Purpose**: Per-job state and resource metrics
**Metrics**: 7
**Features**: Filtering, cardinality management, custom labels

#### Metrics

| Metric | Type | Description | Labels | Notes |
|--------|------|-------------|--------|-------|
| `slurm_job_state` | Gauge | Job operational state | job_id, job_name, user, partition, account, qos | 1.0=RUNNING/COMPLETING, 0.0=other |
| `slurm_job_queue_time_seconds` | Gauge | Time from submit to start | partition, account | Only for completed queuing |
| `slurm_job_run_time_seconds` | Gauge | Time from start to now | partition, account | Only for active jobs (RUNNING/COMPLETING) |
| `slurm_job_cpus` | Gauge | CPUs allocated to job | job_id, job_name, user, partition, account, qos | Sanity check: max 10k |
| `slurm_job_memory_bytes` | Gauge | Memory allocated to job | job_id, job_name, user, partition, account, qos | Converted from MB, max 1TB |
| `slurm_job_nodes` | Gauge | Nodes allocated to job | job_id, job_name, user, partition, account, qos | Default 1 if empty |
| `slurm_job_info` | Gauge | Job metadata | job_id, job_name, user, partition, account, qos | Always 1.0 |

**Label Breakdown**:
- `job_id`: SLURM job ID
- `job_name`: Job name (default: "unknown")
- `user`: Submitting user (default: "unknown")
- `partition`: Partition (default: "unknown")
- `account`: Account name (default: "unknown")
- `qos`: Quality of Service (default: "normal")

**State Definitions**:
- **Active** (state=1.0): RUNNING, COMPLETING
- **Inactive** (state=0.0): PENDING, COMPLETED, FAILED, CANCELLED, TIMEOUT, SUSPENDED

**Filtering Support**:
- Filter by metric name (include/exclude specific metrics)
- Filter by metric type (Gauge, Histogram, Counter)
- Filter by category (timing, resource, info)
- Cardinality management: Samples high-cardinality labels

**Data Source**: Jobs manager API

**Example Output**:
```
slurm_job_state{job_id="12345",job_name="simulation",user="alice",partition="compute",account="proj1",qos="normal"} 1
slurm_job_queue_time_seconds{partition="compute",account="proj1"} 120
slurm_job_run_time_seconds{partition="compute",account="proj1"} 3600
slurm_job_cpus{job_id="12345",job_name="simulation",user="alice",partition="compute",account="proj1",qos="normal"} 64
slurm_job_memory_bytes{job_id="12345",job_name="simulation",user="alice",partition="compute",account="proj1",qos="normal"} 1099511627776
slurm_job_nodes{job_id="12345",job_name="simulation",user="alice",partition="compute",account="proj1",qos="normal"} 4
slurm_job_info{job_id="12345",job_name="simulation",user="alice",partition="compute",account="proj1",qos="normal"} 1
```

---

### 4. Accounts Collector

**File**: `internal/collector/accounts_simple.go`
**Purpose**: Account configuration and resource limits
**Metrics**: 6

#### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_account_info` | Gauge | Account metadata | account, organization, description, parent_account |
| `slurm_account_users_total` | Gauge | Number of users in account | account |
| `slurm_account_cpu_limit` | Gauge | CPU limit for account | account, partition |
| `slurm_account_memory_limit_bytes` | Gauge | Memory limit in bytes | account, partition |
| `slurm_account_job_limit` | Gauge | Job limit for account | account, partition |
| `slurm_account_node_limit` | Gauge | Node limit for account | account, partition |

**Label Breakdown**:
- `account`: Account name
- `organization`: Organization field (default: "unknown")
- `description`: Account description (default: "")
- `parent_account`: Parent account in hierarchy (default: "root")
- `partition`: Partition name (from default partition)

**Data Conversions**:
- Memory limits: From TRES["memory"] converted MB → Bytes (×1,048,576)
- Limit skipping: If limit value > 0, otherwise omitted

**Features**:
- Correlates accounts with user associations
- Reads TRES for memory limits
- Supports account hierarchy

**Data Source**: Accounts manager, Associations for user counts

**Example Output**:
```
slurm_account_info{account="proj1",organization="engineering",description="Project Alpha",parent_account="root"} 1
slurm_account_users_total{account="proj1"} 8
slurm_account_cpu_limit{account="proj1",partition="compute"} 512
slurm_account_memory_limit_bytes{account="proj1",partition="compute"} 1099511627776
slurm_account_job_limit{account="proj1",partition="compute"} 100
slurm_account_node_limit{account="proj1",partition="compute"} 64
```

---

### 5. Partitions Collector

**File**: `internal/collector/partitions_simple.go`
**Purpose**: Partition configuration and job queue status
**Metrics**: 11

#### Metrics

| Metric | Type | Description | Labels | Notes |
|--------|------|-------------|--------|-------|
| `slurm_partition_state` | Gauge | Partition state | partition | 1.0=UP, 0.0=DOWN |
| `slurm_partition_nodes_total` | Gauge | Total nodes in partition | partition | |
| `slurm_partition_nodes_allocated` | Gauge | Allocated nodes | partition | Calculated: total - idle - down |
| `slurm_partition_nodes_idle` | Gauge | Idle nodes | partition | Currently placeholder (TODO) |
| `slurm_partition_nodes_down` | Gauge | Down nodes | partition | Currently placeholder (TODO) |
| `slurm_partition_cpus_total` | Gauge | Total CPUs in partition | partition | |
| `slurm_partition_cpus_allocated` | Gauge | Allocated CPUs | partition | Currently placeholder (TODO) |
| `slurm_partition_cpus_idle` | Gauge | Idle CPUs | partition | Calculated: total - allocated |
| `slurm_partition_jobs_pending` | Gauge | Pending jobs in partition | partition | Currently placeholder (TODO) |
| `slurm_partition_jobs_running` | Gauge | Running jobs in partition | partition | Currently placeholder (TODO) |
| `slurm_partition_info` | Gauge | Partition metadata | partition, max_time, default_time | Includes time limits |

**Label Breakdown**:
- `partition`: Partition name
- `max_time`: Maximum job time (formatted "d-HH:MM" or "HH:MM")
- `default_time`: Default job time (same format)

**State Definitions**:
- `1.0`: Partition is UP and accepting jobs
- `0.0`: Partition is DOWN or otherwise unavailable

**Implementation Notes** (Marked TODO in code):
- Node state calculations (idle, down) return placeholder values
- Job count calculations pending integration with jobs manager
- These will be properly calculated in future releases

**Data Source**: Partitions manager API

**Example Output**:
```
slurm_partition_state{partition="compute"} 1
slurm_partition_nodes_total{partition="compute"} 64
slurm_partition_nodes_allocated{partition="compute"} 48
slurm_partition_nodes_idle{partition="compute"} 16
slurm_partition_nodes_down{partition="compute"} 0
slurm_partition_cpus_total{partition="compute"} 2048
slurm_partition_cpus_allocated{partition="compute"} 1536
slurm_partition_cpus_idle{partition="compute"} 512
slurm_partition_jobs_pending{partition="compute"} 12
slurm_partition_jobs_running{partition="compute"} 48
slurm_partition_info{partition="compute",max_time="7-00:00:00",default_time="1-00:00:00"} 1
```

---

### 6. Users Collector

**File**: `internal/collector/users_simple.go`
**Purpose**: Per-user resource utilization
**Metrics**: 6

#### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_user_info` | Gauge | User metadata | user, account, partition, default_account, admin_level |
| `slurm_user_jobs_running` | Gauge | Running jobs by user | user, account, partition |
| `slurm_user_jobs_pending` | Gauge | Pending jobs by user | user, account, partition |
| `slurm_user_cpus_used` | Gauge | CPUs in use by user | user, account, partition |
| `slurm_user_memory_used_bytes` | Gauge | Memory in use by user | user, account, partition |
| `slurm_user_associations_total` | Gauge | Count of associations | user, account, partition |

**Label Breakdown**:
- `user`: Username
- `account`: Account name (default: "default" if not available)
- `partition`: Partition name
- `default_account`: User's default account
- `admin_level`: Admin level ("None", "Operator", "Administrator")

**State Categorization**:
- **Running**: RUNNING or COMPLETING job states
- **Pending**: PENDING job states

**Data Conversions**:
- Memory: MB → Bytes (×1,048,576)

**Features**:
- Correlates jobs with user associations
- Composite key: (account, partition)
- Counts associations per user

**Data Source**: Users manager, correlated with Jobs API

**Example Output**:
```
slurm_user_info{user="alice",account="proj1",partition="compute",default_account="proj1",admin_level="None"} 1
slurm_user_jobs_running{user="alice",account="proj1",partition="compute"} 3
slurm_user_jobs_pending{user="alice",account="proj1",partition="compute"} 1
slurm_user_cpus_used{user="alice",account="proj1",partition="compute"} 64
slurm_user_memory_used_bytes{user="alice",account="proj1",partition="compute"} 274877906944
slurm_user_associations_total{user="alice",account="proj1",partition="compute"} 1
```

---

### 7. QoS Collector

**File**: `internal/collector/qos.go`
**Purpose**: Quality of Service policy configuration
**Metrics**: 13

#### Metrics

| Metric | Type | Description | Labels | Value Range |
|--------|------|-------------|--------|--------------|
| `slurm_qos_priority` | Gauge | QoS priority level | qos | Any |
| `slurm_qos_usage_factor` | Gauge | Priority factor multiplier | qos | Typically 0.0-2.0 |
| `slurm_qos_max_jobs` | Gauge | Global job limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_jobs_per_user` | Gauge | Per-user job limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_jobs_per_account` | Gauge | Per-account job limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_submit_jobs` | Gauge | Submit queue limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_cpus` | Gauge | Global CPU limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_cpus_per_user` | Gauge | Per-user CPU limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_nodes` | Gauge | Node limit | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_max_wall_time_seconds` | Gauge | Max wall time in seconds | qos | ≥0 or -1 (unlimited) |
| `slurm_qos_min_cpus` | Gauge | Minimum CPUs required | qos | ≥0 |
| `slurm_qos_min_nodes` | Gauge | Minimum nodes required | qos | ≥0 |
| `slurm_qos_info` | Gauge | QoS metadata | qos, description, preempt_mode, flags | Always 1.0 |

**Label Breakdown**:
- `qos`: QoS name (e.g., "normal", "high", "debug")
- `description`: QoS description
- `preempt_mode`: Preemption mode
- `flags`: Comma-separated flags

**Value Conversions**:
- **Infinite values**: Values ≥1,000,000 are converted to -1 (sentinel for unlimited)
- **Wall time**: Converted from minutes to seconds (×60)
- **Omission**: 0 limits often omitted (depends on policy)

**Use Cases**:
- Monitor QoS policies in place
- Verify job resource limits
- Track preemption policies
- Enforce fairness constraints

**Data Source**: QoS manager API

**Example Output**:
```
slurm_qos_priority{qos="normal"} 1000
slurm_qos_usage_factor{qos="normal"} 1.0
slurm_qos_max_jobs{qos="normal"} 100
slurm_qos_max_jobs_per_user{qos="normal"} 10
slurm_qos_max_cpus{qos="normal"} 1024
slurm_qos_max_wall_time_seconds{qos="normal"} 604800
slurm_qos_info{qos="normal",description="Normal QoS",preempt_mode="cluster",flags=""} 1
```

---

### 8. TRES Collector

**File**: `internal/collector/tres_simple.go`
**Purpose**: Trackable Resources (CPU, Memory, GPU, etc.)
**Metrics**: 5
**Status**: Partial (node-level metrics TODO)

#### Metrics

| Metric | Type | Description | Labels | Status |
|--------|------|-------------|--------|--------|
| `slurm_tres_info` | Gauge | TRES metadata | tres_id, tres_type, tres_name, cluster | Production |
| `slurm_tres_count` | Gauge | Number of TRES instances | tres_id, tres_type, tres_name, cluster | Production |
| `slurm_tres_allocated` | Gauge | Currently allocated | tres_id, tres_type, node, cluster | TODO |
| `slurm_tres_available` | Gauge | Available amount | tres_id, tres_type, node, cluster | TODO |
| `slurm_tres_configured` | Gauge | Configured amount | tres_id, tres_type, node, cluster | TODO |

**Label Breakdown**:
- `tres_id`: TRES identifier (e.g., "1", "2", "1001")
- `tres_type`: Resource type (e.g., "cpu", "mem", "node", "gres/gpu")
- `tres_name`: Descriptive name
- `cluster`: Cluster name
- `node`: Node name (for per-node metrics)

**TRES Types** (Examples):
- `1`: CPU
- `2`: Memory
- `3`: Energy
- `4`: Node
- `1001+`: GRES (GPUs, specific hardware)

**Implementation Status**:
- ✅ Type/name enumeration complete
- ⏳ Node-level allocation/availability metrics commented out (pending implementation)

**Data Source**: TRES manager API

**Example Output**:
```
slurm_tres_info{tres_id="1",tres_type="cpu",tres_name="CPU",cluster="prod"} 1
slurm_tres_info{tres_id="2",tres_type="mem",tres_name="Memory",cluster="prod"} 1
slurm_tres_info{tres_id="1001",tres_type="gres/gpu",tres_name="GPU",cluster="prod"} 1
slurm_tres_count{tres_id="1",tres_type="cpu",tres_name="CPU",cluster="prod"} 2048
slurm_tres_count{tres_id="2",tres_type="mem",tres_name="Memory",cluster="prod"} 8192
slurm_tres_count{tres_id="1001",tres_type="gres/gpu",tres_name="GPU",cluster="prod"} 128
```

---

### 9. Licenses Collector

**File**: `internal/collector/licenses_simple.go`
**Purpose**: Software license availability and usage
**Metrics**: 5
**Availability**: v0.0.42+ (requires `GetLicenses()` API)

#### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_licenses_total` | Gauge | Total licenses | feature, cluster |
| `slurm_licenses_used` | Gauge | Currently in use | feature, cluster |
| `slurm_licenses_available` | Gauge | Available for use | feature, cluster |
| `slurm_licenses_free` | Gauge | Free (same as available) | feature, cluster |
| `slurm_licenses_reserved` | Gauge | Reserved for future use | feature, cluster |

**Label Breakdown**:
- `feature`: License feature name (e.g., "matlab", "comsol", "ansys")
- `cluster`: Cluster name

**Relationships**:
```
used + available = total
free ≈ available (for compatibility)
available = total - used - reserved
```

**Features**:
- Tracks software license consumption
- Identifies license bottlenecks
- Enables license-aware scheduling alerts

**Data Source**: License manager API (requires `GetLicenses()`)

**Example Output**:
```
slurm_licenses_total{feature="matlab",cluster="prod"} 50
slurm_licenses_used{feature="matlab",cluster="prod"} 34
slurm_licenses_available{feature="matlab",cluster="prod"} 16
slurm_licenses_free{feature="matlab",cluster="prod"} 16
slurm_licenses_reserved{feature="matlab",cluster="prod"} 0
```

---

### 10. Shares/Fairshare Collector

**File**: `internal/collector/shares_simple.go`
**Purpose**: Fair-share scheduling information
**Metrics**: 7
**Availability**: v0.0.42+ (requires `GetShares()` API)

#### Metrics

| Metric | Type | Description | Labels | Range |
|--------|------|-------------|--------|-------|
| `slurm_shares_raw_shares` | Gauge | Raw share value | user, account, partition, cluster | ≥0 |
| `slurm_shares_normalized_shares` | Gauge | Normalized shares | user, account, partition, cluster | 0.0-1.0 |
| `slurm_shares_raw_usage` | Gauge | Raw usage count | user, account, partition, cluster | ≥0 |
| `slurm_shares_effective_usage` | Gauge | Usage after decay | user, account, partition, cluster | 0.0-1.0 |
| `slurm_shares_usage_factor` | Gauge | Fair-share factor | user, account, partition, cluster | 0.0-1.0 |
| `slurm_shares_level` | Gauge | Position in tree | user, account, partition, cluster | ≥1 |
| `slurm_shares_tree_depth` | Gauge | Max tree depth | partition, cluster | ≥0 |

**Label Breakdown**:
- `user`: Username (for user shares)
- `account`: Account name
- `partition`: Partition name
- `cluster`: Cluster name
- `tree_depth` metric: partition, cluster only (global per partition)

**Fair-Share Interpretation**:
- **usage_factor** (0.0-1.0): User priority metric
  - `1.0`: Perfect fairness
  - `<1.0`: User has consumed more than fair share (lower priority)
  - `>1.0`: User has consumed less (higher priority)
- **normalized_shares** (0.0-1.0): Proportional share allocation
- **effective_usage**: Recent usage weighted by decay function
- **level**: Position in fairshare hierarchy (root=1, children>1)

**Hierarchy Support**:
- User-level shares
- Account-level aggregation
- Multi-partition tracking
- Tree depth tracking

**Data Source**: Fairshare manager API (requires `GetShares()`)

**Example Output**:
```
slurm_shares_raw_shares{user="alice",account="proj1",partition="compute",cluster="prod"} 100
slurm_shares_normalized_shares{user="alice",account="proj1",partition="compute",cluster="prod"} 0.5
slurm_shares_raw_usage{user="alice",account="proj1",partition="compute",cluster="prod"} 120
slurm_shares_effective_usage{user="alice",account="proj1",partition="compute",cluster="prod"} 0.45
slurm_shares_usage_factor{user="alice",account="proj1",partition="compute",cluster="prod"} 0.85
slurm_shares_level{user="alice",account="proj1",partition="compute",cluster="prod"} 2
slurm_shares_tree_depth{partition="compute",cluster="prod"} 3
```

---

### 11. System Collector

**File**: `internal/collector/system_simple.go`
**Purpose**: Exporter and SLURM daemon health metrics
**Metrics**: 14

#### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_system_slurm_daemon_up` | Gauge | Daemon responsiveness | cluster, daemon_type (slurmctld, slurmdbd) |
| `slurm_system_slurm_api_latency_seconds` | Gauge | API call latency | cluster |
| `slurm_system_slurm_api_errors_total` | Counter | Total API errors | cluster, endpoint, error_type |
| `slurm_system_slurm_db_connections` | Gauge | DB connection count | cluster |
| `slurm_system_slurm_db_latency_seconds` | Gauge | Database query latency | cluster |
| `slurm_system_load_average` | Gauge | System load average | cluster, period (1m, 5m, 15m) |
| `slurm_system_memory_usage_bytes` | Gauge | Go runtime memory | cluster, type (allocated, system, heap) |
| `slurm_system_disk_usage_bytes` | Gauge | Disk usage | cluster, mountpoint, type (used, total) |
| `slurm_system_config_last_modified_timestamp` | Gauge | Config modification time | cluster |
| `slurm_system_active_controllers` | Gauge | Active controller count | cluster |
| `slurm_system_api_calls_total` | Counter | Total API calls | cluster, endpoint, status |
| `slurm_system_collection_duration_seconds` | Histogram | Collection latency | collector |

**Label Breakdown**:
- `cluster`: Cluster name
- `daemon_type`: "slurmctld" (main), "slurmdbd" (accounting)
- `endpoint`: API endpoint (e.g., "/slurm/v0.0.43/nodes")
- `error_type`: Error category (e.g., "timeout", "unauthorized")
- `period`: Load average period (1m, 5m, 15m)
- `type`: Subsystem (memory: allocated/system/heap, disk: used/total)
- `mountpoint`: Disk mount point (simulated)
- `status`: HTTP status ("200", "401", "500", etc.)
- `collector`: Collector name

**Features**:
- Tests daemon connectivity (ping with API call)
- Tracks API call latency and error rates
- Monitors exporter's own resource usage
- Provides self-diagnostics
- Custom labels support

**Daemon Health Indicators**:
- `slurm_system_slurm_daemon_up`:
  - `1.0`: Daemon responding to API calls
  - `0.0`: Timeout or connection refused

**Data Sources**:
- Daemon connectivity: Successful API calls
- Load average: System `getloadavg()`
- Memory: Go runtime stats (`runtime.MemStats`)
- Disk: Simulated values (implementation can be customized)
- DB latency: Simulated (implementation pending)

**Example Output**:
```
slurm_system_slurm_daemon_up{cluster="prod",daemon_type="slurmctld"} 1
slurm_system_slurm_api_latency_seconds{cluster="prod"} 0.045
slurm_system_load_average{cluster="prod",period="1m"} 2.34
slurm_system_load_average{cluster="prod",period="5m"} 2.15
slurm_system_load_average{cluster="prod",period="15m"} 2.01
slurm_system_memory_usage_bytes{cluster="prod",type="allocated"} 134217728
slurm_system_memory_usage_bytes{cluster="prod",type="system"} 201326592
slurm_system_memory_usage_bytes{cluster="prod",type="heap"} 67108864
slurm_system_api_calls_total{cluster="prod",endpoint="/slurm/v0.0.43/nodes",status="200"} 12847
slurm_system_collection_duration_seconds_bucket{collector="nodes",le="0.1"} 145
slurm_system_collection_duration_seconds_bucket{collector="nodes",le="0.5"} 156
slurm_system_collection_duration_seconds_bucket{collector="nodes",le="1.0"} 156
slurm_system_collection_duration_seconds_count{collector="nodes"} 156
slurm_system_collection_duration_seconds_sum{collector="nodes"} 18.932
```

---

### 12. Scheduler Collector

**File**: `internal/collector/scheduler.go`
**Purpose**: Exporter collection scheduling and performance
**Metrics**: 3

#### Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `slurm_scheduler_scheduled_collections_total` | Counter | Scheduled collection runs | collector, status (success, error) |
| `slurm_scheduler_missed_collections_total` | Counter | Missed collection runs | collector, reason (latency, behind) |
| `slurm_scheduler_schedule_latency_seconds` | Histogram | Scheduling latency | collector |

**Label Breakdown**:
- `collector`: Collector name (cluster, nodes, jobs, etc.)
- `status`: "success" or "error"
- `reason`: "latency" (too slow), "behind" (skipped due to backlog)

**Features**:
- Tracks collection schedule performance
- Identifies slow collectors
- Detects collection backlog conditions
- Supports per-collector scheduling

**Data Structure** (Internal):
- NextRun, LastRun: Timestamps
- RunCount, ErrorCount: Counters
- Interval, Timeout: Configuration

**Use Cases**:
- Detect collectors running too slowly
- Monitor scheduler health
- Set alerts on missed collections
- Capacity planning for collection interval

**Example Output**:
```
slurm_scheduler_scheduled_collections_total{collector="cluster",status="success"} 1024
slurm_scheduler_scheduled_collections_total{collector="cluster",status="error"} 2
slurm_scheduler_missed_collections_total{collector="jobs",reason="latency"} 5
slurm_scheduler_schedule_latency_seconds_bucket{collector="nodes",le="0.05"} 900
slurm_scheduler_schedule_latency_seconds_bucket{collector="nodes",le="0.1"} 950
slurm_scheduler_schedule_latency_seconds_bucket{collector="nodes",le="0.5"} 1024
slurm_scheduler_schedule_latency_seconds_count{collector="nodes"} 1024
slurm_scheduler_schedule_latency_seconds_sum{collector="nodes"} 48.302
```

---

## Metric Type Summary

### By Prometheus Type

| Type | Count | Examples |
|------|-------|----------|
| **Gauge** | ~85 | Resource amounts, states, allocations, limits, info metrics |
| **Counter** | ~12 | API calls, errors, collection runs, cumulative events |
| **Histogram** | ~10 | Latency, duration, queue time, job run time |

### By Category

| Category | Count | Purpose |
|----------|-------|---------|
| **Resource Metrics** | ~45 | CPU, memory, node allocation |
| **State Metrics** | ~20 | Node state, job state, partition state |
| **Policy/Limit Metrics** | ~25 | QoS limits, account limits, fairshare |
| **Timing/Performance** | ~12 | Latency, collection duration, job timing |
| **Metadata/Info** | ~8 | Cluster info, node info, account info |

### Histogram Metrics (Multi-Series)

Histogram metrics produce multiple time series per metric:
- `_bucket` (one per configured bucket boundary)
- `_count` (total observations)
- `_sum` (sum of observed values)

Examples with buckets:
- `slurm_job_queue_time_seconds` → 3 time series per job
- `slurm_job_run_time_seconds` → 3 time series per job
- `slurm_system_collection_duration_seconds` → 3 time series per collector
- `slurm_scheduler_schedule_latency_seconds` → 3 time series per collector

---

## API Version Dependencies

### Required API Capabilities by Collector

| Collector | Min Version | API Methods | Feature Dependencies |
|-----------|-------------|-------------|----------------------|
| Cluster | v0.0.40 | `Info()`, `Stats()` | Basic cluster info |
| Nodes | v0.0.40 | `Nodes()` | Node listing |
| Jobs | v0.0.40 | `Jobs()` | Job listing and state |
| Accounts | v0.0.40 | `Accounts()`, `Associations()` | Account hierarchy |
| Partitions | v0.0.40 | `Partitions()` | Partition config |
| Users | v0.0.40 | `Users()`, `Jobs()` | User-job correlation |
| QoS | v0.0.40 | `QoS()` | QoS definitions |
| TRES | v0.0.42 | `TRES()` | Trackable resources (v0.0.42+) |
| Licenses | v0.0.42 | `GetLicenses()` | License API (v0.0.42+) |
| Shares | v0.0.42 | `GetShares()` | Fair-share API (v0.0.42+) |
| System | v0.0.40 | `Info()`, system APIs | Daemon connectivity |
| Scheduler | v0.0.40 | Internal scheduling | Timing/latency tracking |

### Features by Version

| Feature | v0.0.40 | v0.0.41 | v0.0.42 | v0.0.43 | v0.0.44 |
|---------|---------|---------|---------|---------|---------|
| Basic metrics (cluster, nodes, jobs) | ✅ | ✅ | ✅ | ✅ | ✅ |
| GetLicenses | ❌ | ❌ | ✅ | ✅ | ✅ |
| GetShares | ❌ | ❌ | ✅ | ✅ | ✅ |
| GetTRES | ❌ | ❌ | ✅ | ✅ | ✅ |
| GetDiagnostics | ❌ | ❌ | ✅ | ✅ | ✅ |
| GetDBDiagnostics | ❌ | ❌ | ✅ | ✅ | ✅ |

---

## Data Conversions

### Unit Conversions

| Conversion | Formula | Example |
|------------|---------|---------|
| Memory: MB → Bytes | value × 1,048,576 | 256 MB → 268,435,456 bytes |
| Wall time: Minutes → Seconds | value × 60 | 1440 min → 86,400 sec |
| Timestamps | Unix epoch (seconds) | 2026-01-26 → 1769347200 |
| CPU allocation ratios | Percentage of total | 50%, 75%, 0% based on state |

### Value Representations

| Type | Representation | Example |
|------|-----------------|---------|
| Infinite/Unlimited | -1 (sentinel) | max_jobs = -1 (unlimited) |
| Boolean states | 1.0 (true), 0.0 (false) | slurm_node_state = 1.0 |
| Percentages | 0.0-1.0 (decimal) | utilization_ratio = 0.75 |
| Metadata labels | String labels | labels: name, description |

### Default Values

| Field | Default | Purpose |
|-------|---------|---------|
| cluster_name | "default_cluster" | Fallback cluster name |
| partition | "unknown" | Unassigned partition |
| user | "unknown" | Unidentified user |
| account | "default" or "unknown" | Fallback account |
| qos | "normal" | Default QoS |
| job_name | "unknown" | Unnamed job |
| arch | "x86_64" | Default architecture |
| os | "linux" | Default OS |

---

## Design Patterns

### 1. Cardinality Management

**Problem**: High-cardinality labels (job_id, user names) create unbounded metric growth.

**Solution**: JobsSimpleCollector implements cardinality awareness
- Samples high-cardinality labels based on configuration
- Uses CardinalityManager to enforce limits
- Filters based on MetricFilter settings

**Example**:
```go
if !collector.cardinality.CheckLimit(labels) {
    // Skip metric emission if cardinality exceeded
}
```

### 2. Metric Filtering

**Pattern**: Collectors support metric-level filtering

**Types**:
- By metric name (include/exclude list)
- By metric type (Gauge, Histogram, Counter)
- By category (timing, resource, info)

**Example**:
```
MetricFilter:
  Include:
    - slurm_job_state
    - slurm_job_cpus
  Exclude:
    - slurm_job_queue_time_seconds
```

### 3. Estimated/Calculated Metrics

**Pattern**: Some metrics are calculated from available data

**Examples**:
- Allocated CPUs: Estimated from node state (50% for ALLOCATED, 75% for COMPLETING)
- Allocated memory: Proportional to CPU allocation
- Partition idle CPUs: Calculated as total - allocated

**Rationale**: SLURM API doesn't directly expose all metrics; calculated values provide useful approximations.

### 4. State Normalization

**Pattern**: Different states mapped to consistent numeric values

**Examples**:
- Node state: 1.0 (UP) vs 0.0 (DOWN/DRAIN)
- Job state: 1.0 (RUNNING/COMPLETING) vs 0.0 (other)
- Partition state: 1.0 (UP) vs 0.0 (DOWN)

**Benefit**: Enables simple alerting rules without complex label matching.

### 5. Info Metrics

**Pattern**: Metadata exposed via always-1.0 Gauge with label values

**Example**:
```
slurm_cluster_info{cluster="prod",version="23.02.5",control_host="ctrl1"} 1
```

**Benefit**: Prometheus scrape metadata without creating extra cardinality.

### 6. Custom Labels Support

**Pattern**: Collectors can add constant labels via configuration

**Enabled For**: JobsSimpleCollector, NodesSimpleCollector, SystemSimpleCollector

**Use Case**: Add environment/region labels in multi-region deployments

### 7. Fallback/Safe Defaults

**Pattern**: Missing fields replaced with reasonable defaults

**Examples**:
- Missing node partition → "unknown"
- Missing job account → "unknown"
- Missing job name → "unknown"

**Benefit**: Prevents metric gaps or errors due to incomplete data.

### 8. Feature Gating

**Pattern**: Optional collectors disabled if API doesn't support them

**Examples**:
- Licenses collector disabled if `GetLicenses()` not available
- Shares collector disabled if `GetShares()` not available
- TRES collector disabled if `GetTRES()` not available

**Implementation**: Version checking in adapter registry

---

## Summary Table

| Metric | Type | Collector | Labels | Status | Version |
|--------|------|-----------|--------|--------|---------|
| slurm_cluster_info | Gauge | Cluster | 5 | ✅ | v0.0.40+ |
| slurm_cluster_nodes_total | Gauge | Cluster | 1 | ✅ | v0.0.40+ |
| slurm_node_state | Gauge | Nodes | 6 | ✅ | v0.0.40+ |
| slurm_job_state | Gauge | Jobs | 6 | ✅ | v0.0.40+ |
| slurm_partition_info | Gauge | Partitions | 3 | ✅ | v0.0.40+ |
| slurm_account_info | Gauge | Accounts | 4 | ✅ | v0.0.40+ |
| slurm_user_info | Gauge | Users | 5 | ✅ | v0.0.40+ |
| slurm_qos_info | Gauge | QoS | 4 | ✅ | v0.0.40+ |
| slurm_licenses_total | Gauge | Licenses | 2 | ✅ | v0.0.42+ |
| slurm_shares_usage_factor | Gauge | Shares | 4 | ✅ | v0.0.42+ |
| slurm_tres_info | Gauge | TRES | 4 | ✅ | v0.0.42+ |
| slurm_system_daemon_up | Gauge | System | 2 | ✅ | v0.0.40+ |
| slurm_scheduler_scheduled_collections_total | Counter | Scheduler | 2 | ✅ | v0.0.40+ |

---

**Complete Specification Version 1.0**
**Total Metrics: ~107 across 12 collectors**
**Metric Types: ~85 Gauge, ~12 Counter, ~10 Histogram**
