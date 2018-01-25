package main

import (
	"fmt"
	"os/exec"

	"github.com/mndrix/tap-go"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
	uuid "github.com/satori/go.uuid"
)

func main() {
	t := tap.New()
	t.Header(0)

	g := util.GetDefaultGenerator()
	g.SetProcessArgs([]string{"true"})
	containerID := uuid.NewV4().String()

	cases := []struct {
		id          string
		action      util.LifecycleAction
		errExpected bool
		err         error
	}{
		{"", util.LifecycleActionNone, false, specerror.NewError(specerror.QueryWithoutIDGenError, fmt.Errorf("state MUST generate an error if it is not provided the ID of a container"), rspecs.Version)},
		{containerID, util.LifecycleActionNone, false, specerror.NewError(specerror.QueryNonExistGenError, fmt.Errorf("state MUST generate an error if a container that does not exist"), rspecs.Version)},
		{containerID, util.LifecycleActionCreate | util.LifecycleActionDelete, true, specerror.NewError(specerror.QueryStateImplement, fmt.Errorf("state MUST return the state of a container as specified in the State section"), rspecs.Version)},
	}

	for _, c := range cases {
		config := util.LifecycleConfig{
			Actions: c.action,
			PreCreate: func(r *util.Runtime) error {
				r.SetID(c.id)
				return nil
			},
			PostCreate: func(r *util.Runtime) error {
				_, err := r.State()
				return err
			},
		}
		err := util.RuntimeLifecycleValidate(g, config)
		t.Ok((err == nil) == c.errExpected, c.err.(*specerror.Error).Err.Err.Error())
		diagnostic := map[string]string{
			"reference": c.err.(*specerror.Error).Err.Reference,
		}
		if err != nil {
			diagnostic["error"] = err.Error()
			if e, ok := err.(*exec.ExitError); ok {
				if len(e.Stderr) > 0 {
					diagnostic["stderr"] = string(e.Stderr)
				}
			}
		}
		t.YAML(diagnostic)
	}

	t.AutoPlan()
}
