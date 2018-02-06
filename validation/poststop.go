package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tap "github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
	uuid "github.com/satori/go.uuid"
)

func main() {
	t := tap.New()
	t.Header(0)

	var output string
	config := util.LifecycleConfig{
		Actions: util.LifecycleActionCreate | util.LifecycleActionStart | util.LifecycleActionDelete,
		PreCreate: func(r *util.Runtime) error {
			r.SetID(uuid.NewV4().String())
			g := util.GetDefaultGenerator()
			output = filepath.Join(r.BundleDir, g.Spec().Root.Path, "output")
			poststop := rspec.Hook{
				Path: fmt.Sprintf("%s/%s/bin/sh", r.BundleDir, g.Spec().Root.Path),
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-stop called' >> %s", output),
				},
			}

			g.AddPostStopHook(poststop)
			g.SetProcessArgs([]string{"sh", "-c", fmt.Sprintf("echo 'process called' >> %s", "/output")})
			r.SetConfig(g)
			return nil
		},
		PostCreate: func(r *util.Runtime) error {
			outputData, err := ioutil.ReadFile(output)
			if err == nil {
				if strings.Contains(string(outputData), "post-stop called") {
					return specerror.NewError(specerror.PoststopTiming, fmt.Errorf("The post-stop hooks MUST be called after the container is deleted"), rspec.Version)
				} else if strings.Contains(string(outputData), "process called") {
					return specerror.NewError(specerror.ProcNotRunAtResRequest, fmt.Errorf("The user-specified program (from process) MUST NOT be run at this time"), rspec.Version)
				}
				return fmt.Errorf("File %v should not exist", output)
			}
			return nil
		},
		PreDelete: func(r *util.Runtime) error {
			util.WaitingForStatus(*r, util.LifecycleStatusStopped, time.Second*10, time.Second)
			outputData, err := ioutil.ReadFile(output)
			if err != nil {
				return specerror.NewError(specerror.ProcImplement, fmt.Errorf("The runtime MUST run the user-specified program, as specified by `process`"), rspec.Version)
			}
			switch string(outputData) {
			case "post-stop called\n":
				return fmt.Errorf("%v\n%v", specerror.NewError(specerror.PoststopTiming, fmt.Errorf("The post-stop hooks MUST be called after the container is deleted"), rspec.Version), specerror.NewError(specerror.ProcImplement, fmt.Errorf("The runtime MUST run the user-specified program, as specified by `process`"), rspec.Version))
			case "process called\n":
				return nil
			case "post-stop called\nprocess called\n", "process called\npost-stop called\n":
				return specerror.NewError(specerror.PoststopTiming, fmt.Errorf("The post-stop hooks MUST be called after the container is deleted"), rspec.Version)
			default:
				return fmt.Errorf("Unsupported output information: %v", string(outputData))
			}
		},
		PostDelete: func(r *util.Runtime) error {
			outputData, err := ioutil.ReadFile(output)
			if err != nil {
				return fmt.Errorf("%v\n%v", specerror.NewError(specerror.PoststopHooksInvoke, fmt.Errorf("The poststop hooks MUST be invoked by the runtime"), rspec.Version), specerror.NewError(specerror.ProcImplement, fmt.Errorf("The runtime MUST run the user-specified program, as specified by `process`"), rspec.Version))
			}
			switch string(outputData) {
			case "post-stop called\n":
				return specerror.NewError(specerror.ProcImplement, fmt.Errorf("The runtime MUST run the user-specified program, as specified by `process`"), rspec.Version)
			case "process called\n":
				fmt.Fprintln(os.Stderr, "WARNING: The poststop hook invoke fails")
				return nil
			case "post-stop called\nprocess called\n":
				return errors.New("The Post-stop should called after the user-specified program command is executed")
			case "process called\npost-stop called\n":
				return nil
			default:
				return fmt.Errorf("Unsupported output information: %v", string(outputData))
			}
		},
	}

	err := util.RuntimeLifecycleValidate(nil, config)
	if err != nil {
		diagnostic := map[string]string{
			"error": err.Error(),
		}
		if e, ok := err.(*exec.ExitError); ok {
			if len(e.Stderr) > 0 {
				diagnostic["stderr"] = string(e.Stderr)
			}
		}
		t.YAML(diagnostic)
	}

	t.AutoPlan()
}
