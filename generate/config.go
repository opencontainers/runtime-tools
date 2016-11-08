package generate

import (
	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

func (g *Generator) InitConfig() {
	if g.Config == nil {
		g.Config = &rspec.Spec{}
	}
}

func (g *Generator) InitConfigAnnotations() {
	g.InitConfig()
	if g.Config.Annotations == nil {
		g.Config.Annotations = make(map[string]string)
	}
}

func (g *Generator) InitConfigLinux() {
	g.InitConfig()
	if g.Config.Linux == nil {
		g.Config.Linux = &rspec.Linux{}
	}
}

func (g *Generator) InitConfigLinuxSeccomp() {
	g.InitConfigLinux()
	if g.Config.Linux.Seccomp == nil {
		g.Config.Linux.Seccomp = &rspec.Seccomp{}
	}
}

func (g *Generator) InitConfigLinuxResources() {
	g.InitConfigLinux()
	if g.Config.Linux.Resources == nil {
		g.Config.Linux.Resources = &rspec.Resources{}
	}
}

func (g *Generator) InitConfigLinuxResourcesCPU() {
	g.InitConfigLinuxResources()
	if g.Config.Linux.Resources.CPU == nil {
		g.Config.Linux.Resources.CPU = &rspec.CPU{}
	}
}

func (g *Generator) InitConfigLinuxResourcesMemory() {
	g.InitConfigLinuxResources()
	if g.Config.Linux.Resources.Memory == nil {
		g.Config.Linux.Resources.Memory = &rspec.Memory{}
	}
}

func (g *Generator) InitConfigLinuxResourcesPids() {
	g.InitConfigLinuxResources()
	if g.Config.Linux.Resources.Pids == nil {
		g.Config.Linux.Resources.Pids = &rspec.Pids{}
	}
}
