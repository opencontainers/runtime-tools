package main

import (
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g := util.GetDefaultGenerator()
	g.AddOrReplaceLinuxNamespace("user", "")
	g.AddLinuxGIDMapping(uint32(1000), uint32(0), uint32(3200))
	err := util.RuntimeInsideValidate(g, nil)
	if err != nil {
		util.Fatal(err)
	}
}
