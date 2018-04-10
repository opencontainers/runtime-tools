package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func getRuntimeToolsNamespace(ns string) string {
	// Deal with exceptional cases of "net" and "mnt", because those strings
	// cannot be recognized by mapStrToNamespace(), which actually expects
	// "network" and "mount" respectively.
	switch ns {
	case "net":
		return "network"
	case "mnt":
		return "mount"
	}

	// In other cases, return just the original string
	return ns
}

func testNamespacePath(t *tap.T, ns string, unshareOpt string) error {
	// Calling 'unshare' (part of util-linux) is easier than doing it from
	// Golang: mnt namespaces cannot be unshared from multithreaded
	// programs.
	cmd := exec.Command("unshare", unshareOpt, "--fork", "sleep", "10000")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("cannot run unshare: %s", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		cmd.Wait()
	}()
	if cmd.Process == nil {
		return fmt.Errorf("process failed to start")
	}
	unsharePid := cmd.Process.Pid

	// Wait until 'unshare' switched its namespace
	// TODO: avoid synchronisation with sleeps.
	time.Sleep(time.Second)

	specialChildren := ""
	if ns == "pid" {
		// Unsharing pidns does not move the process into the new
		// pidns but the next forked process. 'unshare' is called with
		// '--fork' so the pidns will be fully created and populated
		// with a pid 1.
		//
		// However, finding out the pid of the child process is not
		// trivial: it would require to parse
		// /proc/$pid/task/$tid/children but that only works on kernels
		// with CONFIG_PROC_CHILDREN (not all distros have that).
		//
		// It is easier to look at /proc/$pid/ns/pid_for_children on
		// the parent process. Available since Linux 4.12.
		specialChildren = "_for_children"
	}
	unshareNsPath := fmt.Sprintf("/proc/%d/ns/%s", unsharePid, ns+specialChildren)
	unshareNsInode, err := os.Readlink(unshareNsPath)
	if err != nil {
		return fmt.Errorf("cannot read namespace link for the unshare process: %s", err)
	}

	g, err := util.GetDefaultGenerator()
	if err != nil {
		return fmt.Errorf("cannot get the default generator: %v", err)
	}

	rtns := getRuntimeToolsNamespace(ns)
	g.AddOrReplaceLinuxNamespace(rtns, unshareNsPath)

	// The spec is not clear about userns mappings when reusing an
	// existing userns.
	// See https://github.com/opencontainers/runtime-spec/issues/961
	//if ns == "user" {
	//	g.AddLinuxUIDMapping(uint32(1000), uint32(0), uint32(1000))
	//	g.AddLinuxGIDMapping(uint32(1000), uint32(0), uint32(1000))
	//}

	err = util.RuntimeOutsideValidate(g, func(config *rspec.Spec, state *rspec.State) error {
		containerNsPath := fmt.Sprintf("/proc/%d/ns/%s", state.Pid, ns)
		containerNsInode, err := os.Readlink(containerNsPath)
		if err != nil {
			out, err2 := exec.Command("sh", "-c", fmt.Sprintf("ls -la /proc/%d/ns/", state.Pid)).CombinedOutput()
			return fmt.Errorf("cannot read namespace link for the container process: %s\n%v\n%v", err, err2, out)
		}
		if containerNsInode != unshareNsInode {
			return fmt.Errorf("expected: %q, found: %q", unshareNsInode, containerNsInode)
		}
		return nil
	})

	return err
}

func main() {
	t := tap.New()
	t.Header(0)

	cases := []struct {
		name       string
		unshareOpt string
	}{
		{"cgroup", "--cgroup"},
		{"ipc", "--ipc"},
		{"mnt", "--mount"},
		{"net", "--net"},
		{"pid", "--pid"},
		{"user", "--user"},
		{"uts", "--uts"},
	}

	for _, c := range cases {
		if "linux" != runtime.GOOS {
			t.Skip(1, fmt.Sprintf("linux-specific namespace test: %s", c))
		}

		err := testNamespacePath(t, c.name, c.unshareOpt)
		t.Ok(err == nil, fmt.Sprintf("set %s namespace by path", c.name))
		if err != nil {
			diagnostic := map[string]string{
				"error": err.Error(),
			}
			t.YAML(diagnostic)
		}
	}

	t.AutoPlan()
}
