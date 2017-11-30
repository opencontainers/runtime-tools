// Package sanitize removes dangerous and questionably-portable properties from container configurations.
package sanitize

import (
	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

// Santize removes dangerous and questionably-portable properties from container configurations.
func Sanitize(config *rspec.Spec) (err error) {
	config.Process.Terminal = false
	//config.Process.ConsoleSize = nil  // needs runtime-spec#581
	config.Process.User.AdditionalGids = []uint32{}
	config.Process.Capabilities = []string{}
	config.Process.Rlimits = []rspec.Rlimit{}
	config.Process.NoNewPrivileges = false
	config.Process.ApparmorProfile = ""
	config.Process.SelinuxLabel = ""
	config.Root = rspec.Root{
		Path: "rootfs",
	}
	config.Hostname = ""
	//config.Hooks = nil  // needs runtime-spec#427

	for i, _ := range config.Mounts {
		config.Mounts[i].Source = ""
	}

	if config.Linux != nil {
		config.Linux.UIDMappings = []rspec.IDMapping{}
		config.Linux.GIDMappings = []rspec.IDMapping{}
		config.Linux.Sysctl = map[string]string{}
		config.Linux.CgroupsPath = nil
		config.Linux.Namespaces = []rspec.Namespace{}
		config.Linux.Devices = []rspec.Device{}
		config.Linux.Seccomp = nil
		config.Linux.RootfsPropagation = ""
		config.Linux.MaskedPaths = []string{}
		config.Linux.MaskedPaths = []string{}
		config.Linux.MountLabel = ""

		if config.Linux.Resources != nil {
			config.Linux.Resources.Devices = []rspec.DeviceCgroup{}
			config.Linux.Resources.DisableOOMKiller = nil
			config.Linux.Resources.OOMScoreAdj= nil
			config.Linux.Resources.Pids = nil
			config.Linux.Resources.BlockIO = nil
			config.Linux.Resources.HugepageLimits = []rspec.HugepageLimit{}
			config.Linux.Resources.Network = nil

			if config.Linux.Resources.Memory != nil {
				config.Linux.Resources.Memory.Kernel = nil
				config.Linux.Resources.Memory.KernelTCP = nil
				config.Linux.Resources.Memory.Swappiness = nil
			}

			if config.Linux.Resources.CPU != nil {
				config.Linux.Resources.CPU.Quota = nil
				config.Linux.Resources.CPU.Period = nil
				config.Linux.Resources.CPU.RealtimeRuntime = nil
				config.Linux.Resources.CPU.Period = nil
				config.Linux.Resources.CPU.Cpus = nil
				config.Linux.Resources.CPU.Mems = nil
			}
		}
	}

	if config.Solaris != nil {
		config.Solaris.Milestone = ""
		config.Solaris.LimitPriv = ""
		config.Solaris.MaxShmMemory = ""
		config.Solaris.Anet = []rspec.Anet{}
		config.Solaris.CappedCPU = nil
	}

	return nil
}
