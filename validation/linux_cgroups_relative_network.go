package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var id, prio uint32 = 255, 10
	ifName := "lo"
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
	g.SetLinuxResourcesNetworkClassID(id)
	err := util.RuntimeOutsideValidate(g, cgroups.RelCgroupPath, func(pid int, path string) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lnd, err := cg.GetNetworkData(pid, path)
		if err != nil {
			return err
		}
		if *lnd.ClassID != id {
			return fmt.Errorf("network ID is not set correctly, expect: %d, actual: %d", id, lnd.ClassID)
		}
		found := false
		for _, lip := range lnd.Priorities {
			if lip.Name == ifName {
				found = true
				if lip.Priority != prio {
					return fmt.Errorf("network priority for %s is not set correctly, expect: %d, actual: %d", ifName, prio, lip.Priority)
				}
			}
		}
		if !found {
			return fmt.Errorf("network priority for %s is not set correctly", ifName)
		}

		return nil
	})

	if err != nil {
		util.Fatal(err)
	}
}
