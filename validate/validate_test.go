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
		expected ErrorCode
	}{
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: &rspec.Root{}}, "windows", RootOnHyperV},
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: nil}, "windows", NonError},
		{rspec.Spec{Root: nil}, "linux", RootOnNonHyperV},
		{rspec.Spec{Root: &rspec.Root{Path: "maverick-rootfs"}}, "linux", PathName},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs"}}, "linux", NonError},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonExists)}}, "linux", PathExistence},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonDir)}}, "linux", PathExistence},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, "rootfs")}}, "linux", NonError},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs/rootfs"}}, "linux", ArtifactsInSingleDir},
		{rspec.Spec{Root: &rspec.Root{Readonly: true}}, "windows", ReadonlyOnWindows},
	}
	for _, c := range cases {
		v := NewValidator(&c.val, tmpBundle, false, c.platform)
		err := v.CheckRoot()
		assert.Equal(t, c.expected, FindError(err, c.expected), fmt.Sprintf("Fail to check Root: %v %d", err, c.expected))
	}
}

func TestCheckSemVer(t *testing.T) {
	cases := []struct {
		val      string
		expected ErrorCode
	}{
		{rspec.Version, NonError},
		//FIXME: validate currently only handles rpsec.Version
		{"0.0.1", NonRFCError},
		{"invalid", SpecVersion},
	}

	for _, c := range cases {
		v := NewValidator(&rspec.Spec{Version: c.val}, "", false, "linux")
		err := v.CheckSemVer()
		assert.Equal(t, c.expected, FindError(err, c.expected), "Fail to check SemVer "+c.val)
	}
}
