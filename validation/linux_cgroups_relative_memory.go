package main

import (
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var limit int64 = 50593792
	var swappiness uint64 = 50
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
	g.SetLinuxResourcesMemoryLimit(limit)
	g.SetLinuxResourcesMemoryReservation(limit)
	g.SetLinuxResourcesMemorySwap(limit)
	g.SetLinuxResourcesMemoryKernel(limit)
	g.SetLinuxResourcesMemoryKernelTCP(limit)
	g.SetLinuxResourcesMemorySwappiness(swappiness)
	g.SetLinuxResourcesMemoryDisableOOMKiller(true)
	err = util.RuntimeOutsideValidate(g, util.ValidateLinuxResourcesMemory)
	if err != nil {
		util.Fatal(err)
	}
}
