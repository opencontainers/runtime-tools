package generate

import (
	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

func (g *Generator) initConfig() {
	if g.Config == nil {
		g.Config = &rspec.Spec{}
	}
}

func (g *Generator) initConfigAnnotations() {
	g.initConfig()
	if g.Config.Annotations == nil {
		g.Config.Annotations = make(map[string]string)
	}
}

func (g *Generator) initConfigLinux() {
	g.initConfig()
	if g.Config.Linux == nil {
		g.Config.Linux = &rspec.Linux{}
	}
}

func (g *Generator) initConfigLinuxSysctl() {
	g.initConfigLinux()
	if g.Config.Linux.Sysctl == nil {
		g.Config.Linux.Sysctl = make(map[string]string)
	}
}

func (g *Generator) initConfigLinuxSeccomp() {
	g.initConfigLinux()
	if g.Config.Linux.Seccomp == nil {
		g.Config.Linux.Seccomp = &rspec.Seccomp{}
	}
}

func (g *Generator) initConfigLinuxResources() {
	g.initConfigLinux()
	if g.Config.Linux.Resources == nil {
		g.Config.Linux.Resources = &rspec.Resources{}
	}
}

func (g *Generator) initConfigLinuxResourcesCPU() {
	g.initConfigLinuxResources()
	if g.Config.Linux.Resources.CPU == nil {
		g.Config.Linux.Resources.CPU = &rspec.CPU{}
	}
}

func (g *Generator) initConfigLinuxResourcesMemory() {
	g.initConfigLinuxResources()
	if g.Config.Linux.Resources.Memory == nil {
		g.Config.Linux.Resources.Memory = &rspec.Memory{}
	}
}

func (g *Generator) initConfigLinuxResourcesPids() {
	g.initConfigLinuxResources()
	if g.Config.Linux.Resources.Pids == nil {
		g.Config.Linux.Resources.Pids = &rspec.Pids{}
	}
}
