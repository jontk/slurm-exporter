package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockReservationUtilizationSLURMClient struct {
	mock.Mock
}

func (m *MockReservationUtilizationSLURMClient) GetReservationUtilization(ctx context.Context, reservationName string) (*ReservationUtilization, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationUtilization), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) ListReservations(ctx context.Context) ([]*ReservationInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReservationInfo), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationStatistics(ctx context.Context, reservationName string) (*ReservationStatistics, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationStatistics), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationEfficiency(ctx context.Context, reservationName string) (*ReservationEfficiency, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationEfficiency), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationConflicts(ctx context.Context, reservationName string) ([]*ReservationConflict, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReservationConflict), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationAlerts(ctx context.Context, reservationName string) ([]*ReservationAlert, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReservationAlert), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationTrends(ctx context.Context, reservationName string) (*ReservationTrends, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationTrends), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationOverlaps(ctx context.Context, reservationName string) ([]*ReservationOverlap, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReservationOverlap), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationOptimization(ctx context.Context, reservationName string) (*ReservationOptimization, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationOptimization), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationScheduling(ctx context.Context, reservationName string) (*ReservationScheduling, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationScheduling), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationCapacity(ctx context.Context, reservationName string) (*ReservationCapacity, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ReservationCapacity), args.Error(1)
}

func (m *MockReservationUtilizationSLURMClient) GetReservationUsageHistory(ctx context.Context, reservationName string) ([]*ReservationUsageHistory, error) {
	args := m.Called(ctx, reservationName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ReservationUsageHistory), args.Error(1)
}

func TestNewReservationUtilizationCollector(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
}

func TestReservationUtilizationCollector_Describe(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 50, count)
}

func TestReservationUtilizationCollector_Collect_Success(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	startTime := time.Now().Add(-24 * time.Hour)
	endTime := time.Now()

	reservations := []*ReservationInfo{
		{
			Name:      "gpu_reservation",
			State:     "ACTIVE",
			StartTime: startTime,
			EndTime:   endTime,
			Nodes:     []string{"gpu01", "gpu02"},
			Users:     []string{"user1", "user2"},
			Accounts:  []string{"research", "ml"},
			Features:  []string{"gpu", "high_memory"},
			NodeCount: 2,
		},
		{
			Name:      "cpu_reservation",
			State:     "INACTIVE",
			StartTime: startTime,
			EndTime:   endTime,
			Nodes:     []string{"cpu01", "cpu02", "cpu03"},
			Users:     []string{"user3"},
			Accounts:  []string{"compute"},
			Features:  []string{"cpu", "standard"},
			NodeCount: 3,
		},
	}

	utilization := &ReservationUtilization{
		ReservationName:  "gpu_reservation",
		TotalCPUs:        128,
		UsedCPUs:         96,
		TotalMemoryMB:    512000,
		UsedMemoryMB:     384000,
		TotalGPUs:        8,
		UsedGPUs:         6,
		TotalNodes:       2,
		UsedNodes:        2,
		ActiveJobs:       12,
		QueuedJobs:       3,
		UtilizationRate:  0.75,
		EfficiencyScore:  0.85,
		LastUpdated:      time.Now(),
	}

	statistics := &ReservationStatistics{
		ReservationName:    "gpu_reservation",
		TotalJobsRun:       245,
		SuccessfulJobs:     230,
		FailedJobs:         15,
		CancelledJobs:      8,
		AverageJobDuration: 3600,
		AverageWaitTime:    300,
		PeakUtilization:    0.95,
		LowestUtilization:  0.15,
		AverageUtilization: 0.72,
		TotalCPUHours:      15680,
		TotalGPUHours:      3920,
		WastedCPUHours:     1240,
		WastedGPUHours:     310,
		OversubscriptionEvents: 5,
		UndersubscriptionEvents: 23,
	}

	efficiency := &ReservationEfficiency{
		ReservationName:        "gpu_reservation",
		CPUEfficiency:          0.88,
		MemoryEfficiency:       0.75,
		GPUEfficiency:          0.92,
		NetworkEfficiency:      0.68,
		StorageEfficiency:      0.71,
		OverallEfficiency:      0.85,
		ResourceWastePercent:   0.15,
		OptimizationPotential:  0.12,
		BottleneckResources:    []string{"memory", "network"},
		OptimizationSuggestions: []string{"increase_memory", "optimize_network"},
	}

	conflicts := []*ReservationConflict{
		{
			ReservationName:      "gpu_reservation",
			ConflictType:         "RESOURCE_OVERLAP",
			ConflictDescription:  "GPU resources overlap with maintenance window",
			ConflictSeverity:     "HIGH",
			ConflictStartTime:    time.Now().Add(2 * time.Hour),
			ConflictEndTime:      time.Now().Add(4 * time.Hour),
			AffectedResources:    []string{"gpu01", "gpu02"},
			ResolutionRequired:   true,
			ResolutionSuggestion: "Reschedule maintenance or reduce reservation scope",
		},
	}

	alerts := []*ReservationAlert{
		{
			ReservationName: "gpu_reservation",
			AlertType:       "LOW_UTILIZATION",
			AlertMessage:    "Reservation utilization below threshold",
			AlertSeverity:   "WARNING",
			AlertTime:       time.Now(),
			ThresholdValue:  0.5,
			CurrentValue:    0.35,
			AlertActive:     true,
		},
	}

	trends := &ReservationTrends{
		ReservationName:         "gpu_reservation",
		UtilizationTrend:        "INCREASING",
		EfficiencyTrend:         "STABLE",
		DemandTrend:            "INCREASING",
		WasteTrend:             "DECREASING",
		PredictedUtilization:    0.82,
		PredictedEfficiency:     0.87,
		TrendConfidence:         0.78,
		ForecastAccuracy:        0.85,
		SeasonalPatterns:        []string{"weekday_peak", "weekend_low"},
		AnomalyDetected:         false,
		NextMaintenanceWindow:   time.Now().Add(7 * 24 * time.Hour),
		OptimalReservationSize:  140,
		RecommendedAdjustments:  []string{"extend_duration", "add_nodes"},
	}

	overlaps := []*ReservationOverlap{
		{
			ReservationName:    "gpu_reservation",
			OverlapType:        "PARTIAL",
			OverlappingReservation: "shared_gpu_reservation",
			OverlapStartTime:   time.Now().Add(1 * time.Hour),
			OverlapEndTime:     time.Now().Add(3 * time.Hour),
			OverlapSeverity:    "MEDIUM",
			OverlapResolution:  "TIME_SHARING",
			ConflictPotential:  0.3,
		},
	}

	optimization := &ReservationOptimization{
		ReservationName:              "gpu_reservation",
		CurrentEfficiencyScore:       0.85,
		OptimalEfficiencyScore:       0.95,
		ImprovementPotential:         0.10,
		OptimizationRecommendations:  []string{"right_size_memory", "optimize_scheduling"},
		EstimatedCostSavings:         1250.50,
		EstimatedPerformanceGain:     0.15,
		ImplementationComplexity:     "MEDIUM",
		ImplementationTimeframe:      "2_WEEKS",
		RiskAssessment:              "LOW",
		OptimizationPriority:        "HIGH",
	}

	scheduling := &ReservationScheduling{
		ReservationName:      "gpu_reservation",
		SchedulingPolicy:     "FAIR_SHARE",
		SchedulingEfficiency: 0.78,
		AverageQueueTime:     450,
		QueueLength:          8,
		SchedulingConflicts:  2,
		PreemptionEvents:     1,
		BackfillSuccess:      0.85,
		SchedulingLatency:    12,
		ResourceFragmentation: 0.15,
		SchedulingOptimization: []string{"improve_backfill", "reduce_fragmentation"},
	}

	capacity := &ReservationCapacity{
		ReservationName:        "gpu_reservation",
		TotalCapacityCPUs:      128,
		TotalCapacityMemoryMB:  512000,
		TotalCapacityGPUs:      8,
		TotalCapacityNodes:     2,
		UsedCapacityCPUs:       96,
		UsedCapacityMemoryMB:   384000,
		UsedCapacityGPUs:       6,
		UsedCapacityNodes:      2,
		AvailableCapacityCPUs:  32,
		AvailableCapacityMemoryMB: 128000,
		AvailableCapacityGPUs:  2,
		AvailableCapacityNodes: 0,
		CapacityUtilization:    0.75,
		CapacityEfficiency:     0.85,
		CapacityBottlenecks:    []string{"nodes"},
		ExpansionPotential:     0.25,
		ScalingRecommendations: []string{"add_nodes"},
	}

	usageHistory := []*ReservationUsageHistory{
		{
			ReservationName: "gpu_reservation",
			Timestamp:      time.Now().Add(-1 * time.Hour),
			CPUUtilization: 0.78,
			MemoryUtilization: 0.72,
			GPUUtilization: 0.89,
			NodeUtilization: 1.0,
			JobCount:       15,
			UserCount:      8,
			AccountCount:   3,
			EfficiencyScore: 0.83,
		},
	}

	client.On("ListReservations", mock.Anything).Return(reservations, nil)
	client.On("GetReservationUtilization", mock.Anything, "gpu_reservation").Return(utilization, nil)
	client.On("GetReservationUtilization", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationStatistics", mock.Anything, "gpu_reservation").Return(statistics, nil)
	client.On("GetReservationStatistics", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationEfficiency", mock.Anything, "gpu_reservation").Return(efficiency, nil)
	client.On("GetReservationEfficiency", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationConflicts", mock.Anything, "gpu_reservation").Return(conflicts, nil)
	client.On("GetReservationConflicts", mock.Anything, "cpu_reservation").Return([]*ReservationConflict{}, nil)
	client.On("GetReservationAlerts", mock.Anything, "gpu_reservation").Return(alerts, nil)
	client.On("GetReservationAlerts", mock.Anything, "cpu_reservation").Return([]*ReservationAlert{}, nil)
	client.On("GetReservationTrends", mock.Anything, "gpu_reservation").Return(trends, nil)
	client.On("GetReservationTrends", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationOverlaps", mock.Anything, "gpu_reservation").Return(overlaps, nil)
	client.On("GetReservationOverlaps", mock.Anything, "cpu_reservation").Return([]*ReservationOverlap{}, nil)
	client.On("GetReservationOptimization", mock.Anything, "gpu_reservation").Return(optimization, nil)
	client.On("GetReservationOptimization", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationScheduling", mock.Anything, "gpu_reservation").Return(scheduling, nil)
	client.On("GetReservationScheduling", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationCapacity", mock.Anything, "gpu_reservation").Return(capacity, nil)
	client.On("GetReservationCapacity", mock.Anything, "cpu_reservation").Return(nil, errors.New("not found"))
	client.On("GetReservationUsageHistory", mock.Anything, "gpu_reservation").Return(usageHistory, nil)
	client.On("GetReservationUsageHistory", mock.Anything, "cpu_reservation").Return([]*ReservationUsageHistory{}, nil)

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	assert.Greater(t, len(metrics), 0)

	client.AssertExpectations(t)
}

func TestReservationUtilizationCollector_Collect_ListReservationsError(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	client.On("ListReservations", mock.Anything).Return(nil, errors.New("API error"))

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	assert.Equal(t, 0, len(metrics))

	client.AssertExpectations(t)
}

func TestReservationUtilizationCollector_CollectUtilizationMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	utilization := &ReservationUtilization{
		ReservationName:  "test_reservation",
		TotalCPUs:        64,
		UsedCPUs:        48,
		TotalMemoryMB:   256000,
		UsedMemoryMB:    192000,
		TotalGPUs:       4,
		UsedGPUs:        3,
		TotalNodes:      4,
		UsedNodes:       3,
		ActiveJobs:      10,
		QueuedJobs:      2,
		UtilizationRate: 0.75,
		EfficiencyScore: 0.85,
		LastUpdated:     time.Now(),
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectUtilizationMetrics(ch, "test_reservation", utilization)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 12, count)
}

func TestReservationUtilizationCollector_CollectStatisticsMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	statistics := &ReservationStatistics{
		ReservationName:    "test_reservation",
		TotalJobsRun:       100,
		SuccessfulJobs:     95,
		FailedJobs:         5,
		CancelledJobs:      2,
		AverageJobDuration: 3600,
		AverageWaitTime:    300,
		PeakUtilization:    0.95,
		LowestUtilization:  0.15,
		AverageUtilization: 0.75,
		TotalCPUHours:      7200,
		TotalGPUHours:      1800,
		WastedCPUHours:     720,
		WastedGPUHours:     180,
		OversubscriptionEvents: 3,
		UndersubscriptionEvents: 15,
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectStatisticsMetrics(ch, "test_reservation", statistics)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 16, count)
}

func TestReservationUtilizationCollector_CollectEfficiencyMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	efficiency := &ReservationEfficiency{
		ReservationName:        "test_reservation",
		CPUEfficiency:          0.88,
		MemoryEfficiency:       0.75,
		GPUEfficiency:          0.92,
		NetworkEfficiency:      0.68,
		StorageEfficiency:      0.71,
		OverallEfficiency:      0.85,
		ResourceWastePercent:   0.15,
		OptimizationPotential:  0.12,
		BottleneckResources:    []string{"memory", "network"},
		OptimizationSuggestions: []string{"increase_memory", "optimize_network"},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectEfficiencyMetrics(ch, "test_reservation", efficiency)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 8, count)
}

func TestReservationUtilizationCollector_CollectConflictMetrics(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	conflicts := []*ReservationConflict{
		{
			ReservationName:      "test_reservation",
			ConflictType:         "RESOURCE_OVERLAP",
			ConflictDescription:  "Resource overlap detected",
			ConflictSeverity:     "HIGH",
			ConflictStartTime:    time.Now(),
			ConflictEndTime:      time.Now().Add(2 * time.Hour),
			AffectedResources:    []string{"node01", "node02"},
			ResolutionRequired:   true,
			ResolutionSuggestion: "Reschedule conflicting reservation",
		},
		{
			ReservationName:      "test_reservation",
			ConflictType:         "TIME_OVERLAP",
			ConflictDescription:  "Time overlap with maintenance",
			ConflictSeverity:     "MEDIUM",
			ConflictStartTime:    time.Now().Add(4 * time.Hour),
			ConflictEndTime:      time.Now().Add(6 * time.Hour),
			AffectedResources:    []string{"node03"},
			ResolutionRequired:   false,
			ResolutionSuggestion: "Monitor during maintenance window",
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectConflictMetrics(ch, "test_reservation", conflicts)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 4, count)
}

func TestReservationUtilizationCollector_DescribeAll(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	gatherer := prometheus.Gatherer(registry)
	metricFamilies, err := gatherer.Gather()

	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestReservationUtilizationCollector_MetricCount(t *testing.T) {
	client := &MockReservationUtilizationSLURMClient{}
	collector := NewReservationUtilizationCollector(client)

	expectedMetrics := []string{
		"slurm_reservation_total_cpus",
		"slurm_reservation_used_cpus",
		"slurm_reservation_total_memory_bytes",
		"slurm_reservation_used_memory_bytes",
		"slurm_reservation_total_gpus",
		"slurm_reservation_used_gpus",
		"slurm_reservation_total_nodes",
		"slurm_reservation_used_nodes",
		"slurm_reservation_active_jobs",
		"slurm_reservation_queued_jobs",
		"slurm_reservation_utilization_rate",
		"slurm_reservation_efficiency_score",
		"slurm_reservation_jobs_total",
		"slurm_reservation_jobs_successful",
		"slurm_reservation_jobs_failed",
		"slurm_reservation_jobs_cancelled",
		"slurm_reservation_job_duration_seconds",
		"slurm_reservation_job_wait_time_seconds",
		"slurm_reservation_utilization_peak",
		"slurm_reservation_utilization_lowest",
		"slurm_reservation_utilization_average",
		"slurm_reservation_cpu_hours_total",
		"slurm_reservation_gpu_hours_total",
		"slurm_reservation_cpu_hours_wasted",
		"slurm_reservation_gpu_hours_wasted",
		"slurm_reservation_oversubscription_events",
		"slurm_reservation_undersubscription_events",
		"slurm_reservation_cpu_efficiency",
		"slurm_reservation_memory_efficiency",
		"slurm_reservation_gpu_efficiency",
		"slurm_reservation_network_efficiency",
		"slurm_reservation_storage_efficiency",
		"slurm_reservation_overall_efficiency",
		"slurm_reservation_resource_waste_percent",
		"slurm_reservation_optimization_potential",
		"slurm_reservation_conflicts_total",
		"slurm_reservation_conflicts_resolution_required",
		"slurm_reservation_conflicts_by_severity",
		"slurm_reservation_conflicts_by_type",
		"slurm_reservation_alerts_total",
		"slurm_reservation_alerts_active",
		"slurm_reservation_predicted_utilization",
		"slurm_reservation_predicted_efficiency",
		"slurm_reservation_trend_confidence",
		"slurm_reservation_forecast_accuracy",
		"slurm_reservation_anomaly_detected",
		"slurm_reservation_optimal_size",
		"slurm_reservation_overlaps_total",
		"slurm_reservation_overlap_conflict_potential",
		"slurm_reservation_optimization_potential_score",
		"slurm_reservation_estimated_cost_savings",
	}

	assert.Equal(t, len(expectedMetrics), 50)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 50, count)
}