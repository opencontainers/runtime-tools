package cgroups

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
)

// CgroupV2 used for cgroupv2 validation
type CgroupV2 struct{}

// unifiedPath is where the unified hierarchy is mounted.
const unifiedPath = "/sys/fs/cgroup"

// unlimited is how the unified hierarchy spells "no limit".
const unlimited = "max"

func attachError() error {
	return specerror.NewError(specerror.CgroupsPathAttach, fmt.Errorf("The runtime MUST consistently attach to the same place in the cgroups hierarchy given the same value of `cgroupsPath`"), rspec.Version)
}

func parseLimit(value string) (int64, error) {
	if value == unlimited {
		return -1, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

// Version returns the hierarchy version this implementation reads
func (cg *CgroupV2) Version() int {
	return 2
}

// ControllerPath returns the directory backing an absolute cgroupsPath.
// The unified hierarchy keeps every controller in one directory, so the
// controller is ignored.
func (cg *CgroupV2) ControllerPath(controller string, cgPath string) string {
	return filepath.Join(unifiedPath, cgPath)
}

func (cg *CgroupV2) dir(pid int, cgPath string) (string, error) {
	if filepath.IsAbs(cgPath) {
		path := filepath.Join(unifiedPath, cgPath)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return "", specerror.NewError(specerror.CgroupsAbsPathRelToMount, fmt.Errorf("In the case of an absolute path, the runtime MUST take the path to be relative to the cgroups mount point"), rspec.Version)
			}
			return "", err
		}
		return path, nil
	}

	subPath, err := GetCgroupV2Path(pid)
	if err != nil {
		return "", err
	}
	if !strings.Contains(subPath, cgPath) {
		return "", fmt.Errorf("cgroup %s is not mounted as expected", cgPath)
	}

	return filepath.Join(unifiedPath, subPath), nil
}

// readValue returns the trimmed contents of a control file, or an empty
// string if the controller does not provide it.
func (cg *CgroupV2) readValue(pid int, cgPath string, name string) (string, error) {
	dir, err := cg.dir(pid, cgPath)
	if err != nil {
		return "", err
	}

	contents, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(contents)), nil
}

func (cg *CgroupV2) readLimit(pid int, cgPath string, name string) (int64, error) {
	value, err := cg.readValue(pid, cgPath, name)
	if err != nil {
		return 0, err
	}
	if value == "" {
		return 0, attachError()
	}

	return parseLimit(value)
}

// readCPUSet falls back to the effective set because the unified hierarchy
// leaves the configured file empty while the cgroup inherits its parent.
func (cg *CgroupV2) readCPUSet(pid int, cgPath string, name string) (string, error) {
	value, err := cg.readValue(pid, cgPath, name)
	if err != nil || value != "" {
		return value, err
	}

	return cg.readValue(pid, cgPath, name+".effective")
}

// GetBlockIOData gets cgroup blockio data
func (cg *CgroupV2) GetBlockIOData(pid int, cgPath string) (*rspec.LinuxBlockIO, error) {
	value, err := cg.readValue(pid, cgPath, "io.max")
	if err != nil {
		return nil, err
	}

	lb := &rspec.LinuxBlockIO{}
	if value == "" {
		return lb, nil
	}

	for _, line := range strings.Split(value, "\n") {
		if err := addThrottleDevices(lb, line); err != nil {
			return nil, err
		}
	}

	return lb, nil
}

func addThrottleDevices(lb *rspec.LinuxBlockIO, line string) error {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil
	}

	major, minor, err := getDeviceID(fields[0])
	if err != nil {
		return err
	}

	for _, field := range fields[1:] {
		key, value, found := strings.Cut(field, "=")
		if !found || value == unlimited {
			continue
		}
		rate, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		ltd := rspec.LinuxThrottleDevice{}
		ltd.Major = major
		ltd.Minor = minor
		ltd.Rate = rate
		switch key {
		case "rbps":
			lb.ThrottleReadBpsDevice = append(lb.ThrottleReadBpsDevice, ltd)
		case "wbps":
			lb.ThrottleWriteBpsDevice = append(lb.ThrottleWriteBpsDevice, ltd)
		case "riops":
			lb.ThrottleReadIOPSDevice = append(lb.ThrottleReadIOPSDevice, ltd)
		case "wiops":
			lb.ThrottleWriteIOPSDevice = append(lb.ThrottleWriteIOPSDevice, ltd)
		}
	}

	return nil
}

// GetCPUData gets cgroup cpus data
func (cg *CgroupV2) GetCPUData(pid int, cgPath string) (*rspec.LinuxCPU, error) {
	value, err := cg.readValue(pid, cgPath, "cpu.max")
	if err != nil {
		return nil, err
	}
	if value == "" {
		return nil, attachError()
	}
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return nil, fmt.Errorf("unexpected cpu.max content %q", value)
	}
	quota, err := parseLimit(fields[0])
	if err != nil {
		return nil, err
	}
	period, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}

	cpus, err := cg.readCPUSet(pid, cgPath, "cpuset.cpus")
	if err != nil {
		return nil, err
	}
	mems, err := cg.readCPUSet(pid, cgPath, "cpuset.mems")
	if err != nil {
		return nil, err
	}

	// cpu.shares has no unified equivalent: runtimes map it onto cpu.weight
	// with a lossy conversion that the runtime-spec does not define, so the
	// requested share cannot be recovered.
	return &rspec.LinuxCPU{
		Quota:  &quota,
		Period: &period,
		Cpus:   cpus,
		Mems:   mems,
	}, nil
}

// GetDevicesData gets cgroup devices data
func (cg *CgroupV2) GetDevicesData(pid int, cgPath string) ([]rspec.LinuxDeviceCgroup, error) {
	return nil, fmt.Errorf("the unified hierarchy has no devices controller, access is governed by an eBPF program that cannot be read back")
}

// GetHugepageLimitData gets cgroup hugetlb data
func (cg *CgroupV2) GetHugepageLimitData(pid int, cgPath string) ([]rspec.LinuxHugepageLimit, error) {
	pageSizes, err := GetHugePageSize()
	if err != nil {
		return nil, err
	}

	lh := []rspec.LinuxHugepageLimit{}
	for _, pageSize := range pageSizes {
		value, err := cg.readValue(pid, cgPath, strings.Join([]string{"hugetlb", pageSize, "max"}, "."))
		if err != nil {
			return nil, err
		}
		if value == "" {
			continue
		}
		limit := uint64(math.MaxUint64)
		if value != unlimited {
			limit, err = strconv.ParseUint(value, 10, 64)
			if err != nil {
				return nil, err
			}
		}
		pageLimit := rspec.LinuxHugepageLimit{}
		pageLimit.Pagesize = pageSize
		pageLimit.Limit = limit
		lh = append(lh, pageLimit)
	}

	return lh, nil
}

// GetMemoryData gets cgroup memory data
func (cg *CgroupV2) GetMemoryData(pid int, cgPath string) (*rspec.LinuxMemory, error) {
	limit, err := cg.readLimit(pid, cgPath, "memory.max")
	if err != nil {
		return nil, err
	}
	reservation, err := cg.readLimit(pid, cgPath, "memory.low")
	if err != nil {
		return nil, err
	}

	lm := &rspec.LinuxMemory{
		Limit:       &limit,
		Reservation: &reservation,
	}

	value, err := cg.readValue(pid, cgPath, "memory.swap.max")
	if err != nil {
		return nil, err
	}
	if value == "" {
		return lm, nil
	}
	swap, err := parseLimit(value)
	if err != nil {
		return nil, err
	}
	// memory.swap.max limits swap alone while memory.swap is memory plus
	// swap, so the memory limit has to be added back.
	if swap >= 0 && limit >= 0 {
		swap += limit
	}
	lm.Swap = &swap

	return lm, nil
}

// GetNetworkData gets cgroup network data
func (cg *CgroupV2) GetNetworkData(pid int, cgPath string) (*rspec.LinuxNetwork, error) {
	return nil, fmt.Errorf("the unified hierarchy has no net_cls or net_prio controller")
}

// GetPidsData gets cgroup pid ints data
func (cg *CgroupV2) GetPidsData(pid int, cgPath string) (*rspec.LinuxPids, error) {
	limit, err := cg.readLimit(pid, cgPath, "pids.max")
	if err != nil {
		return nil, err
	}

	return &rspec.LinuxPids{Limit: &limit}, nil
}
