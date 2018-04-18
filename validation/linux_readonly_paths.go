package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.AddLinuxReadonlyPaths("/readonly-dir")
	g.AddLinuxReadonlyPaths("/readonly-file")
	err = util.RuntimeInsideValidate(g, func(path string) error {
		testDir := filepath.Join(path, "readonly-dir")
		err = os.MkdirAll(testDir, 0777)
		if err != nil {
			return err
		}
		// create a temp file to make testDir non-empty
		tmpfile, err := ioutil.TempFile(testDir, "tmp")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())

		testFile := filepath.Join(path, "readonly-file")

		// runtimetest cannot check the readability of empty files, so
		// write something.
		return ioutil.WriteFile(testFile, []byte("immutable"), 0777)
	})
	if err != nil {
		util.Fatal(err)
	}
}
