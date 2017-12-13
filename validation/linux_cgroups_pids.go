package main

import (
	"fmt"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var limit int64 = 1000
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
	g.SetLinuxResourcesPidsLimit(limit)
	err := util.RuntimeOutsideValidate(g, func(config *rspec.Spec, state *rspec.State) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lpd, err := cg.GetPidsData(state.Pid, config.Linux.CgroupsPath)
		if err != nil {
			return err
		}
		if lpd.Limit != limit {
			return fmt.Errorf("pids limit is not set correctly, expect: %d, actual: %d", limit, lpd.Limit)
		}
		return nil
	})

	if err != nil {
		util.Fatal(err)
	}
}
