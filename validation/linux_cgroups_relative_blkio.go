package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var weight uint16 = 500
	var leafWeight uint16 = 300
	var major, minor int64 = 8, 0
	var rate uint64 = 102400
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
	g.SetLinuxResourcesBlockIOWeight(weight)
	g.SetLinuxResourcesBlockIOLeafWeight(leafWeight)
	g.AddLinuxResourcesBlockIOWeightDevice(major, minor, weight)
	g.AddLinuxResourcesBlockIOLeafWeightDevice(major, minor, leafWeight)
	g.AddLinuxResourcesBlockIOThrottleReadBpsDevice(major, minor, rate)
	g.AddLinuxResourcesBlockIOThrottleWriteBpsDevice(major, minor, rate)
	g.AddLinuxResourcesBlockIOThrottleReadIOPSDevice(major, minor, rate)
	g.AddLinuxResourcesBlockIOThrottleWriteIOPSDevice(major, minor, rate)
	err := util.RuntimeOutsideValidate(g, cgroups.RelCgroupPath, func(pid int, path string) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lbd, err := cg.GetBlockIOData(pid, path)
		if err != nil {
			return err
		}
		if *lbd.Weight != weight {
			return fmt.Errorf("blkio weight is not set correctly, expect: %d, actual: %d", weight, lbd.Weight)
		}
		if *lbd.LeafWeight != leafWeight {
			return fmt.Errorf("blkio leafWeight is not set correctly, expect: %d, actual: %d", weight, lbd.LeafWeight)
		}

		found := false
		for _, wd := range lbd.WeightDevice {
			if wd.Major == major && wd.Minor == minor {
				found = true
				if *wd.Weight != weight {
					return fmt.Errorf("blkio weight for %d:%d is not set correctly, expect: %d, actual: %d", major, minor, weight, wd.Weight)
				}
				if *wd.LeafWeight != leafWeight {
					return fmt.Errorf("blkio leafWeight for %d:%d is not set correctly, expect: %d, actual: %d", major, minor, leafWeight, wd.LeafWeight)
				}
			}
		}
		if !found {
			return fmt.Errorf("blkio weightDevice for %d:%d is not set", major, minor)
		}

		found = false
		for _, trbd := range lbd.ThrottleReadBpsDevice {
			if trbd.Major == major && trbd.Minor == minor {
				found = true
				if trbd.Rate != rate {
					return fmt.Errorf("blkio read bps for %d:%d is not set correctly, expect: %d, actual: %d", major, minor, rate, trbd.Rate)
				}
			}
		}
		if !found {
			return fmt.Errorf("blkio read bps for %d:%d is not set", major, minor)
		}

		found = false
		for _, twbd := range lbd.ThrottleWriteBpsDevice {
			if twbd.Major == major && twbd.Minor == minor {
				found = true
				if twbd.Rate != rate {
					return fmt.Errorf("blkio write bps for %d:%d is not set correctly, expect: %d, actual: %d", major, minor, rate, twbd.Rate)
				}
			}
		}
		if !found {
			return fmt.Errorf("blkio write bps for %d:%d is not set", major, minor)
		}

		found = false
		for _, trid := range lbd.ThrottleReadIOPSDevice {
			if trid.Major == major && trid.Minor == minor {
				found = true
				if trid.Rate != rate {
					return fmt.Errorf("blkio read iops for %d:%d is not set correctly, expect: %d, actual: %d", major, minor, rate, trid.Rate)
				}
			}
		}
		if !found {
			return fmt.Errorf("blkio read iops for %d:%d is not set", major, minor)
		}

		found = false
		for _, twid := range lbd.ThrottleWriteIOPSDevice {
			if twid.Major == major && twid.Minor == minor {
				found = true
				if twid.Rate != rate {
					return fmt.Errorf("blkio write iops for %d:%d is not set correctly, expect: %d, actual: %d", major, minor, rate, twid.Rate)
				}
			}
		}
		if !found {
			return fmt.Errorf("blkio write iops for %d:%d is not set", major, minor)
		}

		return nil
	})

	if err != nil {
		util.Fatal(err)
	}
}
