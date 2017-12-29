package main

import (
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var id, prio uint32 = 255, 10
	ifName := "lo"
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
	g.SetLinuxResourcesNetworkClassID(id)
	g.AddLinuxResourcesNetworkPriorities(ifName, prio)
	err := util.RuntimeOutsideValidate(g, util.ValidateLinuxResourcesNetwork)
	if err != nil {
		util.Fatal(err)
	}
}
