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

	"github.com/opencontainers/runtime-tools/specerror"
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
		expected specerror.Code
	}{
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: &rspec.Root{}}, "windows", specerror.RootOnHyperV},
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: nil}, "windows", specerror.NonError},
		{rspec.Spec{Root: nil}, "linux", specerror.RootOnNonHyperV},
		{rspec.Spec{Root: &rspec.Root{Path: "maverick-rootfs"}}, "linux", specerror.PathName},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs"}}, "linux", specerror.NonError},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonExists)}}, "linux", specerror.PathExistence},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonDir)}}, "linux", specerror.PathExistence},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, "rootfs")}}, "linux", specerror.NonError},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs/rootfs"}}, "linux", specerror.ArtifactsInSingleDir},
		{rspec.Spec{Root: &rspec.Root{Readonly: true}}, "windows", specerror.ReadonlyOnWindows},
	}
	for _, c := range cases {
		v := NewValidator(&c.val, tmpBundle, false, c.platform)
		err := v.CheckRoot()
		assert.Equal(t, c.expected, specerror.FindError(err, c.expected), fmt.Sprintf("Fail to check Root: %v %d", err, c.expected))
	}
}

func TestCheckSemVer(t *testing.T) {
	cases := []struct {
		val      string
		expected specerror.Code
	}{
		{rspec.Version, specerror.NonError},
		//FIXME: validate currently only handles rpsec.Version
		{"0.0.1", specerror.NonRFCError},
		{"invalid", specerror.SpecVersion},
	}

	for _, c := range cases {
		v := NewValidator(&rspec.Spec{Version: c.val}, "", false, "linux")
		err := v.CheckSemVer()
		assert.Equal(t, c.expected, specerror.FindError(err, c.expected), "Fail to check SemVer "+c.val)
	}
}
