package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/opencontainers/specs"
	"github.com/syndtr/gocapability/capability"
)

var generateFlags = []cli.Flag{
	cli.StringFlag{Name: "rootfs", Value: "rootfs", Usage: "path to the rootfs"},
	cli.BoolFlag{Name: "read-only", Usage: "make the container's rootfs read-only"},
	cli.BoolFlag{Name: "privileged", Usage: "enabled privileged container settings"},
	cli.StringFlag{Name: "hostname", Value: "acme", Usage: "hostname value for the container"},
	cli.IntFlag{Name: "uid", Usage: "uid for the process"},
	cli.IntFlag{Name: "gid", Usage: "gid for the process"},
	cli.StringSliceFlag{Name: "groups", Usage: "supplementary groups for the process"},
	cli.StringSliceFlag{Name: "cap-add", Usage: "add capabilities"},
	cli.StringSliceFlag{Name: "cap-drop", Usage: "drop capabilities"},
	cli.StringFlag{Name: "network", Usage: "network namespace"},
	cli.StringFlag{Name: "mount", Usage: "mount namespace"},
	cli.StringFlag{Name: "pid", Usage: "pid namespace"},
	cli.StringFlag{Name: "ipc", Usage: "ipc namespace"},
	cli.StringFlag{Name: "uts", Usage: "uts namespace"},
	cli.StringFlag{Name: "selinux-label", Usage: "process selinux label"},
	cli.StringSliceFlag{Name: "tmpfs", Usage: "mount tmpfs"},
	cli.StringSliceFlag{Name: "args", Usage: "command to run in the container"},
	cli.StringSliceFlag{Name: "env", Usage: "add environment variable"},
	cli.StringFlag{Name: "mount-cgroups", Value: "ro", Usage: "mount cgroups (rw,ro,no)"},
	cli.StringSliceFlag{Name: "bind", Usage: "bind mount directories src:dest:(rw,ro)"},
	cli.StringSliceFlag{Name: "prestart", Usage: "path to prestart hooks"},
	cli.StringSliceFlag{Name: "poststart", Usage: "path to poststart hooks"},
	cli.StringSliceFlag{Name: "poststop", Usage: "path to poststop hooks"},
	cli.StringFlag{Name: "root-propagation", Usage: "mount propagation for root"},
	cli.StringFlag{Name: "os", Value: runtime.GOOS, Usage: "operating system the container is created for"},
	cli.StringFlag{Name: "arch", Value: runtime.GOARCH, Usage: "architecture the container is created for"},
	cli.StringFlag{Name: "cwd", Value: "/", Usage: "current working directory for the process"},
	cli.StringSliceFlag{Name: "uidmappings", Usage: "add UIDMappings e.g HostID:ContainerID:Size"},
	cli.StringSliceFlag{Name: "gidmappings", Usage: "add GIDMappings e.g HostID:ContainerID:Size"},
	cli.StringFlag{Name: "apparmor", Usage: "specifies the the apparmor profile for the container"},
	cli.StringFlag{Name: "seccomp-default", Usage: "specifies the the defaultaction of Seccomp syscall restrictions"},
	cli.StringSliceFlag{Name: "seccomp-arch", Usage: "specifies Additional architectures permitted to be used for system calls"},
	cli.StringSliceFlag{Name: "seccomp-syscalls", Usage: "specifies Additional architectures permitted to be used for system calls, e.g Name:Action:Arg1_index/Arg1_value/Arg1_valuetwo/Arg1_op, Arg2_index/Arg2_value/Arg2_valuetwo/Arg2_op "},
}

var (
	defaultCaps = []string{
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
		"CAP_AUDIT_READ",
	}
)

var generateCommand = cli.Command{
	Name:  "generate",
	Usage: "generate a OCI spec file",
	Flags: generateFlags,
	Action: func(context *cli.Context) {
		spec, rspec := getDefaultTemplate()
		err := modify(&spec, &rspec, context)
		if err != nil {
			logrus.Fatal(err)
		}
		cName := "config.json"
		rName := "runtime.json"
		data, err := json.MarshalIndent(&spec, "", "\t")
		if err != nil {
			logrus.Fatal(err)
		}
		if err := ioutil.WriteFile(cName, data, 0666); err != nil {
			logrus.Fatal(err)
		}
		rdata, err := json.MarshalIndent(&rspec, "", "\t")
		if err != nil {
			logrus.Fatal(err)
		}
		if err := ioutil.WriteFile(rName, rdata, 0666); err != nil {
			logrus.Fatal(err)
		}
	},
}

func modify(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	spec.Root.Path = context.String("rootfs")
	spec.Root.Readonly = context.Bool("read-only")
	spec.Hostname = context.String("hostname")
	spec.Process.User.UID = uint32(context.Int("uid"))
	spec.Process.User.GID = uint32(context.Int("gid"))
	rspec.Linux.SelinuxProcessLabel = context.String("selinux-label")
	spec.Platform.OS = context.String("os")
	spec.Platform.Arch = context.String("arch")
	spec.Process.Cwd = context.String("cwd")
	rspec.Linux.ApparmorProfile = context.String("apparmor")

	for i, a := range context.StringSlice("args") {
		if a != "" {
			if i == 0 {
				// Replace "sh" from getDefaultTemplate()
				spec.Process.Args[0] = a
			} else {
				spec.Process.Args = append(spec.Process.Args, a)
			}
		}
	}

	for _, e := range context.StringSlice("env") {
		spec.Process.Env = append(spec.Process.Env, e)
	}

	groups := context.StringSlice("groups")
	if groups != nil {
		for _, g := range groups {
			groupID, err := strconv.Atoi(g)
			if err != nil {
				return err
			}
			spec.Process.User.AdditionalGids = append(spec.Process.User.AdditionalGids, uint32(groupID))
		}
	}

	if err := setupCapabilities(spec, rspec, context); err != nil {
		return err
	}
	setupNamespaces(spec, rspec, context)
	if err := addTmpfsMounts(spec, rspec, context); err != nil {
		return err
	}
	if err := mountCgroups(spec, rspec, context); err != nil {
		return err
	}
	if err := addBindMounts(spec, rspec, context); err != nil {
		return err
	}
	if err := addHooks(spec, rspec, context); err != nil {
		return err
	}
	if err := addRootPropagation(spec, rspec, context); err != nil {
		return err
	}
	if err := addIDMappings(spec, rspec, context); err != nil {
		return err
	}
	if err := addSeccomp(spec, rspec, context); err != nil {
		return err
	}

	return nil
}

func addSeccompDefault(rspec *specs.LinuxRuntimeSpec, sdefault string) error {
	switch sdefault {
	case "":
	case "SCMP_ACT_KILL":
	case "SCMP_ACT_TRAP":
	case "SCMP_ACT_ERRNO":
	case "SCMP_ACT_TRACE":
	case "SCMP_ACT_ALLOW":
	default:
		return fmt.Errorf("seccomp-default must be empty or one of " +
			"SCMP_ACT_KILL|SCMP_ACT_TRAP|SCMP_ACT_ERRNO|SCMP_ACT_TRACE|" +
			"SCMP_ACT_ALLOW")
	}
	rspec.Linux.Seccomp.DefaultAction = specs.Action(sdefault)
	return nil
}

func addSeccompArch(rspec *specs.LinuxRuntimeSpec, sArch []string) error {
	for _, archs := range sArch {
		switch archs {
		case "":
		case "SCMP_ARCH_X86":
		case "SCMP_ARCH_X86_64":
		case "SCMP_ARCH_X32":
		case "SCMP_ARCH_ARM":
		case "SCMP_ARCH_AARCH64":
		case "SCMP_ARCH_MIPS":
		case "SCMP_ARCH_MIPS64":
		case "SCMP_ARCH_MIPS64N32":
		case "SCMP_ARCH_MIPSEL":
		case "SCMP_ARCH_MIPSEL64":
		case "SCMP_ARCH_MIPSEL64N32":
		default:
			return fmt.Errorf("seccomp-arch must be empty or one of " +
				"SCMP_ARCH_X86|SCMP_ARCH_X86_64|SCMP_ARCH_X32|SCMP_ARCH_ARM|" +
				"SCMP_ARCH_AARCH64SCMP_ARCH_MIPS|SCMP_ARCH_MIPS64|" +
				"SCMP_ARCH_MIPS64N32|SCMP_ARCH_MIPSEL|SCMP_ARCH_MIPSEL64|" +
				"SCMP_ARCH_MIPSEL64N32")
		}
		rspec.Linux.Seccomp.Architectures = append(rspec.Linux.Seccomp.Architectures, specs.Arch(archs))
	}

	return nil
}

func addSeccompSyscall(rspec *specs.LinuxRuntimeSpec, sSyscall []string) error {
	for _, syscalls := range sSyscall {
		syscall := strings.Split(syscalls, ":")
		if len(syscall) == 3 {
			name := syscall[0]
			switch syscall[1] {
			case "":
			case "SCMP_ACT_KILL":
			case "SCMP_ACT_TRAP":
			case "SCMP_ACT_ERRNO":
			case "SCMP_ACT_TRACE":
			case "SCMP_ACT_ALLOW":
			default:
				return fmt.Errorf("seccomp-syscall action must be empty or " +
					"one of SCMP_ACT_KILL|SCMP_ACT_TRAP|SCMP_ACT_ERRNO|" +
					"SCMP_ACT_TRACE|SCMP_ACT_ALLOW")
			}
			action := specs.Action(syscall[1])

			var Args []*specs.Arg
			if strings.EqualFold(syscall[2], "") {
				Args = nil
			} else {

				argsslice := strings.Split(syscall[2], ",")
				for _, argsstru := range argsslice {
					args := strings.Split(argsstru, "/")
					if len(args) == 4 {
						index, err := strconv.Atoi(args[0])
						value, err := strconv.Atoi(args[1])
						value2, err := strconv.Atoi(args[2])
						if err != nil {
							return err
						}
						switch args[3] {
						case "":
						case "SCMP_CMP_NE":
						case "SCMP_CMP_LT":
						case "SCMP_CMP_LE":
						case "SCMP_CMP_EQ":
						case "SCMP_CMP_GE":
						case "SCMP_CMP_GT":
						case "SCMP_CMP_MASKED_EQ":
						default:
							return fmt.Errorf("seccomp-syscall args must be " +
								"empty or one of SCMP_CMP_NE|SCMP_CMP_LT|" +
								"SCMP_CMP_LE|SCMP_CMP_EQ|SCMP_CMP_GE|" +
								"SCMP_CMP_GT|SCMP_CMP_MASKED_EQ")
						}
						op := specs.Operator(args[3])
						Arg := specs.Arg{
							Index:    uint(index),
							Value:    uint64(value),
							ValueTwo: uint64(value2),
							Op:       op,
						}
						Args = append(Args, &Arg)
					} else {
						return fmt.Errorf("seccomp-sysctl args error: %s", argsstru)
					}
				}
			}

			syscallstruct := specs.Syscall{
				Name:   name,
				Action: action,
				Args:   Args,
			}
			rspec.Linux.Seccomp.Syscalls = append(rspec.Linux.Seccomp.Syscalls, &syscallstruct)
		} else {
			return fmt.Errorf("seccomp sysctl must consist of 3 parameters")
		}
	}

	return nil
}

func addSeccomp(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {

	sd := context.String("seccomp-default")
	sa := context.StringSlice("seccomp-arch")
	ss := context.StringSlice("seccomp-syscalls")

	// Set the DefaultAction of seccomp
	err := addSeccompDefault(rspec, sd)
	if err != nil {
		return err
	}

	// Add the additional architectures permitted to be used for system calls
	err = addSeccompArch(rspec, sa)
	if err != nil {
		return err
	}

	// Set syscall restrict in Seccomp
	// The format of input syscall string is Name:Action:Args[1],Args[2],...,Args[n]
	// The format of Args string is Index/Value/ValueTwo/Operator,and is parsed by function parseArgs()
	err = addSeccompSyscall(rspec, ss)
	if err != nil {
		return err
	}

	return nil
}

func parseArgs(args2parse string) ([]*specs.Arg, error) {
	var Args []*specs.Arg
	argstrslice := strings.Split(args2parse, ",")
	for _, argstr := range argstrslice {
		args := strings.Split(argstr, "/")
		if len(args) == 4 {
			index, err := strconv.Atoi(args[0])
			value, err := strconv.Atoi(args[1])
			value2, err := strconv.Atoi(args[2])
			if err != nil {
				return nil, err
			}
			switch args[3] {
			case "":
			case "SCMP_CMP_NE":
			case "SCMP_CMP_LT":
			case "SCMP_CMP_LE":
			case "SCMP_CMP_EQ":
			case "SCMP_CMP_GE":
			case "SCMP_CMP_GT":
			case "SCMP_CMP_MASKED_EQ":
			default:
				return nil, fmt.Errorf("seccomp-sysctl args must be empty or one of SCMP_CMP_NE|SCMP_CMP_LT|SCMP_CMP_LE|SCMP_CMP_EQ|SCMP_CMP_GE|SCMP_CMP_GT|SCMP_CMP_MASKED_EQ")
			}
			op := specs.Operator(args[3])
			Arg := specs.Arg{
				Index:    uint(index),
				Value:    uint64(value),
				ValueTwo: uint64(value2),
				Op:       op,
			}
			Args = append(Args, &Arg)
		} else {
			return nil, fmt.Errorf("seccomp-sysctl args error: %s", argstr)
		}
	}
	return Args, nil
}

func addIDMappings(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	for _, uidms := range context.StringSlice("uidmappings") {
		idm := strings.Split(uidms, ":")
		if len(idm) == 3 {
			hid, err := strconv.Atoi(idm[0])
			cid, err := strconv.Atoi(idm[1])
			size, err := strconv.Atoi(idm[2])
			if err != nil {
				return err
			}
			uidmapping := specs.IDMapping{
				HostID:      uint32(hid),
				ContainerID: uint32(cid),
				Size:        uint32(size),
			}
			rspec.Linux.UIDMappings = append(rspec.Linux.UIDMappings, uidmapping)
		} else {
			return fmt.Errorf("uidmappings error: %s", uidms)
		}
	}

	for _, gidms := range context.StringSlice("gidmappings") {
		idm := strings.Split(gidms, ":")
		if len(idm) == 3 {
			hid, err := strconv.Atoi(idm[0])
			cid, err := strconv.Atoi(idm[1])
			size, err := strconv.Atoi(idm[2])
			if err != nil {
				return err
			}
			gidmapping := specs.IDMapping{
				HostID:      uint32(hid),
				ContainerID: uint32(cid),
				Size:        uint32(size),
			}
			rspec.Linux.GIDMappings = append(rspec.Linux.GIDMappings, gidmapping)
		} else {
			return fmt.Errorf("gidmappings error: %s", gidms)
		}
	}

	if len(context.StringSlice("uidmappings")) > 0 || len(context.StringSlice("gidmappings")) > 0 {
		rspec.Linux.Namespaces = append(rspec.Linux.Namespaces, specs.Namespace{Type: "user"})
	}

	return nil
}

func addRootPropagation(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	rp := context.String("root-propagation")
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
	rspec.Linux.RootfsPropagation = rp
	return nil
}

func addHooks(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	for _, pre := range context.StringSlice("prestart") {
		parts := strings.Split(pre, ":")
		args := []string{}
		path := parts[0]
		if len(parts) > 1 {
			args = parts[1:]
		}
		rspec.Hooks.Prestart = append(rspec.Hooks.Prestart, specs.Hook{Path: path, Args: args})
	}
	for _, post := range context.StringSlice("poststop") {
		parts := strings.Split(post, ":")
		args := []string{}
		path := parts[0]
		if len(parts) > 1 {
			args = parts[1:]
		}
		rspec.Hooks.Poststop = append(rspec.Hooks.Poststop, specs.Hook{Path: path, Args: args})
	}
	for _, poststart := range context.StringSlice("poststart") {
		parts := strings.Split(poststart, ":")
		args := []string{}
		path := parts[0]
		if len(parts) > 1 {
			args = parts[1:]
		}
		rspec.Hooks.Poststart = append(rspec.Hooks.Poststart, specs.Hook{Path: path, Args: args})
	}
	return nil
}
func addTmpfsMounts(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	for _, dest := range context.StringSlice("tmpfs") {
		name := filepath.Base(dest)
		mntName := fmt.Sprintf("%stmpfs", name)
		mnt := specs.MountPoint{Name: mntName, Path: dest}
		spec.Mounts = append(spec.Mounts, mnt)
		rmnt := specs.Mount{
			Type:    "tmpfs",
			Source:  "tmpfs",
			Options: []string{"nosuid", "nodev", "mode=755"},
		}
		rspec.Mounts[mntName] = rmnt
	}
	return nil
}

func mountCgroups(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	mountCgroupOption := context.String("mount-cgroups")
	switch mountCgroupOption {
	case "ro":
	case "rw":
	case "no":
		return nil
	default:
		return fmt.Errorf("--mount-cgroups should be one of (ro,rw,no)")
	}

	spec.Mounts = append(spec.Mounts, specs.MountPoint{Name: "cgroup", Path: "/sys/fs/cgroup"})
	rspec.Mounts["cgroup"] = specs.Mount{
		Type:    "cgroup",
		Source:  "cgroup",
		Options: []string{"nosuid", "noexec", "nodev", "relatime", mountCgroupOption},
	}

	return nil
}

func addBindMounts(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	for _, b := range context.StringSlice("bind") {
		var source, dest string
		options := "ro"
		bparts := strings.SplitN(b, ":", 3)
		switch len(bparts) {
		case 2:
			source, dest = bparts[0], bparts[1]
		case 3:
			source, dest, options = bparts[0], bparts[1], bparts[2]
		default:
			return fmt.Errorf("--bind should have format src:dest:[options]")
		}
		name := filepath.Base(source)
		mntName := fmt.Sprintf("%sbind", name)
		spec.Mounts = append(spec.Mounts, specs.MountPoint{Name: mntName, Path: dest})
		defaultOptions := []string{"bind"}
		rspec.Mounts[mntName] = specs.Mount{
			Type:    "bind",
			Source:  source,
			Options: append(defaultOptions, options),
		}
	}
	return nil
}

func setupCapabilities(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) error {
	var finalCapList []string

	// Add all capabilities in privileged mode.
	privileged := context.Bool("privileged")
	if privileged {
		for _, cap := range capability.List() {
			finalCapList = append(finalCapList, fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String())))
		}
		spec.Linux.Capabilities = finalCapList
		return nil
	}

	capMappings := make(map[string]bool)
	for _, cap := range capability.List() {
		key := strings.ToUpper(cap.String())
		capMappings[key] = true
	}

	addedCapsMap := make(map[string]bool)
	for _, cap := range defaultCaps {
		addedCapsMap[cap] = true
	}

	addCapList := make([]string, len(defaultCaps))
	copy(addCapList, defaultCaps)
	addCaps := context.StringSlice("cap-add")
	for _, c := range addCaps {
		if !capMappings[c] {
			return fmt.Errorf("Invalid value passed for adding capability")
		}
		cp := fmt.Sprintf("CAP_%s", c)
		if !addedCapsMap[cp] {
			addCapList = append(addCapList, cp)
			addedCapsMap[cp] = true
		}
	}
	dropCaps := context.StringSlice("cap-drop")
	dropCapsMap := make(map[string]bool)
	for _, c := range dropCaps {
		if !capMappings[c] {
			return fmt.Errorf("Invalid value passed for dropping capability")
		}
		cp := fmt.Sprintf("CAP_%s", c)
		dropCapsMap[cp] = true
	}

	for _, c := range addCapList {
		if !dropCapsMap[c] {
			finalCapList = append(finalCapList, c)
		}
	}
	spec.Linux.Capabilities = finalCapList
	return nil
}

func mapStrToNamespace(ns string, path string) specs.Namespace {
	switch ns {
	case "network":
		return specs.Namespace{Type: specs.NetworkNamespace, Path: path}
	case "pid":
		return specs.Namespace{Type: specs.PIDNamespace, Path: path}
	case "mount":
		return specs.Namespace{Type: specs.MountNamespace, Path: path}
	case "ipc":
		return specs.Namespace{Type: specs.IPCNamespace, Path: path}
	case "uts":
		return specs.Namespace{Type: specs.UTSNamespace, Path: path}
	case "user":
		return specs.Namespace{Type: specs.UserNamespace, Path: path}
	default:
		logrus.Fatalf("Should not reach here!")
	}
	return specs.Namespace{}
}

func setupNamespaces(spec *specs.LinuxSpec, rspec *specs.LinuxRuntimeSpec, context *cli.Context) {
	namespaces := []string{"network", "pid", "mount", "ipc", "uts"}
	var linuxNs []specs.Namespace
	for _, nsName := range namespaces {
		nsPath := context.String(nsName)
		if nsPath == "host" {
			continue
		}
		ns := mapStrToNamespace(nsName, nsPath)
		linuxNs = append(linuxNs, ns)
	}
	rspec.Linux.Namespaces = linuxNs
}

func getDefaultTemplate() (specs.LinuxSpec, specs.LinuxRuntimeSpec) {
	spec := specs.LinuxSpec{
		Spec: specs.Spec{
			Version: specs.Version,
			Platform: specs.Platform{
				OS:   runtime.GOOS,
				Arch: runtime.GOARCH,
			},
			Root: specs.Root{
				Path:     "",
				Readonly: false,
			},
			Process: specs.Process{
				Terminal: true,
				User:     specs.User{},
				Args: []string{
					"sh",
				},
				Env: []string{
					"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
					"TERM=xterm",
				},
				Cwd: "/",
			},
			Hostname: "shell",
			Mounts: []specs.MountPoint{
				{
					Name: "proc",
					Path: "/proc",
				},
				{
					Name: "dev",
					Path: "/dev",
				},
				{
					Name: "devpts",
					Path: "/dev/pts",
				},
				{
					Name: "shm",
					Path: "/dev/shm",
				},
				{
					Name: "mqueue",
					Path: "/dev/mqueue",
				},
				{
					Name: "sysfs",
					Path: "/sys",
				},
			},
		},
		Linux: specs.Linux{
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
		},
	}
	rspec := specs.LinuxRuntimeSpec{
		RuntimeSpec: specs.RuntimeSpec{
			Mounts: map[string]specs.Mount{
				"proc": {
					Type:    "proc",
					Source:  "proc",
					Options: nil,
				},
				"dev": {
					Type:    "tmpfs",
					Source:  "tmpfs",
					Options: []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
				},
				"devpts": {
					Type:    "devpts",
					Source:  "devpts",
					Options: []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
				},
				"shm": {
					Type:    "tmpfs",
					Source:  "shm",
					Options: []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
				},
				"mqueue": {
					Type:    "mqueue",
					Source:  "mqueue",
					Options: []string{"nosuid", "noexec", "nodev"},
				},
				"sysfs": {
					Type:    "sysfs",
					Source:  "sysfs",
					Options: []string{"nosuid", "noexec", "nodev"},
				},
			},
		},
		Linux: specs.LinuxRuntime{
			Namespaces: []specs.Namespace{
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
			Rlimits: []specs.Rlimit{
				{
					Type: "RLIMIT_NOFILE",
					Hard: uint64(1024),
					Soft: uint64(1024),
				},
			},
			Devices: []specs.Device{
				{
					Type:        'c',
					Path:        "/dev/null",
					Major:       1,
					Minor:       3,
					Permissions: "rwm",
					FileMode:    0666,
					UID:         0,
					GID:         0,
				},
				{
					Type:        'c',
					Path:        "/dev/random",
					Major:       1,
					Minor:       8,
					Permissions: "rwm",
					FileMode:    0666,
					UID:         0,
					GID:         0,
				},
				{
					Type:        'c',
					Path:        "/dev/full",
					Major:       1,
					Minor:       7,
					Permissions: "rwm",
					FileMode:    0666,
					UID:         0,
					GID:         0,
				},
				{
					Type:        'c',
					Path:        "/dev/tty",
					Major:       5,
					Minor:       0,
					Permissions: "rwm",
					FileMode:    0666,
					UID:         0,
					GID:         0,
				},
				{
					Type:        'c',
					Path:        "/dev/zero",
					Major:       1,
					Minor:       5,
					Permissions: "rwm",
					FileMode:    0666,
					UID:         0,
					GID:         0,
				},
				{
					Type:        'c',
					Path:        "/dev/urandom",
					Major:       1,
					Minor:       9,
					Permissions: "rwm",
					FileMode:    0666,
					UID:         0,
					GID:         0,
				},
			},
			Resources: &specs.Resources{
				Memory: specs.Memory{
					Swappiness: -1,
				},
			},
		},
	}
	return spec, rspec
}
