package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/ocitools/cmd/runtimetest/mount"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/syndtr/gocapability/capability"
	"github.com/urfave/cli"
)

// PR_GET_NO_NEW_PRIVS isn't exposed in Golang so we define it ourselves copying the value from
// the kernel
const PR_GET_NO_NEW_PRIVS = 39

var (
	defaultFS = map[string]string{
		"/proc":    "proc",
		"/sys":     "sysfs",
		"/dev/pts": "devpts",
		"/dev/shm": "tmpfs",
	}

	defaultDevices = []string{
		"/dev/null",
		"/dev/zero",
		"/dev/full",
		"/dev/random",
		"/dev/urandom",
		"/dev/tty",
		"/dev/console",
		"/dev/ptmx",
	}
)

type validation func(*rspec.Spec) error

func loadSpecConfig() (spec *rspec.Spec, err error) {
	cPath := "config.json"
	cf, err := os.Open(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config.json not found")
		}
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		return
	}
	return spec, nil
}

func validateProcess(spec *rspec.Spec) error {
	logrus.Debugf("validating container process")
	uid := os.Getuid()
	if uint32(uid) != spec.Process.User.UID {
		return fmt.Errorf("UID expected: %v, actual: %v", spec.Process.User.UID, uid)
	}
	gid := os.Getgid()
	if uint32(gid) != spec.Process.User.GID {
		return fmt.Errorf("GID expected: %v, actual: %v", spec.Process.User.GID, gid)
	}

	groups, err := os.Getgroups()
	if err != nil {
		return err
	}

	groupsMap := make(map[int]bool)
	for _, g := range groups {
		groupsMap[g] = true
	}

	for _, g := range spec.Process.User.AdditionalGids {
		if !groupsMap[int(g)] {
			return fmt.Errorf("Groups expected: %v, actual (should be superset): %v", spec.Process.User.AdditionalGids, groups)
		}
	}

	if spec.Process.Cwd != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if cwd != spec.Process.Cwd {
			return fmt.Errorf("Cwd expected: %v, actual: %v", spec.Process.Cwd, cwd)
		}
	}

	cmdlineBytes, err := ioutil.ReadFile("/proc/1/cmdline")
	if err != nil {
		return err
	}

	args := bytes.Split(bytes.Trim(cmdlineBytes, "\x00"), []byte("\x00"))
	if len(args) != len(spec.Process.Args) {
		return fmt.Errorf("Process arguments expected: %v, actual: %v", len(spec.Process.Args), len(args))
	}
	for i, a := range args {
		if string(a) != spec.Process.Args[i] {
			return fmt.Errorf("Process arguments expected: %v, actual: %v", string(a), spec.Process.Args[i])
		}
	}

	for _, env := range spec.Process.Env {
		parts := strings.Split(env, "=")
		key := parts[0]
		expectedValue := parts[1]
		actualValue := os.Getenv(key)
		if actualValue != expectedValue {
			return fmt.Errorf("Env %v expected: %v, actual: %v", key, expectedValue, actualValue)
		}
	}

	ret, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0, 0)
	if errno != 0 {
		return errno
	}
	if spec.Process.NoNewPrivileges && ret != 1 {
		return fmt.Errorf("NoNewPrivileges expected: true, actual: false")
	}
	if !spec.Process.NoNewPrivileges && ret != 0 {
		return fmt.Errorf("NoNewPrivileges expected: false, actual: true")
	}

	return nil
}

func validateCapabilities(spec *rspec.Spec) error {
	logrus.Debugf("validating capabilities")

	last := capability.CAP_LAST_CAP
	// workaround for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}

	processCaps, err := capability.NewPid(1)
	if err != nil {
		return err
	}

	expectedCaps := make(map[string]bool)
	for _, ec := range spec.Process.Capabilities {
		expectedCaps[ec] = true
	}

	for _, cap := range capability.List() {
		if cap > last {
			continue
		}

		capKey := fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String()))
		expectedSet := expectedCaps[capKey]
		actuallySet := processCaps.Get(capability.EFFECTIVE, cap)
		if expectedSet != actuallySet {
			if expectedSet {
				return fmt.Errorf("Expected Capability %v not set for process", cap.String())
			}
			return fmt.Errorf("Unexpected Capability %v set for process", cap.String())
		}
	}

	return nil
}

func validateHostname(spec *rspec.Spec) error {
	logrus.Debugf("validating hostname")
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if spec.Hostname != "" && hostname != spec.Hostname {
		return fmt.Errorf("Hostname expected: %v, actual: %v", spec.Hostname, hostname)
	}
	return nil
}

func validateRlimits(spec *rspec.Spec) error {
	logrus.Debugf("validating rlimits")
	for _, r := range spec.Process.Rlimits {
		rl, err := strToRlimit(r.Type)
		if err != nil {
			return err
		}

		var rlimit syscall.Rlimit
		if err := syscall.Getrlimit(rl, &rlimit); err != nil {
			return err
		}

		if rlimit.Cur != r.Soft {
			return fmt.Errorf("%v rlimit soft expected: %v, actual: %v", r.Type, r.Soft, rlimit.Cur)
		}
		if rlimit.Max != r.Hard {
			return fmt.Errorf("%v rlimit hard expected: %v, actual: %v", r.Type, r.Hard, rlimit.Max)
		}
	}
	return nil
}

func validateSysctls(spec *rspec.Spec) error {
	logrus.Debugf("validating sysctls")
	for k, v := range spec.Linux.Sysctl {
		keyPath := filepath.Join("/proc/sys", strings.Replace(k, ".", "/", -1))
		vBytes, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return err
		}
		value := strings.TrimSpace(string(bytes.Trim(vBytes, "\x00")))
		if value != v {
			return fmt.Errorf("Sysctl %v value expected: %v, actual: %v", k, v, value)
		}
	}
	return nil
}

func testWriteAccess(path string) error {
	tmpfile, err := ioutil.TempFile(path, "Test")
	if err != nil {
		return err
	}

	tmpfile.Close()
	os.RemoveAll(filepath.Join(path, tmpfile.Name()))

	return nil
}

func validateRootFS(spec *rspec.Spec) error {
	logrus.Debugf("validating root filesystem")
	if spec.Root.Readonly {
		err := testWriteAccess("/")
		if err == nil {
			return fmt.Errorf("Rootfs should be readonly")
		}
	}

	return nil
}

func validateDefaultFS(spec *rspec.Spec) error {
	logrus.Debugf("validating linux default filesystem")

	mountInfos, err := mount.GetMounts()
	if err != nil {
		return err
	}

	mountsMap := make(map[string]string)
	for _, mountInfo := range mountInfos {
		mountsMap[mountInfo.Mountpoint] = mountInfo.Fstype
	}

	for fs, fstype := range defaultFS {
		if !(mountsMap[fs] == fstype) {
			return fmt.Errorf("%v must exist and expected type is %v", fs, fstype)
		}
	}

	return nil
}

func validateDefaultDevices(spec *rspec.Spec) error {
	logrus.Debugf("validating linux default devices")

	for _, device := range defaultDevices {
		fi, err := os.Stat(device)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("device node %v not found", device)
			}
			return err
		} else {
			if fi.Mode()&os.ModeDevice != os.ModeDevice {
				return fmt.Errorf("file %v is not a device as expected", device)
			}
		}
	}

	return nil
}

func validateMaskedPaths(spec *rspec.Spec) error {
	logrus.Debugf("validating maskedPaths")
	for _, maskedPath := range spec.Linux.MaskedPaths {
		f, err := os.Open(maskedPath)
		if err != nil {
			return err
		}
		defer f.Close()
		b := make([]byte, 1)
		_, err = f.Read(b)
		if err != io.EOF {
			return fmt.Errorf("%v should not be readable", maskedPath)
		}
	}
	return nil
}

func validateROPaths(spec *rspec.Spec) error {
	logrus.Debugf("validating readonlyPaths")
	for _, v := range spec.Linux.ReadonlyPaths {
		err := testWriteAccess(v)
		if err == nil {
			return fmt.Errorf("%v should be readonly", v)
		}
	}

	return nil
}

func mountMatch(specMount rspec.Mount, sysMount rspec.Mount) error {
	if specMount.Destination != sysMount.Destination {
		return fmt.Errorf("mount destination expected: %v, actual: %v", specMount.Destination, sysMount.Destination)
	}

	if specMount.Type != sysMount.Type {
		return fmt.Errorf("mount %v type expected: %v, actual: %v", specMount.Destination, specMount.Type, sysMount.Type)
	}

	if specMount.Source != sysMount.Source {
		return fmt.Errorf("mount %v source expected: %v, actual: %v", specMount.Destination, specMount.Source, sysMount.Source)
	}

	return nil
}

func validateMountsExist(spec *rspec.Spec) error {
	logrus.Debugf("validating mounts exist")

	mountInfos, err := mount.GetMounts()
	if err != nil {
		return err
	}

	mountsMap := make(map[string][]rspec.Mount)
	for _, mountInfo := range mountInfos {
		m := rspec.Mount{
			Destination: mountInfo.Mountpoint,
			Type:        mountInfo.Fstype,
			Source:      mountInfo.Source,
		}
		mountsMap[mountInfo.Mountpoint] = append(mountsMap[mountInfo.Mountpoint], m)
	}

	for _, specMount := range spec.Mounts {
		found := false
		for _, sysMount := range mountsMap[specMount.Destination] {
			if err := mountMatch(specMount, sysMount); err == nil {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("Expected mount %v does not exist", specMount)
		}
	}

	return nil
}

func validate(context *cli.Context) error {
	logLevelString := context.String("log-level")
	logLevel, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		return err
	}
	logrus.SetLevel(logLevel)

	spec, err := loadSpecConfig()
	if err != nil {
		return err
	}

	defaultValidations := []validation{
		validateRootFS,
		validateProcess,
		validateCapabilities,
		validateHostname,
		validateRlimits,
		validateMountsExist,
	}

	linuxValidations := []validation{
		validateDefaultFS,
		validateDefaultDevices,
		validateSysctls,
		validateMaskedPaths,
		validateROPaths,
	}

	for _, v := range defaultValidations {
		if err := v(spec); err != nil {
			return err
		}
	}

	if spec.Platform.OS == "linux" {
		for _, v := range linuxValidations {
			if err := v(spec); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "runtimetest"
	app.Version = "0.0.1"
	app.Usage = "Compare the environment with an OCI configuration"
	app.UsageText = "runtimetest [options]"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "error",
			Usage: "Log level (panic, fatal, error, warn, info, or debug)",
		},
	}

	app.Action = validate
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
