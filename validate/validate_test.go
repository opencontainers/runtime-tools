package validate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-multierror"
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

func TestJSONSchema(t *testing.T) {
	for _, tt := range []struct {
		config *rspec.Spec
		error  string
	}{
		{
			config: &rspec.Spec{},
			error:  "Version string empty",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.1-rc1",
			},
			error: "Could not read schema from HTTP, response status is 404 Not Found",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0",
			},
			error: "",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0",
				Process: &rspec.Process{},
			},
			error: "process.args: Invalid type. Expected: array, given: null",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0",
				Linux:   &rspec.Linux{},
			},
			error: "",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0",
				Linux: &rspec.Linux{
					RootfsPropagation: "",
				},
			},
			error: "",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0",
				Linux: &rspec.Linux{
					RootfsPropagation: "shared",
				},
			},
			error: "",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0",
				Linux: &rspec.Linux{
					RootfsPropagation: "rshared",
				},
			},
			error: "linux.rootfsPropagation: linux.rootfsPropagation must be one of the following: \"private\", \"shared\", \"slave\", \"unbindable\"",
		},
		{
			config: &rspec.Spec{
				Version: "1.0.0-rc5",
			},
			error: "process: process is required",
		},
	} {
		t.Run(tt.error, func(t *testing.T) {
			v := &Validator{spec: tt.config}
			errs := v.CheckJSONSchema()
			if tt.error == "" {
				if errs == nil {
					return
				}
				t.Fatalf("expected no error, but got: %s", errs.Error())
			}
			if errs == nil {
				t.Fatal("failed to raise the expected error")
			}
			merr, ok := errs.(*multierror.Error)
			if !ok {
				t.Fatalf("non-multierror returned by CheckJSONSchema: %s", errs.Error())
			}
			for _, err := range merr.Errors {
				if err.Error() == tt.error {
					return
				}
			}
			assert.Equal(t, tt.error, errs.Error())
		})
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
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: &rspec.Root{}}, "windows", specerror.RootOnHyperVNotSet},
		{rspec.Spec{Windows: &rspec.Windows{HyperV: &rspec.WindowsHyperV{}}, Root: nil}, "windows", specerror.NonError},
		{rspec.Spec{Windows: &rspec.Windows{}, Root: &rspec.Root{Path: filepath.Join(tmpBundle, "rootfs")}}, "windows", specerror.RootPathOnWindowsGUID},
		{rspec.Spec{Windows: &rspec.Windows{}, Root: &rspec.Root{Path: "\\\\?\\Volume{ec84d99e-3f02-11e7-ac6c-00155d7682cf}\\"}}, "windows", specerror.NonError},
		{rspec.Spec{Root: nil}, "linux", specerror.RootOnNonHyperVRequired},
		{rspec.Spec{Root: &rspec.Root{Path: "maverick-rootfs"}}, "linux", specerror.RootPathOnPosixConvention},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs"}}, "linux", specerror.NonError},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonExists)}}, "linux", specerror.RootPathExist},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, rootfsNonDir)}}, "linux", specerror.RootPathExist},
		{rspec.Spec{Root: &rspec.Root{Path: filepath.Join(tmpBundle, "rootfs")}}, "linux", specerror.NonError},
		{rspec.Spec{Root: &rspec.Root{Path: "rootfs/rootfs"}}, "linux", specerror.ArtifactsInSingleDir},
		{rspec.Spec{Root: &rspec.Root{Readonly: true}}, "windows", specerror.RootReadonlyOnWindowsFalse},
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
		{"invalid", specerror.SpecVersionInSemVer},
	}

	for _, c := range cases {
		v := NewValidator(&rspec.Spec{Version: c.val}, "", false, "linux")
		err := v.CheckSemVer()
		assert.Equal(t, c.expected, specerror.FindError(err, c.expected), "Fail to check SemVer "+c.val)
	}
}

func TestCheckProcess(t *testing.T) {
	cases := []struct {
		val      rspec.Spec
		platform string
		expected specerror.Code
	}{
		{
			val: rspec.Spec{
				Version: "1.0.0",
				Process: &rspec.Process{
					Args: []string{"sh"},
					Cwd:  "/",
				},
			},
			platform: "linux",
			expected: specerror.NonError,
		},
		{
			val: rspec.Spec{
				Version: "1.0.0",
				Process: &rspec.Process{
					Args: []string{"sh"},
					Cwd:  "/",
					Rlimits: []rspec.POSIXRlimit{
						{
							Type: "RLIMIT_NOFILE",
							Hard: 1024,
							Soft: 1024,
						},
						{
							Type: "RLIMIT_NPROC",
							Hard: 512,
							Soft: 512,
						},
					},
				},
			},
			platform: "linux",
			expected: specerror.NonError,
		},
		{
			val: rspec.Spec{
				Version: "1.0.0",
				Process: &rspec.Process{
					Args: []string{"sh"},
					Cwd:  "/",
					Rlimits: []rspec.POSIXRlimit{
						{
							Type: "RLIMIT_NOFILE",
							Hard: 1024,
							Soft: 1024,
						},
					},
				},
			},
			platform: "solaris",
			expected: specerror.NonError,
		},
		{
			val: rspec.Spec{
				Version: "1.0.0",
				Process: &rspec.Process{
					Args: []string{"sh"},
					Cwd:  "/",
					Rlimits: []rspec.POSIXRlimit{
						{
							Type: "RLIMIT_DOES_NOT_EXIST",
							Hard: 512,
							Soft: 512,
						},
					},
				},
			},
			platform: "linux",
			expected: specerror.PosixProcRlimitsTypeValueError,
		},
		{
			val: rspec.Spec{
				Version: "1.0.0",
				Process: &rspec.Process{
					Args: []string{"sh"},
					Cwd:  "/",
					Rlimits: []rspec.POSIXRlimit{
						{
							Type: "RLIMIT_NPROC",
							Hard: 512,
							Soft: 512,
						},
					},
				},
			},
			platform: "solaris",
			expected: specerror.PosixProcRlimitsTypeValueError,
		},
	}
	for _, c := range cases {
		v := NewValidator(&c.val, ".", false, c.platform)
		err := v.CheckProcess()
		assert.Equal(t, c.expected, specerror.FindError(err, c.expected), fmt.Sprintf("failed CheckProcess: %v %d", err, c.expected))
	}
}
