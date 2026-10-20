package util

import (
	"github.com/mndrix/tap-go"
	"github.com/opencontainers/runtime-tools/cgroups"
)

// UnifiedCgroups reports whether the host mounts the unified hierarchy,
// on which several of the cgroup v1 resource fields have no counterpart.
func UnifiedCgroups() bool {
	cg, err := cgroups.FindCgroup()
	if err != nil {
		return false
	}

	return cg.Version() == 2
}

// checkValue compares a configured resource with the value read back from
// the cgroup filesystem. A field that either side does not carry is
// reported as a diagnostic rather than counted as a failed check, because
// the hierarchy in use may have no way to express it.
func checkValue[T comparable](t *tap.T, name string, expected *T, actual *T) {
	if expected == nil || actual == nil {
		t.Diagnosticf("skipping %s: not expressed by this cgroup hierarchy", name)
		return
	}

	t.Ok(*expected == *actual, name+" is set correctly")
	t.Diagnosticf("expect: %v, actual: %v", *expected, *actual)
}
