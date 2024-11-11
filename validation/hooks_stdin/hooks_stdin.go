package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/google/uuid"
	tap "github.com/mndrix/tap-go"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func stdinStateCheck(outputDir, hookName string, expectedState rspecs.State) (errs error) {
	var state rspecs.State
	data, err := os.ReadFile(filepath.Join(outputDir, hookName))
	if err != nil {
		errs = errors.Join(errs, err)
		return
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		errs = errors.Join(errs, err)
		return
	}

	if state.ID != expectedState.ID {
		err = fmt.Errorf("wrong container ID %q in the stdin of %s hook, expected %q", state.ID, hookName, expectedState.ID)
		errs = errors.Join(errs, err)
	}

	if state.Bundle != expectedState.Bundle {
		err = fmt.Errorf("wrong bundle directory %q in the stdin of %s hook, expected %q", state.Bundle, hookName, expectedState.Bundle)
		errs = errors.Join(errs, err)
	}

	if hookName != "poststop" && state.Pid != expectedState.Pid {
		err = fmt.Errorf("wrong container process ID %q in the stdin of %s hook, expected %q", state.Version, hookName, expectedState.Version)
		errs = errors.Join(errs, err)
	}

	if !reflect.DeepEqual(state.Annotations, expectedState.Annotations) {
		err = fmt.Errorf("wrong annotations \"%v\" in the stdin of %s hook, expected \"%v\"", state.Annotations, hookName, expectedState.Annotations)
		errs = errors.Join(errs, err)
	}

	switch hookName {
	case "prestart":
		if state.Status != "created" {
			err = fmt.Errorf("wrong status %q in the stdin of %s hook, expected %q", state.Status, hookName, "created")
			errs = errors.Join(errs, err)
		}
	case "poststart":
		if state.Status != "running" {
			err = fmt.Errorf("wrong status %q in the stdin of %s hook, expected %q", state.Status, hookName, "running")
			errs = errors.Join(errs, err)
		}
	case "poststop":
		if state.Status == "" {
			err = fmt.Errorf("status in the stdin of %s hook should not be empty", hookName)
			errs = errors.Join(errs, err)
		}
	default:
		err = fmt.Errorf("internal error, unexpected hook name %q", hookName)
		errs = errors.Join(errs, err)
	}

	return
}

func main() {
	t := tap.New()
	t.Header(0)

	bundleDir, err := util.PrepareBundle()
	if err != nil {
		util.Fatal(err)
	}
	containerID := uuid.NewString()
	defer os.RemoveAll(bundleDir)

	var containerPid int

	annotationKey := "org.opencontainers.runtime-tools"
	annotationValue := "hook stdin test"
	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	outputDir := filepath.Join(bundleDir, g.Config.Root.Path)
	timeout := 1
	g.AddAnnotation(annotationKey, annotationValue)
	g.AddPreStartHook(rspecs.Hook{
		Path: filepath.Join(bundleDir, g.Config.Root.Path, "/bin/sh"),
		Args: []string{
			"sh", "-c", fmt.Sprintf("cat > %s", filepath.Join(outputDir, "prestart")),
		},
		Timeout: &timeout,
	})
	g.AddPostStartHook(rspecs.Hook{
		Path: filepath.Join(bundleDir, g.Config.Root.Path, "/bin/sh"),
		Args: []string{
			"sh", "-c", fmt.Sprintf("cat > %s", filepath.Join(outputDir, "poststart")),
		},
		Timeout: &timeout,
	})
	g.AddPostStopHook(rspecs.Hook{
		Path: filepath.Join(bundleDir, g.Config.Root.Path, "/bin/sh"),
		Args: []string{
			"sh", "-c", fmt.Sprintf("cat > %s", filepath.Join(outputDir, "poststop")),
		},
		Timeout: &timeout,
	})
	g.SetProcessArgs([]string{"true"})
	config := util.LifecycleConfig{
		BundleDir: bundleDir,
		Config:    g,
		Actions:   util.LifecycleActionCreate | util.LifecycleActionStart | util.LifecycleActionDelete,
		PreCreate: func(r *util.Runtime) error {
			r.SetID(containerID)
			return nil
		},
		PostCreate: func(r *util.Runtime) error {
			state, err := r.State()
			if err != nil {
				return err
			}
			containerPid = state.Pid
			return nil
		},
		PreDelete: func(r *util.Runtime) error {
			util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
			return nil
		},
	}

	err = util.RuntimeLifecycleValidate(config)
	if err != nil {
		t.Fail(err.Error())
	}

	expectedState := rspecs.State{
		Pid:         containerPid,
		ID:          containerID,
		Bundle:      bundleDir,
		Annotations: map[string]string{annotationKey: annotationValue},
	}
	for _, file := range []string{"prestart", "poststart", "poststop"} {
		errs := stdinStateCheck(outputDir, file, expectedState)
		var newError error
		if errs == nil {
			newError = errors.New("")
		} else {
			newError = errors.New(errs.Error())
		}
		util.SpecErrorOK(t, errs == nil, specerror.NewError(specerror.PosixHooksStateToStdin, fmt.Errorf("the state of the container MUST be passed to %q hook over stdin", file), rspecs.Version), newError)
	}

	t.AutoPlan()
}
