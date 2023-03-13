package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	tap "github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	t := tap.New()
	t.Header(0)

	bundleDir, err := util.PrepareBundle()
	if err != nil {
		return
	}
	defer os.RemoveAll(bundleDir)

	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	output := filepath.Join(bundleDir, g.Config.Root.Path, "output")
	poststart := rspec.Hook{
		Path: filepath.Join(bundleDir, g.Config.Root.Path, "/bin/false"),
		Args: []string{"false"},
	}
	g.AddPostStartHook(poststart)
	poststartOK := rspec.Hook{
		Path: filepath.Join(bundleDir, g.Config.Root.Path, "/bin/sh"),
		Args: []string{
			"sh", "-c", fmt.Sprintf("echo 'post-start called' >> %s", output),
		},
	}
	g.AddPostStartHook(poststartOK)
	g.SetProcessArgs([]string{"true"})
	config := util.LifecycleConfig{
		Config:    g,
		BundleDir: bundleDir,
		Actions:   util.LifecycleActionCreate | util.LifecycleActionStart | util.LifecycleActionDelete,
		PreCreate: func(r *util.Runtime) error {
			r.SetID(uuid.NewString())
			return nil
		},
		PreDelete: func(r *util.Runtime) error {
			util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
			return nil
		},
	}

	runErr := util.RuntimeLifecycleValidate(config)
	outputData, _ := os.ReadFile(output)

	// if runErr is not nil, it means the runtime generates an error
	// if outputData is not equal to the expected content, it means there is something wrong with the remaining hooks and lifecycle
	if runErr != nil || string(outputData) != "post-start called\n" {
		err = specerror.NewError(specerror.PoststartHookFailGenWarn, fmt.Errorf("if any poststart hook fails, the runtime MUST log a warning, but the remaining hooks and lifecycle continue as if the hook had succeeded"), rspec.Version)
		diagnostic := map[string]string{
			"error": err.Error(),
		}
		_ = t.YAML(diagnostic)
	}

	t.AutoPlan()
}
