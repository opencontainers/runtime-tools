package main

import (
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g := util.GetDefaultGenerator()
	g.AddLinuxMaskedPaths("/masktest")
	err := util.RuntimeInsideValidate(g, func(path string) error {
		pathName := filepath.Join(path, "masktest")
		return os.MkdirAll(pathName, 0700)
	})
	if err != nil {
		util.Fatal(err)
	}
}
