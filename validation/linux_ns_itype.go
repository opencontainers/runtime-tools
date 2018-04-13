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

func testNamespaceInheritType(t *tap.T) error {
	g, err := util.GetDefaultGenerator()
	if err != nil {
		return err
	}

	// Obtain a map for host (runtime) namespace, and remove every namespace
	// from the generated config, to be able to see if each container namespace
	// becomes inherited from its corresponding host namespace.
	hostNsPath := fmt.Sprintf("/proc/%d/ns", os.Getpid())
	hostNsInodes := map[string]string{}
	for _, nsName := range util.ProcNamespaces {
		nsInode, err := os.Readlink(filepath.Join(hostNsPath, nsName))
		if err != nil {
			return err
		}
		hostNsInodes[nsName] = nsInode

		if err := g.RemoveLinuxNamespace(util.GetRuntimeToolsNamespace(nsName)); err != nil {
			return err
		}
	}

	// We need to remove hostname to avoid test failures when not creating UTS namespace
	g.RemoveHostname()

	err = util.RuntimeOutsideValidate(g, func(config *rspec.Spec, state *rspec.State) error {
		containerNsPath := fmt.Sprintf("/proc/%d/ns", state.Pid)

		for _, nsName := range util.ProcNamespaces {
			nsInode, err := os.Readlink(filepath.Join(containerNsPath, nsName))
			if err != nil {
				return err
			}

			t.Ok(hostNsInodes[nsName] == nsInode, fmt.Sprintf("inherit namespace %s without type", nsName))
			if hostNsInodes[nsName] != nsInode {
				specErr := specerror.NewError(specerror.NSInheritWithoutType,
					fmt.Errorf("namespace %s (inode %s) does not inherit runtime namespace %s", nsName, nsInode, hostNsInodes[nsName]),
					rspec.Version)
				diagnostic := map[string]interface{}{
					"expected":       hostNsInodes[nsName],
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

	err := testNamespaceInheritType(t)
	if err != nil {
		util.Fatal(err)
	}

	t.AutoPlan()
}
