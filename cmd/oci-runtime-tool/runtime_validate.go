package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mrunalp/fileutils"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/satori/go.uuid"
	"github.com/urfave/cli"
)

var runtimeValidateFlags = []cli.Flag{
	cli.StringFlag{Name: "runtime", Value: "runc", Usage: "OCI runtime"},
}

var runtimeValidateCommand = cli.Command{
	Name:   "runtime-validate",
	Usage:  "validate an OCI runtime",
	Flags:  runtimeValidateFlags,
	Before: before,
	Action: func(context *cli.Context) error {
		return runtimeValidate(context.String("runtime"))
	},
}

func runtimeValidate(runtime string) error {
	// Find the runtime binary in the PATH
	runtimePath, err := exec.LookPath(runtime)
	if err != nil {
		return err
	}

	// Setup a temporary test directory
	tmpDir, err := ioutil.TempDir("", "ocitest")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Create bundle directory for the test container
	bundleDir := tmpDir + "/busybox"
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return err
	}

	// TODO: Use go package for untar and allow using other root filesystems
	// Untar the root fs
	untarCmd := exec.Command("tar", "-xf", "rootfs.tar.gz", "-C", bundleDir)
	output, err := untarCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
	}

	// Copy the runtimetest binary to the rootfs
	err = fileutils.CopyFile("runtimetest", filepath.Join(bundleDir, "runtimetest"))

	// Generate test configuration
	g := generate.New()
	g.SetRootPath(".")
	g.SetProcessArgs([]string{"/runtimetest"})
	err = g.SaveToFile(filepath.Join(bundleDir, "config.json"), generate.ExportOptions{})
	if err != nil {
		return err
	}

	// TODO: Use a library to split run into create/start
	// Launch the OCI runtime
	containerID := uuid.NewV4()
	runtimeCmd := exec.Command(runtimePath, "run", containerID.String())
	runtimeCmd.Dir = bundleDir
	runtimeCmd.Stdin = os.Stdin
	runtimeCmd.Stdout = os.Stdout
	runtimeCmd.Stderr = os.Stderr
	if err = runtimeCmd.Run(); err != nil {
		return err
	}

	return nil
}
