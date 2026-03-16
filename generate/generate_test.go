package generate_test

import (
	"testing"

	"github.com/opencontainers/runtime-tools/generate"
	"github.com/stretchr/testify/assert"
)

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
