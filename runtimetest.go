package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/opencontainers/ocitools/units"
)

const bundleCacheDir = "./bundles"

var runtimetestFlags = []cli.Flag{
	cli.StringFlag{Name: "runtime, r", Usage: "runtime to be tested"},
	cli.StringFlag{Name: "output, o", Usage: "output format, \n" +
		"-o=all: ouput sucessful details and statics, -o=err-only: ouput failure details and statics"},
	cli.BoolFlag{Name: "debug, d", Usage: "switch of debug mode, defaults to false, with '--debug' to enable debug mode"},
}

var runtimeTestCommand = cli.Command{
	Name:  "runtimetest",
	Usage: "test if a runtime is comlpliant to oci specs",
	Flags: runtimetestFlags,
	Action: func(context *cli.Context) {

		if os.Geteuid() != 0 {
			logrus.Fatalln("runtimetest should be run as root")
		}
		var runtime string
		if runtime = context.String("runtime"); runtime != "runc" {
			logrus.Fatalf("runtimetest have not support %v\n", runtime)
		}
		output := context.String("output")
		setDebugMode(context.Bool("debug"))

		units.LoadTestUnits("./cases.conf")

		if err := os.MkdirAll(bundleCacheDir, os.ModePerm); err != nil {
			logrus.Printf("create cache dir for bundle cases err: %v\ns", bundleCacheDir)
			return
		}

		for _, tu := range *units.Units {
			testTask(tu, runtime)
		}

		units.OutputResult(output)

		if err := os.RemoveAll(bundleCacheDir); err != nil {
			logrus.Fatalf("remove cache dir of bundles %v err: %v\n", bundleCacheDir, err)
		}

		if err := os.Remove("./runtime.json"); err != nil {
			logrus.Fatalf("remove ./runtime.json err: %v\n", err)
		}

		if err := os.Remove("./config.json"); err != nil {
			logrus.Fatalf("remove ./config.json err: %v\n", err)
		}

	},
}

func setDebugMode(debug bool) {
	if !debug {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func testTask(unit *units.TestUnit, runtime string) {
	logrus.Debugf("test bundle name: %v, Test args: %v\n", unit.Name, unit.Args)
	if err := unit.SetRuntime(runtime); err != nil {
		logrus.Fatalf("failed to setup runtime %s , error: %v\n", runtime, err)
	} else {
		unit.Run()
	}
	return
}
