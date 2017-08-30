package validation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/mndrix/tap-go"
	"github.com/mrunalp/fileutils"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/satori/go.uuid"
)

var (
	runtime = "runc"
)

type validation struct {
	test        func(string, string, *rspec.Spec) error
	description string
}

func init() {
	runtimeInEnv := os.Getenv("RUNTIME")
	if runtimeInEnv != "" {
		runtime = runtimeInEnv
	}
}

func TestValidateRuntimeInside(t *testing.T) {
	g, err := getDefaultGenerator()
	if err != nil {
		t.Errorf("%s failed validation: %v", runtime, err)
	}
	g.SetProcessArgs([]string{"/runtimetest"})

	if err := runtimeInsideValidate(runtime, g); err != nil {
		t.Errorf("%s failed validation: %v", runtime, err)
	}
}

func TestValidateRuntimeOutside(t *testing.T) {
	g, err := getDefaultGenerator()
	if err != nil {
		t.Errorf("%s failed validation: %v", runtime, err)
	}

	if err := runtimeOutsideValidate(runtime, g); err != nil {
		t.Errorf("%s failed validation: %v", runtime, err)
	}
}

func runtimeInsideValidate(runtime string, g *generate.Generator) error {
	// Find the runtime binary in the PATH
	runtimePath, err := exec.LookPath(runtime)
	if err != nil {
		return err
	}

	bundleDir, rootfsDir, err := prepareBundle(g)
	if err != nil {
		return err
	}
	defer os.RemoveAll(bundleDir)

	// Copy the runtimetest binary to the rootfs
	err = fileutils.CopyFile("../runtimetest", filepath.Join(rootfsDir, "runtimetest"))
	if err != nil {
		return err
	}

	// TODO: Use a library to split run into create/start
	// Launch the OCI runtime
	containerID := uuid.NewV4()
	runtimeCmd := exec.Command(runtimePath, "run", containerID.String())
	runtimeCmd.Dir = bundleDir
	runtimeCmd.Stdin = os.Stdin
	runtimeCmd.Stdout = os.Stdout
	runtimeCmd.Stderr = os.Stderr
	if err = runtimeCmd.Run(); err != nil {
		return err
	}

	return nil
}

func runtimeOutsideValidate(runtime string, g *generate.Generator) error {
	// Find the runtime binary in the PATH
	runtimePath, err := exec.LookPath(runtime)
	if err != nil {
		return err
	}

	bundleDir, _, err := prepareBundle(g)
	if err != nil {
		return err
	}
	defer os.RemoveAll(bundleDir)

	// Launch the OCI runtime
	containerID := uuid.NewV4()
	runtimeCmd := exec.Command(runtimePath, "create", containerID.String())
	runtimeCmd.Dir = bundleDir
	runtimeCmd.Stdin = os.Stdin
	runtimeCmd.Stdout = os.Stdout
	runtimeCmd.Stderr = os.Stderr
	if err = runtimeCmd.Run(); err != nil {
		return err
	}

	outsideValidations := []validation{
		{
			test:        validateLabels,
			description: "labels",
		},
		// Add more container outside validation
	}

	t := tap.New()
	t.Header(0)

	var validationErrors error
	for _, v := range outsideValidations {
		err := v.test(runtimePath, containerID.String(), g.Spec())
		t.Ok(err == nil, v.description)
		if err != nil {
			validationErrors = multierror.Append(validationErrors, err)
		}
	}
	t.AutoPlan()

	if err = cleanup(runtimePath, containerID.String()); err != nil {
		validationErrors = multierror.Append(validationErrors, err)
	}

	return validationErrors
}

func validateLabels(runtimePath, id string, spec *rspec.Spec) error {
	runtimeCmd := exec.Command(runtimePath, "state", id)
	output, err := runtimeCmd.Output()
	if err != nil {
		return err
	}

	var state rspec.State
	if err := json.NewDecoder(strings.NewReader(string(output))).Decode(&state); err != nil {
		return err
	}
	for key, value := range spec.Annotations {
		if state.Annotations[key] == value {
			continue
		}
		return fmt.Errorf("Expected annotation %s:%s not set", key, value)
	}
	return nil
}

func cleanup(runtimePath, id string) error {
	runtimeCmd := exec.Command(runtimePath, "kill", id, "KILL")
	if err := runtimeCmd.Run(); err != nil {
		return fmt.Errorf("Failed to kill container %s: %v", id, err)
	}

	runtimeCmd = exec.Command(runtimePath, "delete", id)
	if err := runtimeCmd.Run(); err != nil {
		return fmt.Errorf("Failed to kill container %s: %v", id, err)
	}

	return nil
}

func prepareBundle(g *generate.Generator) (string, string, error) {
	// Setup a temporary test directory
	tmpDir, err := ioutil.TempDir("", "ocitest")
	if err != nil {
		return "", "", err
	}

	// Create bundle directory for the test container
	bundleDir := tmpDir
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return "", "", err
	}

	// Create rootfs directory for the test container
	rootfsDir := bundleDir + "/rootfs"
	if err := os.MkdirAll(rootfsDir, 0755); err != nil {
		return "", "", err
	}

	// Untar the root fs
	untarCmd := exec.Command("tar", "-xf", "../rootfs.tar.gz", "-C", rootfsDir)
	output, err := untarCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return "", "", err
	}

	// Generate test configuration
	err = g.SaveToFile(filepath.Join(bundleDir, "config.json"), generate.ExportOptions{})
	if err != nil {
		return "", "", err
	}

	// Copy the configuration file to the rootfs
	err = fileutils.CopyFile(filepath.Join(bundleDir, "config.json"), filepath.Join(rootfsDir, "config.json"))
	if err != nil {
		return "", "", err
	}

	return bundleDir, rootfsDir, nil
}

func getDefaultGenerator() (*generate.Generator, error) {
	// Generate testcase template
	generateCmd := exec.Command("oci-runtime-tool", "generate", "--mount-bind=/tmp:/volume/testing:rw", "--linux-cgroups-path=/tmp/testcgroup", "--linux-device-add=c:80:500:/dev/test:fileMode=438", "--linux-disable-oom-kill=true", "--env=testvar=vartest", "--hostname=localvalidation", "--label=testlabel=nonevar", "--linux-cpu-shares=1024", "--output", "/tmp/config.json")
	output, err := generateCmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return nil, err
	}

	// Get testcase configuration
	g, err := generate.NewFromFile("/tmp/config.json")
	if err != nil {
		return nil, err
	}

	g.SetRootPath("rootfs")

	return &g, nil
}
