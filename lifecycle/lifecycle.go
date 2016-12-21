package lifecycle

import (
	"errors"
	"fmt"
	"os/exec"
)

// LifecycleAction defines lifecycle action according to oci spec
type LifecycleAction string

// Define actions for Lifecycle
const (
	LifecycleState  LifecycleAction = "state"
	LifecycleCreate LifecycleAction = "create"
	LifecycleStart  LifecycleAction = "start"
	LifecycleKill   LifecycleAction = "kill"
	LifecycleDelete LifecycleAction = "delete"
)

// Lifecycle storages the runtime location, bundle path and the container id
type Lifecycle struct {
	ID         string
	Runtime    string
	BundlePath string
}

// NewLifecycle creates a lifecycle with a runtime, bundle path and the container id
//   no need to check the validate of bundlepath or id
func NewLifecycle(runtime, bundlePath, id string) (Lifecycle, error) {
	var lc Lifecycle
	lc.Runtime = runtime
	lc.BundlePath = bundlePath
	lc.ID = id
	return lc, nil
}

// Operate provides lifecycle  operations
func (lc *Lifecycle) Operate(action LifecycleAction, arg ...string) ([]byte, error) {
	var cmd *exec.Cmd

	switch action {
	case LifecycleState:
		cmd = exec.Command(lc.Runtime, string(action), lc.ID)
	case LifecycleCreate:
		// FIXME: '-b' is used in 'runc'
		cmd = exec.Command(lc.Runtime, string(action), "-b", lc.BundlePath, lc.ID)
	case LifecycleStart:
		cmd = exec.Command(lc.Runtime, string(action), lc.ID)
	case LifecycleKill:
		if len(arg) != 1 {
			return nil, errors.New("Should choose a 'signal' to kill a container")
		}
		cmd = exec.Command(lc.Runtime, string(action), lc.ID, arg[0])
	case LifecycleDelete:
		cmd = exec.Command(lc.Runtime, string(action), lc.ID)
	default:
		return nil, fmt.Errorf("'%s' is not supported", string(action))

	}

	return cmd.CombinedOutput()
}
