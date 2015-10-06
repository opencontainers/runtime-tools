package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/opencontainers/specs"
)

var validateFlags = []cli.Flag{
	cli.StringFlag{Name: "rootfs", Usage: "path to the rootfs containing 'cat'"},
	cli.StringFlag{Name: "runtime", Usage: "path to the OCI runtime"},
}

var validateCommand = cli.Command{
	Name:  "validate",
	Usage: "validate a OCI spec file",
	Flags: validateFlags,
	Action: func(context *cli.Context) {
		specDir, err := ioutil.TempDir("", "oci_test")
		if err != nil {
			logrus.Fatal(err)
		}

		if err = os.MkdirAll(specDir, 0700); err != nil {
			logrus.Fatal(err)
		}
		defer os.RemoveAll(specDir)

		rootfs := context.String("rootfs")
		if rootfs == "" {
			logrus.Fatalf("Rootfs path shouldn't be empty")
		}
		runtime := context.String("runtime")
		if runtime == "" {
			logrus.Fatalf("runtime path shouldn't be empty")
		}

		spec, rspec := getDefaultTemplate()
		if err != nil {
			logrus.Fatal(err)
		}

		spec.Process.Args = []string{"cat"}
		spec.Root.Path = rootfs
		cPath := filepath.Join(specDir, "config.json")
		rPath := filepath.Join(specDir, "runtime.json")
		data, err := json.MarshalIndent(&spec, "", "\t")
		if err != nil {
			logrus.Fatal(err)
		}
		if err := ioutil.WriteFile(cPath, data, 0666); err != nil {
			logrus.Fatal(err)
		}
		rdata, err := json.MarshalIndent(&rspec, "", "\t")
		if err != nil {
			logrus.Fatal(err)
		}
		if err := ioutil.WriteFile(rPath, rdata, 0666); err != nil {
			logrus.Fatal(err)
		}

		if err := testRuntime(runtime, specDir, spec, rspec); err != nil {
			logrus.Fatal(err)
		}
		logrus.Infof("Test succeeded.")
	},
}

func testRuntime(runtime string, specDir string, spec specs.LinuxSpec, rspec specs.LinuxRuntimeSpec) error {
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return err
	}

	logrus.Infof("Launcing runtime")
	cmd := exec.Command(runtime, "start")
	cmd.Dir = specDir
	cmd.Stdin = stdinR

	err = cmd.Start()
	if err != nil {
		return err
	}
	stdinR.Close()
	defer stdinW.Close()

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
