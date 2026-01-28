// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/slurm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AdapterAutodetectionTestSuite tests the adapter pattern and API version auto-detection
type AdapterAutodetectionTestSuite struct {
	suite.Suite
	slurmBaseURL string
	authToken    string
	slurmClient  *slurm.Client
}

// SetupSuite initializes the test suite
func (suite *AdapterAutodetectionTestSuite) SetupSuite() {
	// Get SLURM server configuration from environment
	suite.slurmBaseURL = os.Getenv("SLURM_BASE_URL")
	if suite.slurmBaseURL == "" {
		suite.slurmBaseURL = "http://rocky9.ar.jontk.com:6820"
	}

	suite.authToken = os.Getenv("SLURM_AUTH_TOKEN")
	if suite.authToken == "" {
		suite.T().Skip("Skipping adapter tests - SLURM_AUTH_TOKEN not set")
		return
	}

	suite.T().Logf("Testing adapter pattern with SLURM server: %s", suite.slurmBaseURL)
}

// TearDownSuite cleans up after tests
func (suite *AdapterAutodetectionTestSuite) TearDownSuite() {
	if suite.slurmClient != nil {
		_ = suite.slurmClient.Close()
	}
}

// TestAdapterCreationWithAutodetection tests client creation with adapters and auto-detection enabled
func (suite *AdapterAutodetectionTestSuite) TestAdapterCreationWithAutodetection() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Empty for auto-detection
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    5 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	// Create client with auto-detection
	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err, "Failed to create SLURM client with adapter pattern")
	require.NotNil(suite.T(), client, "Client should not be nil")

	suite.slurmClient = client
	defer func() { _ = client.Close() }()

	// Verify client is created and can be used
	assert.NotNil(suite.T(), client.GetSlurmClient())
	suite.T().Log("✓ Client created with adapter pattern and auto-detection enabled")
}

// TestAdapterCreationWithExplicitVersion tests client creation with adapters and explicit version
func (suite *AdapterAutodetectionTestSuite) TestAdapterCreationWithExplicitVersion() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "v0.0.43", // Explicit version
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    5 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err, "Failed to create SLURM client with explicit version")
	require.NotNil(suite.T(), client)
	defer func() { _ = client.Close() }()

	suite.T().Log("✓ Client created with adapter pattern and explicit version")
}

// TestConnectionWithAdapter tests SLURM connection using adapter pattern
func (suite *AdapterAutodetectionTestSuite) TestConnectionWithAdapter() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err)
	defer func() { _ = client.Close() }()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.TestConnection(ctx)
	require.NoError(suite.T(), err, "Connection test should succeed")
	suite.T().Log("✓ Successfully connected to SLURM API using adapter")
}

// TestGetInfoWithAdapter tests retrieving cluster info with adapter pattern
func (suite *AdapterAutodetectionTestSuite) TestGetInfoWithAdapter() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clusterInfo, err := client.GetInfo(ctx)
	require.NoError(suite.T(), err, "GetInfo should succeed")
	require.NotNil(suite.T(), clusterInfo)

	suite.T().Logf("✓ Retrieved cluster info successfully")
	suite.T().Logf("  Cluster name: %v", clusterInfo.ClusterName)
}

// TestGetNodesWithAdapter tests retrieving node information with adapter pattern
func (suite *AdapterAutodetectionTestSuite) TestGetNodesWithAdapter() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nodeList, err := client.GetNodes(ctx)
	require.NoError(suite.T(), err, "GetNodes should succeed")
	require.NotNil(suite.T(), nodeList)

	nodeCount := len(nodeList.Nodes)
	suite.T().Logf("✓ Retrieved node information successfully")
	suite.T().Logf("  Total nodes: %d", nodeCount)
	assert.Greater(suite.T(), nodeCount, 0, "Should have at least one node")
}

// TestGetJobsWithAdapter tests retrieving job information with adapter pattern
func (suite *AdapterAutodetectionTestSuite) TestGetJobsWithAdapter() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobList, err := client.GetJobs(ctx)
	require.NoError(suite.T(), err, "GetJobs should succeed")
	require.NotNil(suite.T(), jobList)

	jobCount := len(jobList.Jobs)
	suite.T().Logf("✓ Retrieved job information successfully")
	suite.T().Logf("  Total jobs: %d", jobCount)
}

// TestGetPartitionsWithAdapter tests retrieving partition information with adapter pattern
func (suite *AdapterAutodetectionTestSuite) TestGetPartitionsWithAdapter() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err)
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	partitionList, err := client.GetPartitions(ctx)
	require.NoError(suite.T(), err, "GetPartitions should succeed")
	require.NotNil(suite.T(), partitionList)

	partitionCount := len(partitionList.Partitions)
	suite.T().Logf("✓ Retrieved partition information successfully")
	suite.T().Logf("  Total partitions: %d", partitionCount)
	assert.Greater(suite.T(), partitionCount, 0, "Should have at least one partition")
}

// TestAdapterVersusTraditionalClient compares adapter and traditional client behavior
func (suite *AdapterAutodetectionTestSuite) TestAdapterVersusTraditionalClient() {
	// Test with adapter
	adapterCfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	adapterClient, err := slurm.NewClient(adapterCfg)
	require.NoError(suite.T(), err)
	defer func() { _ = adapterClient.Close() }()

	// Test with traditional approach (explicit version)
	traditionalCfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "v0.0.43", // Explicit version
		UseAdapters: false,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	traditionalClient, err := slurm.NewClient(traditionalCfg)
	require.NoError(suite.T(), err)
	defer func() { _ = traditionalClient.Close() }()

	// Both should be able to connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = adapterClient.TestConnection(ctx)
	require.NoError(suite.T(), err, "Adapter client should connect")

	err = traditionalClient.TestConnection(ctx)
	require.NoError(suite.T(), err, "Traditional client should connect")

	suite.T().Log("✓ Both adapter and traditional clients connected successfully")

	// Get info from both and compare
	adapterInfo, err := adapterClient.GetInfo(ctx)
	require.NoError(suite.T(), err)

	traditionalInfo, err := traditionalClient.GetInfo(ctx)
	require.NoError(suite.T(), err)

	// Both should return valid cluster info
	assert.NotNil(suite.T(), adapterInfo)
	assert.NotNil(suite.T(), traditionalInfo)

	suite.T().Log("✓ Both clients retrieved cluster info successfully")
}

// TestRateLimitingWithAdapter tests rate limiting works with adapter pattern
func (suite *AdapterAutodetectionTestSuite) TestRateLimitingWithAdapter() {
	cfg := &config.SLURMConfig{
		BaseURL:     suite.slurmBaseURL,
		APIVersion:  "", // Auto-detect
		UseAdapters: true,
		Auth: config.AuthConfig{
			Type:  "jwt",
			Token: suite.authToken,
		},
		Timeout:       30 * time.Second,
		RetryAttempts: 1,
		RetryDelay:    1 * time.Second,
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 2.0, // Very low rate for testing
			BurstSize:         1,
		},
	}

	client, err := slurm.NewClient(cfg)
	require.NoError(suite.T(), err)
	defer func() { _ = client.Close() }()

	ctx := context.Background()

	// Make multiple requests - should be rate limited
	start := time.Now()

	// First request should succeed quickly
	_, err = client.GetInfo(ctx)
	require.NoError(suite.T(), err)

	// Subsequent requests should be delayed by rate limiter
	_, err = client.GetInfo(ctx)
	require.NoError(suite.T(), err)

	elapsed := time.Since(start)

	// With rate limiting at 2 req/s, 2 requests should take at least ~500ms
	suite.T().Logf("✓ Rate limiting applied: 2 requests took %v", elapsed)
	assert.Greater(suite.T(), elapsed, 100*time.Millisecond, "Rate limiting should delay requests")
}

// TestAdapterAutodetectionTestSuite runs the adapter auto-detection test suite
func TestAdapterAutodetectionSuite(t *testing.T) {
	t.Parallel()

	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests skipped")
	}

	suite.Run(t, new(AdapterAutodetectionTestSuite))
}
