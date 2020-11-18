/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package cpus

import (
	"strconv"

	"github.com/fromanirh/topologyinfo/sysfs"
)

// CPUIdList is a list of CPU IDs (integer core identifier)
type CPUIdList []int

// CPUs reports the information about all the CPU found in the system
type CPUs struct {
	Present     CPUIdList         // CPU IDs of all CPUs in the system
	Online      CPUIdList         // CPU IDs of all online CPUs - subset of Present
	CoreCPUs    map[int]CPUIdList // aka thread_siblings
	PackageCPUs map[int]CPUIdList // aka core_siblings
	// you can have more than a package per physical socket (don't assume 1:1)
	Packages CPUIdList // list of package IDs in this system
	// you can have more than a NUMA node per package (don't assume 1:1:)
	NUMANodes    CPUIdList         // list of *online* NUMA nodes in this system
	NUMANodeCPUs map[int]CPUIdList // NUMA node ID -> CPU IDs on this NUMA node
	// internal helpers
	cpuID2NUMANode map[int]int // CPU ID -> NUMA node ID (reverse of NUMANodeCPUs)
}

func (c CPUs) GetNodeIDForCPU(cpuID int) (int, bool) {
	n, ok := c.cpuID2NUMANode[cpuID]
	return n, ok
}

// NewCPUs extracts the CPU information from a given sysfs-like path
func NewCPUs(sysfsPath string) (*CPUs, error) {
	sys := sysfs.New(sysfsPath)
	sysCpu := sys.Join(sysfs.PathDevsSysCPU)

	present, err := sysCpu.ReadList("present")
	if err != nil {
		return nil, err
	}
	online, err := sysCpu.ReadList("online")
	if err != nil {
		return nil, err
	}

	nodes, err := sys.Join(sysfs.PathDevsSysNode).ReadList("online")
	if err != nil {
		return nil, err
	}

	var packageIds CPUIdList
	packages := make(map[string]CPUIdList)
	coreCPUs := make(map[int]CPUIdList)
	packageCPUs := make(map[int]CPUIdList)
	for _, cpuID := range online {
		sysCpuIDTopo := sys.ForCPU(cpuID).Join("topology")

		cpuThreads, err := sysCpuIDTopo.ReadList("thread_siblings_list")
		if err != nil {
			return nil, err
		}
		cpuCores, err := sysCpuIDTopo.ReadList("core_siblings_list")
		if err != nil {
			return nil, err
		}
		physPackageID, err := sysCpuIDTopo.ReadFile("physical_package_id")
		if err != nil {
			return nil, err
		}

		pkgId, err := strconv.Atoi(physPackageID)
		if err != nil {
			return nil, err
		}
		packageIds = append(packageIds, pkgId)

		coreCPUs[cpuID] = cpuThreads
		packageCPUs[cpuID] = cpuCores

		cpusPerPhysPkg := packages[physPackageID]
		cpusPerPhysPkg = append(cpusPerPhysPkg, cpuID)
		packages[physPackageID] = cpusPerPhysPkg
	}

	cpuID2NUMANode := make(map[int]int)
	numaNodeCPUs := make(map[int]CPUIdList)
	for _, node := range nodes {
		cpus, err := sys.ForNode(node).ReadList("cpulist")
		if err != nil {
			return nil, err
		}
		numaNodeCPUs[node] = cpus

		for _, cpu := range cpus {
			cpuID2NUMANode[cpu] = node
		}
	}

	return &CPUs{
		Present:      present,
		Online:       online,
		CoreCPUs:     coreCPUs,
		PackageCPUs:  packageCPUs,
		Packages:     packageIds,
		NUMANodes:    nodes,
		NUMANodeCPUs: numaNodeCPUs,
		// internal helpers
		cpuID2NUMANode: cpuID2NUMANode,
	}, nil
}
