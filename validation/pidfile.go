package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	tap "github.com/mndrix/tap-go"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func main() {
	t := tap.New()
	t.Header(0)

	tempDir, err := ioutil.TempDir("", "oci-pid")
	if err != nil {
		util.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	tempPidFile := filepath.Join(tempDir, "pidfile")

	config := util.LifecycleConfig{
		Actions: util.LifecycleActionCreate | util.LifecycleActionDelete,
		PreCreate: func(r *util.Runtime) error {
			r.PidFile = tempPidFile
			return nil
		},
		PostCreate: func(r *util.Runtime) error {
			pidData, err := ioutil.ReadFile(tempPidFile)
			if err != nil {
				return err
			}
			pid, err := strconv.Atoi(string(pidData))
			if err != nil {
				return err
			}
			state, err := r.State()
			if err != nil {
				return err
			}
			if state.Pid != pid {
				return errors.New("Wrong pid in the pidfile")
			}
			return nil
		},
		PreDelete: func(r *util.Runtime) error {
			return util.WaitingForStatus(*r, util.LifecycleStatusCreated, time.Second*10, time.Second*1)
		},
	}

	g := util.GetDefaultGenerator()
	g.SetProcessArgs([]string{"true"})
	err = util.RuntimeLifecycleValidate(g, config)
	if err != nil {
		util.Fatal(err)
	}

	t.AutoPlan()
}
