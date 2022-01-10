package main

import (
	"github.com/mndrix/tap-go"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func testLinuxRootPropagation(t *tap.T, propMode string) error {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	// Test case validateRootfsPropagation needs CAP_SYS_ADMIN to perform mounts.
	g.AddProcessCapability("CAP_SYS_ADMIN")
	// The generated seccomp profile does not enable mount/umount/umount2 syscalls.
	g.Config.Linux.Seccomp = nil

	g.SetLinuxRootPropagation(propMode)
	g.AddAnnotation("TestName", "check root propagation: "+propMode)
	return util.RuntimeInsideValidate(g, t, nil)
}

func main() {
	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	cases := []string{
		"shared",
		"slave",
		"private",
		"unbindable",
	}

	for _, c := range cases {
		if err := testLinuxRootPropagation(t, c); err != nil {
			t.Fail(err.Error())
		}
	}
}
