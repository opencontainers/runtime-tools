package validation

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
)

// Runtime represents the basic requirement of a container runtime
type Runtime struct {
	RuntimeCommand string
	BundleDir      string
	ID             string
}

// NewRuntime create a runtime by command and the bundle directory
func NewRuntime(runtimeCommand string, bundleDir string) (Runtime, error) {
	var r Runtime
	var err error
	r.RuntimeCommand, err = exec.LookPath(runtimeCommand)
	if err != nil {
		return Runtime{}, err
	}

	r.BundleDir = bundleDir
	return r, err
}

// SetConfig creates a 'config.json' by the generator
func (r *Runtime) SetConfig(g *generate.Generator) error {
	if r.BundleDir == "" {
		return errors.New("Please set the bundle directory first")
	}
	return g.SaveToFile(filepath.Join(r.BundleDir, "config.json"), generate.ExportOptions{})
}

// SetID sets the container ID
func (r *Runtime) SetID(id string) {
	r.ID = id
}

// Create a container
func (r *Runtime) Create() error {
	var args []string
	args = append(args, "create")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	// TODO: following the spec, we need define the bundle, but 'runc' does not..
	//	if r.BundleDir != "" {
	//		args = append(args, r.BundleDir)
	//	}
	cmd := exec.Command(r.RuntimeCommand, args...)
	cmd.Dir = r.BundleDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Start a container
func (r *Runtime) Start() error {
	var args []string
	args = append(args, "start")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	cmd := exec.Command(r.RuntimeCommand, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// State a container information
func (r *Runtime) State() (rspecs.State, error) {
	var args []string
	args = append(args, "state")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	out, err := exec.Command(r.RuntimeCommand, args...).Output()
	if err != nil {
		return rspecs.State{}, err
	}

	var state rspecs.State
	err = json.Unmarshal(out, &state)
	return state, err
}

// Delete a container
func (r *Runtime) Delete() error {
	var args []string
	args = append(args, "delete")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	cmd := exec.Command(r.RuntimeCommand, args...)
	return cmd.Run()
}

// Clean deletes the container and removes the bundle file according to the input parameter
func (r *Runtime) Clean(removeBundle bool) error {
	err := r.Delete()
	if err != nil {
		return err
	}

	if removeBundle {
		os.RemoveAll(r.BundleDir)
	}

	return nil
}
