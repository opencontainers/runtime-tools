package memorypolicy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
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

// MpolNodesValid checks if the provided nodes specification is valid.
func MpolNodesValid(nodes string) error {
	// nodes is a comma-separated list of node IDs or ranges thereof.
	nodeRanges := strings.Split(nodes, ",")
	for _, nodeRange := range nodeRanges {
		nodeRange = strings.TrimSpace(nodeRange)
		if nodeRange == "" {
			continue
		}
		bounds := strings.Split(nodeRange, "-")
		switch len(bounds) {
		case 1:
			// Single node
			number := strings.TrimSpace(bounds[0])
			if _, err := parseNodeID(number); err != nil {
				return err
			}
		case 2:
			// Range of nodes
			startNumber := strings.TrimSpace(bounds[0])
			startID, err := parseNodeID(startNumber)
			if err != nil {
				return err
			}
			endNumber := strings.TrimSpace(bounds[1])
			endID, err := parseNodeID(endNumber)
			if err != nil {
				return err
			}
			if startID > endID {
				return fmt.Errorf("invalid memory policy node range %q: start ID greater than end ID", nodeRange)
			}
		default:
			return fmt.Errorf("invalid memory policy node range %q", nodeRange)
		}
	}
	return nil
}

func parseNodeID(nodeStr string) (int, error) {
	nodeID, err := strconv.Atoi(nodeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid memory policy node %q", nodeStr)
	}
	if nodeID < 0 {
		return 0, fmt.Errorf("memory policy node %d must be non-negative", nodeID)
	}
	return nodeID, nil
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

// MpolValid checks if the provided memory policy configuration is valid.
func MpolValid(mode rspec.MemoryPolicyModeType, nodes string, flags []rspec.MemoryPolicyFlagType) (errs error) {
	if err := MpolModeValid(string(mode)); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := MpolNodesValid(nodes); err != nil {
		errs = multierror.Append(errs, err)
	}
	for _, flag := range flags {
		if err := MpolFlagValid(string(flag)); err != nil {
			multierror.Append(errs, err)
		}
	}
	if errs == nil {
		err := MpolModeNodesValid(mode, nodes)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
