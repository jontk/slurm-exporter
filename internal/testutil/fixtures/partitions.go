package fixtures

import (
	"time"

	"github.com/jontk/slurm-client"
)

// GetTestPartitions returns test partition data
func GetTestPartitions() []slurm.Partition {
	return []slurm.Partition{
		{
			Name:              "compute",
			State:             "up",
			TotalNodes:        4,
			TotalCPUs:         256,
			DefaultTime:       24 * time.Hour,
			MaxTime:           7 * 24 * time.Hour,
			DefaultMemPerCPU:  2048, // 2GB per CPU
			MaxMemPerCPU:      8192, // 8GB per CPU
			Nodes:             "node[01-04]",
			AllowGroups:       []string{"all"},
			AllowAccounts:     []string{"all"},
			DenyAccounts:      []string{},
			QOS:               []string{"normal", "high", "low"},
			DefaultQOS:        "normal",
			Priority:          100,
			PreemptMode:       "off",
			GraceTime:         300, // 5 minutes
		},
		{
			Name:              "gpu",
			State:             "up",
			TotalNodes:        2,
			TotalCPUs:         64,
			DefaultTime:       12 * time.Hour,
			MaxTime:           3 * 24 * time.Hour,
			DefaultMemPerCPU:  4096, // 4GB per CPU
			MaxMemPerCPU:      16384, // 16GB per CPU
			Nodes:             "gpu-node[01-02]",
			AllowGroups:       []string{"gpu-users"},
			AllowAccounts:     []string{"research", "engineering"},
			DenyAccounts:      []string{"finance"},
			QOS:               []string{"normal", "high"},
			DefaultQOS:        "normal",
			Priority:          200,
			PreemptMode:       "suspend",
			GraceTime:         600, // 10 minutes
			Gres:              "gpu:8",
		},
		{
			Name:              "all",
			State:             "up",
			TotalNodes:        6,
			TotalCPUs:         320,
			DefaultTime:       1 * time.Hour,
			MaxTime:           24 * time.Hour,
			DefaultMemPerCPU:  1024, // 1GB per CPU
			MaxMemPerCPU:      4096, // 4GB per CPU
			Nodes:             "node[01-04],gpu-node[01-02]",
			AllowGroups:       []string{"all"},
			AllowAccounts:     []string{"all"},
			DenyAccounts:      []string{},
			QOS:               []string{"normal", "low"},
			DefaultQOS:        "low",
			Priority:          50,
			PreemptMode:       "off",
			GraceTime:         120, // 2 minutes
		},
		{
			Name:              "maintenance",
			State:             "down",
			TotalNodes:        0,
			TotalCPUs:         0,
			DefaultTime:       1 * time.Hour,
			MaxTime:           4 * time.Hour,
			DefaultMemPerCPU:  1024,
			MaxMemPerCPU:      2048,
			Nodes:             "",
			AllowGroups:       []string{"admin"},
			AllowAccounts:     []string{"system"},
			DenyAccounts:      []string{"all"},
			QOS:               []string{"maintenance"},
			DefaultQOS:        "maintenance",
			Priority:          1000,
			PreemptMode:       "off",
			GraceTime:         0,
		},
		{
			Name:              "bigmem",
			State:             "up",
			TotalNodes:        1,
			TotalCPUs:         128,
			DefaultTime:       48 * time.Hour,
			MaxTime:           14 * 24 * time.Hour,
			DefaultMemPerCPU:  16384, // 16GB per CPU
			MaxMemPerCPU:      32768, // 32GB per CPU
			Nodes:             "bigmem01",
			AllowGroups:       []string{"bigmem-users"},
			AllowAccounts:     []string{"research", "bioinformatics"},
			DenyAccounts:      []string{},
			QOS:               []string{"bigmem"},
			DefaultQOS:        "bigmem",
			Priority:          150,
			PreemptMode:       "off",
			GraceTime:         1800, // 30 minutes
		},
	}
}

// GetTestPartitionList returns a complete test partition list
func GetTestPartitionList() *slurm.PartitionList {
	return &slurm.PartitionList{
		Partitions: GetTestPartitions(),
	}
}

// GetActivePartitionList returns only active (up) partitions
func GetActivePartitionList() *slurm.PartitionList {
	partitions := GetTestPartitions()
	activePartitions := []slurm.Partition{}
	for _, partition := range partitions {
		if partition.State == "up" {
			activePartitions = append(activePartitions, partition)
		}
	}
	return &slurm.PartitionList{
		Partitions: activePartitions,
	}
}

// GetEmptyPartitionList returns an empty partition list
func GetEmptyPartitionList() *slurm.PartitionList {
	return &slurm.PartitionList{
		Partitions: []slurm.Partition{},
	}
}