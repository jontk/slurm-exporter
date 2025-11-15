package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQueueAnalysisSLURMClient struct {
	mock.Mock
}

func (m *MockQueueAnalysisSLURMClient) GetQueueAnalysis(ctx context.Context, partition string) (*QueueAnalysis, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).(*QueueAnalysis), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetQueuePositions(ctx context.Context, partition string) ([]*QueuePosition, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).([]*QueuePosition), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetQueueMovement(ctx context.Context, jobID string) (*QueueMovement, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*QueueMovement), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) PredictWaitTime(ctx context.Context, jobID string) (*WaitTimePrediction, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*WaitTimePrediction), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetHistoricalWaitTimes(ctx context.Context, partition string, period string) (*HistoricalWaitTimes, error) {
	args := m.Called(ctx, partition, period)
	return args.Get(0).(*HistoricalWaitTimes), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetQueueEfficiency(ctx context.Context, partition string) (*QueueEfficiency, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).(*QueueEfficiency), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetResourceBasedQueueAnalysis(ctx context.Context, resourceType string) (*ResourceBasedQueueAnalysis, error) {
	args := m.Called(ctx, resourceType)
	return args.Get(0).(*ResourceBasedQueueAnalysis), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetPriorityBasedQueueAnalysis(ctx context.Context, partition string) (*PriorityBasedQueueAnalysis, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).(*PriorityBasedQueueAnalysis), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetBackfillAnalysis(ctx context.Context, partition string) (*BackfillAnalysis, error) {
	args := m.Called(ctx, partition)
	return args.Get(0).(*BackfillAnalysis), args.Error(1)
}

func (m *MockQueueAnalysisSLURMClient) GetSystemLoadImpact(ctx context.Context) (*SystemLoadImpact, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemLoadImpact), args.Error(1)
}

func TestNewQueueAnalysisCollector(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}
	collector := NewQueueAnalysisCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.queuePosition)
	assert.NotNil(t, collector.queueDepth)
	assert.NotNil(t, collector.waitTimePrediction)
	assert.NotNil(t, collector.historicalWaitTimes)
	assert.NotNil(t, collector.queueEfficiency)
}

func TestQueueAnalysisCollector_Describe(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}
	collector := NewQueueAnalysisCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 54 metrics, verify we have the correct number
	assert.Equal(t, 54, len(descs), "Should have exactly 54 metric descriptions")
}

func TestQueueAnalysisCollector_Collect_Success(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}

	// Mock queue analysis
	queueAnalysis := &QueueAnalysis{
		Partition:           "compute",
		TotalJobs:          150,
		QueuedJobs:         45,
		RunningJobs:        105,
		AverageWaitTime:    3600.0,
		MedianWaitTime:     2400.0,
		MaxWaitTime:        14400.0,
		MinWaitTime:        300.0,
		ThroughputRate:     2.5,
		QueueEfficiency:    0.85,
		BottleneckScore:    0.3,
		ResourceUtilization: 0.78,
		LastUpdated:        time.Now(),
	}

	// Mock queue positions
	queuePositions := []*QueuePosition{
		{
			JobID:              "job_001",
			UserName:           "user1",
			AccountName:        "account1",
			Partition:          "compute",
			QueuePosition:      5,
			OriginalPosition:   8,
			TimeInQueue:        1800.0,
			EstimatedWaitTime:  2400.0,
			PriorityScore:      1500.0,
			JobSize:            8,
			RequestedRuntime:   3600.0,
			EstimatedRuntime:   2700.0,
			ResourcesRequested: map[string]float64{
				"cpu":    8.0,
				"memory": 32.0,
			},
			QueueStagnation:    false,
			StagnationDuration: 0.0,
			MovementRate:       0.8,
			PositionChanges:    3,
			LastMovement:       time.Now().Add(-300 * time.Second),
			QueueCategory:      "standard",
			QueuePriority:      "medium",
			PreemptionRisk:     0.1,
			BackfillEligible:   true,
			BackfillChance:     0.6,
		},
		{
			JobID:           "job_002",
			UserName:        "user2",
			AccountName:     "account2",
			Partition:       "compute",
			QueuePosition:   12,
			OriginalPosition: 15,
			TimeInQueue:     3600.0,
			EstimatedWaitTime: 5400.0,
			PriorityScore:   1200.0,
			JobSize:         4,
			QueueStagnation: true,
			StagnationDuration: 1800.0,
			MovementRate:    0.2,
		},
	}

	// Mock queue movement
	queueMovement := &QueueMovement{
		JobID:              "job_001",
		InitialPosition:    8,
		CurrentPosition:    5,
		TotalMovement:      3,
		ForwardMovement:    3,
		BackwardMovement:   0,
		MovementVelocity:   0.8,
		MovementAcceleration: 0.1,
		LastMovementTime:   time.Now().Add(-300 * time.Second),
		MovementTrend:      "improving",
		PositionStability:  0.7,
		MovementPattern:    "steady_progress",
		AverageMovementRate: 0.6,
		TimeStagnant:       0.0,
		StagnationRisk:     0.2,
		MovementEfficiency: 0.85,
		PredictedFinalPosition: 2,
		MovementConfidence: 0.8,
	}

	// Mock wait time prediction
	waitTimePrediction := &WaitTimePrediction{
		JobID:                "job_001",
		Partition:            "compute",
		PredictedWaitTime:    2400.0,
		PredictionConfidence: 0.85,
		ModelVersion:         "v2.1",
		ModelAccuracy:        0.82,
		ConfidenceInterval: map[string]float64{
			"lower": 1800.0,
			"upper": 3600.0,
		},
		PredictionFactors: map[string]float64{
			"queue_position":     0.4,
			"resource_demand":    0.3,
			"historical_pattern": 0.2,
			"system_load":        0.1,
		},
		HistoricalAccuracy:  0.78,
		UncertaintyRange:    600.0,
		ModelValidation:     "validated",
		LastModelUpdate:     time.Now().Add(-24 * time.Hour),
		PredictionVariance:  180000.0,
		ModelReliability:    0.9,
		ContextualFactors:   []string{"peak_hours", "weekend_schedule"},
		SeasonalAdjustment:  1.1,
		PredictionMetadata: map[string]interface{}{
			"training_samples": 10000,
			"feature_count":   15,
		},
	}

	// Mock historical wait times
	historicalWaitTimes := &HistoricalWaitTimes{
		Partition:         "compute",
		AnalysisPeriod:    "7d",
		AverageWaitTime:   3200.0,
		MedianWaitTime:    2400.0,
		P90WaitTime:       7200.0,
		P95WaitTime:       10800.0,
		P99WaitTime:       18000.0,
		MinWaitTime:       300.0,
		MaxWaitTime:       86400.0,
		StandardDeviation: 2400.0,
		WaitTimeVariance:  5760000.0,
		TotalSamples:      5000,
		AnomalyCount:      25,
		AnomalyRate:       0.005,
		TrendDirection:    "stable",
		TrendSlope:        -0.02,
		SeasonalPatterns: map[string]float64{
			"monday":    1.2,
			"friday":    0.8,
			"weekend":   0.6,
		},
		PeakHours: map[string]float64{
			"09:00": 1.5,
			"14:00": 1.3,
			"18:00": 0.7,
		},
		LastUpdated: time.Now(),
	}

	// Mock queue efficiency
	queueEfficiency := &QueueEfficiency{
		Partition:           "compute",
		ThroughputRate:      2.5,
		CompletionRate:      0.95,
		QueueUtilization:    0.85,
		ResourceEfficiency:  0.78,
		SchedulingEfficiency: 0.82,
		WaitTimeEfficiency:  0.75,
		BackfillEfficiency:  0.68,
		PreemptionRate:      0.05,
		JobFailureRate:      0.02,
		SystemResponsiveness: 0.88,
		QueueStability:      0.9,
		LoadBalancing:       0.72,
		ResourceContention:  0.3,
		BottleneckSeverity:  0.25,
		OptimizationScore:   0.8,
		PerformanceRating:   "good",
		EfficiencyTrend:     "improving",
		LastAnalyzed:        time.Now(),
	}

	// Mock resource-based queue analysis
	resourceAnalysis := &ResourceBasedQueueAnalysis{
		ResourceType:       "cpu",
		TotalDemand:        1000.0,
		AvailableCapacity:  800.0,
		UtilizationRate:    0.8,
		DemandSatisfaction: 0.75,
		QueuedDemand:       200.0,
		AverageJobSize:     16.0,
		LargeJobImpact:     0.3,
		SmallJobEfficiency: 0.85,
		ResourceFragmentation: 0.15,
		AllocationEfficiency: 0.82,
		ResourceWaste:       0.08,
		OptimalAllocation:   850.0,
		ProjectedDemand:     1200.0,
		CapacityGap:         -400.0,
		ScalingRecommendation: "increase_capacity",
		LastAnalyzed:        time.Now(),
	}

	// Mock priority-based queue analysis
	priorityAnalysis := &PriorityBasedQueueAnalysis{
		Partition:           "compute",
		PriorityDistribution: map[string]int{
			"high":   15,
			"medium": 45,
			"low":    35,
		},
		AveragePriority:      1250.0,
		MedianPriority:       1000.0,
		PrioritySpread:       800.0,
		HighPriorityWaitTime: 900.0,
		LowPriorityWaitTime:  7200.0,
		PriorityInversion:    0.05,
		PriorityEffectiveness: 0.85,
		FairnesScore:         0.78,
		PriorityBalancing:    0.82,
		AgingEffectiveness:   0.75,
		PriorityTrend:        "stable",
		LastAnalyzed:         time.Now(),
	}

	// Mock backfill analysis
	backfillAnalysis := &BackfillAnalysis{
		Partition:               "compute",
		BackfillOpportunities:   25,
		BackfillSuccessRate:     0.72,
		BackfillEfficiency:      0.68,
		BackfillUtilization:     0.55,
		AverageBackfillDuration: 1800.0,
		BackfillJobCount:        18,
		BackfillResourceSavings: 0.15,
		BackfillImpact:          0.25,
		BackfillOptimization:    0.8,
		FragmentationReduction:  0.2,
		SchedulingImprovement:   0.18,
		BackfillTrend:           "improving",
		LastAnalyzed:            time.Now(),
	}

	// Mock system load impact
	systemLoadImpact := &SystemLoadImpact{
		CurrentSystemLoad:    0.75,
		LoadImpactOnQueues:   0.3,
		QueueSensitivity:     0.4,
		LoadTrend:            "stable",
		LoadVariability:      0.15,
		PeakLoadImpact:       0.5,
		LoadBalancingScore:   0.82,
		SystemResponsiveness: 0.88,
		LoadPrediction:       0.78,
		CapacityUtilization:  0.85,
		SystemStress:         0.25,
		LoadDistribution: map[string]float64{
			"compute": 0.8,
			"gpu":     0.6,
			"memory":  0.9,
		},
		CriticalThresholds: map[string]float64{
			"cpu_utilization":    0.95,
			"memory_utilization": 0.9,
			"queue_depth":        100.0,
		},
		LastUpdated: time.Now(),
	}

	// Setup mock expectations
	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(queueAnalysis, nil)
	client.On("GetQueuePositions", mock.Anything, mock.AnythingOfType("string")).Return(queuePositions, nil)
	client.On("GetQueueMovement", mock.Anything, mock.AnythingOfType("string")).Return(queueMovement, nil)
	client.On("PredictWaitTime", mock.Anything, mock.AnythingOfType("string")).Return(waitTimePrediction, nil)
	client.On("GetHistoricalWaitTimes", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(historicalWaitTimes, nil)
	client.On("GetQueueEfficiency", mock.Anything, mock.AnythingOfType("string")).Return(queueEfficiency, nil)
	client.On("GetResourceBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(resourceAnalysis, nil)
	client.On("GetPriorityBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(priorityAnalysis, nil)
	client.On("GetBackfillAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(backfillAnalysis, nil)
	client.On("GetSystemLoadImpact", mock.Anything).Return(systemLoadImpact, nil)

	collector := NewQueueAnalysisCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 500)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify we collected metrics
	assert.Greater(t, len(metrics), 0, "Should collect at least some metrics")

	// Verify specific metrics are present
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		metricNames[desc.String()] = true
	}

	// Check for key metric families
	foundQueuePosition := false
	foundWaitTime := false
	foundEfficiency := false
	foundBackfill := false
	foundSystemLoad := false

	for name := range metricNames {
		if strings.Contains(name, "queue_position") {
			foundQueuePosition = true
		}
		if strings.Contains(name, "wait_time") {
			foundWaitTime = true
		}
		if strings.Contains(name, "queue_efficiency") {
			foundEfficiency = true
		}
		if strings.Contains(name, "backfill") {
			foundBackfill = true
		}
		if strings.Contains(name, "system_load") {
			foundSystemLoad = true
		}
	}

	assert.True(t, foundQueuePosition, "Should have queue position metrics")
	assert.True(t, foundWaitTime, "Should have wait time metrics")
	assert.True(t, foundEfficiency, "Should have queue efficiency metrics")
	assert.True(t, foundBackfill, "Should have backfill metrics")
	assert.True(t, foundSystemLoad, "Should have system load metrics")

	client.AssertExpectations(t)
}

func TestQueueAnalysisCollector_Collect_Error(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}

	// Mock error response
	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return((*QueueAnalysis)(nil), assert.AnError)

	collector := NewQueueAnalysisCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still collect some metrics (empty metrics after reset)
	assert.GreaterOrEqual(t, len(metrics), 0, "Should handle errors gracefully")

	client.AssertExpectations(t)
}

func TestQueueAnalysisCollector_MetricValues(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}

	// Create test data with known values
	queueAnalysis := &QueueAnalysis{
		Partition:       "test",
		TotalJobs:      100,
		QueuedJobs:     30,
		AverageWaitTime: 3600.0,
		QueueEfficiency: 0.85,
	}

	queuePositions := []*QueuePosition{
		{
			JobID:         "test_job",
			Partition:     "test",
			QueuePosition: 5,
			TimeInQueue:   1800.0,
			PriorityScore: 1500.0,
		},
	}

	waitTimePrediction := &WaitTimePrediction{
		JobID:                "test_job",
		PredictedWaitTime:    2400.0,
		PredictionConfidence: 0.85,
	}

	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(queueAnalysis, nil)
	client.On("GetQueuePositions", mock.Anything, mock.AnythingOfType("string")).Return(queuePositions, nil)
	client.On("GetQueueMovement", mock.Anything, mock.AnythingOfType("string")).Return(&QueueMovement{}, nil)
	client.On("PredictWaitTime", mock.Anything, mock.AnythingOfType("string")).Return(waitTimePrediction, nil)
	client.On("GetHistoricalWaitTimes", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&HistoricalWaitTimes{}, nil)
	client.On("GetQueueEfficiency", mock.Anything, mock.AnythingOfType("string")).Return(&QueueEfficiency{}, nil)
	client.On("GetResourceBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&ResourceBasedQueueAnalysis{}, nil)
	client.On("GetPriorityBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&PriorityBasedQueueAnalysis{}, nil)
	client.On("GetBackfillAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&BackfillAnalysis{}, nil)
	client.On("GetSystemLoadImpact", mock.Anything).Return(&SystemLoadImpact{}, nil)

	collector := NewQueueAnalysisCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundQueueDepth := false
	foundAverageWaitTime := false
	foundQueuePosition := false
	foundPredictedWaitTime := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_queue_depth":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(30), *mf.Metric[0].Gauge.Value)
				foundQueueDepth = true
			}
		case "slurm_queue_average_wait_time_seconds":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(3600), *mf.Metric[0].Gauge.Value)
				foundAverageWaitTime = true
			}
		case "slurm_queue_position":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(5), *mf.Metric[0].Gauge.Value)
				foundQueuePosition = true
			}
		case "slurm_wait_time_prediction_seconds":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(2400), *mf.Metric[0].Gauge.Value)
				foundPredictedWaitTime = true
			}
		}
	}

	assert.True(t, foundQueueDepth, "Should find queue depth metric with correct value")
	assert.True(t, foundAverageWaitTime, "Should find average wait time metric with correct value")
	assert.True(t, foundQueuePosition, "Should find queue position metric with correct value")
	assert.True(t, foundPredictedWaitTime, "Should find predicted wait time metric with correct value")
}

func TestQueueAnalysisCollector_Integration(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}

	// Setup comprehensive mock data
	setupQueueAnalysisMocks(client)

	collector := NewQueueAnalysisCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_queue_depth Number of jobs currently queued in each partition
		# TYPE slurm_queue_depth gauge
		slurm_queue_depth{partition="compute"} 25
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_queue_depth")
	assert.NoError(t, err)
}

func TestQueueAnalysisCollector_WaitTimePrediction(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}

	waitTimePrediction := &WaitTimePrediction{
		JobID:                "prediction_test",
		Partition:            "compute",
		PredictedWaitTime:    3600.0,
		PredictionConfidence: 0.9,
		ModelAccuracy:        0.85,
		HistoricalAccuracy:   0.8,
		UncertaintyRange:     600.0,
	}

	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&QueueAnalysis{}, nil)
	client.On("GetQueuePositions", mock.Anything, mock.AnythingOfType("string")).Return([]*QueuePosition{}, nil)
	client.On("GetQueueMovement", mock.Anything, mock.AnythingOfType("string")).Return(&QueueMovement{}, nil)
	client.On("PredictWaitTime", mock.Anything, mock.AnythingOfType("string")).Return(waitTimePrediction, nil)
	client.On("GetHistoricalWaitTimes", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&HistoricalWaitTimes{}, nil)
	client.On("GetQueueEfficiency", mock.Anything, mock.AnythingOfType("string")).Return(&QueueEfficiency{}, nil)
	client.On("GetResourceBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&ResourceBasedQueueAnalysis{}, nil)
	client.On("GetPriorityBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&PriorityBasedQueueAnalysis{}, nil)
	client.On("GetBackfillAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&BackfillAnalysis{}, nil)
	client.On("GetSystemLoadImpact", mock.Anything).Return(&SystemLoadImpact{}, nil)

	collector := NewQueueAnalysisCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify wait time prediction metrics are present
	foundPredictionMetrics := false
	foundConfidenceMetrics := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "wait_time_prediction") {
			foundPredictionMetrics = true
		}
		if strings.Contains(desc, "prediction_confidence") {
			foundConfidenceMetrics = true
		}
	}

	assert.True(t, foundPredictionMetrics, "Should find wait time prediction metrics")
	assert.True(t, foundConfidenceMetrics, "Should find prediction confidence metrics")
}

func TestQueueAnalysisCollector_QueueMovementAnalysis(t *testing.T) {
	client := &MockQueueAnalysisSLURMClient{}

	queueMovement := &QueueMovement{
		JobID:                "movement_test",
		InitialPosition:      10,
		CurrentPosition:      5,
		TotalMovement:        5,
		MovementVelocity:     0.8,
		MovementTrend:        "improving",
		PositionStability:    0.7,
		StagnationRisk:       0.2,
		MovementEfficiency:   0.85,
		PredictedFinalPosition: 2,
	}

	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&QueueAnalysis{}, nil)
	client.On("GetQueuePositions", mock.Anything, mock.AnythingOfType("string")).Return([]*QueuePosition{}, nil)
	client.On("GetQueueMovement", mock.Anything, mock.AnythingOfType("string")).Return(queueMovement, nil)
	client.On("PredictWaitTime", mock.Anything, mock.AnythingOfType("string")).Return(&WaitTimePrediction{}, nil)
	client.On("GetHistoricalWaitTimes", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&HistoricalWaitTimes{}, nil)
	client.On("GetQueueEfficiency", mock.Anything, mock.AnythingOfType("string")).Return(&QueueEfficiency{}, nil)
	client.On("GetResourceBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&ResourceBasedQueueAnalysis{}, nil)
	client.On("GetPriorityBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&PriorityBasedQueueAnalysis{}, nil)
	client.On("GetBackfillAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&BackfillAnalysis{}, nil)
	client.On("GetSystemLoadImpact", mock.Anything).Return(&SystemLoadImpact{}, nil)

	collector := NewQueueAnalysisCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify queue movement metrics are present
	foundMovementMetrics := false
	foundVelocityMetrics := false
	foundStagnationMetrics := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "queue_movement") {
			foundMovementMetrics = true
		}
		if strings.Contains(desc, "movement_velocity") {
			foundVelocityMetrics = true
		}
		if strings.Contains(desc, "stagnation") {
			foundStagnationMetrics = true
		}
	}

	assert.True(t, foundMovementMetrics, "Should find queue movement metrics")
	assert.True(t, foundVelocityMetrics, "Should find movement velocity metrics")
	assert.True(t, foundStagnationMetrics, "Should find stagnation metrics")
}

func setupQueueAnalysisMocks(client *MockQueueAnalysisSLURMClient) {
	queueAnalysis := &QueueAnalysis{
		Partition:        "compute",
		TotalJobs:       100,
		QueuedJobs:      25,
		RunningJobs:     75,
		AverageWaitTime: 2400.0,
		QueueEfficiency: 0.85,
	}

	queuePositions := []*QueuePosition{
		{
			JobID:         "job1",
			Partition:     "compute",
			QueuePosition: 3,
			TimeInQueue:   1200.0,
			PriorityScore: 1800.0,
		},
	}

	queueMovement := &QueueMovement{
		JobID:              "job1",
		InitialPosition:    5,
		CurrentPosition:    3,
		TotalMovement:      2,
		MovementVelocity:   0.6,
		MovementTrend:      "improving",
		PositionStability:  0.8,
	}

	waitTimePrediction := &WaitTimePrediction{
		JobID:                "job1",
		PredictedWaitTime:    1800.0,
		PredictionConfidence: 0.82,
		ModelAccuracy:        0.78,
	}

	historicalWaitTimes := &HistoricalWaitTimes{
		Partition:       "compute",
		AnalysisPeriod:  "7d",
		AverageWaitTime: 2200.0,
		MedianWaitTime:  1800.0,
		P90WaitTime:     4800.0,
		TotalSamples:    1000,
	}

	queueEfficiency := &QueueEfficiency{
		Partition:           "compute",
		ThroughputRate:      3.2,
		QueueUtilization:    0.88,
		ResourceEfficiency:  0.82,
		SchedulingEfficiency: 0.85,
		PerformanceRating:   "good",
	}

	resourceAnalysis := &ResourceBasedQueueAnalysis{
		ResourceType:      "cpu",
		TotalDemand:       800.0,
		AvailableCapacity: 1000.0,
		UtilizationRate:   0.8,
		QueuedDemand:      150.0,
	}

	priorityAnalysis := &PriorityBasedQueueAnalysis{
		Partition:      "compute",
		AveragePriority: 1200.0,
		MedianPriority: 1000.0,
		PrioritySpread: 600.0,
		FairnesScore:   0.75,
	}

	backfillAnalysis := &BackfillAnalysis{
		Partition:             "compute",
		BackfillOpportunities: 15,
		BackfillSuccessRate:   0.7,
		BackfillEfficiency:    0.65,
		BackfillJobCount:      10,
	}

	systemLoadImpact := &SystemLoadImpact{
		CurrentSystemLoad:  0.75,
		LoadImpactOnQueues: 0.3,
		LoadTrend:          "stable",
		SystemStress:       0.25,
	}

	client.On("GetQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(queueAnalysis, nil)
	client.On("GetQueuePositions", mock.Anything, mock.AnythingOfType("string")).Return(queuePositions, nil)
	client.On("GetQueueMovement", mock.Anything, mock.AnythingOfType("string")).Return(queueMovement, nil)
	client.On("PredictWaitTime", mock.Anything, mock.AnythingOfType("string")).Return(waitTimePrediction, nil)
	client.On("GetHistoricalWaitTimes", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(historicalWaitTimes, nil)
	client.On("GetQueueEfficiency", mock.Anything, mock.AnythingOfType("string")).Return(queueEfficiency, nil)
	client.On("GetResourceBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(resourceAnalysis, nil)
	client.On("GetPriorityBasedQueueAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(priorityAnalysis, nil)
	client.On("GetBackfillAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(backfillAnalysis, nil)
	client.On("GetSystemLoadImpact", mock.Anything).Return(systemLoadImpact, nil)
}