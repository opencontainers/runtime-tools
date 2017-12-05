package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var limit int64 = 1000
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath("/test")
	g.SetLinuxResourcesPidsLimit(limit)
	err := util.RuntimeOutsideValidate(g, func(path string) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lpd, err := cg.GetPidsData(path)
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
