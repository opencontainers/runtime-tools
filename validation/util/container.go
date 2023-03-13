package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/opencontainers/runtime-tools/specerror"
)

// Runtime represents the basic requirement of a container runtime
type Runtime struct {
	RuntimeCommand string
	BundleDir      string
	PidFile        string
	ID             string
	stdout         *os.File
	stderr         *os.File
}

// DefaultSignal represents the default signal sends to a container
const DefaultSignal = "TERM"

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

// bundleDir returns the bundle directory.  Generally this is
// BundleDir, but when BundleDir is the empty string, it falls back to
// ., as specified in the CLI spec.
func (r *Runtime) bundleDir() (bundleDir string) {
	if r.BundleDir == "" {
		return "."
	}
	return r.BundleDir
}

// SetConfig creates a 'config.json' by the generator
func (r *Runtime) SetConfig(g *generate.Generator) error {
	if g == nil {
		return errors.New("cannot set a nil config")
	}
	return g.SaveToFile(filepath.Join(r.bundleDir(), "config.json"), generate.ExportOptions{})
}

// SetID sets the container ID
func (r *Runtime) SetID(id string) {
	r.ID = id
}

// Create a container
func (r *Runtime) Create() (err error) {
	var args []string
	args = append(args, "create")
	if r.PidFile != "" {
		args = append(args, "--pid-file", r.PidFile)
	}
	if r.BundleDir != "" {
		args = append(args, "--bundle", r.BundleDir)
	}
	if r.ID != "" {
		args = append(args, r.ID)
	}
	cmd := exec.Command(r.RuntimeCommand, args...)
	id := uuid.NewString()
	r.stdout, err = os.OpenFile(filepath.Join(r.bundleDir(), fmt.Sprintf("stdout-%s", id)), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	cmd.Stdout = r.stdout
	r.stderr, err = os.OpenFile(filepath.Join(r.bundleDir(), fmt.Sprintf("stderr-%s", id)), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	cmd.Stderr = r.stderr

	err = cmd.Run()
	if err == nil {
		return err
	}

	if e, ok := err.(*exec.ExitError); ok {
		stdout, stderr, _ := r.ReadStandardStreams()
		if len(stderr) == 0 {
			stderr = stdout
		}
		e.Stderr = stderr
		return e
	}
	return err
}

// ReadStandardStreams collects content from the stdout and stderr buffers.
func (r *Runtime) ReadStandardStreams() (stdout []byte, stderr []byte, err error) {
	stdout, err = os.ReadFile(r.stdout.Name())
	stderr, err2 := os.ReadFile(r.stderr.Name())
	if err == nil && err2 != nil {
		err = err2
	}
	return
}

// Start a container
func (r *Runtime) Start() (err error) {
	var args []string
	args = append(args, "start")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	cmd := exec.Command(r.RuntimeCommand, args...)
	return execWithStderrFallbackToStdout(cmd)
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
		if e, ok := err.(*exec.ExitError); ok {
			if len(e.Stderr) == 0 {
				e.Stderr = out
				return rspecs.State{}, e
			}
		}
		return rspecs.State{}, err
	}

	var state rspecs.State
	err = json.Unmarshal(out, &state)
	if err != nil {
		return rspecs.State{}, specerror.NewError(specerror.DefaultStateJSONPattern, fmt.Errorf("when serialized in JSON, the format MUST adhere to the default pattern"), rspecs.Version)
	}
	return state, err
}

// Kill a container
func (r *Runtime) Kill(sig string) (err error) {
	var args []string
	args = append(args, "kill")
	if r.ID != "" {
		args = append(args, r.ID)
	}
	if sig != "" {
		// TODO: runc does not support this
		//	args = append(args, "--signal", sig)
		args = append(args, sig)
	} else {
		args = append(args, DefaultSignal)
	}

	cmd := exec.Command(r.RuntimeCommand, args...)
	return execWithStderrFallbackToStdout(cmd)
}

// Delete removes a (stopped) container.
func (r *Runtime) Delete() error {
	return r.del(false)
}

// ForceDelete removes a container (killing it if necessary).
func (r *Runtime) ForceDelete() error {
	return r.del(true)
}

func (r *Runtime) del(force bool) (err error) {
	var args []string
	args = append(args, "delete")
	if force {
		args = append(args, "--force")
	}
	if r.ID != "" {
		args = append(args, r.ID)
	}

	cmd := exec.Command(r.RuntimeCommand, args...)
	return execWithStderrFallbackToStdout(cmd)
}

// Clean kills and removes the container and its bundle directory.
func (r *Runtime) Clean() {
	err := r.ForceDelete()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Clean: Delete: ", err)
	}

	err = os.RemoveAll(r.bundleDir())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Clean: ", err)
	}
}

func execWithStderrFallbackToStdout(cmd *exec.Cmd) (err error) {
	stdout, err := cmd.Output()
	if err == nil {
		return err
	}

	if e, ok := err.(*exec.ExitError); ok {
		if len(e.Stderr) == 0 {
			e.Stderr = stdout
			return e
		}
	}
	return err
}
