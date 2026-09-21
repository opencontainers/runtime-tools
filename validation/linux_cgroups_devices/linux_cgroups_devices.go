package main

import (
	"os"

	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
	"github.com/opencontainers/runtime-tools/validation/util"
)

func deviceNode(path string, major int64, minor int64) rspec.LinuxDevice {
	mode := os.FileMode(0o666)
	owner := uint32(0)

	return rspec.LinuxDevice{
		Path:     path,
		Type:     "c",
		Major:    major,
		Minor:    minor,
		FileMode: &mode,
		UID:      &owner,
		GID:      &owner,
	}
}

func main() {
	var major1, minor1, major2, minor2, major3, minor3 int64 = 10, 229, 8, 20, 10, 200

	t := tap.New()
	t.Header(0)
	defer t.AutoPlan()

	g, err := util.GetDefaultGenerator()
	if err != nil {
		util.Fatal(err)
	}
	g.SetLinuxCgroupsPath(cgroups.AbsCgroupPath)
	g.AddLinuxResourcesDevice(true, "c", &major1, &minor1, "rwm")
	g.AddLinuxResourcesDevice(true, "b", &major2, &minor2, "rw")
	g.AddLinuxResourcesDevice(true, "b", &major3, &minor3, "r")

	// The rules above name no device that the bundle puts in the
	// container, so nothing about them can be observed from the inside.
	// Add a node the first rule allows and one that only the leading
	// deny-all covers, which leaves the rules themselves untouched.
	g.AddDevice(deviceNode("/dev/allowed", major1, minor1))
	g.AddDevice(deviceNode("/dev/denied", major1, minor1+1))

	// Reading the rules back is a cgroup v1 affair. The unified hierarchy
	// applies the same rules but publishes no list to compare against, so
	// there they are checked by trying the two nodes above.
	if util.UnifiedCgroups() {
		g.AddAnnotation("TestName", "check linux.resources.devices")
		err = util.RuntimeInsideValidate(g, t, nil)
	} else {
		err = util.RuntimeOutsideValidate(g, t, util.ValidateLinuxResourcesDevices)
	}
	if err != nil {
		t.Fail(err.Error())
	}
}
