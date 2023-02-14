package main

import (
	"fmt"
	"os"
	"os/exec"
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

	var output string
	config := util.LifecycleConfig{
		Actions: util.LifecycleActionCreate | util.LifecycleActionStart | util.LifecycleActionDelete,
		PreCreate: func(r *util.Runtime) error {
			r.SetID(uuid.NewString())
			g, err := util.GetDefaultGenerator()
			if err != nil {
				util.Fatal(err)
			}
			output = filepath.Join(r.BundleDir, g.Config.Root.Path, "output")
			shPath := filepath.Join(r.BundleDir, g.Config.Root.Path, "/bin/sh")
			g.AddPreStartHook(rspec.Hook{
				Path: shPath,
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'pre-start1 called' >> %s", output),
				},
			})
			g.AddPreStartHook(rspec.Hook{
				Path: shPath,
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'pre-start2 called' >> %s", output),
				},
			})
			g.AddPostStartHook(rspec.Hook{
				Path: shPath,
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-start1 called' >> %s", output),
				},
			})
			g.AddPostStartHook(rspec.Hook{
				Path: shPath,
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-start2 called' >> %s", output),
				},
			})
			g.AddPostStopHook(rspec.Hook{
				Path: shPath,
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-stop1 called' >> %s", output),
				},
			})
			g.AddPostStopHook(rspec.Hook{
				Path: shPath,
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-stop2 called' >> %s", output),
				},
			})
			g.SetProcessArgs([]string{"true"})
			return r.SetConfig(g)
		},
		PreDelete: func(r *util.Runtime) error {
			util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
			return nil
		},
	}

	err := util.RuntimeLifecycleValidate(config)
	outputData, _ := os.ReadFile(output)
	if err == nil && string(outputData) != "pre-start1 called\npre-start2 called\npost-start1\npost-start2\npost-stop1\npost-stop2\n" {
		err = specerror.NewError(specerror.PosixHooksCalledInOrder, fmt.Errorf("Hooks MUST be called in the listed order"), rspec.Version)
		diagnostic := map[string]string{
			"error": err.Error(),
		}
		_ = t.YAML(diagnostic)
	} else {
		diagnostic := map[string]string{
			"error": err.Error(),
		}
		if e, ok := err.(*exec.ExitError); ok {
			if len(e.Stderr) > 0 {
				diagnostic["stderr"] = string(e.Stderr)
			}
		}
		_ = t.YAML(diagnostic)
	}

	t.AutoPlan()
}
