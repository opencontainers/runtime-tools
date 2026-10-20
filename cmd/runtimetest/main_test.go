package main

import (
	"path/filepath"
	"syscall"
	"testing"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

func devicePtr(v int64) *int64 {
	return &v
}

func TestDeviceReadAllowed(t *testing.T) {
	// The list linux_cgroups_devices builds: the generator's leading
	// deny-all followed by the three rules the test adds.
	rules := []rspec.LinuxDeviceCgroup{
		{Allow: false, Access: "rwm"},
		{Allow: true, Type: "c", Major: devicePtr(10), Minor: devicePtr(229), Access: "rwm"},
		{Allow: true, Type: "b", Major: devicePtr(8), Minor: devicePtr(20), Access: "rw"},
		{Allow: true, Type: "b", Major: devicePtr(10), Minor: devicePtr(200), Access: "r"},
	}

	for _, test := range []struct {
		description string
		devType     string
		rules       []rspec.LinuxDeviceCgroup
		major       int64
		minor       int64
		expected    bool
	}{
		{description: "explicitly allowed", rules: rules, devType: "c", major: 10, minor: 229, expected: true},
		{description: "a minor the allow does not cover", rules: rules, devType: "c", major: 10, minor: 230, expected: false},
		{description: "right numbers, wrong type", rules: rules, devType: "b", major: 10, minor: 229, expected: false},
		{description: "allowed block device", rules: rules, devType: "b", major: 8, minor: 20, expected: true},
		{description: "read-only allow still permits reading", rules: rules, devType: "b", major: 10, minor: 200, expected: true},
		{description: "covered only by the deny-all", rules: rules, devType: "c", major: 1, minor: 3, expected: false},
		{description: "no rules at all", rules: nil, devType: "c", major: 1, minor: 3, expected: true},
		{
			description: "a deny that does not mention read",
			rules:       []rspec.LinuxDeviceCgroup{{Allow: false, Access: "w"}},
			devType:     "c", major: 1, minor: 3, expected: true,
		},
		{
			description: "a later rule overrides an earlier one",
			rules: []rspec.LinuxDeviceCgroup{
				{Allow: true, Access: "rwm"},
				{Allow: false, Type: "c", Major: devicePtr(10), Minor: devicePtr(229), Access: "rwm"},
			},
			devType: "c", major: 10, minor: 229, expected: false,
		},
	} {
		got := deviceReadAllowed(test.rules, test.devType, test.major, test.minor)
		if got != test.expected {
			t.Errorf("%s: %s %d:%d allowed == %t, expected %t",
				test.description, test.devType, test.major, test.minor, got, test.expected)
		}
	}
}

func TestIsDefaultDevicePath(t *testing.T) {
	if !isDefaultDevicePath("/dev/null") {
		t.Error("/dev/null is a default device")
	}
	if isDefaultDevicePath("/dev/denied") {
		t.Error("/dev/denied is not a default device")
	}
}

func TestDeviceCgroupType(t *testing.T) {
	for _, test := range []struct {
		devType  string
		expected string
		governed bool
	}{
		{devType: "c", expected: "c", governed: true},
		{devType: "u", expected: "c", governed: true},
		{devType: "b", expected: "b", governed: true},
		{devType: "p", governed: false},
	} {
		got, governed := deviceCgroupType(test.devType)
		if got != test.expected || governed != test.governed {
			t.Errorf("%s: type == %q governed == %t, expected %q %t",
				test.devType, got, governed, test.expected, test.governed)
		}
	}
}

func TestDeviceDeniedByCgroupFifo(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "fifo")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	// Opening a FIFO without a writer blocks unless the open is
	// non-blocking.
	if deviceDeniedByCgroup(fifo) {
		t.Error("a FIFO outside any device rule was reported as denied")
	}
}
