package main

import (
	"fmt"

	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	page := "1GB"
	var limit uint64 = 56892210544640
	g := util.GetDefaultGenerator()
	g.SetLinuxCgroupsPath("/test")
	g.AddLinuxResourcesHugepageLimit(page, limit)
	err := util.RuntimeOutsideValidate(g, func(path string) error {
		cg, err := cgroups.FindCgroup()
		if err != nil {
			return err
		}
		lhd, err := cg.GetHugepageLimitData(path)
		if err != nil {
			return err
		}
		for _, lhl := range lhd {
			if lhl.Pagesize == page && lhl.Limit != limit {
				return fmt.Errorf("hugepage %s limit is not set correctly, expect: %d, actual: %d", page, limit, lhl.Limit)
			}
		}
		return nil
	})
	if err != nil {
		util.Fatal(err)
	}
}
