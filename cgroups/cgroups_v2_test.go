package cgroups

import (
	"os"
	"path/filepath"
	"testing"
)

var _ Cgroup = (*CgroupV2)(nil)

const testCgroupPath = "/cgrouptest"

func newTestCgroupV2(t *testing.T, files map[string]string) *CgroupV2 {
	t.Helper()

	mountPath := t.TempDir()
	cgDir := filepath.Join(mountPath, testCgroupPath)
	if err := os.Mkdir(cgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(cgDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return &CgroupV2{MountPath: mountPath}
}

func TestCgroupV2Version(t *testing.T) {
	cg := &CgroupV2{}
	if cg.Version() != 2 {
		t.Errorf("Version() == %d, expected 2", cg.Version())
	}
}

func TestCgroupV2MissingCgroup(t *testing.T) {
	cg := &CgroupV2{MountPath: t.TempDir()}
	if _, err := cg.GetPidsData(0, testCgroupPath); err == nil {
		t.Error("expected an error for a cgroup that was never created")
	}
}

func TestCgroupV2GetPidsData(t *testing.T) {
	for _, test := range []struct {
		content  string
		expected int64
	}{
		{content: "1000\n", expected: 1000},
		{content: "max\n", expected: -1},
	} {
		cg := newTestCgroupV2(t, map[string]string{"pids.max": test.content})
		lp, err := cg.GetPidsData(0, testCgroupPath)
		if err != nil {
			t.Fatal(err)
		}
		if *lp.Limit != test.expected {
			t.Errorf("pids.max %q: limit == %d, expected %d", test.content, *lp.Limit, test.expected)
		}
	}
}

func TestCgroupV2GetMemoryData(t *testing.T) {
	for _, test := range []struct {
		files        map[string]string
		description  string
		reservation  int64
		limit        int64
		swap         int64
		swapReported bool
	}{
		{
			description: "limits in bytes",
			files: map[string]string{
				"memory.max":      "50593792\n",
				"memory.low":      "33554432\n",
				"memory.swap.max": "16515072\n",
			},
			limit:        50593792,
			reservation:  33554432,
			swap:         67108864,
			swapReported: true,
		},
		{
			description: "unlimited memory and swap",
			files: map[string]string{
				"memory.max":      "max\n",
				"memory.low":      "0\n",
				"memory.swap.max": "max\n",
			},
			limit:        -1,
			reservation:  0,
			swap:         -1,
			swapReported: true,
		},
		{
			description: "no swap accounting",
			files: map[string]string{
				"memory.max": "50593792\n",
				"memory.low": "0\n",
			},
			limit:       50593792,
			reservation: 0,
		},
	} {
		cg := newTestCgroupV2(t, test.files)
		lm, err := cg.GetMemoryData(0, testCgroupPath)
		if err != nil {
			t.Fatalf("%s: %v", test.description, err)
		}
		if *lm.Limit != test.limit {
			t.Errorf("%s: limit == %d, expected %d", test.description, *lm.Limit, test.limit)
		}
		if *lm.Reservation != test.reservation {
			t.Errorf("%s: reservation == %d, expected %d", test.description, *lm.Reservation, test.reservation)
		}
		if test.swapReported != (lm.Swap != nil) {
			t.Fatalf("%s: swap reported == %t, expected %t", test.description, lm.Swap != nil, test.swapReported)
		}
		if test.swapReported && *lm.Swap != test.swap {
			t.Errorf("%s: swap == %d, expected %d", test.description, *lm.Swap, test.swap)
		}
		if lm.Swappiness != nil || lm.DisableOOMKiller != nil || lm.KernelTCP != nil {
			t.Errorf("%s: reported a value the unified hierarchy does not have", test.description)
		}
	}
}

func TestCgroupV2GetCPUData(t *testing.T) {
	for _, test := range []struct {
		files       map[string]string
		description string
		cpus        string
		mems        string
		quota       int64
		period      uint64
	}{
		{
			description: "quota and explicit cpuset",
			files: map[string]string{
				"cpu.max":     "50000 100000\n",
				"cpuset.cpus": "0-1\n",
				"cpuset.mems": "0\n",
				"cpu.weight":  "100\n",
			},
			quota:  50000,
			period: 100000,
			cpus:   "0-1",
			mems:   "0",
		},
		{
			description: "no quota, inherited cpuset",
			files: map[string]string{
				"cpu.max":               "max 100000\n",
				"cpuset.cpus":           "\n",
				"cpuset.cpus.effective": "0-3\n",
				"cpuset.mems":           "\n",
				"cpuset.mems.effective": "0\n",
			},
			quota:  -1,
			period: 100000,
			cpus:   "0-3",
			mems:   "0",
		},
	} {
		cg := newTestCgroupV2(t, test.files)
		lc, err := cg.GetCPUData(0, testCgroupPath)
		if err != nil {
			t.Fatalf("%s: %v", test.description, err)
		}
		if *lc.Quota != test.quota {
			t.Errorf("%s: quota == %d, expected %d", test.description, *lc.Quota, test.quota)
		}
		if *lc.Period != test.period {
			t.Errorf("%s: period == %d, expected %d", test.description, *lc.Period, test.period)
		}
		if lc.Cpus != test.cpus {
			t.Errorf("%s: cpus == %q, expected %q", test.description, lc.Cpus, test.cpus)
		}
		if lc.Mems != test.mems {
			t.Errorf("%s: mems == %q, expected %q", test.description, lc.Mems, test.mems)
		}
		if lc.Shares != nil {
			t.Errorf("%s: reported shares, which cannot be recovered from cpu.weight", test.description)
		}
	}
}

func TestCgroupV2GetBlockIOData(t *testing.T) {
	cg := newTestCgroupV2(t, map[string]string{
		"io.max": "8:0 rbps=102400 wbps=204800 riops=max wiops=1000\n259:0 rbps=max wbps=max riops=max wiops=max\n",
	})
	lb, err := cg.GetBlockIOData(0, testCgroupPath)
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
	if lb.ThrottleReadBpsDevice[0].Major != 8 || lb.ThrottleReadBpsDevice[0].Minor != 0 {
		t.Errorf("read bps device == %d:%d, expected 8:0", lb.ThrottleReadBpsDevice[0].Major, lb.ThrottleReadBpsDevice[0].Minor)
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

	var limit uint64 = 2 * (1 << 30)
	files := map[string]string{
		"hugetlb." + pageSizes[0] + ".max": "2147483648\n",
	}
	cg := newTestCgroupV2(t, files)
	lh, err := cg.GetHugepageLimitData(0, testCgroupPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(lh) != 1 {
		t.Fatalf("hugepage limits == %v, expected a single entry", lh)
	}
	if lh[0].Pagesize != pageSizes[0] || lh[0].Limit != limit {
		t.Errorf("hugepage limit == %s/%d, expected %s/%d", lh[0].Pagesize, lh[0].Limit, pageSizes[0], limit)
	}
}

func TestCgroupV2UnreadableControllers(t *testing.T) {
	cg := newTestCgroupV2(t, nil)
	if _, err := cg.GetDevicesData(0, testCgroupPath); err == nil {
		t.Error("expected an error for the devices controller")
	}
	if _, err := cg.GetNetworkData(0, testCgroupPath); err == nil {
		t.Error("expected an error for the network controllers")
	}
}
