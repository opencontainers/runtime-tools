package main

import (
	"os"
	"runtime"

	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	if runtime.GOOS != "linux" && runtime.GOOS != "solaris" {
		util.Skip("POSIX-specific process.rlimits test", map[string]string{"OS": runtime.GOOS})
		os.Exit(0)
	}

	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}

	large, err := util.LargeRlimitsSupported()
	if err != nil {
		util.Fatal(err)
	}

	var gigaBytes uint64 = 1024 * 1024 * 1024
	for _, r := range []struct {
		rlimit     string
		hard, soft uint64
	}{
		{"RLIMIT_AS", 2 * gigaBytes, 1 * gigaBytes},
		{"RLIMIT_CORE", 4 * gigaBytes, 3 * gigaBytes},
		{"RLIMIT_DATA", 6 * gigaBytes, 5 * gigaBytes},
		{"RLIMIT_FSIZE", 8 * gigaBytes, 7 * gigaBytes},
		{"RLIMIT_STACK", 10 * gigaBytes, 9 * gigaBytes},
		{"RLIMIT_CPU", 120, 60},       // seconds
		{"RLIMIT_NOFILE", 4000, 3000}, // number of files
	} {
		if !large && r.hard >= 4*gigaBytes {
			continue
		}
		g.AddProcessRlimits(r.rlimit, r.hard, r.soft)
	}

	err = util.RuntimeInsideValidate(g, nil, nil)
	if err != nil {
		util.Fatal(err)
	}
}
