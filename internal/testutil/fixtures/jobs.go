package fixtures

import (
	"fmt"
	"time"

	"github.com/jontk/slurm-client"
)

// GetTestJobs returns test job data
func GetTestJobs() []slurm.Job {
	now := time.Now()
	return []slurm.Job{
		{
			ID:          "12345",
			Name:        "test-job-1",
			State:       "RUNNING",
			UserID:      "user1",
			UserName:    "testuser1",
			Partition:   "compute",
			Account:     "research",
			CPUs:        16,
			Memory:      32768, // 32GB in MB
			Priority:    1000,
			SubmitTime:  now.Add(-2 * time.Hour),
			StartTime:   now.Add(-1 * time.Hour),
			TimeLimit:   24 * time.Hour,
			NodeList:    "node[01-02]",
			ArrayJobID:  0,
			ArrayTaskID: 0,
			QOS:         "normal",
		},
		{
			ID:          "12346",
			Name:        "test-job-2",
			State:       "PENDING",
			UserID:      "user2",
			UserName:    "testuser2",
			Partition:   "compute",
			Account:     "engineering",
			CPUs:        32,
			Memory:      65536, // 64GB in MB
			Priority:    500,
			SubmitTime:  now.Add(-30 * time.Minute),
			TimeLimit:   48 * time.Hour,
			ArrayJobID:  0,
			ArrayTaskID: 0,
			QOS:         "high",
			Reason:      "Resources",
		},
		{
			ID:          "12347",
			Name:        "test-job-3",
			State:       "COMPLETED",
			UserID:      "user1",
			UserName:    "testuser1",
			Partition:   "gpu",
			Account:     "research",
			CPUs:        8,
			Memory:      16384, // 16GB in MB
			GPUs:        2,
			Priority:    750,
			SubmitTime:  now.Add(-4 * time.Hour),
			StartTime:   now.Add(-3 * time.Hour),
			EndTime:     now.Add(-1 * time.Hour),
			TimeLimit:   4 * time.Hour,
			NodeList:    "gpu-node01",
			ArrayJobID:  0,
			ArrayTaskID: 0,
			QOS:         "normal",
			ExitCode:    0,
		},
		{
			ID:          "12348",
			Name:        "test-job-4",
			State:       "FAILED",
			UserID:      "user3",
			UserName:    "testuser3",
			Partition:   "compute",
			Account:     "finance",
			CPUs:        4,
			Memory:      8192, // 8GB in MB
			Priority:    250,
			SubmitTime:  now.Add(-2 * time.Hour),
			StartTime:   now.Add(-90 * time.Minute),
			EndTime:     now.Add(-30 * time.Minute),
			TimeLimit:   2 * time.Hour,
			NodeList:    "node03",
			ArrayJobID:  0,
			ArrayTaskID: 0,
			QOS:         "low",
			ExitCode:    1,
			Reason:      "NonZeroExitCode",
		},
		{
			ID:          "12349",
			Name:        "array-job",
			State:       "RUNNING",
			UserID:      "user2",
			UserName:    "testuser2",
			Partition:   "compute",
			Account:     "engineering",
			CPUs:        2,
			Memory:      4096, // 4GB in MB
			Priority:    600,
			SubmitTime:  now.Add(-1 * time.Hour),
			StartTime:   now.Add(-45 * time.Minute),
			TimeLimit:   1 * time.Hour,
			NodeList:    "node04",
			ArrayJobID:  12349,
			ArrayTaskID: 1,
			ArrayTaskStr: "1-10",
			QOS:         "normal",
		},
	}
}

// GetEmptyJobList returns an empty job list
func GetEmptyJobList() *slurm.JobList {
	return &slurm.JobList{
		Jobs: []slurm.Job{},
	}
}

// GetRunningJobList returns only running jobs
func GetRunningJobList() *slurm.JobList {
	jobs := GetTestJobs()
	runningJobs := []slurm.Job{}
	for _, job := range jobs {
		if job.State == "RUNNING" {
			runningJobs = append(runningJobs, job)
		}
	}
	return &slurm.JobList{
		Jobs: runningJobs,
	}
}

// GetTestJobList returns a complete test job list
func GetTestJobList() *slurm.JobList {
	return &slurm.JobList{
		Jobs: GetTestJobs(),
	}
}

// GenerateLargeJobList generates a large number of jobs for performance testing
func GenerateLargeJobList(count int) *slurm.JobList {
	jobs := make([]slurm.Job, count)

	states := []string{"PENDING", "RUNNING", "COMPLETED", "FAILED", "CANCELLED"}
	partitions := []string{"compute", "gpu", "bigmem", "debug", "interactive"}
	users := []string{"user1", "user2", "user3", "user4", "user5", "admin", "researcher"}

	for i := 0; i < count; i++ {
		jobID := fmt.Sprintf("%d", 100000+i)
		state := states[i%len(states)]
		partition := partitions[i%len(partitions)]
		user := users[i%len(users)]

		// Create realistic job with varying resource requirements
		cpus := 1 + (i%64)           // 1-64 CPUs
		memory := 1024 + (i%32)*1024 // 1GB-32GB memory
		runtime := time.Duration(i%3600) * time.Second // 0-1 hour runtime

		jobs[i] = slurm.Job{
			ID:         jobID,
			Name:       fmt.Sprintf("job_%d", i),
			State:      state,
			Partition:  partition,
			UserName:   user,
			UserID:     fmt.Sprintf("%d", 1000+i%100),
			Account:    "default",
			Priority:   uint32(1000 + (i % 1000)),
			CPUs:       uint32(cpus),
			Memory:     uint64(memory),
			TimeLimit:  24 * time.Hour,
			SubmitTime: time.Now().Add(-time.Duration(i) * time.Minute),
			StartTime:  time.Now().Add(-runtime),
			EndTime:    time.Now(),
			NodeList:   fmt.Sprintf("node%02d", (i%10)+1),
			QOS:        "normal",
			ExitCode:   0,
		}

		// Add some variety to failed jobs
		if state == "FAILED" {
			jobs[i].ExitCode = 1
		}

		// Add GPU requirements for GPU partition
		if partition == "gpu" {
			jobs[i].GPUs = uint32(1 + (i % 4))
		}

		// Add array job information for some jobs
		if i%10 == 0 {
			jobs[i].ArrayJobID = uint64(100000 + i)
			jobs[i].ArrayTaskID = uint64(i % 5)
		}
	}

	return &slurm.JobList{
		Jobs: jobs,
	}
}

// GenerateJobBatch generates a batch of jobs with specific characteristics
func GenerateJobBatch(count int, stateFilter string, partitionFilter string) *slurm.JobList {
	jobs := make([]slurm.Job, count)

	users := []string{"user1", "user2", "user3", "user4", "user5"}

	for i := 0; i < count; i++ {
		jobID := fmt.Sprintf("%d", 200000+i)
		user := users[i%len(users)]

		jobs[i] = slurm.Job{
			ID:         jobID,
			Name:       fmt.Sprintf("batch_job_%d", i),
			State:      stateFilter,
			Partition:  partitionFilter,
			UserName:   user,
			UserID:     fmt.Sprintf("%d", 2000+i%50),
			Account:    "batch",
			Priority:   uint32(500 + i),
			CPUs:       uint32(1 + (i % 32)),
			Memory:     uint64(2048 + (i%16)*1024),
			TimeLimit:  12 * time.Hour,
			SubmitTime: time.Now().Add(-time.Duration(i) * time.Second),
			StartTime:  time.Now().Add(-time.Duration(i) * time.Second),
			EndTime:    time.Now(),
			NodeList:   fmt.Sprintf("node%02d", (i%20)+1),
			QOS:        "normal",
			ExitCode:   0,
		}
	}

	return &slurm.JobList{
		Jobs: jobs,
	}
}