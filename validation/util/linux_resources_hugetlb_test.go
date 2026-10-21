package util

import (
	"strings"
	"testing"
)

const procCgroups = `#subsys_name	hierarchy	num_cgroups	enabled
cpuset	0	12	1
cpu	0	12	1
memory	0	12	0
pids	0	12	1
`

func TestControllerEnabled(t *testing.T) {
	for _, tc := range []struct {
		name    string
		enabled bool
	}{
		{"cpu", true},
		{"pids", true},
		{"memory", false},
		{"hugetlb", false},
	} {
		enabled, err := controllerEnabled(strings.NewReader(procCgroups), tc.name)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if enabled != tc.enabled {
			t.Errorf("%s: enabled %v, want %v", tc.name, enabled, tc.enabled)
		}
	}
}
