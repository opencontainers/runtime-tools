package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/syndtr/gocapability/capability"
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
	fmt.Println("validating container process")
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

	args := strings.Split(string(bytes.Trim(cmdlineBytes, "\x00")), " ")
	if len(args) != len(spec.Process.Args) {
		return fmt.Errorf("Process arguments expected: %v, actual: %v", len(spec.Process.Args), len(args))
	}
	for i, a := range args {
		if a != spec.Process.Args[i] {
			return fmt.Errorf("Process arguments expected: %v, actual: %v", a, spec.Process.Args[i])
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

	return nil
}

func validateCapabilities(spec *rspec.Spec) error {
	fmt.Println("validating capabilities")
	capabilityMap := make(map[string]capability.Cap)
	expectedCaps := make(map[capability.Cap]bool)
	last := capability.CAP_LAST_CAP
	// workaround for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}
	for _, cap := range capability.List() {
		if cap > last {
			continue
		}
		capKey := fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String()))
		capabilityMap[capKey] = cap
		expectedCaps[cap] = false
	}

	for _, ec := range spec.Process.Capabilities {
		cap := capabilityMap[ec]
		expectedCaps[cap] = true
	}

	processCaps, err := capability.NewPid(1)
	if err != nil {
		return err
	}

	for _, cap := range capability.List() {
		expectedSet := expectedCaps[cap]
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
	fmt.Println("validating hostname")
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
	fmt.Println("validating rlimits")
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
	fmt.Println("validating sysctls")
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
	fmt.Println("validating root")
	if spec.Root.Readonly {
		err := testWriteAccess("/")
		if err == nil {
			return fmt.Errorf("Rootfs should be readonly")
		}
	}

	return nil
}

func main() {
	spec, err := loadSpecConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %q", err)
	}

	validations := []validation{
		validateRootFS,
		validateProcess,
		validateCapabilities,
		validateHostname,
		validateRlimits,
		validateSysctls,
	}

	for _, v := range validations {
		if err := v(spec); err != nil {
			logrus.Fatalf("Validation failed: %q", err)
		}
	}
}
