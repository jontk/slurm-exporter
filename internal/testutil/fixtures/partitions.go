package fixtures

import (
	slurm "github.com/jontk/slurm-client"
)

// GetTestPartitions returns a list of test partitions for testing
func GetTestPartitions() []slurm.Partition {
	return []slurm.Partition{
		{
			Name: strPtr("compute"),
		},
		{
			Name: strPtr("gpu"),
		},
		{
			Name: strPtr("highmem"),
		},
	}
}
