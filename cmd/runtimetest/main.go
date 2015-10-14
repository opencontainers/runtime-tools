package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/specs"
	"github.com/syndtr/gocapability/capability"
)

func loadSpecConfig() (spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, err error) {
	cPath := "config.json"
	cf, err := os.Open(cPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("config.json not found")
		}
	}
	defer cf.Close()

	rPath := "runtime.json"
	rf, err := os.Open(rPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("runtime.json not found")
		}
	}
	defer rf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		return
	}
	if err = json.NewDecoder(rf).Decode(&rspec); err != nil {
		return
	}
	return spec, rspec, nil
}

func validateProcess(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec) error {
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

	if len(groups) != len(spec.Process.User.AdditionalGids) {
		return fmt.Errorf("Groups expected: %v, actual: %v", spec.Process.User.AdditionalGids, groups)
	}

	groupsMap := make(map[int]bool)
	for _, g := range spec.Process.User.AdditionalGids {
		groupsMap[int(g)] = true
	}

	for _, g := range groups {
		if !groupsMap[g] {
			return fmt.Errorf("Groups expected: %v, actual: %v", spec.Process.User.AdditionalGids, groups)
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
		return fmt.Errorf("Process arguments expected: %v, actual: %v")
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
			return fmt.Errorf("Env %v expected: %v, actual: %v", expectedValue, actualValue)
		}
	}

	return nil
}

func validateCapabilities(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec) error {
	capabilityMap := make(map[string]capability.Cap)
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
	}

	processCaps, err := capability.NewPid(1)
	if err != nil {
		return err
	}

	for _, ec := range spec.Linux.Capabilities {
		cap := capabilityMap[ec]
		if !processCaps.Get(capability.EFFECTIVE, cap) {
			return fmt.Errorf("Expected Capability %v not set for process")
		}
	}

	//TODO check that unexpected caps aren't set

	return nil
}

func main() {
	spec, rspec, err := loadSpecConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %q", err)
	}
	if err := validateProcess(spec, rspec); err != nil {
		logrus.Fatalf("Validation failed: %q", err)
	}
	if err := validateCapabilities(spec, rspec); err != nil {
		logrus.Fatalf("Validation failed: %q", err)
	}
}
