package main

import (
	"fmt"

	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	pageSizes, err := cgroups.GetHugePageSize()
	if err != nil {
		t.Fail(fmt.Sprintf("error when getting hugepage sizes: %+v", err))
	}
	// When setting the limit just for checking if writing works, the amount of memory
	// requested does not matter, as all insigned integers will be accepted.
	// Use 2GiB as an example
	var limit uint64 = 2 * (1 << 30)

	for _, pageSize := range pageSizes {
		g, err := util.GetDefaultGenerator()
		if err != nil {
			util.Fatal(err)
		}
		g.SetLinuxCgroupsPath(cgroups.RelCgroupPath)
		g.AddLinuxResourcesHugepageLimit(pageSize, limit)
		err = util.RuntimeOutsideValidate(g, t, func(config *rspec.Spec, t *tap.T, state *rspec.State) error {
			cg, err := cgroups.FindCgroup()
			t.Ok((err == nil), "find hugetlb cgroup")
			if err != nil {
				t.Diagnostic(err.Error())
				return nil
			}

			lhd, err := cg.GetHugepageLimitData(state.Pid, config.Linux.CgroupsPath)
			t.Ok((err == nil), "get hugetlb cgroup data")
			if err != nil {
				t.Diagnostic(err.Error())
				return nil
			}

			found := false
			for _, lhl := range lhd {
				if lhl.Pagesize == pageSize {
					found = true
					t.Ok(lhl.Limit == limit, fmt.Sprintf("hugepage limit is set correctly for size: %s", pageSize))
					t.Diagnosticf("expect: %d, actual: %d", limit, lhl.Limit)
				}
			}
			t.Ok(found, "hugepage limit found")

			return nil
		})

		if err != nil {
			t.Fail(err.Error())
		}
	}
}
