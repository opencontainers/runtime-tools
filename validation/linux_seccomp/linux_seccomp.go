package main

import (
	tap "github.com/mndrix/tap-go"
	"github.com/opencontainers/runtime-tools/generate/seccomp"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	syscallArgs := seccomp.SyscallOpts{
		Action:  "errno",
		Syscall: "getcwd",
	}
	g.SetDefaultSeccompAction("allow")
	g.SetSyscallAction(syscallArgs)
	err = util.RuntimeInsideValidate(g, t, nil)
	t.Ok(err == nil, "seccomp action is added correctly")
	if err != nil {
		t.Fail(err.Error())
	}
}
