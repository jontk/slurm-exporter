package fixtures

import (
	"fmt"

	"github.com/jontk/slurm-client"
)

// GetTestNodes returns test node data
func GetTestNodes() []slurm.Node {
	return []slurm.Node{
		{
			Name:           "node01",
			State:          "idle",
			CPUs:           64,
			AllocatedCPUs:  0,
			Memory:         131072, // 128GB in MB
			AllocatedMemory: 0,
			TmpDisk:        1000000, // 1TB in MB
			AllocatedTmpDisk: 0,
			Partitions:     []string{"compute", "all"},
			Features:       []string{"haswell", "ib"},
			Gres:           "",
			Reason:         "",
			ReasonTime:     nil,
			Architecture:   "x86_64",
			OS:             "Linux",
			KernelVersion:  "5.15.0",
		},
		{
			Name:           "node02",
			State:          "allocated",
			CPUs:           64,
			AllocatedCPUs:  64,
			Memory:         131072,
			AllocatedMemory: 65536, // 64GB allocated
			TmpDisk:        1000000,
			AllocatedTmpDisk: 500000,
			Partitions:     []string{"compute", "all"},
			Features:       []string{"haswell", "ib"},
			Gres:           "",
			Reason:         "",
			ReasonTime:     nil,
			Architecture:   "x86_64",
			OS:             "Linux",
			KernelVersion:  "5.15.0",
		},
		{
			Name:           "node03",
			State:          "down",
			CPUs:           64,
			AllocatedCPUs:  0,
			Memory:         131072,
			AllocatedMemory: 0,
			TmpDisk:        1000000,
			AllocatedTmpDisk: 0,
			Partitions:     []string{"compute", "all"},
			Features:       []string{"haswell", "ib"},
			Gres:           "",
			Reason:         "Hardware failure",
			ReasonTime:     nil,
			Architecture:   "x86_64",
			OS:             "Linux",
			KernelVersion:  "5.15.0",
		},
		{
			Name:           "node04",
			State:          "mixed",
			CPUs:           64,
			AllocatedCPUs:  32,
			Memory:         131072,
			AllocatedMemory: 32768, // 32GB allocated
			TmpDisk:        1000000,
			AllocatedTmpDisk: 250000,
			Partitions:     []string{"compute", "all"},
			Features:       []string{"haswell", "ib"},
			Gres:           "",
			Reason:         "",
			ReasonTime:     nil,
			Architecture:   "x86_64",
			OS:             "Linux",
			KernelVersion:  "5.15.0",
		},
		{
			Name:           "gpu-node01",
			State:          "idle",
			CPUs:           32,
			AllocatedCPUs:  0,
			Memory:         262144, // 256GB
			AllocatedMemory: 0,
			TmpDisk:        2000000, // 2TB
			AllocatedTmpDisk: 0,
			Partitions:     []string{"gpu", "all"},
			Features:       []string{"skylake", "ib", "gpu"},
			Gres:           "gpu:4",
			GresUsed:       "gpu:0",
			Reason:         "",
			ReasonTime:     nil,
			Architecture:   "x86_64",
			OS:             "Linux",
			KernelVersion:  "5.15.0",
		},
		{
			Name:           "gpu-node02",
			State:          "drain",
			CPUs:           32,
			AllocatedCPUs:  0,
			Memory:         262144,
			AllocatedMemory: 0,
			TmpDisk:        2000000,
			AllocatedTmpDisk: 0,
			Partitions:     []string{"gpu", "all"},
			Features:       []string{"skylake", "ib", "gpu"},
			Gres:           "gpu:4",
			GresUsed:       "gpu:0",
			Reason:         "Maintenance",
			ReasonTime:     nil,
			Architecture:   "x86_64",
			OS:             "Linux",
			KernelVersion:  "5.15.0",
		},
	}
}

// GetTestNodeList returns a complete test node list
func GetTestNodeList() *slurm.NodeList {
	return &slurm.NodeList{
		Nodes: GetTestNodes(),
	}
}

// GetIdleNodeList returns only idle nodes
func GetIdleNodeList() *slurm.NodeList {
	nodes := GetTestNodes()
	idleNodes := []slurm.Node{}
	for _, node := range nodes {
		if node.State == "idle" {
			idleNodes = append(idleNodes, node)
		}
	}
	return &slurm.NodeList{
		Nodes: idleNodes,
	}
}

// GetEmptyNodeList returns an empty node list
func GetEmptyNodeList() *slurm.NodeList {
	return &slurm.NodeList{
		Nodes: []slurm.Node{},
	}
}

// GenerateLargeNodeList generates a large number of nodes for performance testing
func GenerateLargeNodeList(count int) *slurm.NodeList {
	nodes := make([]slurm.Node, count)

	states := []string{"idle", "allocated", "mixed", "down", "drain", "draining"}
	partitionSets := [][]string{
		{"compute", "all"},
		{"gpu", "all"},
		{"bigmem", "all"},
		{"debug", "all"},
		{"interactive"},
	}

	for i := 0; i < count; i++ {
		nodeName := fmt.Sprintf("node%04d", i+1)
		state := states[i%len(states)]
		partitions := partitionSets[i%len(partitionSets)]

		// Create realistic node with varying specifications
		cpus := 16 + (i%9)*8        // 16, 24, 32, 40, 48, 56, 64, 72, 80 CPUs
		memory := 32768 + (i%16)*8192 // 32GB-128GB memory in MB
		tmpDisk := 500000 + (i%20)*100000 // 500GB-2.5TB temp disk

		// Calculate allocation based on state
		var allocatedCPUs uint32
		var allocatedMemory uint64
		var allocatedTmpDisk uint64

		switch state {
		case "allocated":
			allocatedCPUs = uint32(cpus)
			allocatedMemory = uint64(memory)
			allocatedTmpDisk = uint64(tmpDisk)
		case "mixed":
			allocatedCPUs = uint32(cpus / 2)
			allocatedMemory = uint64(memory / 2)
			allocatedTmpDisk = uint64(tmpDisk / 2)
		default:
			allocatedCPUs = 0
			allocatedMemory = 0
			allocatedTmpDisk = 0
		}

		node := slurm.Node{
			Name:             nodeName,
			State:            state,
			CPUs:             uint32(cpus),
			AllocatedCPUs:    allocatedCPUs,
			Memory:           uint64(memory),
			AllocatedMemory:  allocatedMemory,
			TmpDisk:          uint64(tmpDisk),
			AllocatedTmpDisk: allocatedTmpDisk,
			Partitions:       partitions,
			Features:         []string{"haswell", "ib"},
			Architecture:     "x86_64",
			OS:               "Linux",
			KernelVersion:    "5.15.0",
		}

		// Add GPU nodes for some nodes
		if i%10 == 0 {
			node.Gres = "gpu:4"
			node.GresUsed = fmt.Sprintf("gpu:%d", allocatedCPUs/8) // Assume 8 CPUs per GPU
			node.Features = append(node.Features, "gpu")
		}

		// Add reasons for down/drain nodes
		if state == "down" || state == "drain" || state == "draining" {
			reasons := []string{"Hardware failure", "Maintenance", "Power issue", "Network problem"}
			node.Reason = reasons[i%len(reasons)]
		}

		nodes[i] = node
	}

	return &slurm.NodeList{
		Nodes: nodes,
	}
}