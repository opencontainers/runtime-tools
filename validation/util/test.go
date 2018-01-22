package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mndrix/tap-go"
	"github.com/mrunalp/fileutils"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/satori/go.uuid"
)

var (
	// RuntimeCommand is the default runtime command.
	RuntimeCommand = "runc"
)

// LifecycleAction defines the phases will be called.
type LifecycleAction int

const (
	// LifecycleActionCreate creates a container
	LifecycleActionCreate = 1 << iota
	// LifecycleActionStart starts a container
	LifecycleActionStart
	// LifecycleActionDelete deletes a container
	LifecycleActionDelete
)

// LifecycleStatus follows https://github.com/opencontainers/runtime-spec/blob/master/runtime.md#state
type LifecycleStatus int

const (
	// LifecycleStatusCreating "creating"
	LifecycleStatusCreating = 1 << iota
	// LifecycleStatusCreated "created"
	LifecycleStatusCreated
	// LifecycleStatusRunning "running"
	LifecycleStatusRunning
	// LifecycleStatusStopped "stopped"
	LifecycleStatusStopped
)

var lifecycleStatusMap = map[string]LifecycleStatus{
	"creating": LifecycleStatusCreating,
	"created":  LifecycleStatusCreated,
	"running":  LifecycleStatusRunning,
	"stopped":  LifecycleStatusStopped,
}

// LifecycleConfig includes
// 1. Actions to define the default running lifecycles.
// 2. Four phases for user to add his/her own operations.
type LifecycleConfig struct {
	Actions    LifecycleAction
	PreCreate  func(runtime *Runtime) error
	PostCreate func(runtime *Runtime) error
	PreDelete  func(runtime *Runtime) error
	PostDelete func(runtime *Runtime) error
}

// PreFunc initializes the test environment after preparing the bundle
// but before creating the container.
type PreFunc func(string) error

// AfterFunc validate container's outside environment after created
type AfterFunc func(config *rspec.Spec, state *rspec.State) error

func init() {
	runtimeInEnv := os.Getenv("RUNTIME")
	if runtimeInEnv != "" {
		RuntimeCommand = runtimeInEnv
	}
}

// Fatal prints a warning to stderr and exits.
func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}

// Skip skips a full TAP suite.
func Skip(message string, diagnostic interface{}) {
	t := tap.New()
	t.Header(1)
	t.Skip(1, message)
	if diagnostic != nil {
		t.YAML(diagnostic)
	}
}

// PrepareBundle creates a test bundle in a temporary directory.
func PrepareBundle() (string, error) {
	bundleDir, err := ioutil.TempDir("", "ocitest")
	if err != nil {
		return "", err
	}

	// Untar the root fs
	untarCmd := exec.Command("tar", "-xf", fmt.Sprintf("rootfs-%s.tar.gz", runtime.GOARCH), "-C", bundleDir)
	output, err := untarCmd.CombinedOutput()
	if err != nil {
		os.Stderr.Write(output)
		os.RemoveAll(bundleDir)
		return "", err
	}

	return bundleDir, nil
}

// GetDefaultGenerator creates a default configuration generator.
func GetDefaultGenerator() *generate.Generator {
	g := generate.New()
	g.SetRootPath(".")
	g.SetProcessArgs([]string{"/runtimetest", "--path=/"})
	return &g
}

// WaitingForStatus waits an expected runtime status, return error if
// 1. fail to query the status
// 2. timeout
func WaitingForStatus(r Runtime, status LifecycleStatus, retryTimeout time.Duration, pollInterval time.Duration) error {
	for start := time.Now(); time.Since(start) < retryTimeout; time.Sleep(pollInterval) {
		state, err := r.State()
		if err != nil {
			return err
		}
		if v, ok := lifecycleStatusMap[state.Status]; ok {
			if status&v != 0 {
				return nil
			}
		} else {
			// In spec, it says 'Additional values MAY be defined by the runtime'.
			continue
		}
	}

	return errors.New("timeout in waiting for the container status")
}

// RuntimeInsideValidate runs runtimetest inside a container.
func RuntimeInsideValidate(g *generate.Generator, f PreFunc) (err error) {
	bundleDir, err := PrepareBundle()
	if err != nil {
		return err
	}

	if f != nil {
		if err := f(bundleDir); err != nil {
			return err
		}
	}

	r, err := NewRuntime(RuntimeCommand, bundleDir)
	if err != nil {
		os.RemoveAll(bundleDir)
		return err
	}
	defer r.Clean(true, true)
	err = r.SetConfig(g)
	if err != nil {
		return err
	}
	err = fileutils.CopyFile("runtimetest", filepath.Join(r.BundleDir, "runtimetest"))
	if err != nil {
		return err
	}

	r.SetID(uuid.NewV4().String())
	stderr, err := r.Create()
	if err != nil {
		os.Stderr.WriteString("failed to create the container\n")
		os.Stderr.Write(stderr)
		return err
	}

	// FIXME: why do we need this?  Without a sleep here, I get:
	//   failed to start the container
	//   container "..." does not exist
	time.Sleep(1 * time.Second)

	stderr, err = r.Start()
	if err != nil {
		os.Stderr.WriteString("failed to start the container\n")
		os.Stderr.Write(stderr)
		return err
	}

	// FIXME: wait until the container exits and collect its exit code.
	time.Sleep(1 * time.Second)

	stdout, stderr, err := r.ReadStandardStreams()
	if err != nil {
		if len(stderr) == 0 {
			stderr = stdout
		}
		os.Stderr.WriteString("failed to read standard streams\n")
		os.Stderr.Write(stderr)
		return err
	}

	os.Stdout.Write(stdout)
	return nil
}

// RuntimeOutsideValidate validate runtime outside a container.
func RuntimeOutsideValidate(g *generate.Generator, f AfterFunc) error {
	bundleDir, err := PrepareBundle()
	if err != nil {
		return err
	}

	r, err := NewRuntime(RuntimeCommand, bundleDir)
	if err != nil {
		os.RemoveAll(bundleDir)
		return err
	}
	defer r.Clean(true, true)
	err = r.SetConfig(g)
	if err != nil {
		return err
	}
	err = fileutils.CopyFile("runtimetest", filepath.Join(r.BundleDir, "runtimetest"))
	if err != nil {
		return err
	}

	r.SetID(uuid.NewV4().String())
	stderr, err := r.Create()
	if err != nil {
		os.Stderr.WriteString("failed to create the container\n")
		os.Stderr.Write(stderr)
		return err
	}

	if f != nil {
		state, err := r.State()
		if err != nil {
			return err
		}
		if err := f(g.Spec(), &state); err != nil {
			return err
		}
	}
	return nil
}

// RuntimeLifecycleValidate validates runtime lifecycle.
func RuntimeLifecycleValidate(g *generate.Generator, config LifecycleConfig) error {
	bundleDir, err := PrepareBundle()
	if err != nil {
		return err
	}
	r, err := NewRuntime(RuntimeCommand, bundleDir)
	if err != nil {
		os.RemoveAll(bundleDir)
		return err
	}
	defer r.Clean(true, true)
	err = r.SetConfig(g)
	if err != nil {
		return err
	}
	r.SetID(uuid.NewV4().String())

	if config.PreCreate != nil {
		if err := config.PreCreate(&r); err != nil {
			return err
		}
	}

	if config.Actions&LifecycleActionCreate != 0 {
		stderr, err := r.Create()
		if err != nil {
			os.Stderr.WriteString("failed to create the container\n")
			os.Stderr.Write(stderr)
			return err
		}
	}

	if config.PostCreate != nil {
		if err := config.PostCreate(&r); err != nil {
			return err
		}
	}

	if config.Actions&LifecycleActionStart != 0 {
		stderr, err := r.Start()
		if err != nil {
			os.Stderr.WriteString("failed to start the container\n")
			os.Stderr.Write(stderr)
			return err
		}
	}

	if config.PreDelete != nil {
		if err := config.PreDelete(&r); err != nil {
			return err
		}
	}

	if config.Actions&LifecycleActionDelete != 0 {
		stderr, err := r.Delete()
		if err != nil {
			os.Stderr.WriteString("failed to delete the container\n")
			os.Stderr.Write(stderr)
			return err
		}
	}

	if config.PostDelete != nil {
		if err := config.PostDelete(&r); err != nil {
			return err
		}
	}
	return nil
}
