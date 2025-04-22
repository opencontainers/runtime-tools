package memorypolicy

import (
	"fmt"
	"strings"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

var (
	knownModes map[rspec.MemoryPolicyModeType]struct{} = map[rspec.MemoryPolicyModeType]struct{}{
		rspec.MpolDefault:            {},
		rspec.MpolBind:               {},
		rspec.MpolInterleave:         {},
		rspec.MpolWeightedInterleave: {},
		rspec.MpolPreferred:          {},
		rspec.MpolPreferredMany:      {},
		rspec.MpolLocal:              {},
	}

	knownModeFlags map[rspec.MemoryPolicyFlagType]struct{} = map[rspec.MemoryPolicyFlagType]struct{}{
		rspec.MpolFNumaBalancing: {},
		rspec.MpolFRelativeNodes: {},
		rspec.MpolFStaticNodes:   {},
	}
)

// MpolModeValid checks if the provided memory policy mode is valid.
func MpolModeValid(mode string) error {
	if !strings.HasPrefix(mode, "MPOL_") {
		return fmt.Errorf("memory policy mode %q must start with 'MPOL_'", mode)
	}
	if _, ok := knownModes[rspec.MemoryPolicyModeType(mode)]; !ok {
		return fmt.Errorf("invalid memory policy mode %q", mode)
	}
	return nil
}

// MpolModeNodesValid checks if the nodes specification is valid for the given memory policy mode.
func MpolModeNodesValid(mode rspec.MemoryPolicyModeType, nodes string) error {
	switch mode {
	case rspec.MpolDefault, rspec.MpolLocal:
		if nodes != "" {
			return fmt.Errorf("memory policy mode %q must not have nodes specified", mode)
		}
	case rspec.MpolBind, rspec.MpolInterleave, rspec.MpolWeightedInterleave, rspec.MpolPreferred, rspec.MpolPreferredMany:
		if nodes == "" {
			return fmt.Errorf("memory policy mode %q must have nodes specified", mode)
		}
	case "":
		return fmt.Errorf("memory policy mode must be specified")
	default:
		return fmt.Errorf("unknown memory policy mode %q ", mode)
	}
	return nil
}

// MpolFlagValid checks if the provided memory policy flag is valid.
func MpolFlagValid(flag string) error {
	if !strings.HasPrefix(flag, "MPOL_F_") {
		return fmt.Errorf("memory policy flag %q must start with 'MPOL_F_'", flag)
	}
	if _, ok := knownModeFlags[rspec.MemoryPolicyFlagType(flag)]; !ok {
		return fmt.Errorf("invalid memory policy flag %q", flag)
	}
	return nil
}
