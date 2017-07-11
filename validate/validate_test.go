package validate

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

func checkErrors(t *testing.T, title string, msgs []string, valid bool) {
	if valid && len(msgs) > 0 {
		t.Fatalf("%s: expected not to get error, but get %d errors:\n%s", title, len(msgs), strings.Join(msgs, "\n"))
	} else if !valid && len(msgs) == 0 {
		t.Fatalf("%s: expected to get error, but actually not", title)
	}
}

func TestCheckRootfsPath(t *testing.T) {
	tmpBundle, err := ioutil.TempDir("", "oci-check-rootfspath")
	if err != nil {
		t.Fatalf("Failed to create a TempDir in 'CheckRootfsPath'")
	}
	defer os.RemoveAll(tmpBundle)

	rootfsDir := "rootfs"
	rootfsNonDir := "rootfsfile"
	rootfsNonExists := "rootfsnil"
	if err := os.MkdirAll(filepath.Join(tmpBundle, rootfsDir), 0700); err != nil {
		t.Fatalf("Failed to create a rootfs directory in 'CheckRootfsPath'")
	}
	if _, err := os.Create(filepath.Join(tmpBundle, rootfsNonDir)); err != nil {
		t.Fatalf("Failed to create a non-directory rootfs in 'CheckRootfsPath'")
	}

	cases := []struct {
		val      string
		expected bool
	}{
		{rootfsDir, true},
		{rootfsNonDir, false},
		{rootfsNonExists, false},
		{filepath.Join(tmpBundle, rootfsDir), true},
		{filepath.Join(tmpBundle, rootfsNonDir), false},
		{filepath.Join(tmpBundle, rootfsNonExists), false},
	}
	for _, c := range cases {
		v := NewValidator(&rspec.Spec{Root: &rspec.Root{Path: c.val}}, tmpBundle, false, "linux")
		checkErrors(t, "CheckRootfsPath "+c.val, v.CheckRootfsPath(), c.expected)
	}
}

func TestCheckSemVer(t *testing.T) {
	cases := []struct {
		val      string
		expected bool
	}{
		{rspec.Version, true},
		//FIXME: validate currently only handles rpsec.Version
		{"0.0.1", false},
		{"invalid", false},
	}

	for _, c := range cases {
		v := NewValidator(&rspec.Spec{Version: c.val}, "", false, "linux")
		checkErrors(t, "CheckSemVer "+c.val, v.CheckSemVer(), c.expected)
	}
}
