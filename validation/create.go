package main

import (
	"fmt"

	"github.com/mndrix/tap-go"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validation/util"
	"github.com/satori/go.uuid"
)

func main() {
	t := tap.New()
	t.Header(0)

	g := generate.New()
	g.SetRootPath(".")
	g.SetProcessArgs([]string{"ls"})

	bundleDir, err := util.PrepareBundle()
	if err != nil {
		util.Fatal(err)
	}

	r, err := util.NewRuntime(util.RuntimeCommand, bundleDir)
	if err != nil {
		util.Fatal(err)
	}
	defer r.Clean(true, true)

	err = r.SetConfig(&g)
	if err != nil {
		util.Fatal(err)
	}

	containerID := uuid.NewV4().String()
	cases := []struct {
		id          string
		errExpected bool
		err         error
	}{
		{"", false, specerror.NewError(specerror.CreateWithBundlePathAndID, fmt.Errorf("create MUST generate an error if the ID is not provided"), rspecs.Version)},
		{containerID, true, specerror.NewError(specerror.CreateNewContainer, fmt.Errorf("create MUST create a new container"), rspecs.Version)},
		{containerID, false, specerror.NewError(specerror.CreateWithUniqueID, fmt.Errorf("create MUST generate an error if the ID provided is not unique"), rspecs.Version)},
	}

	for _, c := range cases {
		r.SetID(c.id)
		stderr, err := r.Create()
		t.Ok((err == nil) == c.errExpected, c.err.(*specerror.Error).Err.Err.Error())
		diagnostic := map[string]string{
			"reference": c.err.(*specerror.Error).Err.Reference,
		}
		if err != nil {
			diagnostic["error"] = err.Error()
		}
		if len(stderr) > 0 {
			diagnostic["stderr"] = string(stderr)
		}
		t.YAML(diagnostic)

		if err == nil {
			state, _ := r.State()
			t.Ok(state.ID == c.id, "")
			t.YAML(map[string]string{
				"container ID": c.id,
				"state ID":     state.ID,
			})
		}
	}

	t.AutoPlan()
}
