package validate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"

	rerr "github.com/opencontainers/runtime-tools/error"
)

func TestNewValidator(t *testing.T) {
	testSpec := &rspec.Spec{}
	testBundle := ""
	testPlatform := "not" + runtime.GOOS
	cases := []struct {
		val      Validator
		expected Validator
	}{
		{Validator{testSpec, testBundle, true, testPlatform}, Validator{testSpec, testBundle, true, runtime.GOOS}},
		{Validator{testSpec, testBundle, true, runtime.GOOS}, Validator{testSpec, testBundle, true, runtime.GOOS}},
		{Validator{testSpec, testBundle, false, testPlatform}, Validator{testSpec, testBundle, false, testPlatform}},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, NewValidator(c.val.spec, c.val.bundlePath, c.val.HostSpecific, c.val.platform))
	}
}

func TestCheckRoot(t *testing.T) {
	tmpBundle, err := ioutil.TempDir("", "oci-check-rootfspath")
	if err != nil {
		t.Fatalf("Failed to create a TempDir in 'CheckRoot'")
	}
	defer os.RemoveAll(tmpBundle)

	rootfsDir := "rootfs/rootfs"
	rootfsNonDir := "rootfsfile"
	rootfsNonExists := "rootfsnil"
	if err := os.MkdirAll(filepath.Join(tmpBundle, rootfsDir), 0700); err != nil {
		t.Fatalf("Failed to create a rootfs directory in 'CheckRoot'")
	}
	if _, err := os.Create(filepath.Join(tmpBundle, rootfsNonDir)); err != nil {
		t.Fatalf("Failed to create a non-directory rootfs in 'CheckRoot'")
	}

	// Note: Abs error is not tested
	cases := []struct {
		val      rspec.Spec
		platform string
		expected rerr.SpecErrorCode
	}{
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: &rspec.Root{}}, "windows", rerr.RootOnHyperV},
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: nil}, "windows", rerr.NonError},
		{rspec.Spec{Root: nil}, "linux", rerr.RootOnNonHyperV},
		{rspec.Spec{Root: &rspec.Root{Path: "maverick-rootfs"}}, "linux", rerr.PathName},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs"}}, "linux", rerr.NonError},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonExists)}}, "linux", rerr.PathExistence},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonDir)}}, "linux", rerr.PathExistence},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, "rootfs")}}, "linux", rerr.NonError},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs/rootfs"}}, "linux", rerr.ArtifactsInSingleDir},
		{rspec.Spec{Root: &rspec.Root{Readonly: true}}, "windows", rerr.ReadonlyOnWindows},
	}
	for _, c := range cases {
		v := NewValidator(&c.val, tmpBundle, false, c.platform)
		err := v.CheckRoot()
		assert.Equal(t, c.expected, rerr.FindError(err, c.expected), fmt.Sprintf("Fail to check Root: %v %d", err, c.expected))
	}
}

func TestCheckSemVer(t *testing.T) {
	cases := []struct {
		val      string
		expected rerr.SpecErrorCode
	}{
		{rspec.Version, rerr.NonError},
		//FIXME: validate currently only handles rpsec.Version
		{"0.0.1", rerr.NonRFCError},
		{"invalid", rerr.SpecVersion},
	}

	for _, c := range cases {
		v := NewValidator(&rspec.Spec{Version: c.val}, "", false, "linux")
		err := v.CheckSemVer()
		assert.Equal(t, c.expected, rerr.FindError(err, c.expected), "Fail to check SemVer "+c.val)
	}
}
