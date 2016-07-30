package seccomp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	types "github.com/opencontainers/runtime-spec/specs-go"
)

// ParseSyscallFlag takes the name of the action, the arguments (syscalls) that were
// passed with it at the command line and a pointer to the config struct. It parses
// the action and syscalls and updates the config accordingly
func ParseSyscallFlag(action string, syscallArgs []string, config *types.Seccomp) error {

	if syscallArgs == nil {
		return errors.New("Error: Nil syscall slice")
	}

	correctedAction, err := parseAction(action)
	if err != nil {
		return err
	}

	if correctedAction == config.DefaultAction {
		logrus.Info("Action is already set as default")
		return nil
	}

	for _, syscallArg := range syscallArgs {
		delimArgs := strings.Split(syscallArg, ":")
		argSlice, err := parseArguments(delimArgs)
		if err != nil {
			return err
		}

		newSyscall := newSyscallStruct(delimArgs[0], correctedAction, *argSlice)
		descison, err := decideCourseOfAction(&newSyscall, config.Syscalls)
		if err != nil {
			fmt.Println(err)
			return err
		}
		delimDescison := strings.Split(descison, ":")

		if delimDescison[0] == nothing {
			logrus.Info("No action taken: ", newSyscall)
		}

		if delimDescison[0] == seccompAppend {
			config.Syscalls = append(config.Syscalls, newSyscall)
		}

		if delimDescison[0] == seccompOverwrite {
			indexForOverwrite, err := strconv.ParseInt(delimDescison[1], 10, 32)
			if err != nil {
				return err
			}
			config.Syscalls[indexForOverwrite] = newSyscall
		}
	}
	return nil
}

var actions = map[string]types.Action{
	"allow": types.ActAllow,
	"errno": types.ActErrno,
	"kill":  types.ActKill,
	"trace": types.ActTrace,
	"trap":  types.ActTrap,
}

// Take passed action, return the SCMP_ACT_<ACTION> version of it
func parseAction(action string) (types.Action, error) {
	a, ok := actions[action]
	if !ok {
		return "", fmt.Errorf("Unrecognized action: %s", action)
	}
	return a, nil
}

//ParseDefaultAction simply sets the default action of the seccomp configuration
func ParseDefaultAction(action string, config *types.Seccomp) error {
	if action == "" {
		return nil
	}

	defaultAction, err := parseAction(action)
	if err != nil {
		return err
	}
	config.DefaultAction = defaultAction
	return nil
}

func newSyscallStruct(name string, action types.Action, args []types.Arg) types.Syscall {
	syscallStruct := types.Syscall{
		Name:   name,
		Action: action,
		Args:   args,
	}
	return syscallStruct
}
