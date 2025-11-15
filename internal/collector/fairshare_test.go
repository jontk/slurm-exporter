package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFairShareSlurmClient for testing
type MockFairShareSlurmClient struct {
	mock.Mock
}

func (m *MockFairShareSlurmClient) Jobs() slurm.JobManager {
	args := m.Called()
	return args.Get(0).(slurm.JobManager)
}

func (m *MockFairShareSlurmClient) Nodes() slurm.NodeManager {
	args := m.Called()
	return args.Get(0).(slurm.NodeManager)
}

func (m *MockFairShareSlurmClient) Partitions() slurm.PartitionManager {
	args := m.Called()
	return args.Get(0).(slurm.PartitionManager)
}

func (m *MockFairShareSlurmClient) Info() slurm.InfoManager {
	args := m.Called()
	return args.Get(0).(slurm.InfoManager)
}

// MockFairShareJobManager for testing
type MockFairShareJobManager struct {
	mock.Mock
}

func (m *MockFairShareJobManager) List(ctx context.Context, opts ...slurm.JobListOption) (*slurm.JobList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*slurm.JobList), args.Error(1)
}

func (m *MockFairShareJobManager) Get(ctx context.Context, jobID string, opts ...slurm.JobGetOption) (*slurm.Job, error) {
	args := m.Called(ctx, jobID, opts)
	return args.Get(0).(*slurm.Job), args.Error(1)
}

func (m *MockFairShareJobManager) Submit(ctx context.Context, job *slurm.JobSubmissionRequest) (*slurm.JobSubmissionResponse, error) {
	args := m.Called(ctx, job)
	return args.Get(0).(*slurm.JobSubmissionResponse), args.Error(1)
}

func (m *MockFairShareJobManager) Cancel(ctx context.Context, jobID string, opts ...slurm.JobCancelOption) error {
	args := m.Called(ctx, jobID, opts)
	return args.Error(0)
}

func TestNewFairShareCollector(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}
	config := &FairShareConfig{
		CollectionInterval:       30 * time.Second,
		EnableUserFairShare:     true,
		EnableAccountHierarchy:  true,
		EnablePriorityAnalysis:  true,
		EnableViolationDetection: true,
		EnableTrendAnalysis:     true,
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)

	require.NoError(t, err)
	assert.NotNil(t, collector)
	assert.Equal(t, mockClient, collector.slurmClient)
	assert.Equal(t, config, collector.config)
	assert.NotNil(t, collector.metrics)
	assert.NotNil(t, collector.userFairShares)
	assert.NotNil(t, collector.priorityFactors)
	assert.NotNil(t, collector.violationDetector)
	assert.NotNil(t, collector.policyAnalyzer)
	assert.NotNil(t, collector.trendAnalyzer)
	assert.NotNil(t, collector.queueAnalyzer)
	assert.NotNil(t, collector.behaviorAnalyzer)
}

func TestNewFairShareCollector_DefaultConfig(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}

	collector, err := NewFairShareCollector(mockClient, nil, nil)

	require.NoError(t, err)
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.config)
	assert.Equal(t, 30*time.Second, collector.config.CollectionInterval)
	assert.Equal(t, 24*time.Hour, collector.config.FairShareRetention)
	assert.Equal(t, 6*time.Hour, collector.config.PriorityRetention)
	assert.True(t, collector.config.EnableUserFairShare)
	assert.True(t, collector.config.EnableAccountHierarchy)
	assert.True(t, collector.config.EnablePriorityAnalysis)
	assert.True(t, collector.config.EnableViolationDetection)
	assert.True(t, collector.config.EnableTrendAnalysis)
	assert.Equal(t, 0.2, collector.config.ViolationThreshold)
	assert.Equal(t, 24*time.Hour, collector.config.DecayPeriod)
	assert.Equal(t, 7*24*time.Hour, collector.config.ResetCycle)
}

func TestFairShareCollector_PriorityWeights(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}

	collector, err := NewFairShareCollector(mockClient, nil, nil)
	require.NoError(t, err)

	weights := collector.config.PriorityWeights
	assert.Equal(t, 0.2, weights.AgeWeight)
	assert.Equal(t, 0.5, weights.FairShareWeight)
	assert.Equal(t, 0.2, weights.QoSWeight)
	assert.Equal(t, 0.05, weights.PartitionWeight)
	assert.Equal(t, 0.03, weights.AssocWeight)
	assert.Equal(t, 0.02, weights.JobSizeWeight)
	
	// Test that weights sum to approximately 1.0
	total := weights.AgeWeight + weights.FairShareWeight + weights.QoSWeight + 
		weights.PartitionWeight + weights.AssocWeight + weights.JobSizeWeight
	assert.InDelta(t, 1.0, total, 0.01)
}

func TestFairShareCollector_Describe(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}
	config := &FairShareConfig{
		CollectionInterval: 30 * time.Second,
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 100)
	collector.Describe(ch)
	close(ch)

	// Count the number of metrics described
	count := 0
	for range ch {
		count++
	}

	// Should have multiple metrics described
	assert.Greater(t, count, 30)
}

func TestFairShareCollector_Collect(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}
	mockJobManager := &MockFairShareJobManager{}

	// Create test data
	now := time.Now()
	endTime := now.Add(-30 * time.Minute)
	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "COMPLETED",
				UserName:  "user1",
				Account:   "account1",
				Partition: "partition1",
				StartTime: now.Add(-time.Hour),
				EndTime:   &endTime,
				SubmitTime: now.Add(-65 * time.Minute),
				CPUs:      8,
				Memory:    "16GB",
			},
			{
				JobID:     "12346",
				JobState:  "PENDING",
				UserName:  "user2",
				Account:   "account2",
				Partition: "partition1",
				SubmitTime: now.Add(-30 * time.Minute),
				CPUs:      4,
				Memory:    "8GB",
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &FairShareConfig{
		CollectionInterval:      30 * time.Second,
		MaxUsersPerCollection:   100,
		EnableUserFairShare:     true,
		EnablePriorityAnalysis:  true,
		EnableViolationDetection: true,
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	require.NoError(t, err)

	// Create a registry and register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Collect metrics
	ch := make(chan prometheus.Metric, 500)
	collector.Collect(ch)
	close(ch)

	// Count collected metrics
	metricCount := 0
	for range ch {
		metricCount++
	}

	// Should collect multiple metrics
	assert.Greater(t, metricCount, 0)

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestFairShareCollector_collectUserFairShares(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}
	mockJobManager := &MockFairShareJobManager{}

	now := time.Now()
	endTime := now.Add(-30 * time.Minute)
	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "COMPLETED",
				UserName:  "testuser",
				Account:   "testaccount",
				Partition: "partition1",
				StartTime: now.Add(-time.Hour),
				EndTime:   &endTime,
				CPUs:      8,
				Memory:    "16GB",
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &FairShareConfig{
		CollectionInterval:    30 * time.Second,
		MaxUsersPerCollection: 100,
		ViolationThreshold:    0.2,
		PriorityWeights: PriorityWeights{
			FairShareWeight: 0.5,
		},
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	require.NoError(t, err)

	ctx := context.Background()
	err = collector.collectUserFairShares(ctx)

	assert.NoError(t, err)
	assert.Len(t, collector.userFairShares, 1)

	userKey := "testuser_testaccount"
	fairShare, exists := collector.userFairShares[userKey]
	require.True(t, exists)
	
	assert.Equal(t, "testuser", fairShare.UserName)
	assert.Equal(t, "testaccount", fairShare.Account)
	assert.Greater(t, fairShare.FairShareFactor, 0.0)
	assert.Greater(t, fairShare.RawShares, int64(0))
	assert.Greater(t, fairShare.NormalizedShares, 0.0)
	assert.GreaterOrEqual(t, fairShare.EffectiveUsage, 0.0)
	assert.Greater(t, fairShare.CPUUsage, 0.0)
	assert.Greater(t, fairShare.TargetUsage, 0.0)
	assert.Greater(t, fairShare.UsageRatio, 0.0)
	assert.Contains(t, []string{"improving", "degrading", "stable"}, fairShare.TrendDirection)
	assert.GreaterOrEqual(t, fairShare.DataQuality, 0.0)
	assert.LessOrEqual(t, fairShare.DataQuality, 1.0)

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestFairShareCollector_calculateUserFairShare(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}
	config := &FairShareConfig{
		ViolationThreshold: 0.2,
		DecayPeriod:       24 * time.Hour,
		PriorityWeights: PriorityWeights{
			FairShareWeight: 0.5,
		},
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	require.NoError(t, err)

	usage := &UserUsageData{
		UserName:      "testuser",
		Account:       "testaccount",
		TotalCPUUsage: 100.0, // 100 CPU-hours
		JobCount:      5,
	}

	fairShare := collector.calculateUserFairShare(usage)

	assert.NotNil(t, fairShare)
	assert.Equal(t, "testuser", fairShare.UserName)
	assert.Equal(t, "testaccount", fairShare.Account)
	assert.Greater(t, fairShare.FairShareFactor, 0.0)
	assert.Equal(t, int64(100), fairShare.RawShares)
	assert.Equal(t, 0.1, fairShare.NormalizedShares)
	assert.Equal(t, usage.TotalCPUUsage/168.0, fairShare.EffectiveUsage) // Weekly average
	assert.Equal(t, usage.TotalCPUUsage, fairShare.CPUUsage)
	assert.Equal(t, 100.0, fairShare.TargetUsage) // 0.1 * 1000.0
	assert.Greater(t, fairShare.UsageRatio, 0.0)
	assert.Greater(t, fairShare.FairSharePriority, 0.0)
	assert.Equal(t, "root.testaccount", fairShare.AccountPath)
	assert.Equal(t, "root", fairShare.ParentAccount)
	assert.Equal(t, 1, fairShare.Level)
	assert.Equal(t, 0.85, fairShare.DataQuality)
	assert.Greater(t, fairShare.TrendStrength, 0.0)
	assert.Greater(t, fairShare.PredictedFactor, 0.0)
}

func TestFairShareCollector_calculateJobPriority(t *testing.T) {
	mockClient := &MockFairShareSlurmClient{}
	config := &FairShareConfig{
		PriorityWeights: PriorityWeights{
			AgeWeight:       0.2,
			FairShareWeight: 0.5,
			QoSWeight:       0.1,
			PartitionWeight: 0.05,
		},
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	require.NoError(t, err)

	now := time.Now()
	job := &slurm.Job{
		JobID:     "12345",
		UserName:  "testuser",
		Account:   "testaccount",
		Partition: "partition1",
		SubmitTime: now.Add(-2 * time.Hour), // 2 hours old
		CPUs:      8,
	}

	priority := collector.calculateJobPriority(job)

	assert.NotNil(t, priority)
	assert.Equal(t, "12345", priority.JobID)
	assert.Equal(t, "testuser", priority.UserName)
	assert.Equal(t, "testaccount", priority.Account)
	assert.Equal(t, "partition1", priority.Partition)
	assert.Greater(t, priority.TotalPriority, int64(0))
	assert.GreaterOrEqual(t, priority.AgeFactor, 0.0)
	assert.LessOrEqual(t, priority.AgeFactor, 0.2) // Max age weight
	assert.Equal(t, 0.5, priority.FairShareFactor)
	assert.Equal(t, 0.1, priority.QoSFactor)
	assert.Equal(t, 0.05, priority.PartitionFactor)
	assert.Greater(t, priority.NormalizedAge, 0.0)
	assert.Equal(t, 1.0, priority.NormalizedFairShare) // 0.5 / 0.5
	assert.Equal(t, 1.0, priority.NormalizedQoS)       // 0.1 / 0.1
	assert.Equal(t, 1.0, priority.NormalizedPartition) // 0.05 / 0.05
	assert.Greater(t, priority.PriorityScore, 0.0)
	assert.LessOrEqual(t, priority.PriorityScore, 1.0)
	assert.Equal(t, 1, priority.PriorityRank)
	assert.Greater(t, priority.EstimatedWaitTime, time.Duration(0))
	assert.Equal(t, 8, priority.JobSize)
	assert.Equal(t, job.SubmitTime, priority.SubmitTime)
	assert.Equal(t, 0.9, priority.CalculationQuality)
}

func TestUserFairShare_DataStructure(t *testing.T) {
	now := time.Now()
	fairShare := &UserFairShare{
		UserName:          "testuser",
		Account:           "testaccount",
		Partition:         "partition1",
		Timestamp:         now,
		FairShareFactor:   0.8,
		RawShares:         200,
		NormalizedShares:  0.15,
		EffectiveUsage:    150.0,
		CPUUsage:          120.0,
		MemoryUsage:       80.0,
		GPUUsage:          10.0,
		NodeUsage:         5.0,
		TargetUsage:       100.0,
		UsageRatio:        1.5,
		FairSharePriority: 4000.0,
		LastDecay:         now.Add(-24 * time.Hour),
		LastReset:         now.Add(-7 * 24 * time.Hour),
		DecayHalfLife:     24 * time.Hour,
		AccountPath:       "root.testaccount",
		ParentAccount:     "root",
		Level:             1,
		DataQuality:       0.95,
		LastUpdated:       now,
		UpdateCount:       5,
		TrendDirection:    "improving",
		TrendStrength:     0.8,
		PredictedFactor:   0.85,
		IsViolating:       true,
		ViolationSeverity: 0.3,
		ViolationDuration: 2 * time.Hour,
	}

	assert.Equal(t, "testuser", fairShare.UserName)
	assert.Equal(t, "testaccount", fairShare.Account)
	assert.Equal(t, "partition1", fairShare.Partition)
	assert.Equal(t, 0.8, fairShare.FairShareFactor)
	assert.Equal(t, int64(200), fairShare.RawShares)
	assert.Equal(t, 0.15, fairShare.NormalizedShares)
	assert.Equal(t, 150.0, fairShare.EffectiveUsage)
	assert.Equal(t, 1.5, fairShare.UsageRatio)
	assert.Equal(t, 4000.0, fairShare.FairSharePriority)
	assert.Equal(t, "root.testaccount", fairShare.AccountPath)
	assert.Equal(t, 1, fairShare.Level)
	assert.Equal(t, "improving", fairShare.TrendDirection)
	assert.True(t, fairShare.IsViolating)
	assert.Equal(t, 0.3, fairShare.ViolationSeverity)
}

func TestJobPriorityFactors_DataStructure(t *testing.T) {
	now := time.Now()
	submitTime := now.Add(-2 * time.Hour)
	predictedStart := now.Add(30 * time.Minute)
	
	priority := &JobPriorityFactors{
		JobID:               "12345",
		UserName:            "testuser",
		Account:             "testaccount",
		Partition:           "partition1",
		QoS:                 "normal",
		Timestamp:           now,
		TotalPriority:       15000,
		BasePriority:        1000,
		AgeFactor:           0.15,
		FairShareFactor:     0.45,
		QoSFactor:           0.08,
		PartitionFactor:     0.04,
		AssocFactor:         0.025,
		JobSizeFactor:       0.015,
		NormalizedAge:       0.75,
		NormalizedFairShare: 0.9,
		NormalizedQoS:       0.8,
		NormalizedPartition: 0.8,
		NormalizedAssoc:     0.83,
		NormalizedJobSize:   0.75,
		PriorityScore:       0.75,
		PriorityRank:        25,
		PriorityPercentile:  80.0,
		EstimatedWaitTime:   45 * time.Minute,
		QueuePosition:       15,
		PredictedStartTime:  predictedStart,
		JobSize:             16,
		SubmitTime:          submitTime,
		RequestedRuntime:    4 * time.Hour,
		RequestedResources:  map[string]float64{"cpu": 16, "memory": 32},
		CalculationQuality:  0.92,
		DataFreshness:       2 * time.Minute,
	}

	assert.Equal(t, "12345", priority.JobID)
	assert.Equal(t, "testuser", priority.UserName)
	assert.Equal(t, int64(15000), priority.TotalPriority)
	assert.Equal(t, 0.15, priority.AgeFactor)
	assert.Equal(t, 0.45, priority.FairShareFactor)
	assert.Equal(t, 0.75, priority.PriorityScore)
	assert.Equal(t, 25, priority.PriorityRank)
	assert.Equal(t, 45*time.Minute, priority.EstimatedWaitTime)
	assert.Equal(t, 16, priority.JobSize)
	assert.Equal(t, submitTime, priority.SubmitTime)
	assert.Equal(t, 0.92, priority.CalculationQuality)
}

func TestFairShareViolationDetector_analyzeViolations(t *testing.T) {
	config := &FairShareConfig{
		ViolationThreshold: 0.2,
	}
	
	detector := &FairShareViolationDetector{
		config:     config,
		violations: make(map[string]*FairShareViolation),
		thresholds: &ViolationThresholds{
			MinorThreshold:    0.1,
			ModerateThreshold: 0.2,
			MajorThreshold:    0.4,
			CriticalThreshold: 0.6,
		},
	}

	userFairShares := map[string]*UserFairShare{
		"user1_account1": {
			UserName:          "user1",
			Account:           "account1",
			FairShareFactor:   0.5,
			UsageRatio:        1.5, // Violating (>1.2)
			IsViolating:       true,
			ViolationSeverity: 0.25,
			ViolationDuration: 2 * time.Hour,
		},
		"user2_account2": {
			UserName:        "user2",
			Account:         "account2",
			FairShareFactor: 1.2,
			UsageRatio:      0.8, // Not violating
			IsViolating:     false,
		},
	}

	detector.analyzeViolations(userFairShares)

	// Should detect one violation
	assert.Len(t, detector.violations, 1)
	
	// Find the violation
	var violation *FairShareViolation
	for _, v := range detector.violations {
		violation = v
		break
	}
	
	require.NotNil(t, violation)
	assert.Equal(t, "user1", violation.UserName)
	assert.Equal(t, "account1", violation.Account)
	assert.Equal(t, "overuse", violation.ViolationType)
	assert.Equal(t, "moderate", violation.Severity)
	assert.Equal(t, 0.5, violation.CurrentFactor)
	assert.Equal(t, 1.0, violation.TargetFactor)
	assert.Greater(t, violation.Deviation, 0.0)
	assert.Greater(t, violation.ClusterImpact, 0.0)
	assert.False(t, violation.IsResolved)
}

func TestFairSharePolicyAnalyzer_analyzePolicyEffectiveness(t *testing.T) {
	analyzer := &FairSharePolicyAnalyzer{
		policyMetrics: &PolicyEffectivenessMetrics{},
	}

	userFairShares := map[string]*UserFairShare{
		"user1_account1": {
			UserName:        "user1",
			FairShareFactor: 0.8,
			UsageRatio:      1.1,
			IsViolating:     false,
		},
		"user2_account2": {
			UserName:        "user2",
			FairShareFactor: 1.2,
			UsageRatio:      0.9,
			IsViolating:     false,
		},
		"user3_account3": {
			UserName:        "user3",
			FairShareFactor: 0.5,
			UsageRatio:      1.6, // Violating
			IsViolating:     true,
		},
	}

	analyzer.analyzePolicyEffectiveness(userFairShares)

	metrics := analyzer.policyMetrics
	assert.Greater(t, metrics.OverallScore, 0.0)
	assert.LessOrEqual(t, metrics.OverallScore, 1.0)
	assert.Greater(t, metrics.BalanceScore, 0.0)
	assert.LessOrEqual(t, metrics.BalanceScore, 1.0)
	assert.Greater(t, metrics.FairnessScore, 0.0)
	assert.LessOrEqual(t, metrics.FairnessScore, 1.0)
	assert.Greater(t, metrics.EfficiencyScore, 0.0)
	assert.LessOrEqual(t, metrics.EfficiencyScore, 1.0)
	assert.InDelta(t, 1.0/3.0, metrics.ViolationRate, 0.01) // 1 out of 3 users violating
	assert.Equal(t, 0.3, metrics.GiniCoefficient)
}

func TestFairShareCollector_Integration(t *testing.T) {
	// This test verifies the complete workflow
	mockClient := &MockFairShareSlurmClient{}
	mockJobManager := &MockFairShareJobManager{}

	now := time.Now()
	endTime := now.Add(-30 * time.Minute)
	testJobs := &slurm.JobList{
		Jobs: []*slurm.Job{
			{
				JobID:     "12345",
				JobState:  "COMPLETED",
				UserName:  "user1",
				Account:   "account1",
				Partition: "partition1",
				StartTime: now.Add(-time.Hour),
				EndTime:   &endTime,
				SubmitTime: now.Add(-65 * time.Minute),
				CPUs:      8,
				Memory:    "16GB",
			},
			{
				JobID:     "12346",
				JobState:  "PENDING",
				UserName:  "user2",
				Account:   "account2",
				Partition: "partition1",
				SubmitTime: now.Add(-30 * time.Minute),
				CPUs:      4,
				Memory:    "8GB",
			},
		},
	}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &FairShareConfig{
		CollectionInterval:         30 * time.Second,
		FairShareRetention:        24 * time.Hour,
		EnableUserFairShare:       true,
		EnableAccountHierarchy:    true,
		EnablePriorityAnalysis:    true,
		EnableViolationDetection:  true,
		EnableTrendAnalysis:       true,
		EnableQueueAnalysis:       true,
		EnableBehaviorAnalysis:    true,
		MaxUsersPerCollection:     100,
		EnableParallelProcessing:  true,
		MaxConcurrentAnalyses:     3,
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	require.NoError(t, err)

	// Test metric collection
	registry := prometheus.NewRegistry()
	err = registry.Register(collector)
	require.NoError(t, err)

	// Collect metrics to verify no errors
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	assert.Greater(t, len(metricFamilies), 0)

	// Verify that data was collected
	assert.Greater(t, len(collector.userFairShares), 0)
	assert.Greater(t, len(collector.priorityFactors), 0)

	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestViolationThresholds_Validation(t *testing.T) {
	thresholds := &ViolationThresholds{
		MinorThreshold:    0.1,
		ModerateThreshold: 0.2,
		MajorThreshold:    0.4,
		CriticalThreshold: 0.6,
		DurationThreshold: time.Hour,
	}

	// Verify threshold ordering
	assert.Less(t, thresholds.MinorThreshold, thresholds.ModerateThreshold)
	assert.Less(t, thresholds.ModerateThreshold, thresholds.MajorThreshold)
	assert.Less(t, thresholds.MajorThreshold, thresholds.CriticalThreshold)
	assert.Greater(t, thresholds.DurationThreshold, time.Duration(0))
}

func TestUserBehaviorProfile_DataStructure(t *testing.T) {
	profile := &UserBehaviorProfile{
		UserName:           "testuser",
		Account:            "testaccount",
		AnalysisPeriod:     7 * 24 * time.Hour,
		LastUpdated:        time.Now(),
		BehaviorScore:      0.85,
		FairnessScore:      0.9,
		EfficiencyScore:    0.8,
		PredictabilityScore: 0.75,
		ResourceEfficiency:  0.82,
		TimeEfficiency:     0.88,
		QueueEfficiency:    0.9,
		OverallEfficiency:  0.85,
	}

	assert.Equal(t, "testuser", profile.UserName)
	assert.Equal(t, "testaccount", profile.Account)
	assert.Equal(t, 7*24*time.Hour, profile.AnalysisPeriod)
	assert.Equal(t, 0.85, profile.BehaviorScore)
	assert.Equal(t, 0.9, profile.FairnessScore)
	assert.Equal(t, 0.8, profile.EfficiencyScore)
	assert.Equal(t, 0.75, profile.PredictabilityScore)
	assert.Equal(t, 0.82, profile.ResourceEfficiency)
	assert.Equal(t, 0.88, profile.TimeEfficiency)
	assert.Equal(t, 0.9, profile.QueueEfficiency)
	assert.Equal(t, 0.85, profile.OverallEfficiency)
}

func TestQueueAnalysisData_DataStructure(t *testing.T) {
	queueData := &QueueAnalysisData{
		PartitionName:        "normal",
		Timestamp:           time.Now(),
		QueueLength:         25,
		TotalJobs:           100,
		RunningJobs:         40,
		PendingJobs:         25,
		AverageWaitTime:     45 * time.Minute,
		MedianWaitTime:      30 * time.Minute,
		MaxWaitTime:         2 * time.Hour,
		MinWaitTime:         5 * time.Minute,
		PriorityDistribution: map[string]int{
			"high":   5,
			"medium": 15,
			"low":    5,
		},
		HighPriorityJobs:     5,
		MediumPriorityJobs:   15,
		LowPriorityJobs:      5,
		TotalCPURequest:      800,
		TotalMemoryRequest:   1600,
		TotalGPURequest:      20,
		JobStartRate:         8.5,
		JobCompletionRate:    8.0,
		ThroughputTrend:      "stable",
		QueueEfficiency:      0.85,
		ResourceUtilization:  0.78,
		QueueBalance:         0.9,
	}

	assert.Equal(t, "normal", queueData.PartitionName)
	assert.Equal(t, 25, queueData.QueueLength)
	assert.Equal(t, 100, queueData.TotalJobs)
	assert.Equal(t, 40, queueData.RunningJobs)
	assert.Equal(t, 25, queueData.PendingJobs)
	assert.Equal(t, 45*time.Minute, queueData.AverageWaitTime)
	assert.Equal(t, 5, queueData.PriorityDistribution["high"])
	assert.Equal(t, int64(800), queueData.TotalCPURequest)
	assert.Equal(t, 8.5, queueData.JobStartRate)
	assert.Equal(t, "stable", queueData.ThroughputTrend)
	assert.Equal(t, 0.85, queueData.QueueEfficiency)
}

// Benchmark tests
func BenchmarkFairShareCollector_Collect(b *testing.B) {
	mockClient := &MockFairShareSlurmClient{}
	mockJobManager := &MockFairShareJobManager{}

	// Create large dataset for benchmarking
	jobs := make([]*slurm.Job, 100)
	now := time.Now()
	for i := 0; i < 100; i++ {
		endTime := now.Add(-30 * time.Minute)
		jobs[i] = &slurm.Job{
			JobID:     fmt.Sprintf("job%d", i),
			JobState:  "COMPLETED",
			UserName:  fmt.Sprintf("user%d", i%20),
			Account:   fmt.Sprintf("account%d", i%10),
			Partition: "partition1",
			StartTime: now.Add(-time.Hour),
			EndTime:   &endTime,
			SubmitTime: now.Add(-65 * time.Minute),
			CPUs:      8,
			Memory:    "16GB",
		}
	}

	testJobs := &slurm.JobList{Jobs: jobs}

	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(testJobs, nil)

	config := &FairShareConfig{
		CollectionInterval:    30 * time.Second,
		MaxUsersPerCollection: 1000,
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan prometheus.Metric, 1000)
		collector.Collect(ch)
		close(ch)
		// Drain the channel
		for range ch {
		}
	}
}

func BenchmarkFairShareCollector_calculateUserFairShare(b *testing.B) {
	mockClient := &MockFairShareSlurmClient{}
	config := &FairShareConfig{
		ViolationThreshold: 0.2,
		DecayPeriod:       24 * time.Hour,
		PriorityWeights: PriorityWeights{
			FairShareWeight: 0.5,
		},
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	if err != nil {
		b.Fatal(err)
	}

	usage := &UserUsageData{
		UserName:      "testuser",
		Account:       "testaccount",
		TotalCPUUsage: 100.0,
		JobCount:      5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.calculateUserFairShare(usage)
	}
}

func BenchmarkFairShareCollector_calculateJobPriority(b *testing.B) {
	mockClient := &MockFairShareSlurmClient{}
	config := &FairShareConfig{
		PriorityWeights: PriorityWeights{
			AgeWeight:       0.2,
			FairShareWeight: 0.5,
			QoSWeight:       0.1,
			PartitionWeight: 0.05,
		},
	}

	collector, err := NewFairShareCollector(mockClient, nil, config)
	if err != nil {
		b.Fatal(err)
	}

	now := time.Now()
	job := &slurm.Job{
		JobID:     "12345",
		UserName:  "testuser",
		Account:   "testaccount",
		Partition: "partition1",
		SubmitTime: now.Add(-2 * time.Hour),
		CPUs:      8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.calculateJobPriority(job)
	}
}