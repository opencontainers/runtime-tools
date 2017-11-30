package validation

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mrunalp/fileutils"
	rspecs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/opencontainers/runtime-tools/generate"
	"github.com/opencontainers/runtime-tools/specerror"
)

var (
	runtimeCommand = "runc"
)

// build test environment before running container
type preFunc func(string) error

func init() {
	runtimeInEnv := os.Getenv("RUNTIME")
	if runtimeInEnv != "" {
		runtimeCommand = runtimeInEnv
	}
}

func prepareBundle() (string, error) {
	// Setup a temporary test directory
	bundleDir, err := ioutil.TempDir("", "ocitest")
	if err != nil {
		return "", err
	}

	// Untar the root fs
	untarCmd := exec.Command("tar", "-xf", fmt.Sprintf("../rootfs-%s.tar.gz", runtime.GOARCH), "-C", bundleDir)
	_, err = untarCmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(bundleDir)
		return "", err
	}

	return bundleDir, nil
}

func getDefaultGenerator() *generate.Generator {
	g := generate.New()
	g.SetRootPath(".")
	g.SetProcessArgs([]string{"/runtimetest", "--path=/"})
	return &g
}

func runtimeInsideValidate(g *generate.Generator, f preFunc) error {
	bundleDir, err := prepareBundle()
	if err != nil {
		return err
	}

	if f != nil {
		if err := f(bundleDir); err != nil {
			return err
		}
	}

	r, err := NewRuntime(runtimeCommand, bundleDir)
	if err != nil {
		os.RemoveAll(bundleDir)
		return err
	}
	defer r.Clean(true)
	err = r.SetConfig(g)
	if err != nil {
		return err
	}
	err = fileutils.CopyFile("../runtimetest", filepath.Join(bundleDir, "runtimetest"))
	if err != nil {
		return err
	}

	r.SetID(uuid.NewV4().String())
	err = r.Create()
	if err != nil {
		return err
	}
	return r.Start()
}

func TestValidateBasic(t *testing.T) {
	g := getDefaultGenerator()

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test whether rootfs Readonly can be applied as false
func TestValidateRootFSReadWrite(t *testing.T) {
	g := getDefaultGenerator()
	g.SetRootReadonly(false)

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test whether rootfs Readonly can be applied as true
func TestValidateRootFSReadonly(t *testing.T) {
	if "windows" == runtime.GOOS {
		t.Skip("skip this test on windows platform")
	}

	g := getDefaultGenerator()
	g.SetRootReadonly(true)

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test Process
func TestValidateProcess(t *testing.T) {
	g := getDefaultGenerator()
	g.SetProcessCwd("/test")
	g.AddProcessEnv("testa", "valuea")
	g.AddProcessEnv("testb", "123")

	assert.Nil(t, runtimeInsideValidate(g, func(path string) error {
		pathName := filepath.Join(path, "test")
		return os.MkdirAll(pathName, 0700)
	}))
}

// Test whether Capabilites can be applied or not
func TestValidateCapabilities(t *testing.T) {
	if "linux" != runtime.GOOS {
		t.Skip("skip linux-specific capabilities test")
	}

	g := getDefaultGenerator()
	g.SetupPrivileged(true)

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test whether hostname can be applied or not
func TestValidateHostname(t *testing.T) {
	g := getDefaultGenerator()
	g.SetHostname("hostname-specific")

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

func TestValidateRootfsPropagationPrivate(t *testing.T) {
	t.Skip("has not been implemented yet")
}

func TestValidateRootfsPropagationSlave(t *testing.T) {
	t.Skip("has not been implemented yet")
}

func TestValidateRootfsPropagationShared(t *testing.T) {
	g := getDefaultGenerator()
	g.SetupPrivileged(true)
	g.SetLinuxRootPropagation("shared")

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

func TestValidateRootfsPropagationUnbindable(t *testing.T) {
	g := getDefaultGenerator()
	g.SetupPrivileged(true)
	g.SetLinuxRootPropagation("unbindable")

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

func TestValidateLinuxDevices(t *testing.T) {
	g := getDefaultGenerator()

	// add char device
	cdev := rspecs.LinuxDevice{}
	cdev.Path = "/dev/test1"
	cdev.Type = "c"
	cdev.Major = 10
	cdev.Minor = 666
	cmode := os.FileMode(int32(432))
	cdev.FileMode = &cmode
	cuid := uint32(0)
	cdev.UID = &cuid
	cgid := uint32(0)
	cdev.GID = &cgid
	g.AddDevice(cdev)
	// add block device
	bdev := rspecs.LinuxDevice{}
	bdev.Path = "/dev/test2"
	bdev.Type = "b"
	bdev.Major = 8
	bdev.Minor = 666
	bmode := os.FileMode(int32(432))
	bdev.FileMode = &bmode
	uid := uint32(0)
	bdev.UID = &uid
	gid := uint32(0)
	bdev.GID = &gid
	g.AddDevice(bdev)
	// add fifo device
	pdev := rspecs.LinuxDevice{}
	pdev.Path = "/dev/test3"
	pdev.Type = "p"
	pdev.Major = 8
	pdev.Minor = 666
	pmode := os.FileMode(int32(432))
	pdev.FileMode = &pmode
	g.AddDevice(pdev)

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

func TestValidateMaskedPaths(t *testing.T) {
	g := getDefaultGenerator()
	g.AddLinuxMaskedPaths("/masktest")

	assert.Nil(t, runtimeInsideValidate(g, func(path string) error {
		pathName := filepath.Join(path, "masktest")
		return os.MkdirAll(pathName, 0700)
	}))
}

func TestValidateROPaths(t *testing.T) {
	g := getDefaultGenerator()
	g.AddLinuxReadonlyPaths("readonlytest")

	assert.Nil(t, runtimeInsideValidate(g, func(path string) error {
		pathName := filepath.Join(path, "readonlytest")
		return os.MkdirAll(pathName, 0700)
	}))
}

func TestValidateOOMScoreAdj(t *testing.T) {
	g := getDefaultGenerator()
	g.SetProcessOOMScoreAdj(500)

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

func TestValidateUIDMappings(t *testing.T) {
	g := getDefaultGenerator()
	g.AddLinuxUIDMapping(uint32(1000), uint32(0), uint32(3200))

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

func TestValidateGIDMappings(t *testing.T) {
	g := getDefaultGenerator()
	g.AddLinuxGIDMapping(uint32(1000), uint32(0), uint32(3200))

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test whether mounts are correctly mounted
func TestValidateMounts(t *testing.T) {
	// TODO mounts generation options have not been implemented
	// will add it after 'mounts generate' done
}

// Test whether rlimits can be applied or not
func TestValidateRlimits(t *testing.T) {
	g := getDefaultGenerator()
	g.AddProcessRlimits("RLIMIT_NOFILE", 1024, 1024)

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test whether sysctls can be applied or not
func TestValidateSysctls(t *testing.T) {
	g := getDefaultGenerator()
	g.AddLinuxSysctl("net.ipv4.ip_forward", "1")

	assert.Nil(t, runtimeInsideValidate(g, nil))
}

// Test Create operation
func TestValidateCreate(t *testing.T) {
	g := generate.New()
	g.SetRootPath(".")
	g.SetProcessArgs([]string{"ls"})

	bundleDir, err := prepareBundle()
	assert.Nil(t, err)

	r, err := NewRuntime(runtimeCommand, bundleDir)
	assert.Nil(t, err)
	defer r.Clean(true)

	err = r.SetConfig(&g)
	assert.Nil(t, err)

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
		err := r.Create()
		assert.Equal(t, c.errExpected, err == nil, c.err.Error())

		if err == nil {
			state, _ := r.State()
			assert.Equal(t, c.id, state.ID, c.err.Error())
		}
	}
}
