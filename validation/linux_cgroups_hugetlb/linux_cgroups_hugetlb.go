package main

import (
	"fmt"
	"runtime"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/util/tap"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func testHugetlbCgroups() error {
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
			return err
		}
		g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
		g.AddLinuxResourcesHugepageLimit(pageSize, limit)
		err = util.RuntimeOutsideValidate(g, t, func(config *rspec.Spec, t *tap.T, state *rspec.State) error {
			cg, err := cgroups.FindCgroup()
			if err != nil {
				return err
			}
			lhd, err := cg.GetHugepageLimitData(state.Pid, config.Linux.CgroupsPath)
			if err != nil {
				return err
			}
			for _, lhl := range lhd {
				if lhl.Pagesize != pageSize {
					continue
				}
				t.Ok(lhl.Limit == limit, fmt.Sprintf("hugepage limit is set correctly for size: %s", pageSize))
				t.Diagnosticf("expect: %d, actual: %d", limit, lhl.Limit)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func testWrongHugetlb() error {
	// We deliberately set the page size to a wrong value, "3MB", to see
	// if the container really returns an error. Page sizes will always be a
	// on the format 2^(integer)
	page := "3MB"
	var limit uint64 = 100 * 3 * 1024 * 1024

	g, err := util.GetDefaultGenerator()
	if err != nil {
		return err
	}

	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
	g.AddLinuxResourcesHugepageLimit(page, limit)

	err = util.RuntimeOutsideValidate(g, t, func(config *rspec.Spec, t *tap.T, state *rspec.State) error {
		return nil
	})
	t.Ok(err != nil, "hugepage invalid pagesize results in an errror")
	if err == nil {
		t.Diagnosticf("expect: err != nil, actual: err == nil")
	}
	return err
}

func main() {
	if "linux" != runtime.GOOS {
		util.Fatal(fmt.Errorf("linux-specific cgroup test"))
	}

	if err := testHugetlbCgroups(); err != nil {
		util.Fatal(err)
	}

	if err := testWrongHugetlb(); err == nil {
		util.Fatal(err)
	}
}
