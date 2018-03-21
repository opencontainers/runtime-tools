package main

import (
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetProcessOOMScoreAdj(500)
	g.SetProcessNoNewPrivileges(true)
	g.SetupPrivileged(true)
	g.SetProcessApparmorProfile("acme_secure_profile")
	g.SetProcessSelinuxLabel("system_u:system_r:svirt_lxc_net_t:s0:c124,c675")

	err = util.RuntimeInsideValidate(g, nil)
	if err != nil {
		util.Fatal(err)
	}
}
