package validation

import (
	"fmt"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
)

// ContainerOperation represents common funtions for container testing
type ContainerOperation interface {
	// SetConfig creates a 'config.json' by the generator
	SetConfig(g *generate.Generator) error

	// SetID sets the container ID
	SetID(id string)

	// Create a container
	Create() error

	// Start a container
	Start() error

	// State a container information
	State() (rspecs.State, error)

	// Delete a container
	Delete() error

	// Clean deletes the container and removes the bundle file according to the input parameter
	Clean(removeBundle bool) error
}

// Runtime represents the basic requirement of a container runtime
type Runtime struct {
	RuntimeCommand string
	BundleDir      string
	ID             string
}

// NewRuntime creates different runtime based on runtimeCommand
func NewRuntime(runtimeCommand string, bundleDir string) (ContainerOperation, error) {
	switch runtimeCommand {
	case "runc":
		r, err := NewRunc(runtimeCommand, bundleDir)
		return r, err
	}

	return nil, fmt.Errorf("%s is not supported yet", runtimeCommand)
}
