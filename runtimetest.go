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

var runtimetestFlags = []cli.Flag{
	cli.StringFlag{Name: "runtime, r", Usage: "runtime to be tested"},
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

		if err := os.MkdirAll(TestCacheDir, os.ModePerm); err != nil {
			logrus.Fatalf("Failed to create cache dir: %s", TestCacheDir)
		}
		defer os.RemoveAll(TestCacheDir)

		logrus.Info("Start to test runtime lifecycle...")
		if _, err := testLifecycle(runtime); err != nil {
			os.RemoveAll(TestCacheDir)
			logrus.Fatal(err)
		}
		logrus.Info("Runtime lifecycle test succeeded.")

		logrus.Info("Start to test runtime state...")
		if _, err := testState(runtime); err != nil {
			os.RemoveAll(TestCacheDir)
			logrus.Fatal(err)
		}
		logrus.Info("Runtime state test succeeded.")

		logrus.Info("Start to test runtime main config...")
		if output, err := testMainConfigs(runtime); err != nil {
			os.RemoveAll(TestCacheDir)
			logrus.Info(output)
			logrus.Fatal(err)
		} else if output != "" {
			logrus.Info(output)
		}
		logrus.Info("Runtime main config test succeeded.")

	},
}

func testState(runtime string) (string, error) {
	testConfig := getDefaultConfig()
	testConfig.Process.Args = []string{"sleep", "60"}
	testID := GetFreeUUID(runtime)
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
		return "", fmt.Errorf("Expected container ID: %s to match: %s", state.ID, testID)
	}
	if state.BundlePath != unit.GetBundlePath() {
		return "", fmt.Errorf("Expected container bundle path: %s to match: %s", state.BundlePath, unit.GetBundlePath())
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
		return output, errors.New("Expected to get an error when start with a duplicated container ID")
	}
}

func testLifecycle(runtime string) (string, error) {
	OKArgs := []string{"true"}
	OKHooks := []rspec.Hook{{Path: "/bin/true", Args: []string{"true"}}}
	FailHooks := []rspec.Hook{{Path: "/bin/false", Args: []string{"false"}}}

	allOK := getDefaultConfig()
	allOK.Process.Args = OKArgs
	allOK.Hooks.Prestart = OKHooks
	allOK.Hooks.Poststart = OKHooks
	allOK.Hooks.Poststop = OKHooks
	allOKUnit := TestUnit{
		Name:    "allOK",
		Runtime: runtime,
		Config:  allOK,
	}
	allOKUnit.Start()
	defer allOKUnit.Stop()
	if output, err := allOKUnit.GetOutput(); err != nil {
		return output, err
	}

	prestartFailed := allOK
	prestartFailed.Hooks.Prestart = FailHooks
	poststartFailed := allOK
	poststartFailed.Hooks.Poststart = FailHooks
	poststopFailed := allOK
	poststopFailed.Hooks.Poststop = FailHooks
	hookFailedUnit := []TestUnit{
		{Name: "prestart", Runtime: runtime, Config: prestartFailed},
		{Name: "poststart", Runtime: runtime, Config: poststartFailed},
		{Name: "poststop", Runtime: runtime, Config: poststopFailed},
	}
	for _, unit := range hookFailedUnit {
		unit.Start()
		defer unit.Stop()
		if output, err := unit.GetOutput(); err == nil {
			return output, fmt.Errorf("Expected to get an error when %s fails", unit.Name)
		}
	}

	return "", nil
}

func testMainConfigs(runtime string) (string, error) {
	testConfig := getDefaultConfig()
	testConfig.Process.Args = []string{"./runtimetest"}

	defaultUnit := TestUnit{
		Name:           "configs",
		Runtime:        runtime,
		Config:         testConfig,
		ExpectedResult: true,
	}

	defaultUnit.Prepare()
	defer defaultUnit.Clean()

	// Copy runtimtest from plugins to rootfs
	src := "./runtimetest"
	dest := path.Join(defaultUnit.GetBundlePath(), "rootfs", "runtimetest")
	if err := copyFile(dest, src); err != nil {
		return "", fmt.Errorf("Failed to copy '%s' to '%s': %v\n", src, dest, err)
	}
	if err := os.Chmod(dest, os.ModePerm); err != nil {
		return "", fmt.Errorf("Failed to chmod runtimetest: %v\n", err)
	}

	src = path.Join(defaultUnit.GetBundlePath(), configFile)
	dest = path.Join(defaultUnit.GetBundlePath(), "rootfs", configFile)
	if err := copyFile(dest, src); err != nil {
		return "", fmt.Errorf("Failed to copy '%s' to '%s': %v\n", src, dest, err)
	}

	defaultUnit.Start()
	defer defaultUnit.Stop()
	if output, err := defaultUnit.GetOutput(); err != nil {
		return output, fmt.Errorf("Failed to test main config '%s' case: %v", defaultUnit.Name, err)
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
