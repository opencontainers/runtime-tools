package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"time"

	tap "github.com/mndrix/tap-go"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
	uuid "github.com/satori/go.uuid"
)

func stdinStateCheck(outputDir, hookName string, expectedState rspecs.State) error {
	var state rspecs.State
	data, err := ioutil.ReadFile(filepath.Join(outputDir, hookName))
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &state)
	if err != nil {
		return err
	}

	if state.ID != expectedState.ID {
		return fmt.Errorf("wrong container ID %q in the stdin of %s hook, expected %q", state.ID, hookName, expectedState.ID)
	}

	if state.Bundle != expectedState.Bundle {
		return fmt.Errorf("wrong bundle directory %q in the stdin of %s hook, expected %q", state.Bundle, hookName, expectedState.Bundle)
	}

	if hookName != "poststop" && state.Pid != expectedState.Pid {
		return fmt.Errorf("wrong container process ID %q in the stdin of %s hook, expected %q", state.Version, hookName, expectedState.Version)
	}

	if !reflect.DeepEqual(state.Annotations, expectedState.Annotations) {
		return fmt.Errorf("wrong annotations \"%v\" in the stdin of %s hook, expected \"%v\"", state.Annotations, hookName, expectedState.Annotations)
	}
	return nil
}

func main() {
	t := tap.New()
	t.Header(0)

	bundleDir, err := util.PrepareBundle()
	if err != nil {
		util.Fatal(err)
	}
	containerID := uuid.NewV4().String()
	defer os.RemoveAll(bundleDir)

	var containerPid int

	annotationKey := "org.opencontainers.runtime-tools"
	annotationValue := "hook stdin test"
	g := util.GetDefaultGenerator()
	outputDir := filepath.Join(bundleDir, g.Spec().Root.Path)
	timeout := 1
	g.AddAnnotation(annotationKey, annotationValue)
	g.AddPreStartHook(rspecs.Hook{
		Path: filepath.Join(bundleDir, g.Spec().Root.Path, "/bin/sh"),
		Args: []string{
			"sh", "-c", fmt.Sprintf("cat > %s", filepath.Join(outputDir, "prestart")),
		},
		Timeout: &timeout,
	})
	g.AddPostStartHook(rspecs.Hook{
		Path: filepath.Join(bundleDir, g.Spec().Root.Path, "/bin/sh"),
		Args: []string{
			"sh", "-c", fmt.Sprintf("cat > %s", filepath.Join(outputDir, "poststart")),
		},
		Timeout: &timeout,
	})
	g.AddPostStopHook(rspecs.Hook{
		Path: filepath.Join(bundleDir, g.Spec().Root.Path, "/bin/sh"),
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
		PreDelete: func(r *util.Runtime) error {
			state, err := r.State()
			if err != nil {
				return err
			}
			containerPid = state.Pid
			util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
			return nil
		},
	}

	err = util.RuntimeLifecycleValidate(config)
	if err != nil {
		util.Fatal(err)
	}

	expectedState := rspecs.State{
		Pid:         containerPid,
		ID:          containerID,
		Bundle:      bundleDir,
		Annotations: map[string]string{annotationKey: annotationValue},
	}
	for _, file := range []string{"prestart", "poststart", "poststop"} {
		err := stdinStateCheck(outputDir, file, expectedState)
		util.SpecErrorOK(t, err == nil, specerror.NewError(specerror.PosixHooksStateToStdin, fmt.Errorf("the state of the container MUST be passed to %q hook over stdin", file), rspecs.Version), err)
	}

	t.AutoPlan()
}
