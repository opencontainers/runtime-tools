package cgroups

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

// Cgroup represents interfaces for cgroup validation
type Cgroup interface {
	GetBlockIOData(cgPath string) (*rspec.LinuxBlockIO, error)
	GetCPUData(cgPath string) (*rspec.LinuxCPU, error)
	GetDevicesData(cgPath string) ([]rspec.LinuxDeviceCgroup, error)
	GetHugepageLimitData(cgPath string) ([]rspec.LinuxHugepageLimit, error)
	GetMemoryData(cgPath string) (*rspec.LinuxMemory, error)
	GetNetworkData(cgPath string) (*rspec.LinuxNetwork, error)
	GetPidsData(cgPath string) (*rspec.LinuxPids, error)
}

// FindCgroup gets cgroup root mountpoint
func FindCgroup() (Cgroup, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cgroupv2 := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Split(text, " ")
		// Safe as mountinfo encodes mountpoints with spaces as \040.
		index := strings.Index(text, " - ")
		postSeparatorFields := strings.Fields(text[index+3:])
		numPostFields := len(postSeparatorFields)

		// This is an error as we can't detect if the mount is for "cgroup"
		if numPostFields == 0 {
			return nil, fmt.Errorf("Found no fields post '-' in %q", text)
		}

		if postSeparatorFields[0] == "cgroup" {
			// Check that the mount is properly formated.
			if numPostFields < 3 {
				return nil, fmt.Errorf("Error found less than 3 fields post '-' in %q", text)
			}

			cg := &CgroupV1{
				MountPath: filepath.Dir(fields[4]),
			}
			return cg, nil
		} else if postSeparatorFields[0] == "cgroup2" {
			cgroupv2 = true
			continue
			//TODO cgroupv2 unimplemented
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if cgroupv2 {
		return nil, fmt.Errorf("cgroupv2 is not supported yet")
	}
	return nil, fmt.Errorf("cgroup is not found")
}
