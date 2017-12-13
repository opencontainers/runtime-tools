package main

import (
	"fmt"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var id, prio uint32 = 255, 10
	ifName := "lo"
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
	g.SetLinuxResourcesNetworkClassID(id)
	err := util.RuntimeOutsideValidate(g, func(config *rspec.Spec, state *rspec.State) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lnd, err := cg.GetNetworkData(state.Pid, config.Linux.CgroupsPath)
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
