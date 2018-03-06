package main

import (
	"os"
	"runtime"

	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	if "linux" != runtime.GOOS && "solaris" != runtime.GOOS {
		util.Skip("POSIX-specific process.rlimits test", map[string]string{"OS": runtime.GOOS})
		os.Exit(0)
	}

	g := util.GetDefaultGenerator()
	g.AddProcessRlimits("RLIMIT_NOFILE", 1024, 1024)
	err := util.RuntimeInsideValidate(g, nil)
	if err != nil {
		util.Fatal(err)
	}
}
