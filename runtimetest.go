package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
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

		logrus.Info("Start to test runtime operation...")
		if _, err := testOperation(runtime); err != nil {
			os.RemoveAll(TestCacheDir)
			logrus.Fatal(err)
		}
		logrus.Info("Runtime operation test succeeded.")

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

func testOperation(runtime string) (string, error) {
	testRunningConfig := getDefaultConfig()
	testRunningConfig.Process.Args = []string{"sleep", "60"}
	runningUnit := TestUnit{
		Name:    "running",
		Runtime: runtime,
		Config:  testRunningConfig,
	}
	if _, err := runningUnit.GetState(); err == nil {
		return "", ErrStateWithoutID
	}

	runningID := GetFreeUUID(runtime)
	runningUnit.ID = runningID
	// Start a running container (terminated in 60s)
	go func() {
		runningUnit.Prepare()
		runningUnit.Start()
	}()
	var state rspec.State
	var err error
	for t := time.Now(); time.Since(t) < time.Minute; time.Sleep(time.Second * 5) {
		if state, err = runningUnit.GetState(); err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	defer runningUnit.Stop()
	if err := checkState(state, runningUnit); err != nil {
		return "", err
	}

	type testOperationUnit struct {
		Unit        TestUnit
		prepare     bool
		expectedErr error
	}

	testConfig := getDefaultConfig()
	testConfig.Process.Args = []string{"true"}
	startOperUnits := []testOperationUnit{
		{Unit: TestUnit{Name: "start-with-dup-id", Runtime: runtime, Config: testConfig, ID: runningID}, prepare: true, expectedErr: ErrStartWithDupID},
		{Unit: TestUnit{Name: "start-without-id", Runtime: runtime, Config: testConfig}, prepare: true, expectedErr: ErrStartWithoutID},
		{Unit: TestUnit{Name: "start-without-bundle", Runtime: runtime, Config: testConfig, ID: GetFreeUUID(runtime)}, prepare: false, expectedErr: ErrStartWithoutBundle},
	}
	for _, operUnit := range startOperUnits {
		if operUnit.prepare {
			operUnit.Unit.Prepare()
		}
		err := operUnit.Unit.Start()
		defer operUnit.Unit.Stop()
		if err != nil && operUnit.expectedErr == nil {
			return "", err
		} else if err == nil && operUnit.expectedErr != nil {
			return "", operUnit.expectedErr
		}
	}
	return "", nil
}

func testLifecycle(runtime string) (string, error) {
	OKHooks := []rspec.Hook{{Path: "/bin/true", Args: []string{"true"}}}
	FailHooks := []rspec.Hook{{Path: "/bin/false", Args: []string{"false"}}}

	processOutput := "hello, ocitools"
	allOK := getDefaultConfig()
	allOK.Process.Args = []string{"echo", processOutput}
	allOK.Hooks.Prestart = OKHooks
	allOK.Hooks.Poststart = OKHooks
	allOK.Hooks.Poststop = OKHooks
	allOKUnit := TestUnit{
		Name:    "allOK",
		Runtime: runtime,
		Config:  allOK,
		ID:      GetFreeUUID(runtime),
	}
	allOKUnit.Prepare()
	allOKUnit.Start()
	defer allOKUnit.Stop()
	if output, err := allOKUnit.GetOutput(); err != nil {
		return output, err
	} else if processOutput != strings.TrimSpace(output) {
		return "", fmt.Errorf("Failed to run 'Process' successfully")
	}

	prestartFailed := allOK
	prestartFailed.Hooks.Prestart = FailHooks
	poststartFailed := allOK
	poststartFailed.Hooks.Poststart = FailHooks
	poststopFailed := allOK
	poststopFailed.Hooks.Poststop = FailHooks
	hookFailedUnits := []TestUnit{
		{Name: "prestart", Runtime: runtime, Config: prestartFailed, ID: GetFreeUUID(runtime)},
		{Name: "poststart", Runtime: runtime, Config: poststartFailed, ID: GetFreeUUID(runtime)},
		{Name: "poststop", Runtime: runtime, Config: poststopFailed, ID: GetFreeUUID(runtime)},
	}
	for _, unit := range hookFailedUnits {
		unit.Prepare()
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
		Name:    "configs",
		Runtime: runtime,
		Config:  testConfig,
		ID:      GetFreeUUID(runtime),
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

func checkState(state rspec.State, unit TestUnit) error {
	if state.ID != unit.ID {
		return fmt.Errorf("Expected container ID: %s to match: %s", state.ID, unit.ID)
	}
	if state.BundlePath != unit.GetBundlePath() {
		return fmt.Errorf("Expected container bundle path: %s to match: %s", state.BundlePath, unit.GetBundlePath())
	}
	return nil
}
