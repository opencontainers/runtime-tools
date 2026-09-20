package cgroups

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

var (
	// AbsCgroupPath is absolute path for container's cgroup mount
	AbsCgroupPath = "/cgrouptest"
	// RelCgroupPath is relative path for container's cgroup mount
	RelCgroupPath = "testdir/cgrouptest/container"
)

// Cgroup represents interfaces for cgroup validation
type Cgroup interface {
	Version() int
	GetBlockIOData(pid int, cgPath string) (*rspec.LinuxBlockIO, error)
	GetCPUData(pid int, cgPath string) (*rspec.LinuxCPU, error)
	GetDevicesData(pid int, cgPath string) ([]rspec.LinuxDeviceCgroup, error)
	GetHugepageLimitData(pid int, cgPath string) ([]rspec.LinuxHugepageLimit, error)
	GetMemoryData(pid int, cgPath string) (*rspec.LinuxMemory, error)
	GetNetworkData(pid int, cgPath string) (*rspec.LinuxNetwork, error)
	GetPidsData(pid int, cgPath string) (*rspec.LinuxPids, error)
}

// FindCgroup gets cgroup root mountpoint
func FindCgroup() (Cgroup, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	unifiedPath := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Split(text, " ")
		// Safe as mountinfo encodes mountpoints with spaces as \040.
		index := strings.Index(text, " - ")
		postSeparatorFields := strings.Split(text[index+3:], " ")
		numPostFields := len(postSeparatorFields)

		// This is an error as we can't detect if the mount is for "cgroup"
		if numPostFields == 0 {
			return nil, fmt.Errorf("Found no fields post '-' in %q", text)
		}

		if postSeparatorFields[0] == "cgroup" {
			// No need to parse the rest of the postSeparatorFields

			cg := &CgroupV1{
				MountPath: filepath.Dir(fields[4]),
			}
			return cg, nil
		} else if postSeparatorFields[0] == "cgroup2" {
			// A unified hierarchy is only used when no legacy
			// controller is mounted, so keep looking.
			unifiedPath = fields[4]
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if unifiedPath != "" {
		cg := &CgroupV2{
			MountPath: unifiedPath,
		}
		return cg, nil
	}
	return nil, fmt.Errorf("cgroup is not found")
}

// GetSubsystemPath gets path of subsystem
func GetSubsystemPath(pid int, subsystem string) (string, error) {
	contents, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return "", err
	}

	parts := strings.Split(strings.TrimSpace(string(contents)), "\n")
	for _, part := range parts {
		elem := strings.SplitN(part, ":", 3)
		if len(elem) < 3 {
			continue
		}
		subelems := strings.Split(elem[1], ",")
		if slices.Contains(subelems, subsystem) {
			return elem[2], nil
		}
	}

	return "", fmt.Errorf("subsystem %s not found", subsystem)
}
