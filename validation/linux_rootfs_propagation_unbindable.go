package main

import (
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetupPrivileged(true)
	g.SetLinuxRootPropagation("unbindable")
	err = util.RuntimeInsideValidate(g, nil)
	if err != nil {
		util.Fatal(err)
	}
}
