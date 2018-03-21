package main

import (
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g := util.GetDefaultGenerator()
	g.SetProcessUID(10)
	g.SetProcessGID(10)
	g.AddProcessAdditionalGid(5)

	err := util.RuntimeInsideValidate(g, nil)
	if err != nil {
		util.Fatal(err)
	}
}
