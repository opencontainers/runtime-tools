package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func testNamespaceNoPath(t *tap.T) error {
	hostNsPath := fmt.Sprintf("/proc/%d/ns", os.Getpid())
	hostNsInodes := map[string]string{}
	for _, nsName := range util.ProcNamespaces {
		nsInode, err := os.Readlink(filepath.Join(hostNsPath, nsName))
		if err != nil {
			return err
		}
		hostNsInodes[nsName] = nsInode
	}

	g, err := util.GetDefaultGenerator()
	if err != nil {
		return err
	}

	// As the namespaces, cgroups and user, are not set by GetDefaultGenerator(),
	// others are set by default. We just set them explicitly to avoid confusion.
	g.AddOrReplaceLinuxNamespace("cgroup", "")
	g.AddOrReplaceLinuxNamespace("ipc", "")
	g.AddOrReplaceLinuxNamespace("mount", "")
	g.AddOrReplaceLinuxNamespace("network", "")
	g.AddOrReplaceLinuxNamespace("pid", "")
	g.AddOrReplaceLinuxNamespace("user", "")
	g.AddOrReplaceLinuxNamespace("uts", "")

	// For user namespaces, we need to set uid/gid maps to create a container
	g.AddLinuxUIDMapping(uint32(1000), uint32(0), uint32(1000))
	g.AddLinuxGIDMapping(uint32(1000), uint32(0), uint32(1000))

	err = util.RuntimeOutsideValidate(g, func(config *rspec.Spec, state *rspec.State) error {
		containerNsPath := fmt.Sprintf("/proc/%d/ns", state.Pid)

		for _, nsName := range util.ProcNamespaces {
			nsInode, err := os.Readlink(filepath.Join(containerNsPath, nsName))
			if err != nil {
				return err
			}

			t.Ok(hostNsInodes[nsName] != nsInode, fmt.Sprintf("create namespace %s without path", nsName))
			if hostNsInodes[nsName] == nsInode {
				specErr := specerror.NewError(specerror.NSNewNSWithoutPath,
					fmt.Errorf("both namespaces for %s have the same inode %s", nsName, nsInode),
					rspec.Version)
				diagnostic := map[string]interface{}{
					"expected":       fmt.Sprintf("!= %s", hostNsInodes[nsName]),
					"actual":         nsInode,
					"namespace type": nsName,
					"level":          specErr.(*specerror.Error).Err.Level,
					"reference":      specErr.(*specerror.Error).Err.Reference,
				}
				t.YAML(diagnostic)

				continue
			}
		}

		return nil
	})

	return err
}

func main() {
	t := tap.New()
	t.Header(0)

	if "linux" != runtime.GOOS {
		t.Skip(1, fmt.Sprintf("linux-specific namespace test"))
	}

	err := testNamespaceNoPath(t)
	if err != nil {
		util.Fatal(err)
	}

	t.AutoPlan()
}
