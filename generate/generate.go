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
						Access: StrPtr("rwm"),
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

// AddProcessAdditionalGid adds an additional gid into g.Config.Process.AdditionalGids.
func (g *Generator) AddProcessAdditionalGid(gid uint32) {
	g.InitConfig()
	for _, group := range g.Config.Process.User.AdditionalGids {
		if group == gid {
			return
		}
	}
	g.Config.Process.User.AdditionalGids = append(g.Config.Process.User.AdditionalGids, gid)
}

// SetupPrivileged sets up the priviledge-related fields inside g.Config.
func (g *Generator) SetupPrivileged() {
	// Add all capabilities in privileged mode.
	var finalCapList []string
	for _, cap := range capability.List() {
		if g.HostSpecific && cap > lastCap() {
			continue
		}
		finalCapList = append(finalCapList, fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String())))
	}
	g.InitConfigLinux()
	g.Config.Process.Capabilities = finalCapList
	g.Config.Process.SelinuxLabel = ""
	g.Config.Process.ApparmorProfile = ""
	g.Config.Linux.Seccomp = nil
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

// AddProcessCapability adds a process capability into g.Config.Process.Capabilities.
func (g *Generator) AddProcessCapability(c string) error {
	if err := checkCap(c, g.HostSpecific); err != nil {
		return err
	}

	cp := fmt.Sprintf("CAP_%s", strings.ToUpper(c))

	g.InitConfig()
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

	g.InitConfig()
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

// AddOrReplaceLinuxNamespace adds or replaces a namespace inside
// g.Config.Linux.Namespaces.
func (g *Generator) AddOrReplaceLinuxNamespace(ns string, path string) error {
	namespace, err := mapStrToNamespace(ns, path)
	if err != nil {
		return err
	}

	g.InitConfigLinux()
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

// BoolPtr returns the pointer pointing to the boolean b.
func BoolPtr(b bool) *bool { return &b }

// IntPtr returns the pointer pointing to the int i.
func IntPtr(i int) *int { return &i }

// StrPtr returns the pointer pointing to the string s.
func StrPtr(s string) *string { return &s }

// Uint64Ptr returns the pointer pointing to the uint64 i.
func Uint64Ptr(i uint64) *uint64 { return &i }

// SetSyscallAction adds rules for syscalls with the specified action
func (g *Generator) SetSyscallAction(arguments seccomp.SyscallOpts) error {
	g.InitConfigLinuxSeccomp()
	return seccomp.ParseSyscallFlag(arguments, g.Config.Linux.Seccomp)
}

// SetDefaultSeccompAction sets the default action for all syscalls not defined
// and then removes any syscall rules with this action already specified.
func (g *Generator) SetDefaultSeccompAction(action string) error {
	g.InitConfigLinuxSeccomp()
	return seccomp.ParseDefaultAction(action, g.Config.Linux.Seccomp)
}

// SetDefaultSeccompActionForce only sets the default action for all syscalls not defined
func (g *Generator) SetDefaultSeccompActionForce(action string) error {
	g.InitConfigLinuxSeccomp()
	return seccomp.ParseDefaultActionForce(action, g.Config.Linux.Seccomp)
}

// SetSeccompArchitecture sets the supported seccomp architectures
func (g *Generator) SetSeccompArchitecture(architecture string) error {
	g.InitConfigLinuxSeccomp()
	return seccomp.ParseArchitectureFlag(architecture, g.Config.Linux.Seccomp)
}

// RemoveSeccompRule removes rules for any specified syscalls
func (g *Generator) RemoveSeccompRule(arguments string) error {
	g.InitConfigLinuxSeccomp()
	return seccomp.RemoveAction(arguments, g.Config.Linux.Seccomp)
}

// RemoveAllSeccompRules removes all syscall rules
func (g *Generator) RemoveAllSeccompRules() error {
	g.InitConfigLinuxSeccomp()
	return seccomp.RemoveAllSeccompRules(g.Config.Linux.Seccomp)
}
