package cgroups

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

var _ Cgroup = (*CgroupV2)(nil)

func requireCgroupV2(t *testing.T) {
	t.Helper()

	var st unix.Statfs_t
	if err := unix.Statfs(unifiedPath, &st); err != nil || st.Type != unix.CGROUP2_SUPER_MAGIC {
		t.Skip("cgroup v2 is not mounted at " + unifiedPath)
	}
}

// newTestCgroup creates a cgroup right below the root of the unified
// hierarchy, where the controllers are enabled, and returns it as an
// absolute cgroupsPath.
func newTestCgroup(t *testing.T, controllers ...string) string {
	t.Helper()
	requireCgroupV2(t)

	subtreeControl := filepath.Join(unifiedPath, "cgroup.subtree_control")
	enabled, err := os.ReadFile(subtreeControl)
	if err != nil {
		t.Skip(err)
	}
	for _, controller := range controllers {
		controller := controller
		if slices.Contains(strings.Fields(string(enabled)), controller) {
			continue
		}
		if err := os.WriteFile(subtreeControl, []byte("+"+controller), 0o644); err != nil {
			t.Skipf("cannot enable the %s controller at %s: %v", controller, unifiedPath, err)
		}
		t.Cleanup(func() {
			if err := os.WriteFile(subtreeControl, []byte("-"+controller), 0o644); err != nil {
				t.Logf("cannot disable the %s controller again: %v", controller, err)
			}
		})
	}

	dir, err := os.MkdirTemp(unifiedPath, "runtime-tools-test-")
	if errors.Is(err, os.ErrPermission) || errors.Is(err, unix.EROFS) {
		t.Skipf("cannot create a cgroup: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(dir); err != nil {
			t.Error(err)
		}
	})

	return "/" + filepath.Base(dir)
}

func writeControl(t *testing.T, cgPath string, name string, value string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(unifiedPath, cgPath, name), []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readControl(t *testing.T, cgPath string, name string) string {
	t.Helper()

	contents, err := os.ReadFile(filepath.Join(unifiedPath, cgPath, name))
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(contents))
}

func TestCgroupV2Version(t *testing.T) {
	cg := &CgroupV2{}
	if cg.Version() != 2 {
		t.Errorf("Version() == %d, expected 2", cg.Version())
	}
}

func TestFindCgroupV2(t *testing.T) {
	requireCgroupV2(t)

	cg, err := FindCgroup()
	if err != nil {
		t.Fatal(err)
	}
	if cg.Version() != 2 {
		t.Errorf("FindCgroup() found a version %d hierarchy, expected 2", cg.Version())
	}
}

func TestGetCgroupV2Path(t *testing.T) {
	requireCgroupV2(t)

	path, err := GetCgroupV2Path(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("cgroup path %q is not absolute", path)
	}
	if _, err := os.Stat(filepath.Join(unifiedPath, path)); err != nil {
		t.Error(err)
	}
}

func TestCgroupV2RelativePath(t *testing.T) {
	requireCgroupV2(t)

	path, err := GetCgroupV2Path(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(unifiedPath, path, "pids.max")); err != nil {
		t.Skipf("the pids controller is not enabled for %s", path)
	}

	cg := &CgroupV2{}
	lp, err := cg.GetPidsData(os.Getpid(), filepath.Base(path))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := parseLimit(readControl(t, path, "pids.max"))
	if err != nil {
		t.Fatal(err)
	}
	if *lp.Limit != expected {
		t.Errorf("limit == %d, expected %d", *lp.Limit, expected)
	}
}

func TestCgroupV2MissingCgroup(t *testing.T) {
	requireCgroupV2(t)

	cg := &CgroupV2{}
	if _, err := cg.GetPidsData(0, "/runtime-tools-test-missing"); err == nil {
		t.Error("expected an error for a cgroup that was never created")
	}
}

func TestCgroupV2GetPidsData(t *testing.T) {
	cgPath := newTestCgroup(t, "pids")
	cg := &CgroupV2{}

	for _, test := range []struct {
		value    string
		expected int64
	}{
		{value: "1000", expected: 1000},
		{value: "max", expected: -1},
	} {
		writeControl(t, cgPath, "pids.max", test.value)
		lp, err := cg.GetPidsData(0, cgPath)
		if err != nil {
			t.Fatal(err)
		}
		if *lp.Limit != test.expected {
			t.Errorf("pids.max %q: limit == %d, expected %d", test.value, *lp.Limit, test.expected)
		}
	}
}

func TestCgroupV2GetMemoryData(t *testing.T) {
	cgPath := newTestCgroup(t, "memory")
	cg := &CgroupV2{}
	_, err := os.Stat(filepath.Join(unifiedPath, cgPath, "memory.swap.max"))
	swapAccounting := err == nil

	for _, test := range []struct {
		description string
		max         string
		low         string
		swapMax     string
		limit       int64
		reservation int64
		swap        int64
	}{
		{
			description: "limits in bytes",
			max:         "50593792",
			low:         "33554432",
			swapMax:     "16515072",
			limit:       50593792,
			reservation: 33554432,
			swap:        67108864,
		},
		{
			description: "unlimited memory and swap",
			max:         "max",
			low:         "0",
			swapMax:     "max",
			limit:       -1,
			reservation: 0,
			swap:        -1,
		},
	} {
		writeControl(t, cgPath, "memory.max", test.max)
		writeControl(t, cgPath, "memory.low", test.low)
		if swapAccounting {
			writeControl(t, cgPath, "memory.swap.max", test.swapMax)
		}

		lm, err := cg.GetMemoryData(0, cgPath)
		if err != nil {
			t.Fatalf("%s: %v", test.description, err)
		}
		if *lm.Limit != test.limit {
			t.Errorf("%s: limit == %d, expected %d", test.description, *lm.Limit, test.limit)
		}
		if *lm.Reservation != test.reservation {
			t.Errorf("%s: reservation == %d, expected %d", test.description, *lm.Reservation, test.reservation)
		}
		if swapAccounting != (lm.Swap != nil) {
			t.Fatalf("%s: swap reported == %t, expected %t", test.description, lm.Swap != nil, swapAccounting)
		}
		if swapAccounting && *lm.Swap != test.swap {
			t.Errorf("%s: swap == %d, expected %d", test.description, *lm.Swap, test.swap)
		}
		if lm.Swappiness != nil || lm.DisableOOMKiller != nil || lm.KernelTCP != nil {
			t.Errorf("%s: reported a value the unified hierarchy does not have", test.description)
		}
	}
}

func TestCgroupV2GetCPUData(t *testing.T) {
	cgPath := newTestCgroup(t, "cpu", "cpuset")
	cg := &CgroupV2{}

	writeControl(t, cgPath, "cpu.max", "50000 100000")
	writeControl(t, cgPath, "cpuset.cpus", "0")
	writeControl(t, cgPath, "cpuset.mems", "0")
	lc, err := cg.GetCPUData(0, cgPath)
	if err != nil {
		t.Fatal(err)
	}
	if *lc.Quota != 50000 || *lc.Period != 100000 {
		t.Errorf("quota/period == %d/%d, expected 50000/100000", *lc.Quota, *lc.Period)
	}
	if lc.Cpus != "0" || lc.Mems != "0" {
		t.Errorf("cpus/mems == %q/%q, expected \"0\"/\"0\"", lc.Cpus, lc.Mems)
	}
	if lc.Shares != nil {
		t.Error("reported shares, which cannot be recovered from cpu.weight")
	}

	writeControl(t, cgPath, "cpu.max", "max 100000")
	writeControl(t, cgPath, "cpuset.cpus", "")
	writeControl(t, cgPath, "cpuset.mems", "")
	lc, err = cg.GetCPUData(0, cgPath)
	if err != nil {
		t.Fatal(err)
	}
	if *lc.Quota != -1 {
		t.Errorf("quota == %d, expected -1", *lc.Quota)
	}
	if cpus := readControl(t, cgPath, "cpuset.cpus.effective"); lc.Cpus != cpus {
		t.Errorf("inherited cpus == %q, expected %q", lc.Cpus, cpus)
	}
	if mems := readControl(t, cgPath, "cpuset.mems.effective"); lc.Mems != mems {
		t.Errorf("inherited mems == %q, expected %q", lc.Mems, mems)
	}
}

func TestCgroupV2GetBlockIOData(t *testing.T) {
	cgPath := newTestCgroup(t, "io")
	cg := &CgroupV2{}

	devices, err := filepath.Glob("/sys/block/*/dev")
	if err != nil {
		t.Fatal(err)
	}
	device := ""
	for _, dev := range devices {
		contents, err := os.ReadFile(dev)
		if err != nil {
			continue
		}
		candidate := strings.TrimSpace(string(contents))
		limit := candidate + " rbps=102400 wbps=204800 wiops=1000"
		if os.WriteFile(filepath.Join(unifiedPath, cgPath, "io.max"), []byte(limit), 0o644) == nil {
			device = candidate
			break
		}
	}
	if device == "" {
		t.Skip("no block device accepts an io.max limit")
	}

	lb, err := cg.GetBlockIOData(0, cgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(lb.ThrottleReadBpsDevice) != 1 || lb.ThrottleReadBpsDevice[0].Rate != 102400 {
		t.Errorf("read bps devices == %v, expected a single 102400 entry", lb.ThrottleReadBpsDevice)
	}
	if len(lb.ThrottleWriteBpsDevice) != 1 || lb.ThrottleWriteBpsDevice[0].Rate != 204800 {
		t.Errorf("write bps devices == %v, expected a single 204800 entry", lb.ThrottleWriteBpsDevice)
	}
	if len(lb.ThrottleReadIOPSDevice) != 0 {
		t.Errorf("read iops devices == %v, expected none", lb.ThrottleReadIOPSDevice)
	}
	if len(lb.ThrottleWriteIOPSDevice) != 1 || lb.ThrottleWriteIOPSDevice[0].Rate != 1000 {
		t.Errorf("write iops devices == %v, expected a single 1000 entry", lb.ThrottleWriteIOPSDevice)
	}
	if len(lb.ThrottleReadBpsDevice) == 1 {
		major, minor, err := getDeviceID(device)
		if err != nil {
			t.Fatal(err)
		}
		if lb.ThrottleReadBpsDevice[0].Major != major || lb.ThrottleReadBpsDevice[0].Minor != minor {
			t.Errorf("read bps device == %d:%d, expected %s", lb.ThrottleReadBpsDevice[0].Major, lb.ThrottleReadBpsDevice[0].Minor, device)
		}
	}
	if lb.Weight != nil || lb.LeafWeight != nil {
		t.Error("reported a weight, which is written to a scheduler dependent file")
	}
}

func TestCgroupV2GetHugepageLimitData(t *testing.T) {
	pageSizes, err := GetHugePageSize()
	if err != nil || len(pageSizes) == 0 {
		t.Skip("no hugepage sizes on this host")
	}
	cgPath := newTestCgroup(t, "hugetlb")
	cg := &CgroupV2{}

	var limit uint64 = 2 * (1 << 30)
	writeControl(t, cgPath, "hugetlb."+pageSizes[0]+".max", "2147483648")
	lh, err := cg.GetHugepageLimitData(0, cgPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, pageLimit := range lh {
		if pageLimit.Pagesize != pageSizes[0] {
			continue
		}
		if pageLimit.Limit != limit {
			t.Errorf("hugepage limit == %d, expected %d", pageLimit.Limit, limit)
		}
		return
	}
	t.Errorf("hugepage limits == %v, expected an entry for %s", lh, pageSizes[0])
}

func TestCgroupV2UnreadableControllers(t *testing.T) {
	cg := &CgroupV2{}
	if _, err := cg.GetDevicesData(0, AbsCgroupPath); err == nil {
		t.Error("expected an error for the devices controller")
	}
	if _, err := cg.GetNetworkData(0, AbsCgroupPath); err == nil {
		t.Error("expected an error for the network controllers")
	}
}
