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

func main() {
	spec, rspec, err := loadSpecConfig()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %q", err)
	}
	logrus.Infof("%+v", spec)
	logrus.Infof("%+v", rspec)
}
