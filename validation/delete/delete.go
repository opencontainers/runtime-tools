package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mndrix/tap-go"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	bundleDir, err := util.PrepareBundle()
	if err != nil {
		util.Fatal(err)
	}
	defer os.RemoveAll(bundleDir)

	stoppedConfig, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	stoppedConfig.SetProcessArgs([]string{"true"})
	runningConfig, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	runningConfig.SetProcessArgs([]string{"sleep", "30"})
	containerID := uuid.NewString()
	testRuntime, _ := util.NewRuntime(util.RuntimeCommand, bundleDir)
	cases := []struct {
		config      *generate.Generator
		id          string
		action      util.LifecycleAction
		errExpected bool
		// set true to check whether 'delete' takes effect and just check if 'delete' DOES NOT take effect
		effectCheck bool
		// set true if 'delete' may either fail and have no effect, or succeed and delete the container
		mayDelete bool
		err       error
	}{
		// Note: the nil config test case should run first since we are re-using the bundle
		// delete without id
		{nil, "", util.LifecycleActionNone, false, false, false, specerror.NewError(specerror.DeleteWithoutIDGenError, fmt.Errorf("`delete` operation MUST generate an error if it is not provided the container ID"), rspecs.Version)},
		// delete a created container
		// TODO: bump runtime-spec to v1.3.1 once it's out (for the reference).
		{stoppedConfig, containerID, util.LifecycleActionCreate, false, false, true, specerror.NewError(specerror.DeleteCreatedNoEffectOrDelete, fmt.Errorf("attempting to `delete` a container that is `created` MAY either have no effect on the container and generate an error, or terminate the container process and then delete the container"), rspecs.Version)},
		// delete a running container
		{runningConfig, containerID, util.LifecycleActionCreate | util.LifecycleActionStart, false, true, false, specerror.NewError(specerror.DeleteNonStopGenError, fmt.Errorf("attempting to `delete` a container that is neither `created` nor `stopped` MUST generate an error"), rspecs.Version)},
	}

	for _, c := range cases {
		config := util.LifecycleConfig{
			Config:    c.config,
			BundleDir: bundleDir,
			Actions:   c.action,
			PreCreate: func(r *util.Runtime) error {
				r.SetID(c.id)
				return nil
			},
		}
		util.RuntimeLifecycleValidate(config)
		// waiting the 'stoppedConfig' testcase to stop
		// the 'runningConfig' testcase sleeps 30 seconds, so 10 seconds are enough for this case
		testRuntime.SetID(c.id)
		util.WaitingForStatus(testRuntime, util.LifecycleStatusCreated|util.LifecycleStatusStopped, time.Second*10, time.Second*1)
		deletedErr := testRuntime.Delete()
		if c.mayDelete {
			if deletedErr == nil {
				// waiting for the error of State, just in case the delete operation takes time
				util.WaitingForStatus(testRuntime, util.LifecycleActionNone, time.Second*3, time.Second/2)
			}
			_, err = testRuntime.State()
			// either 'delete' failed and the container is still there,
			// or 'delete' succeeded and the container is gone
			util.SpecErrorOK(t, (deletedErr == nil) == (err != nil), c.err, deletedErr)
		} else {
			util.SpecErrorOK(t, (deletedErr == nil) == c.errExpected, c.err, deletedErr)
		}

		if c.effectCheck {
			// waiting for the error of State, just in case the delete operation takes time
			util.WaitingForStatus(testRuntime, util.LifecycleActionNone, time.Second*3, time.Second/2)
			_, err = testRuntime.State()
			// err == nil means the 'delete' operation does NOT take effect
			util.SpecErrorOK(t, err == nil, specerror.NewError(specerror.DeleteNonStopHaveNoEffect, fmt.Errorf("attempting to `delete` a container that is neither `created` nor `stopped` MUST have no effect on the container"), rspecs.Version), err)
		}

		// created and but deleted
		if (c.action&util.LifecycleActionCreate != 0) && (deletedErr != nil) {
			testRuntime.Kill("KILL")
			// waiting for the container to be killed, just in case the kill operation takes time
			util.WaitingForStatus(testRuntime, util.LifecycleStatusStopped, time.Second*10, time.Second*1)
			err = testRuntime.Delete()
			if err != nil {
				t.Fail(err.Error())
			}
		}
	}
}
