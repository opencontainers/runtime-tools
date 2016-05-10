package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

const bundleCacheDir = "./bundles"

var runtimetestFlags = []cli.Flag{
	cli.StringFlag{Name: "runtime, r", Usage: "runtime to be tested"},
	cli.BoolFlag{Name: "debug, d", Usage: "switch of debug mode, default to 'false', with '--debug' to enable debug mode"},
}

var runtimeTestCommand = cli.Command{
	Name:  "runtimetest",
	Usage: "test if a runtime is compliant to OCI Runtime Specification",
	Flags: runtimetestFlags,
	Action: func(context *cli.Context) {
		if os.Geteuid() != 0 {
			logrus.Fatalln("Should be run as 'root'")
		}
		var runtime string
		if runtime = context.String("runtime"); runtime != "runc" {
			logrus.Fatalf("'%s' is currently not supported", runtime)
		}

		if err := os.MkdirAll(bundleCacheDir, os.ModePerm); err != nil {
			logrus.Fatalf("Failed to create cache dir: %s", bundleCacheDir)
		}
		_, err := testState(runtime)
		if err != nil {
			os.RemoveAll(bundleCacheDir)
			logrus.Fatalf("\n%v", err)
		}
		logrus.Info("Runtime state test succeeded.")

		output, err := testMainConfigs(runtime)
		if err != nil {
			os.RemoveAll(bundleCacheDir)
			logrus.Infof("\n%s", output)
			logrus.Fatalf("\n%v", err)
		}
		if output != "" {
			logrus.Infof("\n%s", output)
		}
		logrus.Info("Runtime main config test succeeded.")

	},
}

func setDebugMode(debug bool) {
	if !debug {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func testState(runtime string) (string, error) {
	testConfig := getDefaultConfig()
	testConfig.Process.Args = []string{"sleep", "60"}
	//TODO: use UUID
	testID := "12345678"
	unit := TestUnit{
		Name:    "state",
		Runtime: runtime,
		Config:  testConfig,
		ID:      testID,
	}
	go func() {
		unit.Start()
	}()
	var state rspec.State
	var err error
	for t := time.Now(); time.Since(t) < time.Minute; time.Sleep(time.Second * 5) {
		if state, err = unit.GetState(); err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	defer unit.Stop()
	if state.ID != testID {
		return "", fmt.Errorf("Expect container ID: %s to match: %s", state.ID, testID)
	}
	if state.BundlePath != unit.GetBundlePath() {
		return "", fmt.Errorf("Expect container bundle path: %s to match: %s", state.BundlePath, unit.GetBundlePath())
	}

	unitDup := TestUnit{
		Name:    "state-dup",
		Runtime: runtime,
		Config:  testConfig,
		ID:      testID,
	}
	unitDup.Start()
	defer unitDup.Stop()
	// Expected to get error
	if output, err := unitDup.GetOutput(); err != nil {
		return output, nil
	} else {
		return output, errors.New("Failed to popup error with duplicated container ID")
	}
}

func testMainConfigs(runtime string) (string, error) {
	testConfig := getDefaultConfig()
	testConfig.Process.Args = []string{"./runtimetest"}
	testConfig.Hostname = "zenlin"

	hostnameUnit := TestUnit{
		Name:           "configs",
		Runtime:        runtime,
		Config:         testConfig,
		ExpectedResult: true,
	}

	hostnameUnit.Prepare()
	defer hostnameUnit.Clean()

	// Copy runtimtest from plugins to rootfs
	src := "./runtimetest"
	dest := path.Join(hostnameUnit.GetBundlePath(), "rootfs", "runtimetest")
	if err := copyFile(dest, src); err != nil {
		return "", fmt.Errorf("Failed to copy '%s' to '%s': %v\n", src, dest, err)
	}
	if err := os.Chmod(dest, os.ModePerm); err != nil {
		return "", fmt.Errorf("Failed to chmod runtimetest: %v\n", err)
	}

	src = path.Join(hostnameUnit.GetBundlePath(), configFile)
	dest = path.Join(hostnameUnit.GetBundlePath(), "rootfs", configFile)
	if err := copyFile(dest, src); err != nil {
		return "", fmt.Errorf("Failed to copy '%s' to '%s': %v\n", src, dest, err)
	}

	hostnameUnit.Start()
	defer hostnameUnit.Stop()
	if output, err := hostnameUnit.GetOutput(); err != nil {
		return output, fmt.Errorf("Failed to test main config '%s' case: %v", hostnameUnit.Name, err)
	} else {
		return output, nil
	}
}

func copyFile(dst string, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func getDefaultConfig() *rspec.Spec {
	config := GetDefaultTemplate()
	config.Root.Path = "rootfs"
	config.Platform.OS = runtime.GOOS
	config.Platform.Arch = runtime.GOARCH
	config.Process.Cwd = "/"

	return config
}
