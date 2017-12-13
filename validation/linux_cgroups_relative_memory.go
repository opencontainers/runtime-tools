package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var limit int64 = 50593792
	var swappiness uint64 = 50
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
	g.SetLinuxResourcesMemoryLimit(limit)
	g.SetLinuxResourcesMemoryReservation(limit)
	g.SetLinuxResourcesMemorySwap(limit)
	g.SetLinuxResourcesMemoryKernel(limit)
	g.SetLinuxResourcesMemoryKernelTCP(limit)
	g.SetLinuxResourcesMemorySwappiness(swappiness)
	g.SetLinuxResourcesMemoryDisableOOMKiller(true)
	err := util.RuntimeOutsideValidate(g, cgroups.RelCgroupPath, func(pid int, path string) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lm, err := cg.GetMemoryData(pid, path)
		if err != nil {
			return err
		}
		if limit != *lm.Limit {
			return fmt.Errorf("memory limit is not set correctly, expect: %d, actual: %d", limit, *lm.Limit)
		}
		if limit != *lm.Reservation {
			return fmt.Errorf("memory reservation is not set correctly, expect: %d, actual: %d", limit, *lm.Reservation)
		}
		if limit != *lm.Swap {
			return fmt.Errorf("memory swap is not set correctly, expect: %d, actual: %d", limit, *lm.Reservation)
		}
		if limit != *lm.Kernel {
			return fmt.Errorf("memory kernel is not set correctly, expect: %d, actual: %d", limit, *lm.Kernel)
		}
		if limit != *lm.KernelTCP {
			return fmt.Errorf("memory kernelTCP is not set correctly, expect: %d, actual: %d", limit, *lm.Kernel)
		}
		if swappiness != *lm.Swappiness {
			return fmt.Errorf("memory swappiness is not set correctly, expect: %d, actual: %d", swappiness, *lm.Swappiness)
		}
		if true != *lm.DisableOOMKiller {
			return fmt.Errorf("memory oom is not set correctly, expect: %t, actual: %t", true, *lm.DisableOOMKiller)
		}
		return nil
	})
	if err != nil {
		util.Fatal(err)
	}
}
