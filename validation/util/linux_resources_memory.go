package util

import (
	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
)

// ValidateLinuxResourcesMemory validates linux.resources.memory.
func ValidateLinuxResourcesMemory(config *rspec.Spec, t *tap.T, state *rspec.State) error {
	cg, err := cgroups.FindCgroup()
	t.Ok((err == nil), "find memory cgroup")
	if err != nil {
		t.Diagnostic(err.Error())
		return nil
	}

	lm, err := cg.GetMemoryData(state.Pid, config.Linux.CgroupsPath)
	t.Ok((err == nil), "get memory cgroup data")
	if err != nil {
		t.Diagnostic(err.Error())
		return nil
	}

	lrm := config.Linux.Resources.Memory
	checkValue(t, "memory limit", lrm.Limit, lm.Limit)
	checkValue(t, "memory reservation", lrm.Reservation, lm.Reservation)
	checkValue(t, "memory swap", lrm.Swap, lm.Swap)
	checkValue(t, "memory kernel", lrm.Kernel, lm.Kernel) //nolint:staticcheck // Ignore SA1019: Kernel is deprecated
	checkValue(t, "memory kernelTCP", lrm.KernelTCP, lm.KernelTCP)
	checkValue(t, "memory swappiness", lrm.Swappiness, lm.Swappiness)
	checkValue(t, "memory oom", lrm.DisableOOMKiller, lm.DisableOOMKiller)

	return nil
}
