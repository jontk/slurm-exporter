package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
)

// BenchmarkCollectorPerformance benchmarks collector performance with different data sizes
func BenchmarkCollectorPerformance(b *testing.B) {
	scenarios := []struct {
		name       string
		jobs       int
		nodes      int
		partitions int
	}{
		{"small_cluster", 100, 10, 2},
		{"medium_cluster", 1000, 100, 5},
		{"large_cluster", 10000, 1000, 10},
		{"xlarge_cluster", 100000, 5000, 20},
	}
	
	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			benchmarkCollectorWithData(b, scenario.jobs, scenario.nodes, scenario.partitions)
		})
	}
}

func benchmarkCollectorWithData(b *testing.B, jobCount, nodeCount, partitionCount int) {
	// Setup
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockClient := testutil.NewMockSLURMClient(ctrl)
	registry := testutil.NewTestMetricsRegistry()
	logger := testutil.NewTestLogger().GetEntry()
	
	// Generate test data
	jobs := testutil.Generator.GenerateJobs(jobCount)
	nodes := testutil.Generator.GenerateNodes(nodeCount)
	partitions := testutil.Generator.GeneratePartitions(partitionCount)
	
	// Setup mock responses
	testutil.SetupMockJobList(mockClient, jobs)
	testutil.SetupMockNodeList(mockClient, nodes)
	testutil.SetupMockPartitionList(mockClient, partitions)
	
	collector := collector.NewJobsSimpleCollector(
		mockClient,
		registry,
		logger,
		config.CollectorConfig{Enabled: true},
	)
	
	ctx := context.Background()
	
	// Benchmark
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 1000)
		err := collector.Collect(ctx, ch)
		close(ch)
		
		if err != nil {
			b.Fatal(err)
		}
		
		// Drain channel
		for range ch {
		}
	}
	
	// Report custom metrics
	opsPerSec := float64(b.N) / b.Elapsed().Seconds()
	jobsPerSec := opsPerSec * float64(jobCount)
	
	b.ReportMetric(opsPerSec, "collections/sec")
	b.ReportMetric(jobsPerSec, "jobs/sec")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(jobCount), "ns/job")
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []int{100, 1000, 10000, 50000}
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("jobs_%d", size), func(b *testing.B) {
			benchmarkMemoryUsage(b, size)
		})
	}
}

func benchmarkMemoryUsage(b *testing.B, jobCount int) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockClient := testutil.NewMockSLURMClient(ctrl)
	registry := testutil.NewTestMetricsRegistry()
	logger := testutil.NewTestLogger().GetEntry()
	
	jobs := testutil.Generator.GenerateJobs(jobCount)
	testutil.SetupMockJobList(mockClient, jobs)
	
	collector := collector.NewJobsSimpleCollector(
		mockClient,
		registry,
		logger,
		config.CollectorConfig{Enabled: true},
	)
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	var m1, m2 runtime.MemStats
	for i := 0; i < b.N; i++ {
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		ch := make(chan prometheus.Metric, 1000)
		err := collector.Collect(ctx, ch)
		close(ch)
		
		if err != nil {
			b.Fatal(err)
		}
		
		// Drain channel
		for range ch {
		}
		
		runtime.ReadMemStats(&m2)
		
		// Report memory usage
		allocBytes := m2.TotalAlloc - m1.TotalAlloc
		b.ReportMetric(float64(allocBytes), "bytes/op")
		b.ReportMetric(float64(m2.Mallocs-m1.Mallocs), "allocs/op")
	}
}

// BenchmarkConcurrentCollection benchmarks concurrent collection performance
func BenchmarkConcurrentCollection(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8, 16}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			benchmarkConcurrentCollection(b, concurrency)
		})
	}
}

func benchmarkConcurrentCollection(b *testing.B, concurrency int) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockClient := testutil.NewMockSLURMClient(ctrl)
	registry := testutil.NewTestMetricsRegistry()
	logger := testutil.NewTestLogger().GetEntry()
	
	jobs := testutil.Generator.GenerateJobs(1000)
	testutil.SetupMockJobList(mockClient, jobs)
	
	collector := collector.NewJobsSimpleCollector(
		mockClient,
		registry,
		logger,
		config.CollectorConfig{Enabled: true},
	)
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		
		for pb.Next() {
			ch := make(chan prometheus.Metric, 1000)
			err := collector.Collect(ctx, ch)
			close(ch)
			
			if err != nil {
				b.Fatal(err)
			}
			
			// Drain channel
			for range ch {
			}
		}
	})
}

// BenchmarkMetricCardinality benchmarks metric cardinality impact
func BenchmarkMetricCardinality(b *testing.B) {
	cardinalityLevels := []struct {
		name        string
		partitions  int
		users       int
		accounts    int
	}{
		{"low_cardinality", 5, 10, 5},
		{"medium_cardinality", 20, 100, 20},
		{"high_cardinality", 50, 500, 50},
		{"extreme_cardinality", 100, 1000, 100},
	}
	
	for _, level := range cardinalityLevels {
		b.Run(level.name, func(b *testing.B) {
			benchmarkMetricCardinality(b, level.partitions, level.users, level.accounts)
		})
	}
}

func benchmarkMetricCardinality(b *testing.B, partitions, users, accounts int) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockClient := testutil.NewMockSLURMClient(ctrl)
	registry := testutil.NewTestMetricsRegistry()
	logger := testutil.NewTestLogger().GetEntry()
	
	// Generate high cardinality data
	jobs := generateHighCardinalityJobs(1000, partitions, users, accounts)
	testutil.SetupMockJobList(mockClient, jobs)
	
	collector := collector.NewJobsSimpleCollector(
		mockClient,
		registry,
		logger,
		config.CollectorConfig{Enabled: true},
	)
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 10000)
		err := collector.Collect(ctx, ch)
		close(ch)
		
		if err != nil {
			b.Fatal(err)
		}
		
		// Count metrics
		metricCount := 0
		for range ch {
			metricCount++
		}
		
		b.ReportMetric(float64(metricCount), "metrics/op")
	}
}

// BenchmarkCachePerformance benchmarks caching performance
func BenchmarkCachePerformance(b *testing.B) {
	scenarios := []struct {
		name    string
		cacheHitRate float64
	}{
		{"no_cache", 0.0},
		{"low_hit_rate", 0.3},
		{"medium_hit_rate", 0.6},
		{"high_hit_rate", 0.9},
	}
	
	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			benchmarkCachePerformance(b, scenario.cacheHitRate)
		})
	}
}

func benchmarkCachePerformance(b *testing.B, hitRate float64) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()
	
	mockClient := testutil.NewMockSLURMClient(ctrl)
	registry := testutil.NewTestMetricsRegistry()
	logger := testutil.NewTestLogger().GetEntry()
	
	jobs := testutil.Generator.GenerateJobs(1000)
	
	// Setup mock to simulate cache behavior
	callCount := 0
	mockClient.EXPECT().
		ListJobs(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, opts interface{}) ([]slurm.Job, error) {
			callCount++
			// Simulate cache hit by occasionally returning quickly
			if float64(callCount%100)/100.0 < hitRate {
				// Cache hit - return quickly
				time.Sleep(1 * time.Millisecond)
			} else {
				// Cache miss - simulate API call
				time.Sleep(10 * time.Millisecond)
			}
			return jobs, nil
		}).
		AnyTimes()
	
	collector := collector.NewJobsSimpleCollector(
		mockClient,
		registry,
		logger,
		config.CollectorConfig{Enabled: true},
	)
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 1000)
		err := collector.Collect(ctx, ch)
		close(ch)
		
		if err != nil {
			b.Fatal(err)
		}
		
		// Drain channel
		for range ch {
		}
	}
}

// BenchmarkSerializationPerformance benchmarks metric serialization
func BenchmarkSerializationPerformance(b *testing.B) {
	metricCounts := []int{100, 1000, 10000, 50000}
	
	for _, count := range metricCounts {
		b.Run(fmt.Sprintf("metrics_%d", count), func(b *testing.B) {
			benchmarkSerializationPerformance(b, count)
		})
	}
}

func benchmarkSerializationPerformance(b *testing.B, metricCount int) {
	// Create test metrics
	registry := prometheus.NewRegistry()
	
	// Create multiple gauge metrics
	gauges := make([]prometheus.Gauge, metricCount)
	for i := 0; i < metricCount; i++ {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("test_metric_%d", i),
			Help: "Test metric for benchmarking",
		})
		gauges[i] = gauge
		registry.MustRegister(gauge)
		gauge.Set(float64(i))
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Gather metrics (this includes serialization)
		families, err := registry.Gather()
		if err != nil {
			b.Fatal(err)
		}
		
		// Count total metrics
		totalMetrics := 0
		for _, family := range families {
			totalMetrics += len(family.GetMetric())
		}
		
		if totalMetrics != metricCount {
			b.Fatalf("Expected %d metrics, got %d", metricCount, totalMetrics)
		}
	}
	
	b.ReportMetric(float64(metricCount), "metrics/op")
}

// Helper functions

func generateHighCardinalityJobs(count, partitions, users, accounts int) []slurm.Job {
	jobs := make([]slurm.Job, count)
	
	partitionNames := make([]string, partitions)
	for i := 0; i < partitions; i++ {
		partitionNames[i] = fmt.Sprintf("partition_%d", i)
	}
	
	userNames := make([]string, users)
	for i := 0; i < users; i++ {
		userNames[i] = fmt.Sprintf("user_%d", i)
	}
	
	accountNames := make([]string, accounts)
	for i := 0; i < accounts; i++ {
		accountNames[i] = fmt.Sprintf("account_%d", i)
	}
	
	states := []string{"RUNNING", "PENDING", "COMPLETED", "FAILED"}
	
	for i := 0; i < count; i++ {
		jobs[i] = slurm.Job{
			ID:        fmt.Sprintf("%d", 10000+i),
			State:     states[i%len(states)],
			Partition: partitionNames[i%partitions],
			UserID:    userNames[i%users],
			GroupID:   fmt.Sprintf("group_%d", i%10), // Lower cardinality for groups
			Metadata: map[string]interface{}{
				"account": accountNames[i%accounts],
				"qos":     fmt.Sprintf("qos_%d", i%5), // Even lower cardinality for QoS
			},
		}
	}
	
	return jobs
}

