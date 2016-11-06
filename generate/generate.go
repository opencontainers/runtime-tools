// Package generate implements functions generating container config files.
package generate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/generate/seccomp"
	"github.com/syndtr/gocapability/capability"
)

var (
	// Namespaces include the names of supported namespaces.
	Namespaces = []string{"network", "pid", "mount", "ipc", "uts", "user", "cgroup"}
)

// Generator represents a generator for a container config.
type Generator struct {
	Config       *rspec.Spec
	HostSpecific bool
}

// ExportOptions have toggles for exporting only certain parts of the specification
type ExportOptions struct {
	Seccomp bool // seccomp toggles if only seccomp should be exported
}

// New creates a config Generator with the default config.
func New() Generator {
	config := rspec.Spec{
		Version: rspec.Version,
		Platform: rspec.Platform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Root: rspec.Root{
			Path:     "",
			Readonly: false,
		},
		Process: rspec.Process{
			Terminal: false,
			User:     rspec.User{},
			Args: []string{
				"sh",
			},
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"TERM=xterm",
			},
			Cwd: "/",
			Capabilities: []string{
				"CAP_CHOWN",
				"CAP_DAC_OVERRIDE",
				"CAP_FSETID",
				"CAP_FOWNER",
				"CAP_MKNOD",
				"CAP_NET_RAW",
				"CAP_SETGID",
				"CAP_SETUID",
				"CAP_SETFCAP",
				"CAP_SETPCAP",
				"CAP_NET_BIND_SERVICE",
				"CAP_SYS_CHROOT",
				"CAP_KILL",
				"CAP_AUDIT_WRITE",
			},
			Rlimits: []rspec.Rlimit{
				{
					Type: "RLIMIT_NOFILE",
					Hard: uint64(1024),
					Soft: uint64(1024),
				},
			},
		},
		Hostname: "mrsdalloway",
		Mounts: []rspec.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
				Options:     nil,
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
			},
			{
				Destination: "/dev/mqueue",
				Type:        "mqueue",
				Source:      "mqueue",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev", "ro"},
			},
		},
		Linux: &rspec.Linux{
			Resources: &rspec.Resources{
				Devices: []rspec.DeviceCgroup{
					{
						Allow:  false,
						Access: strPtr("rwm"),
					},
				},
			},
			Namespaces: []rspec.Namespace{
				{
					Type: "pid",
				},
				{
					Type: "network",
				},
				{
					Type: "ipc",
				},
				{
					Type: "uts",
				},
				{
					Type: "mount",
				},
			},
			Devices: []rspec.Device{},
		},
	}
	config.Linux.Seccomp = seccomp.DefaultProfile(&config)
	return Generator{
		Config: &config,
	}
}

// NewFromConfig creates a config Generator from a given config.
func NewFromConfig(config *rspec.Spec) Generator {
	return Generator{
		Config: config,
	}
}

// NewFromFile loads the template specifed in a file into a config Generator.
func NewFromFile(path string) (Generator, error) {
	cf, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Generator{}, fmt.Errorf("template configuration at %s not found", path)
		}
	}
	defer cf.Close()

	return NewFromTemplate(cf)
}

// NewFromTemplate loads the template from io.Reader into a config Generator.
func NewFromTemplate(r io.Reader) (Generator, error) {
	var config rspec.Spec
	if err := json.NewDecoder(r).Decode(&config); err != nil {
		return Generator{}, err
	}
	return Generator{
		Config: &config,
	}, nil
}

// Save writes the config into w.
func (g *Generator) Save(w io.Writer, exportOpts ExportOptions) (err error) {
	var data []byte

	if exportOpts.Seccomp {
		data, err = json.MarshalIndent(g.Config.Linux.Seccomp, "", "\t")
	} else {
		data, err = json.MarshalIndent(g.Config, "", "\t")
	}
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// SaveToFile writes the config into a file.
func (g *Generator) SaveToFile(path string, exportOpts ExportOptions) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return g.Save(f, exportOpts)
}

// SetVersion sets g.Config.Version.
func (g *Generator) SetVersion(version string) {
	g.initConfig()
	g.Config.Version = version
}

// SetRootPath sets g.Config.Root.Path.
func (g *Generator) SetRootPath(path string) {
	g.initConfig()
	g.Config.Root.Path = path
}

// SetRootReadonly sets g.Config.Root.Readonly.
func (g *Generator) SetRootReadonly(b bool) {
	g.initConfig()
	g.Config.Root.Readonly = b
}

// SetHostname sets g.Config.Hostname.
func (g *Generator) SetHostname(s string) {
	g.initConfig()
	g.Config.Hostname = s
}

// ClearAnnotations clears g.Config.Annotations.
func (g *Generator) ClearAnnotations() {
	if g.Config == nil {
		return
	}
	g.Config.Annotations = make(map[string]string)
}

// AddAnnotation adds an annotation into g.Config.Annotations.
func (g *Generator) AddAnnotation(key, value string) {
	g.initConfigAnnotations()
	g.Config.Annotations[key] = value
}

// RemoveAnnotation remove an annotation from g.Config.Annotations.
func (g *Generator) RemoveAnnotation(key string) {
	if g.Config == nil || g.Config.Annotations == nil {
		return
	}
	delete(g.Config.Annotations, key)
}

// SetPlatformOS sets g.Config.Process.OS.
func (g *Generator) SetPlatformOS(os string) {
	g.initConfig()
	g.Config.Platform.OS = os
}

// SetPlatformArch sets g.Config.Platform.Arch.
func (g *Generator) SetPlatformArch(arch string) {
	g.initConfig()
	g.Config.Platform.Arch = arch
}

// SetProcessUID sets g.Config.Process.User.UID.
func (g *Generator) SetProcessUID(uid uint32) {
	g.initConfig()
	g.Config.Process.User.UID = uid
}

// SetProcessGID sets g.Config.Process.User.GID.
func (g *Generator) SetProcessGID(gid uint32) {
	g.initConfig()
	g.Config.Process.User.GID = gid
}

// SetProcessCwd sets g.Config.Process.Cwd.
func (g *Generator) SetProcessCwd(cwd string) {
	g.initConfig()
	g.Config.Process.Cwd = cwd
}

// SetProcessNoNewPrivileges sets g.Config.Process.NoNewPrivileges.
func (g *Generator) SetProcessNoNewPrivileges(b bool) {
	g.initConfig()
	g.Config.Process.NoNewPrivileges = b
}

// SetProcessTerminal sets g.Config.Process.Terminal.
func (g *Generator) SetProcessTerminal(b bool) {
	g.initConfig()
	g.Config.Process.Terminal = b
}

// SetProcessApparmorProfile sets g.Config.Process.ApparmorProfile.
func (g *Generator) SetProcessApparmorProfile(prof string) {
	g.initConfig()
	g.Config.Process.ApparmorProfile = prof
}

// SetProcessArgs sets g.Config.Process.Args.
func (g *Generator) SetProcessArgs(args []string) {
	g.initConfig()
	g.Config.Process.Args = args
}

// ClearProcessEnv clears g.Config.Process.Env.
func (g *Generator) ClearProcessEnv() {
	if g.Config == nil {
		return
	}
	g.Config.Process.Env = []string{}
}

// AddProcessEnv adds env into g.Config.Process.Env.
func (g *Generator) AddProcessEnv(env string) {
	g.initConfig()
	g.Config.Process.Env = append(g.Config.Process.Env, env)
}

// ClearProcessAdditionalGids clear g.Config.Process.AdditionalGids.
func (g *Generator) ClearProcessAdditionalGids() {
	if g.Config == nil {
		return
	}
	g.Config.Process.User.AdditionalGids = []uint32{}
}

// AddProcessAdditionalGid adds an additional gid into g.Config.Process.AdditionalGids.
func (g *Generator) AddProcessAdditionalGid(gid uint32) {
	g.initConfig()
	for _, group := range g.Config.Process.User.AdditionalGids {
		if group == gid {
			return
		}
	}
	g.Config.Process.User.AdditionalGids = append(g.Config.Process.User.AdditionalGids, gid)
}

// SetProcessSelinuxLabel sets g.Config.Process.SelinuxLabel.
func (g *Generator) SetProcessSelinuxLabel(label string) {
	g.initConfig()
	g.Config.Process.SelinuxLabel = label
}

// SetLinuxCgroupsPath sets g.Config.Linux.CgroupsPath.
func (g *Generator) SetLinuxCgroupsPath(path string) {
	g.initConfigLinux()
	g.Config.Linux.CgroupsPath = strPtr(path)
}

// SetLinuxMountLabel sets g.Config.Linux.MountLabel.
func (g *Generator) SetLinuxMountLabel(label string) {
	g.initConfigLinux()
	g.Config.Linux.MountLabel = label
}

// SetLinuxResourcesDisableOOMKiller sets g.Config.Linux.Resources.DisableOOMKiller.
func (g *Generator) SetLinuxResourcesDisableOOMKiller(disable bool) {
	g.initConfigLinuxResources()
	g.Config.Linux.Resources.DisableOOMKiller = &disable
}

// SetLinuxResourcesOOMScoreAdj sets g.Config.Linux.Resources.OOMScoreAdj.
func (g *Generator) SetLinuxResourcesOOMScoreAdj(adj int) {
	g.initConfigLinuxResources()
	g.Config.Linux.Resources.OOMScoreAdj = &adj
}

// SetLinuxResourcesCPUShares sets g.Config.Linux.Resources.CPU.Shares.
func (g *Generator) SetLinuxResourcesCPUShares(shares uint64) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.Shares = &shares
}

// SetLinuxResourcesCPUQuota sets g.Config.Linux.Resources.CPU.Quota.
func (g *Generator) SetLinuxResourcesCPUQuota(quota uint64) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.Quota = &quota
}

// SetLinuxResourcesCPUPeriod sets g.Config.Linux.Resources.CPU.Period.
func (g *Generator) SetLinuxResourcesCPUPeriod(period uint64) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.Period = &period
}

// SetLinuxResourcesCPURealtimeRuntime sets g.Config.Linux.Resources.CPU.RealtimeRuntime.
func (g *Generator) SetLinuxResourcesCPURealtimeRuntime(time uint64) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.RealtimeRuntime = &time
}

// SetLinuxResourcesCPURealtimePeriod sets g.Config.Linux.Resources.CPU.RealtimePeriod.
func (g *Generator) SetLinuxResourcesCPURealtimePeriod(period uint64) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.RealtimePeriod = &period
}

// SetLinuxResourcesCPUCpus sets g.Config.Linux.Resources.CPU.Cpus.
func (g *Generator) SetLinuxResourcesCPUCpus(cpus string) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.Cpus = &cpus
}

// SetLinuxResourcesCPUMems sets g.Config.Linux.Resources.CPU.Mems.
func (g *Generator) SetLinuxResourcesCPUMems(mems string) {
	g.initConfigLinuxResourcesCPU()
	g.Config.Linux.Resources.CPU.Mems = &mems
}

// SetLinuxResourcesMemoryLimit sets g.Config.Linux.Resources.Memory.Limit.
func (g *Generator) SetLinuxResourcesMemoryLimit(limit uint64) {
	g.initConfigLinuxResourcesMemory()
	g.Config.Linux.Resources.Memory.Limit = &limit
}

// SetLinuxResourcesMemoryReservation sets g.Config.Linux.Resources.Memory.Reservation.
func (g *Generator) SetLinuxResourcesMemoryReservation(reservation uint64) {
	g.initConfigLinuxResourcesMemory()
	g.Config.Linux.Resources.Memory.Reservation = &reservation
}

// SetLinuxResourcesMemorySwap sets g.Config.Linux.Resources.Memory.Swap.
func (g *Generator) SetLinuxResourcesMemorySwap(swap uint64) {
	g.initConfigLinuxResourcesMemory()
	g.Config.Linux.Resources.Memory.Swap = &swap
}

// SetLinuxResourcesMemoryKernel sets g.Config.Linux.Resources.Memory.Kernel.
func (g *Generator) SetLinuxResourcesMemoryKernel(kernel uint64) {
	g.initConfigLinuxResourcesMemory()
	g.Config.Linux.Resources.Memory.Kernel = &kernel
}

// SetLinuxResourcesMemoryKernelTCP sets g.Config.Linux.Resources.Memory.KernelTCP.
func (g *Generator) SetLinuxResourcesMemoryKernelTCP(kernelTCP uint64) {
	g.initConfigLinuxResourcesMemory()
	g.Config.Linux.Resources.Memory.KernelTCP = &kernelTCP
}

// SetLinuxResourcesMemorySwappiness sets g.Config.Linux.Resources.Memory.Swappiness.
func (g *Generator) SetLinuxResourcesMemorySwappiness(swappiness uint64) {
	g.initConfigLinuxResourcesMemory()
	g.Config.Linux.Resources.Memory.Swappiness = &swappiness
}

// SetLinuxResourcesPidsLimit sets g.Config.Linux.Resources.Pids.Limit.
func (g *Generator) SetLinuxResourcesPidsLimit(limit int64) {
	g.initConfigLinuxResourcesPids()
	g.Config.Linux.Resources.Pids.Limit = &limit
}

// ClearLinuxSysctl clears g.Config.Linux.Sysctl.
func (g *Generator) ClearLinuxSysctl() {
	if g.Config == nil || g.Config.Linux == nil {
		return
	}
	g.Config.Linux.Sysctl = make(map[string]string)
}

// AddLinuxSysctl adds a new sysctl config into g.Config.Linux.Sysctl.
func (g *Generator) AddLinuxSysctl(key, value string) {
	g.initConfigLinuxSysctl()
	g.Config.Linux.Sysctl[key] = value
}

// RemoveLinuxSysctl removes a sysctl config from g.Config.Linux.Sysctl.
func (g *Generator) RemoveLinuxSysctl(key string) {
	if g.Config == nil || g.Config.Linux == nil || g.Config.Linux.Sysctl == nil {
		return
	}
	delete(g.Config.Linux.Sysctl, key)
}

// ClearLinuxUIDMappings clear g.Config.Linux.UIDMappings.
func (g *Generator) ClearLinuxUIDMappings() {
	if g.Config == nil || g.Config.Linux == nil {
		return
	}
	g.Config.Linux.UIDMappings = []rspec.IDMapping{}
}

// AddLinuxUIDMapping adds uidMap into g.Config.Linux.UIDMappings.
func (g *Generator) AddLinuxUIDMapping(hid, cid, size uint32) {
	idMapping := rspec.IDMapping{
		HostID:      hid,
		ContainerID: cid,
		Size:        size,
	}

	g.initConfigLinux()
	g.Config.Linux.UIDMappings = append(g.Config.Linux.UIDMappings, idMapping)
}

// ClearLinuxGIDMappings clear g.Config.Linux.GIDMappings.
func (g *Generator) ClearLinuxGIDMappings() {
	if g.Config == nil || g.Config.Linux == nil {
		return
	}
	g.Config.Linux.GIDMappings = []rspec.IDMapping{}
}

// AddLinuxGIDMapping adds gidMap into g.Config.Linux.GIDMappings.
func (g *Generator) AddLinuxGIDMapping(hid, cid, size uint32) {
	idMapping := rspec.IDMapping{
		HostID:      hid,
		ContainerID: cid,
		Size:        size,
	}

	g.initConfigLinux()
	g.Config.Linux.GIDMappings = append(g.Config.Linux.GIDMappings, idMapping)
}

// SetLinuxRootPropagation sets g.Config.Linux.RootfsPropagation.
func (g *Generator) SetLinuxRootPropagation(rp string) error {
	switch rp {
	case "":
	case "private":
	case "rprivate":
	case "slave":
	case "rslave":
	case "shared":
	case "rshared":
	default:
		return fmt.Errorf("rootfs-propagation must be empty or one of private|rprivate|slave|rslave|shared|rshared")
	}
	g.initConfigLinux()
	g.Config.Linux.RootfsPropagation = rp
	return nil
}

// ClearPreStartHooks clear g.Config.Hooks.Prestart.
func (g *Generator) ClearPreStartHooks() {
	if g.Config == nil {
		return
	}
	g.Config.Hooks.Prestart = []rspec.Hook{}
}

// AddPreStartHook add a prestart hook into g.Config.Hooks.Prestart.
func (g *Generator) AddPreStartHook(path string, args []string) {
	g.initConfig()
	hook := rspec.Hook{Path: path, Args: args}
	g.Config.Hooks.Prestart = append(g.Config.Hooks.Prestart, hook)
}

// ClearPostStopHooks clear g.Config.Hooks.Poststop.
func (g *Generator) ClearPostStopHooks() {
	if g.Config == nil {
		return
	}
	g.Config.Hooks.Poststop = []rspec.Hook{}
}

// AddPostStopHook adds a poststop hook into g.Config.Hooks.Poststop.
func (g *Generator) AddPostStopHook(path string, args []string) {
	g.initConfig()
	hook := rspec.Hook{Path: path, Args: args}
	g.Config.Hooks.Poststop = append(g.Config.Hooks.Poststop, hook)
}

// ClearPostStartHooks clear g.Config.Hooks.Poststart.
func (g *Generator) ClearPostStartHooks() {
	if g.Config == nil {
		return
	}
	g.Config.Hooks.Poststart = []rspec.Hook{}
}

// AddPostStartHook adds a poststart hook into g.Config.Hooks.Poststart.
func (g *Generator) AddPostStartHook(path string, args []string) {
	g.initConfig()
	hook := rspec.Hook{Path: path, Args: args}
	g.Config.Hooks.Poststart = append(g.Config.Hooks.Poststart, hook)
}

// AddTmpfsMount adds a tmpfs mount into g.Config.Mounts.
func (g *Generator) AddTmpfsMount(dest string, options []string) {
	mnt := rspec.Mount{
		Destination: dest,
		Type:        "tmpfs",
		Source:      "tmpfs",
		Options:     options,
	}

	g.initConfig()
	g.Config.Mounts = append(g.Config.Mounts, mnt)
}

// AddCgroupsMount adds a cgroup mount into g.Config.Mounts.
func (g *Generator) AddCgroupsMount(mountCgroupOption string) error {
	switch mountCgroupOption {
	case "ro":
	case "rw":
		break
	case "no":
		return nil
	default:
		return fmt.Errorf("--mount-cgroups should be one of (ro,rw,no)")
	}

	mnt := rspec.Mount{
		Destination: "/sys/fs/cgroup",
		Type:        "cgroup",
		Source:      "cgroup",
		Options:     []string{"nosuid", "noexec", "nodev", "relatime", mountCgroupOption},
	}
	g.initConfig()
	g.Config.Mounts = append(g.Config.Mounts, mnt)

	return nil
}

// AddBindMount adds a bind mount into g.Config.Mounts.
func (g *Generator) AddBindMount(source, dest, options string) {
	if options == "" {
		options = "ro"
	}

	defaultOptions := []string{"bind"}

	mnt := rspec.Mount{
		Destination: dest,
		Type:        "bind",
		Source:      source,
		Options:     append(defaultOptions, options),
	}
	g.initConfig()
	g.Config.Mounts = append(g.Config.Mounts, mnt)
}

// SetupPrivileged sets up the priviledge-related fields inside g.Config.
func (g *Generator) SetupPrivileged(privileged bool) {
	if privileged {
		// Add all capabilities in privileged mode.
		var finalCapList []string
		for _, cap := range capability.List() {
			if g.HostSpecific && cap > lastCap() {
				continue
			}
			finalCapList = append(finalCapList, fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String())))
		}
		g.initConfigLinux()
		g.Config.Process.Capabilities = finalCapList
		g.Config.Process.SelinuxLabel = ""
		g.Config.Process.ApparmorProfile = ""
		g.Config.Linux.Seccomp = nil
	}
}

func lastCap() capability.Cap {
	last := capability.CAP_LAST_CAP
	// hack for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}

	return last
}

func checkCap(c string, hostSpecific bool) error {
	isValid := false
	cp := strings.ToUpper(c)

	for _, cap := range capability.List() {
		if cp == strings.ToUpper(cap.String()) {
			if hostSpecific && cap > lastCap() {
				return fmt.Errorf("CAP_%s is not supported on the current host", cp)
			}
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("Invalid value passed for adding capability")
	}
	return nil
}

// ClearProcessCapabilities clear g.Config.Process.Capabilities.
func (g *Generator) ClearProcessCapabilities() {
	if g.Config == nil {
		return
	}
	g.Config.Process.Capabilities = []string{}
}

// AddProcessCapability adds a process capability into g.Config.Process.Capabilities.
func (g *Generator) AddProcessCapability(c string) error {
	if err := checkCap(c, g.HostSpecific); err != nil {
		return err
	}

	cp := fmt.Sprintf("CAP_%s", strings.ToUpper(c))

	g.initConfig()
	for _, cap := range g.Config.Process.Capabilities {
		if strings.ToUpper(cap) == cp {
			return nil
		}
	}

	g.Config.Process.Capabilities = append(g.Config.Process.Capabilities, cp)
	return nil
}

// DropProcessCapability drops a process capability from g.Config.Process.Capabilities.
func (g *Generator) DropProcessCapability(c string) error {
	if err := checkCap(c, g.HostSpecific); err != nil {
		return err
	}

	cp := fmt.Sprintf("CAP_%s", strings.ToUpper(c))

	g.initConfig()
	for i, cap := range g.Config.Process.Capabilities {
		if strings.ToUpper(cap) == cp {
			g.Config.Process.Capabilities = append(g.Config.Process.Capabilities[:i], g.Config.Process.Capabilities[i+1:]...)
			return nil
		}
	}

	return nil
}

func mapStrToNamespace(ns string, path string) (rspec.Namespace, error) {
	switch ns {
	case "network":
		return rspec.Namespace{Type: rspec.NetworkNamespace, Path: path}, nil
	case "pid":
		return rspec.Namespace{Type: rspec.PIDNamespace, Path: path}, nil
	case "mount":
		return rspec.Namespace{Type: rspec.MountNamespace, Path: path}, nil
	case "ipc":
		return rspec.Namespace{Type: rspec.IPCNamespace, Path: path}, nil
	case "uts":
		return rspec.Namespace{Type: rspec.UTSNamespace, Path: path}, nil
	case "user":
		return rspec.Namespace{Type: rspec.UserNamespace, Path: path}, nil
	case "cgroup":
		return rspec.Namespace{Type: rspec.CgroupNamespace, Path: path}, nil
	default:
		return rspec.Namespace{}, fmt.Errorf("Should not reach here!")
	}
}

// ClearLinuxNamespaces clear g.Config.Linux.Namespaces.
func (g *Generator) ClearLinuxNamespaces() {
	if g.Config == nil || g.Config.Linux == nil {
		return
	}
	g.Config.Linux.Namespaces = []rspec.Namespace{}
}

// AddOrReplaceLinuxNamespace adds or replaces a namespace inside
// g.Config.Linux.Namespaces.
func (g *Generator) AddOrReplaceLinuxNamespace(ns string, path string) error {
	namespace, err := mapStrToNamespace(ns, path)
	if err != nil {
		return err
	}

	g.initConfigLinux()
	for i, ns := range g.Config.Linux.Namespaces {
		if ns.Type == namespace.Type {
			g.Config.Linux.Namespaces[i] = namespace
			return nil
		}
	}
	g.Config.Linux.Namespaces = append(g.Config.Linux.Namespaces, namespace)
	return nil
}

// RemoveLinuxNamespace removes a namespace from g.Config.Linux.Namespaces.
func (g *Generator) RemoveLinuxNamespace(ns string) error {
	namespace, err := mapStrToNamespace(ns, "")
	if err != nil {
		return err
	}

	if g.Config == nil || g.Config.Linux == nil {
		return nil
	}
	for i, ns := range g.Config.Linux.Namespaces {
		if ns.Type == namespace.Type {
			g.Config.Linux.Namespaces = append(g.Config.Linux.Namespaces[:i], g.Config.Linux.Namespaces[i+1:]...)
			return nil
		}
	}
	return nil
}

// strPtr returns the pointer pointing to the string s.
func strPtr(s string) *string { return &s }

// SetSyscallAction adds rules for syscalls with the specified action
func (g *Generator) SetSyscallAction(arguments seccomp.SyscallOpts) error {
	g.initConfigLinuxSeccomp()
	return seccomp.ParseSyscallFlag(arguments, g.Config.Linux.Seccomp)
}

// SetDefaultSeccompAction sets the default action for all syscalls not defined
// and then removes any syscall rules with this action already specified.
func (g *Generator) SetDefaultSeccompAction(action string) error {
	g.initConfigLinuxSeccomp()
	return seccomp.ParseDefaultAction(action, g.Config.Linux.Seccomp)
}

// SetDefaultSeccompActionForce only sets the default action for all syscalls not defined
func (g *Generator) SetDefaultSeccompActionForce(action string) error {
	g.initConfigLinuxSeccomp()
	return seccomp.ParseDefaultActionForce(action, g.Config.Linux.Seccomp)
}

// SetSeccompArchitecture sets the supported seccomp architectures
func (g *Generator) SetSeccompArchitecture(architecture string) error {
	g.initConfigLinuxSeccomp()
	return seccomp.ParseArchitectureFlag(architecture, g.Config.Linux.Seccomp)
}

// RemoveSeccompRule removes rules for any specified syscalls
func (g *Generator) RemoveSeccompRule(arguments string) error {
	g.initConfigLinuxSeccomp()
	return seccomp.RemoveAction(arguments, g.Config.Linux.Seccomp)
}

// RemoveAllSeccompRules removes all syscall rules
func (g *Generator) RemoveAllSeccompRules() error {
	g.initConfigLinuxSeccomp()
	return seccomp.RemoveAllSeccompRules(g.Config.Linux.Seccomp)
}

// AddLinuxMaskedPaths adds masked paths into g.Config.Linux.MaskedPaths.
func (g *Generator) AddLinuxMaskedPaths(path string) {
	g.initConfigLinux()
	g.Config.Linux.MaskedPaths = append(g.Config.Linux.MaskedPaths, path)
}

// AddLinuxReadonlyPaths adds readonly paths into g.Config.Linux.MaskedPaths.
func (g *Generator) AddLinuxReadonlyPaths(path string) {
	g.initConfigLinux()
	g.Config.Linux.ReadonlyPaths = append(g.Config.Linux.ReadonlyPaths, path)
}
