package main

import (
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/util/tap"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	var limit int64 = 1000

	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
	g.SetLinuxResourcesPidsLimit(limit)
	err = util.RuntimeOutsideValidate(g, t, util.ValidateLinuxResourcesPids)
	if err != nil {
		t.Fail(err.Error())
	}
}
