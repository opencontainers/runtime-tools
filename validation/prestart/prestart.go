package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	tap "github.com/opencontainers/runtime-tools/util/tap"
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
			g.AddPreStartHook(rspec.Hook{
				Path: filepath.Join(r.BundleDir, g.Config.Root.Path, "/bin/sh"),
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'pre-start called' >> %s", output),
				},
			})
			g.SetProcessArgs([]string{"sh", "-c", fmt.Sprintf("echo 'process called' >> %s", "/output")})
			return r.SetConfig(g)
		},
		PostCreate: func(r *util.Runtime) error {
			outputData, err := os.ReadFile(output)
			if err == nil {
				if strings.Contains(string(outputData), "pre-start called") {
					return specerror.NewError(specerror.PrestartTiming, fmt.Errorf("Pre-start hooks MUST be called after the `start` operation is called"), rspec.Version)
				} else if strings.Contains(string(outputData), "process called") {
					return specerror.NewError(specerror.ProcNotRunAtResRequest, fmt.Errorf("The user-specified program (from process) MUST NOT be run at this time"), rspec.Version)
				}

				return fmt.Errorf("File %v should not exist", output)
			}
			return nil
		},
		PreDelete: func(r *util.Runtime) error {
			util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
			outputData, err := os.ReadFile(output)
			if err != nil {
				return fmt.Errorf("%v\n%v", specerror.NewError(specerror.PrestartHooksInvoke, fmt.Errorf("The prestart hooks MUST be invoked by the runtime"), rspec.Version), specerror.NewError(specerror.ProcImplement, fmt.Errorf("The runtime MUST run the user-specified program, as specified by `process`"), rspec.Version))
			}
			switch string(outputData) {
			case "pre-start called\n":
				return specerror.NewError(specerror.ProcImplement, fmt.Errorf("The runtime MUST run the user-specified program, as specified by `process`"), rspec.Version)
			case "process called\n":
				return specerror.NewError(specerror.PrestartHooksInvoke, fmt.Errorf("The prestart hooks MUST be invoked by the runtime"), rspec.Version)
			case "pre-start called\nprocess called\n":
				return nil
			case "process called\npre-start called\n":
				return specerror.NewError(specerror.PrestartTiming, fmt.Errorf("Pre-start hooks MUST be called before the user-specified program command is executed"), rspec.Version)
			default:
				return fmt.Errorf("Unsupported output information: %v", string(outputData))
			}
		},
	}

	err := util.RuntimeLifecycleValidate(config)
	if err != nil {
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
