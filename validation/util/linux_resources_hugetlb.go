package util

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/opencontainers/runtime-tools/cgroups"
)

// GetHugetlbPageSizes returns the huge page sizes of the host. The list is
// empty when the kernel offers no huge pages or no hugetlb cgroup
// controller, in which case no runtime can apply hugepage limits.
func GetHugetlbPageSizes() ([]string, error) {
	enabled, err := cgroupControllerEnabled("hugetlb")
	if err != nil || !enabled {
		return nil, err
	}

	pageSizes, err := cgroups.GetHugePageSize()
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return pageSizes, err
}

// cgroupControllerEnabled reports whether the kernel provides the named
// cgroup controller, as listed in /proc/cgroups for both cgroup versions.
func cgroupControllerEnabled(name string) (bool, error) {
	f, err := os.Open("/proc/cgroups")
	if err != nil {
		return false, err
	}
	defer f.Close()

	return controllerEnabled(f, name)
}

func controllerEnabled(r io.Reader, name string) (bool, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) == 4 && fields[0] == name {
			return fields[3] == "1", nil
		}
	}
	return false, s.Err()
}
