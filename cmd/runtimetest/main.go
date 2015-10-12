package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/opencontainers/specs"
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
		return fmt.Errorf("UID expected: %q, actual: %q", spec.Process.User.UID, uid)
	}
	gid := os.Getgid()
	if uint32(gid) != spec.Process.User.GID {
		return fmt.Errorf("GID expected: %q, actual: %q", spec.Process.User.GID, gid)
	}

	if spec.Process.Cwd != "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		if wd != spec.Process.Cwd {
			return fmt.Errorf("Cwd expected: %q, actual: %q", spec.Process.Cwd, wd)
		}
	}

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
}
