package generate_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	rfc2119 "github.com/opencontainers/runtime-tools/error"
	"github.com/opencontainers/runtime-tools/generate"
	"github.com/opencontainers/runtime-tools/specerror"
	"github.com/opencontainers/runtime-tools/validate"
	"github.com/stretchr/testify/assert"
)

// Smoke test to ensure that _at the very least_ our default configuration
// passes the validation tests. If this test fails, something is _very_ wrong
// and needs to be fixed immediately (as it will break downstreams that depend
// on us for a "sane default" and do compliance testing -- such as umoci).
func TestGenerateValid(t *testing.T) {
	plat := "linux"
	if runtime.GOOS == "windows" {
		plat = "windows"
	}

	isolations := []string{"process", "hyperv"}
	for _, isolation := range isolations {
		if plat == "linux" && isolation == "hyperv" {
			// Combination doesn't make sense.
			continue
		}

		bundle, err := os.MkdirTemp("", "TestGenerateValid_bundle")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(bundle)

		// Create our toy bundle.
		rootfsPath := filepath.Join(bundle, "rootfs")
		if err := os.Mkdir(rootfsPath, 0o755); err != nil {
			t.Fatal(err)
		}
		configPath := filepath.Join(bundle, "config.json")
		g, err := generate.New(plat)
		if err != nil {
			t.Fatal(err)
		}
		if runtime.GOOS == "windows" {
			g.AddWindowsLayerFolders("C:\\fakelayer")
			g.AddWindowsLayerFolders("C:\\fakescratch")
			if isolation == "process" {
				// Add the Rootfs section (note: fake volume guid)
				g.SetRootPath("\\\\?\\Volume{ec84d99e-3f02-11e7-ac6c-00155d7682cf}\\")
			} else {
				// Add the Hyper-V section
				g.SetWindowsHypervUntilityVMPath("")
			}
		}
		if err := (&g).SaveToFile(configPath, generate.ExportOptions{Seccomp: false}); err != nil {
			t.Fatal(err)
		}

		// Validate the bundle.
		v, err := validate.NewValidatorFromPath(bundle, true, runtime.GOOS)
		if err != nil {
			t.Errorf("unexpected NewValidatorFromPath error: %+v", err)
		}
		if err := v.CheckAll(); err != nil {
			levelErrors, err := specerror.SplitLevel(err, rfc2119.Must)
			if err != nil {
				t.Errorf("unexpected non-multierror: %+v", err)
				return
			}
			for _, e := range levelErrors.Warnings {
				t.Logf("unexpected warning: %v", e)
			}
			if err := levelErrors.Error; err != nil {
				t.Errorf("unexpected MUST error(s): %+v", err)
			}
		}
	}
}

func TestRemoveMount(t *testing.T) {
	g, err := generate.New("linux")
	if err != nil {
		t.Fatal(err)
	}
	size := len(g.Mounts())
	g.RemoveMount("/dev/shm")
	if size-1 != len(g.Mounts()) {
		t.Errorf("Unable to remove /dev/shm from mounts")
	}
}

func TestEnvCaching(t *testing.T) {
	// Start with empty ENV and add a few
	g, err := generate.New("windows")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"k1=v1", "k2=v2"}
	g.AddProcessEnv("k1", "v1")
	g.AddProcessEnv("k2", "v2")
	assert.Equal(t, expected, g.Config.Process.Env)

	// Test override and existing ENV
	g, err = generate.New("linux")
	if err != nil {
		t.Fatal(err)
	}
	expected = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "TERM=xterm", "k1=v1", "k2=v4", "k3=v3"}
	g.AddProcessEnv("k1", "v1")
	g.AddProcessEnv("k2", "v2")
	g.AddProcessEnv("k3", "v3")
	g.AddProcessEnv("k2", "v4")
	assert.Equal(t, expected, g.Config.Process.Env)

	// Test empty ENV
	g, err = generate.New("windows")
	if err != nil {
		t.Fatal(err)
	}
	g.AddProcessEnv("", "")
	assert.Equal(t, []string(nil), g.Config.Process.Env)
}

func TestMultipleEnvCaching(t *testing.T) {
	// Start with empty ENV and add a few
	g, err := generate.New("windows")
	if err != nil {
		t.Fatal(err)
	}
	newEnvs := []string{"k1=v1", "k2=v2"}
	expected := []string{"k1=v1", "k2=v2"}
	g.AddMultipleProcessEnv(newEnvs)
	assert.Equal(t, expected, g.Config.Process.Env)

	// Test override and existing ENV
	g, err = generate.New("linux")
	if err != nil {
		t.Fatal(err)
	}
	newEnvs = []string{"k1=v1", "k2=v2", "k3=v3", "k2=v4"}
	expected = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "TERM=xterm", "k1=v1", "k2=v4", "k3=v3"}
	g.AddMultipleProcessEnv(newEnvs)
	assert.Equal(t, expected, g.Config.Process.Env)

	// Test empty ENV
	g, err = generate.New("windows")
	if err != nil {
		t.Fatal(err)
	}
	g.AddMultipleProcessEnv([]string{})
	assert.Equal(t, []string(nil), g.Config.Process.Env)
}
