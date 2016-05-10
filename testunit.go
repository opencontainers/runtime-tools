package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pborman/uuid"
)

const (
	TestCacheDir = "./bundles/"
	configFile   = "config.json"
	ociTools     = "ocitools"
	TEST_READY   = "test ready"
)

// TestUnit for storage testcase
type TestUnit struct {
	ID             string
	Name           string
	Runtime        string
	Config         *rspec.Spec
	ExpectedOutput string
	ExpectedResult bool

	output     string
	err        error
	bundlePath string
	ready      string
}

// Prepare a bundle for a test unit
func (unit *TestUnit) Prepare() error {
	if unit.ready == TEST_READY {
		return nil
	}
	if unit.Name == "" || unit.Runtime == "" || unit.Config == nil {
		return errors.New("Could not prepare a test unit which does not have 'Name', 'Runtime' or 'Config'.")
	}

	if err := unit.prepareBundle(); err != nil {
		return errors.New("Failed to prepare bundle")
	}

	unit.ready = TEST_READY
	return nil
}

// Clean the generated bundle of a test unit
func (unit *TestUnit) Clean() error {
	if unit.ready == TEST_READY {
		unit.ready = ""
		return os.RemoveAll(unit.bundlePath)
	}
	return nil
}

// Start a test unit
// Generate a bundle from unit's args and start it by unit's runtime
func (unit *TestUnit) Start() error {
	if unit.ready != TEST_READY {
		if err := unit.Prepare(); err != nil {
			return err
		}
	}

	if unit.ID == "" {
		unit.ID = GetFreeUUID(unit.Runtime)
	}

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	//FIXME: it is runc preferred.
	cmd := exec.Command(unit.Runtime, "start", "-b", unit.bundlePath, unit.ID)
	cmd.Stdin = os.Stdin
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		unit.err = errors.New(stderr.String())
	}
	unit.output = stdout.String()

	return unit.err
}

// Stop a test unit and remove the generated bundle
func (unit *TestUnit) Stop() ([]byte, error) {
	if unit.ID == "" {
		return nil, errors.New("Could not stop a test unit which does not have 'ID'.")
	}
	cmd := exec.Command(unit.Runtime, "stop", unit.ID)
	cmd.Stdin = os.Stdin
	output, err := cmd.CombinedOutput()

	unit.Clean()
	return output, err
}

// GetState return the state of a running test unit
func (unit *TestUnit) GetState() (rspec.State, error) {
	if unit.ID == "" {
		return rspec.State{}, errors.New("Could not get the state of a test unit which does not have 'ID'.")
	}
	cmd := exec.Command(unit.Runtime, "state", unit.ID)
	cmd.Stdin = os.Stdin
	output, err := cmd.CombinedOutput()

	var state rspec.State
	if err == nil {
		err = json.Unmarshal(output, &state)
	}
	return state, err
}

// GetBundlePath return the bundle of a test unit
func (unit *TestUnit) GetBundlePath() string {
	return unit.bundlePath
}

// IsPass checks whether a test is successful
func (unit *TestUnit) GetOutput() (string, error) {
	return unit.output, unit.err
}

func (unit *TestUnit) prepareBundle() error {
	// Create bundle follder
	cwd, _ := os.Getwd()
	unit.bundlePath = path.Join(cwd, TestCacheDir, unit.Name)
	if err := os.RemoveAll(unit.bundlePath); err != nil {
		return err
	}

	if err := os.Mkdir(unit.bundlePath, os.ModePerm); err != nil {
		return err
	}

	// Create rootfs for bundle
	rootfs := unit.bundlePath + "/rootfs"
	if err := untarRootfs(rootfs); err != nil {
		return err
	}

	data, err := json.MarshalIndent(&unit.Config, "", "\t")
	if err != nil {
		return err
	}
	cName := path.Join(unit.bundlePath, configFile)
	if err := ioutil.WriteFile(cName, data, 0666); err != nil {
		return err
	}

	return nil
}

func untarRootfs(rootfs string) error {
	// Create rootfs folder to bundle
	if err := os.Mkdir(rootfs, os.ModePerm); err != nil {
		return fmt.Errorf("Failed to create rootfs for bundle '%s': %v\n", rootfs, err)
	}

	cmd := exec.Command("tar", "-xf", "rootfs.tar.gz", "-C", rootfs)
	cmd.Dir = ""
	cmd.Stdin = os.Stdin
	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}
	return nil
}

// GetFreeUUID provides a free uuid
func GetFreeUUID(runtime string) string {
	id := uuid.NewUUID()

	unit := TestUnit{ID: id.String()}
	if _, err := unit.GetState(); err != nil {
		return id.String()
	} else {
		return GetFreeUUID(runtime)
	}
}
