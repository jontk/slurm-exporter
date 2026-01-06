package benchmark

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/testutil/fixtures"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
	"github.com/jontk/slurm-exporter/internal/testutil"
)

// BenchmarkJobsCollectorSmall tests performance with small job sets
func BenchmarkJobsCollectorSmall(b *testing.B) {
	benchmarkJobsCollector(b, 100) // 100 jobs
}

// BenchmarkJobsCollectorMedium tests performance with medium job sets
func BenchmarkJobsCollectorMedium(b *testing.B) {
	benchmarkJobsCollector(b, 1000) // 1,000 jobs
}

// BenchmarkJobsCollectorLarge tests performance with large job sets
func BenchmarkJobsCollectorLarge(b *testing.B) {
	benchmarkJobsCollector(b, 10000) // 10,000 jobs
}

// BenchmarkJobsCollectorXLarge tests performance with extra large job sets
func BenchmarkJobsCollectorXLarge(b *testing.B) {
	benchmarkJobsCollector(b, 50000) // 50,000 jobs
}

func benchmarkJobsCollector(b *testing.B, jobCount int) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientWithJobs(jobCount)

	collector := collector.NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, jobCount*10) // Buffer for all potential metrics
		err := collector.Collect(ctx, ch)
		if err != nil {
			b.Fatalf("Collection failed: %v", err)
		}
		close(ch)

		// Drain the channel to simulate real usage
		for range ch {
		}
	}
}

// BenchmarkNodesCollectorScale tests node collector performance
func BenchmarkNodesCollectorScale(b *testing.B) {
	sizes := []int{10, 100, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Nodes_%d", size), func(b *testing.B) {
			benchmarkNodesCollector(b, size)
		})
	}
}

func benchmarkNodesCollector(b *testing.B, nodeCount int) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientWithNodes(nodeCount)

	collector := collector.NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, nodeCount*20) // Buffer for all potential metrics
		err := collector.Collect(ctx, ch)
		if err != nil {
			b.Fatalf("Collection failed: %v", err)
		}
		close(ch)

		// Drain the channel
		for range ch {
		}
	}
}

// BenchmarkRegistryCollectionConcurrent tests concurrent collection performance
func BenchmarkRegistryCollectionConcurrent(b *testing.B) {
	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, level := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent_%d", level), func(b *testing.B) {
			benchmarkRegistryConcurrent(b, level)
		})
	}
}

func benchmarkRegistryConcurrent(b *testing.B, concurrency int) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientFull()

	registry := collector.NewRegistry(mockClient, logger)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch := make(chan prometheus.Metric, 10000)
			err := registry.Collect(ctx, ch)
			if err != nil {
				b.Fatalf("Collection failed: %v", err)
			}
			close(ch)

			// Drain the channel
			for range ch {
			}
		}
	})
}

// BenchmarkMemoryUsage tests memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientFull()

	registry := collector.NewRegistry(mockClient, logger)
	ctx := context.Background()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 10000)
		err := registry.Collect(ctx, ch)
		if err != nil {
			b.Fatalf("Collection failed: %v", err)
		}
		close(ch)

		// Drain the channel
		for range ch {
		}
	}

	b.StopTimer()
	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.Alloc-m1.Alloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
}

// BenchmarkMetricFiltering tests filtering performance impact
func BenchmarkMetricFiltering(b *testing.B) {
	filterScenarios := []struct {
		name    string
		enabled bool
		filters int
	}{
		{"NoFiltering", false, 0},
		{"LightFiltering", true, 5},
		{"HeavyFiltering", true, 20},
		{"ExtremeFiltering", true, 100},
	}

	for _, scenario := range filterScenarios {
		b.Run(scenario.name, func(b *testing.B) {
			benchmarkFiltering(b, scenario.enabled, scenario.filters)
		})
	}
}

func benchmarkFiltering(b *testing.B, enableFiltering bool, filterCount int) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientWithJobs(1000)

	collector := collector.NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	if enableFiltering {
		// Create filter configuration with multiple patterns
		filters := make([]string, filterCount)
		for i := 0; i < filterCount; i++ {
			filters[i] = fmt.Sprintf("slurm_job_*_%d", i)
		}

		// Note: This would need actual filter configuration implementation
		// For now, just measure the baseline
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 10000)
		err := collector.Collect(ctx, ch)
		if err != nil {
			b.Fatalf("Collection failed: %v", err)
		}
		close(ch)

		// Drain the channel
		for range ch {
		}
	}
}

// BenchmarkCardinalityManagement tests cardinality management performance
func BenchmarkCardinalityManagement(b *testing.B) {
	cardinalityLimits := []int{100, 1000, 10000, 50000}

	for _, limit := range cardinalityLimits {
		b.Run(fmt.Sprintf("Cardinality_%d", limit), func(b *testing.B) {
			benchmarkCardinality(b, limit)
		})
	}
}

func benchmarkCardinality(b *testing.B, cardinalityLimit int) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientWithJobs(cardinalityLimit / 10) // 10 metrics per job

	collector := collector.NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Note: This would need actual cardinality management implementation

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, cardinalityLimit*2)
		err := collector.Collect(ctx, ch)
		if err != nil {
			b.Fatalf("Collection failed: %v", err)
		}
		close(ch)

		// Drain the channel
		for range ch {
		}
	}
}

// BenchmarkLabelGeneration tests label creation performance
func BenchmarkLabelGeneration(b *testing.B) {
	logger := testutil.GetTestLogger()
	mockClient := setupMockSlurmClientWithJobs(1000)

	collector := collector.NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Set custom labels to test label merging performance
	customLabels := map[string]string{
		"cluster_name": "test-cluster",
		"environment":  "benchmark",
		"region":       "us-east-1",
		"datacenter":   "dc1",
		"zone":         "zone-a",
	}
	collector.SetCustomLabels(customLabels)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 10000)
		err := collector.Collect(ctx, ch)
		if err != nil {
			b.Fatalf("Collection failed: %v", err)
		}
		close(ch)

		// Drain the channel
		for range ch {
		}
	}
}

// BenchmarkTimeSeriesLoad tests time series creation load
func BenchmarkTimeSeriesLoad(b *testing.B) {
	b.Run("TimeSeriesCreation", func(b *testing.B) {
		registry := prometheus.NewRegistry()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Create multiple gauge vectors to simulate metric creation
			for j := 0; j < 100; j++ {
				gauge := prometheus.NewGaugeVec(
					prometheus.GaugeOpts{
						Name: fmt.Sprintf("test_metric_%d_%d", i, j),
						Help: "Test metric for benchmarking",
					},
					[]string{"job_id", "user", "partition", "state"},
				)

				// Add some sample data
				gauge.WithLabelValues("12345", "testuser", "compute", "running").Set(1)
				gauge.WithLabelValues("12346", "testuser2", "gpu", "pending").Set(0)

				registry.Register(gauge)
			}
		}
	})
}

// Helper functions for setting up mock clients

func setupMockSlurmClientWithJobs(jobCount int) *mocks.MockSlurmClient {
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Generate large job list
	jobList := fixtures.GenerateLargeJobList(jobCount)

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(jobList, nil)

	return mockClient
}

func setupMockSlurmClientWithNodes(nodeCount int) *mocks.MockSlurmClient {
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Generate large node list
	nodeList := fixtures.GenerateLargeNodeList(nodeCount)

	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(nodeList, nil)

	return mockClient
}

func setupMockSlurmClientFull() *mocks.MockSlurmClient {
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)
	mockNodeManager := new(mocks.MockNodeManager)
	mockPartitionManager := new(mocks.MockPartitionManager)

	mockClient.On("Jobs").Return(mockJobManager)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockClient.On("Partitions").Return(mockPartitionManager)

	mockJobManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestJobList(), nil)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	return mockClient
}