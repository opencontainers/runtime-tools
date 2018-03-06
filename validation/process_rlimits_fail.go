package main

import (
	"fmt"

	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g := util.GetDefaultGenerator()
	g.AddProcessRlimits("RLIMIT_TEST", 1024, 1024)
	err := util.RuntimeInsideValidate(g, nil)
	if err == nil {
		util.Fatal(specerror.NewError(specerror.PosixProcRlimitsTypeGenError, fmt.Errorf("The runtime MUST generate an error for any values which cannot be mapped to a relevant kernel interface"), rspecs.Version))
	}
}
