package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var shares uint64 = 1024
	var period uint64 = 100000
	var quota int64 = 50000
	var cpus, mems string = "0-1", "0"
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
	g.SetLinuxResourcesCPUShares(shares)
	g.SetLinuxResourcesCPUQuota(quota)
	g.SetLinuxResourcesCPUPeriod(period)
	g.SetLinuxResourcesCPUCpus(cpus)
	g.SetLinuxResourcesCPUMems(mems)
	err := util.RuntimeOutsideValidate(g, cgroups.AbsCgroupPath, func(pid int, path string) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lcd, err := cg.GetCPUData(pid, path)
		if err != nil {
			return err
		}
		if *lcd.Shares != shares {
			return fmt.Errorf("cpus shares limit is not set correctly, expect: %d, actual: %d", shares, lcd.Shares)
		}
		if *lcd.Quota != quota {
			return fmt.Errorf("cpus quota is not set correctly, expect: %d, actual: %d", quota, lcd.Quota)
		}
		if *lcd.Period != period {
			return fmt.Errorf("cpus period is not set correctly, expect: %d, actual: %d", period, lcd.Period)
		}
		if lcd.Cpus != cpus {
			return fmt.Errorf("cpus cpus is not set correctly, expect: %s, actual: %s", cpus, lcd.Cpus)
		}
		if lcd.Mems != mems {
			return fmt.Errorf("cpus mems is not set correctly, expect: %s, actual: %s", mems, lcd.Mems)
		}
		return nil
	})

	if err != nil {
		util.Fatal(err)
	}
}
