package main

import (
	"github.com/mndrix/tap-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	const (
		shares     uint64 = 1024
		period     uint64 = 100000
		quota      int64  = 50000
		cpus, mems        = "0-1", "0"
	)

	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
	// cpu.shares is mapped onto cpu.weight by a conversion the runtime-spec
	// does not define, and the mapping is lossy, so the configured value
	// cannot be recovered.
	if !util.UnifiedCgroups() {
		g.SetLinuxResourcesCPUShares(shares)
	}
	g.SetLinuxResourcesCPUQuota(quota)
	g.SetLinuxResourcesCPUPeriod(period)
	g.SetLinuxResourcesCPUCpus(cpus)
	g.SetLinuxResourcesCPUMems(mems)
	err = util.RuntimeOutsideValidate(g, t, util.ValidateLinuxResourcesCPU)

	if err != nil {
		t.Fail(err.Error())
	}
}
