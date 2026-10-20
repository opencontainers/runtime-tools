package main

import (
	"fmt"
	"runtime"

	"github.com/mndrix/tap-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func testBlkioCgroups(rate uint64, isEmpty bool) error {
	var weight uint16 = 500
	var leafWeight uint16 = 300

	major, minor, err := util.GetBlockDevice()
	if err != nil {
		return err
	}

	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	g, err := util.GetDefaultGenerator()
	if err != nil {
		return err
	}

	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)

	if !isEmpty {
		g.SetLinuxResourcesBlockIOWeight(weight)
		g.SetLinuxResourcesBlockIOLeafWeight(leafWeight)
	}

	g.AddLinuxResourcesBlockIOWeightDevice(major, minor, weight)
	g.AddLinuxResourcesBlockIOLeafWeightDevice(major, minor, leafWeight)
	g.AddLinuxResourcesBlockIOThrottleReadBpsDevice(major, minor, rate)
	g.AddLinuxResourcesBlockIOThrottleWriteBpsDevice(major, minor, rate)
	g.AddLinuxResourcesBlockIOThrottleReadIOPSDevice(major, minor, rate)
	g.AddLinuxResourcesBlockIOThrottleWriteIOPSDevice(major, minor, rate)

	err = util.RuntimeOutsideValidate(g, t, util.ValidateLinuxResourcesBlockIO)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if runtime.GOOS != "linux" {
		util.Fatal(fmt.Errorf("linux-specific cgroup test"))
	}

	cases := []struct {
		rate    uint64
		isEmpty bool
	}{
		{102400, false},
		{204800, false},
		{102400, true},
		{204800, true},
	}

	for _, c := range cases {
		if err := testBlkioCgroups(c.rate, c.isEmpty); err != nil {
			util.Fatal(err)
		}
	}
}
