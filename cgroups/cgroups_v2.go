package cgroups

import (
	"fmt"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

// CgroupV2 used for cgroupv2 validation
type CgroupV2 struct {
	MountPath string
}

// GetBlockIOData gets cgroup blockio data
func GetBlockIOData(cgPath string) (*rspec.LinuxBlockIO, error) {
	return nil, fmt.Errorf("unimplemented yet")
}

// GetCPUData gets cgroup cpus data
func GetCPUData(cgPath string) (*rspec.LinuxCPU, error) {
	return nil, fmt.Errorf("unimplemented yet")
}

// GetDevicesData gets cgroup devices data
func GetDevicesData(cgPath string) ([]rspec.LinuxDeviceCgroup, error) {
	return nil, fmt.Errorf("unimplemented yet")
}

// GetHugepageLimitData gets cgroup hugetlb data
func GetHugepageLimitData(cgPath string) ([]rspec.LinuxHugepageLimit, error) {
	return nil, fmt.Errorf("unimplemented yet")
}

// GetMemoryData gets cgroup memory data
func (cg *CgroupV2) GetMemoryData(cgPath string) (*rspec.LinuxMemory, error) {
	return nil, fmt.Errorf("unimplemented yet")
}

// GetNetworkData gets cgroup network data
func GetNetworkData(cgPath string) (*rspec.LinuxNetwork, error) {
	return nil, fmt.Errorf("unimplemented yet")
}

// GetPidsData gets cgroup pids data
func GetPidsData(cgPath string) (*rspec.LinuxPids, error) {
	return nil, fmt.Errorf("unimplemented yet")
}
