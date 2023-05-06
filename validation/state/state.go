package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/google/uuid"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	rfc2119 "github.com/opencontainers/runtime-tools/error"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/util/tap"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	t := tap.New()
	t.Header(0)

	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetProcessArgs([]string{"true"})
	containerID := uuid.NewString()

	cases := []struct {
		id          string
		action      util.LifecycleAction
		errExpected bool
		err         *rfc2119.Error
	}{
		{"", util.LifecycleActionNone, false, specerror.NewRFCErrorOrPanic(specerror.QueryWithoutIDGenError, fmt.Errorf("state MUST generate an error if it is not provided the ID of a container"), rspecs.Version)},
		{containerID, util.LifecycleActionNone, false, specerror.NewRFCErrorOrPanic(specerror.QueryNonExistGenError, fmt.Errorf("state MUST generate an error if a container that does not exist"), rspecs.Version)},
		{containerID, util.LifecycleActionCreate | util.LifecycleActionDelete, true, specerror.NewRFCErrorOrPanic(specerror.QueryStateImplement, fmt.Errorf("state MUST return the state of a container as specified in the State section"), rspecs.Version)},
	}

	for _, c := range cases {
		config := util.LifecycleConfig{
			Config:  g,
			Actions: c.action,
			PreCreate: func(r *util.Runtime) error {
				r.SetID(c.id)
				return nil
			},
			PostCreate: func(r *util.Runtime) error {
				_, err = r.State()
				r.Kill("KILL")
				util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
				return err
			},
		}
		err = util.RuntimeLifecycleValidate(config)
		// DefaultStateJSONPattern might returns
		if e, ok := err.(*specerror.Error); ok {
			diagnostic := map[string]string{
				"reference": e.Err.Reference,
				"error":     e.Err.Error(),
			}
			_ = t.YAML(diagnostic)
			continue
		}

		t.Ok((err == nil) == c.errExpected, c.err.Error())
		diagnostic := map[string]string{
			"reference": c.err.Reference,
		}
		if err != nil {
			diagnostic["error"] = err.Error()
			if e, ok := err.(*exec.ExitError); ok {
				if len(e.Stderr) > 0 {
					diagnostic["stderr"] = string(e.Stderr)
				}
			}
		}
		_ = t.YAML(diagnostic)
	}

	t.AutoPlan()
}
