package util

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/satori/go.uuid"
)

// Runtime represents the basic requirement of a container runtime
type Runtime struct {
	RuntimeCommand string
	BundleDir      string
	ID             string
	stdout         *os.File
	stderr         *os.File
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
	return g.SaveToFile(filepath.Join(r.bundleDir(), "config.json"), generate.ExportOptions{})
}

// SetID sets the container ID
func (r *Runtime) SetID(id string) {
	r.ID = id
}

// Create a container
func (r *Runtime) Create() (stderr []byte, err error) {
	var args []string
	args = append(args, "create")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	if r.BundleDir != "" {
		args = append(args, "--bundle", r.BundleDir)
	}
	cmd := exec.Command(r.RuntimeCommand, args...)
	id := uuid.NewV4().String()
	r.stdout, err = os.OpenFile(filepath.Join(r.bundleDir(), fmt.Sprintf("stdout-%s", id)), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return []byte(""), err
	}
	cmd.Stdout = r.stdout
	r.stderr, err = os.OpenFile(filepath.Join(r.bundleDir(), fmt.Sprintf("stderr-%s", id)), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		return []byte(""), err
	}
	cmd.Stderr = r.stderr

	err = cmd.Run()
	if err == nil {
		return []byte(""), err
	}

	stdout, stderr, _ := r.ReadStandardStreams()
	if len(stderr) == 0 {
		stderr = stdout
	}
	return stderr, err
}

// ReadStandardStreams collects content from the stdout and stderr buffers.
func (r *Runtime) ReadStandardStreams() (stdout []byte, stderr []byte, err error) {
	_, err = r.stdout.Seek(0, io.SeekStart)
	stdout, err2 := ioutil.ReadAll(r.stdout)
	if err == nil && err2 != nil {
		err = err2
	}
	_, err = r.stderr.Seek(0, io.SeekStart)
	stderr, err2 = ioutil.ReadAll(r.stderr)
	if err == nil && err2 != nil {
		err = err2
	}
	return stdout, stderr, err
}

// Start a container
func (r *Runtime) Start() (stderr []byte, err error) {
	var args []string
	args = append(args, "start")
	if r.ID != "" {
		args = append(args, r.ID)
	}

	cmd := exec.Command(r.RuntimeCommand, args...)
	stdout, err := cmd.Output()
	if e, ok := err.(*exec.ExitError); ok {
		stderr = e.Stderr
	}
	if err != nil && len(stderr) == 0 {
		stderr = stdout
	}

	return stderr, err
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

// Clean deletes the container.  If removeBundle is set, the bundle
// directory is removed after the container is deleted succesfully or, if
// forceRemoveBundle is true, after the deletion attempt regardless of
// whether it was successful or not.
func (r *Runtime) Clean(removeBundle bool, forceRemoveBundle bool) error {
	err := r.Delete()

	if removeBundle && (err == nil || forceRemoveBundle) {
		err2 := os.RemoveAll(r.bundleDir())
		if err2 != nil && err == nil {
			err = err2
		}
	}

	return err
}
