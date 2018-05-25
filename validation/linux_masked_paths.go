package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/opencontainers/runtime-tools/validation/util"
)

func checkMaskedPaths() error {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		return err
	}

	maskedDir := "masked-dir"
	maskedSubDir := "masked-subdir"
	maskedFile := "masked-file"

	maskedDirTop := filepath.Join("/", maskedDir)
	maskedFileTop := filepath.Join("/", maskedFile)

	maskedDirSub := filepath.Join(maskedDirTop, maskedSubDir)
	maskedFileSub := filepath.Join(maskedDirTop, maskedFile)
	maskedFileSubSub := filepath.Join(maskedDirSub, maskedFile)

	g.AddLinuxMaskedPaths(maskedDirTop)
	g.AddLinuxMaskedPaths(maskedFileTop)
	g.AddLinuxMaskedPaths(maskedDirSub)
	g.AddLinuxMaskedPaths(maskedFileSub)
	g.AddLinuxMaskedPaths(maskedFileSubSub)
	err = util.RuntimeInsideValidate(g, func(path string) error {
		testDir := filepath.Join(path, maskedDirSub)
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

		// runtimetest cannot check the readability of empty files, so
		// write something.
		testSubSubFile := filepath.Join(path, maskedFileSubSub)
		if err := ioutil.WriteFile(testSubSubFile, []byte("secrets"), 0777); err != nil {
			return err
		}

		testSubFile := filepath.Join(path, maskedFileSub)
		if err := ioutil.WriteFile(testSubFile, []byte("secrets"), 0777); err != nil {
			return err
		}

		testFile := filepath.Join(path, maskedFile)
		return ioutil.WriteFile(testFile, []byte("secrets"), 0777)
	})
	return err
}

func checkMaskedRelPaths() error {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		return err
	}

	// Deliberately set a relative path to be masked, and expect an error
	maskedRelPath := "masked-relpath"

	g.AddLinuxMaskedPaths(maskedRelPath)
	err = util.RuntimeInsideValidate(g, func(path string) error {
		testFile := filepath.Join(path, maskedRelPath)
		if _, err := os.Stat(testFile); err != nil && os.IsNotExist(err) {
			return err
		}

		return nil
	})
	if err != nil {
		return nil
	}
	return fmt.Errorf("expected: err != nil, actual: err == nil")
}

func main() {
	if err := checkMaskedPaths(); err != nil {
		util.Fatal(err)
	}

	if err := checkMaskedRelPaths(); err != nil {
		util.Fatal(err)
	}
}
