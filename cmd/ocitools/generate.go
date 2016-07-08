package main

import (
	"os"
	"runtime"

	"github.com/opencontainers/ocitools/generate"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

var generateFlags = []cli.Flag{
	cli.StringFlag{Name: "output", Value: "output", Usage: "output file (defaults to stdout)"},
	cli.StringFlag{Name: "rootfs", Value: "rootfs", Usage: "path to the rootfs"},
	cli.BoolFlag{Name: "read-only", Usage: "make the container's rootfs read-only"},
	cli.BoolFlag{Name: "privileged", Usage: "enabled privileged container settings"},
	cli.BoolFlag{Name: "no-new-privileges", Usage: "set no new privileges bit for the container process"},
	cli.BoolFlag{Name: "tty", Usage: "allocate a new tty for the container process"},
	cli.StringFlag{Name: "hostname", Usage: "hostname value for the container"},
	cli.IntFlag{Name: "uid", Usage: "uid for the process"},
	cli.IntFlag{Name: "gid", Usage: "gid for the process"},
	cli.StringSliceFlag{Name: "groups", Usage: "supplementary groups for the process"},
	cli.StringSliceFlag{Name: "cap-add", Usage: "add capabilities"},
	cli.StringSliceFlag{Name: "cap-drop", Usage: "drop capabilities"},
	cli.StringFlag{Name: "cgroup", Usage: "cgroup namespace"},
	cli.StringFlag{Name: "network", Usage: "network namespace"},
	cli.StringFlag{Name: "mount", Usage: "mount namespace"},
	cli.StringFlag{Name: "pid", Usage: "pid namespace"},
	cli.StringFlag{Name: "ipc", Usage: "ipc namespace"},
	cli.StringFlag{Name: "user", Usage: "user namespace"},
	cli.StringFlag{Name: "uts", Usage: "uts namespace"},
	cli.StringFlag{Name: "selinux-label", Usage: "process selinux label"},
	cli.StringFlag{Name: "mount-label", Usage: "selinux mount context label"},
	cli.StringSliceFlag{Name: "tmpfs", Usage: "mount tmpfs"},
	cli.StringSliceFlag{Name: "args", Usage: "command to run in the container"},
	cli.StringSliceFlag{Name: "env", Usage: "add environment variable"},
	cli.StringFlag{Name: "cgroups-path", Usage: "specify the path to the cgroups"},
	cli.StringFlag{Name: "mount-cgroups", Value: "no", Usage: "mount cgroups (rw,ro,no)"},
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
	cli.StringSliceFlag{Name: "sysctl", Usage: "add sysctl settings e.g net.ipv4.forward=1"},
	cli.StringFlag{Name: "apparmor", Usage: "specifies the the apparmor profile for the container"},
	cli.StringFlag{Name: "seccomp-default", Usage: "specifies the the defaultaction of Seccomp syscall restrictions"},
	cli.StringSliceFlag{Name: "seccomp-arch", Usage: "specifies Additional architectures permitted to be used for system calls"},
	cli.StringSliceFlag{Name: "seccomp-syscalls", Usage: "specifies Additional architectures permitted to be used for system calls, e.g Name:Action:Arg1_index/Arg1_value/Arg1_valuetwo/Arg1_op, Arg2_index/Arg2_value/Arg2_valuetwo/Arg2_op "},
	cli.StringSliceFlag{Name: "seccomp-allow", Usage: "specifies syscalls to be added to allowed"},
	cli.StringSliceFlag{Name: "seccomp-errno", Usage: "specifies syscalls to be added to list that returns an error"},
	cli.StringFlag{Name: "template", Usage: "base template to use for creating the configuration"},
	cli.StringSliceFlag{Name: "label", Usage: "add annotations to the configuration e.g. key=value"},
}

var generateCommand = cli.Command{
	Name:   "generate",
	Usage:  "generate an OCI spec file",
	Flags:  generateFlags,
	Before: before,
	Action: func(context *cli.Context) error {
		// Start from the default template.
		specgen := generate.New()

		var template string
		if context.IsSet("template") {
			template = context.String("template")
		}
		if template != "" {
			var err error
			specgen, err = generate.NewFromFile(template)
			if err != nil {
				return err
			}
		}

		err := setupSpec(specgen, context)
		if err != nil {
			return err
		}

		if context.IsSet("output") {
			output := context.String("output")
			err = specgen.SaveToFile(output)
		} else {
			err = specgen.Save(os.Stdout)
		}
		if err != nil {
			return err
		}
		return nil
	},
}

func setupSpec(g generate.Generator, context *cli.Context) error {
	spec := g.GetSpec()

	if len(spec.Version) == 0 {
		g.SetVersion(rspec.Version)
	}

	if context.IsSet("hostname") {
		g.SetHostname(context.String("hostname"))
	}

	g.SetPlatformOS(context.String("os"))
	g.SetPlatformArch(context.String("arch"))

	if context.IsSet("label") {
		annotations := context.StringSlice("label")
		for _, s := range annotations {
			if err := g.AddAnnotation(s); err != nil {
				return err
			}
		}
	}

	g.SetRootPath(context.String("rootfs"))

	if context.IsSet("read-only") {
		g.SetRootReadonly(context.Bool("read-only"))
	}

	if context.IsSet("uid") {
		g.SetProcessUID(uint32(context.Int("uid")))
	}

	if context.IsSet("gid") {
		g.SetProcessGID(uint32(context.Int("gid")))
	}

	if context.IsSet("selinux-label") {
		g.SetProcessSelinuxLabel(context.String("selinux-label"))
	}

	g.SetProcessCwd(context.String("cwd"))

	if context.IsSet("apparmor") {
		g.SetProcessApparmorProfile(context.String("apparmor"))
	}

	if context.IsSet("no-new-privileges") {
		g.SetProcessNoNewPrivileges(context.Bool("no-new-privileges"))
	}

	if context.IsSet("tty") {
		g.SetProcessTerminal(context.Bool("tty"))
	}

	if context.IsSet("args") {
		g.SetProcessArgs(context.StringSlice("args"))
	}

	if context.IsSet("env") {
		envs := context.StringSlice("env")
		for _, env := range envs {
			g.AddProcessEnv(env)
		}
	}

	if context.IsSet("groups") {
		groups := context.StringSlice("groups")
		for _, group := range groups {
			g.AddProcessAdditionalGid(group)
		}
	}

	if context.IsSet("cgroups-path") {
		g.SetLinuxCgroupsPath(context.String("cgroups-path"))
	}

	if context.IsSet("mount-label") {
		g.SetLinuxMountLabel(context.String("mount-label"))
	}

	if context.IsSet("sysctl") {
		sysctls := context.StringSlice("sysctl")
		for _, s := range sysctls {
			g.AddLinuxSysctl(s)
		}
	}

	privileged := false
	if context.IsSet("privileged") {
		privileged = context.Bool("privileged")
	}
	g.SetupPrivileged(privileged)

	if context.IsSet("cap-add") {
		addCaps := context.StringSlice("cap-add")
		for _, cap := range addCaps {
			if err := g.AddProcessCapability(cap); err != nil {
				return err
			}
		}
	}

	if context.IsSet("cap-drop") {
		dropCaps := context.StringSlice("cap-drop")
		for _, cap := range dropCaps {
			if err := g.DropProcessCapability(cap); err != nil {
				return err
			}
		}
	}

	needsNewUser := false

	var uidMaps, gidMaps []string

	if context.IsSet("uidmappings") {
		uidMaps = context.StringSlice("uidmappings")
	}

	if context.IsSet("gidmappings") {
		gidMaps = context.StringSlice("gidmappings")
	}

	if len(uidMaps) > 0 || len(gidMaps) > 0 {
		needsNewUser = true
	}

	nsMaps := map[string]string{}
	for _, nsName := range generate.Namespaces {
		if context.IsSet(nsName) {
			nsMaps[nsName] = context.String(nsName)
		}
	}
	setupLinuxNamespaces(g, needsNewUser, nsMaps)

	if context.IsSet("tmpfs") {
		tmpfsSlice := context.StringSlice("tmpfs")
		for _, s := range tmpfsSlice {
			if err := g.AddTmpfsMount(s); err != nil {
				return err
			}
		}
	}

	mountCgroupOption := context.String("mount-cgroups")
	if err := g.AddCgroupsMount(mountCgroupOption); err != nil {
		return err
	}

	if context.IsSet("bind") {
		binds := context.StringSlice("bind")
		for _, bind := range binds {
			if err := g.AddBindMount(bind); err != nil {
				return err
			}
		}
	}

	if context.IsSet("prestart") {
		preStartHooks := context.StringSlice("prestart")
		for _, hook := range preStartHooks {
			if err := g.AddPreStartHook(hook); err != nil {
				return err
			}
		}
	}

	if context.IsSet("poststop") {
		postStopHooks := context.StringSlice("poststop")
		for _, hook := range postStopHooks {
			if err := g.AddPostStopHook(hook); err != nil {
				return err
			}
		}
	}

	if context.IsSet("poststart") {
		postStartHooks := context.StringSlice("poststart")
		for _, hook := range postStartHooks {
			if err := g.AddPostStartHook(hook); err != nil {
				return err
			}
		}
	}

	if context.IsSet("root-propagation") {
		rp := context.String("root-propagation")
		if err := g.SetLinuxRootPropagation(rp); err != nil {
			return err
		}
	}

	for _, uidMap := range uidMaps {
		if err := g.AddLinuxUIDMapping(uidMap); err != nil {
			return err
		}
	}

	for _, gidMap := range gidMaps {
		if err := g.AddLinuxGIDMapping(gidMap); err != nil {
			return err
		}
	}

	var sd string
	var sa, ss []string

	if context.IsSet("seccomp-default") {
		sd = context.String("seccomp-default")
	}

	if context.IsSet("seccomp-arch") {
		sa = context.StringSlice("seccomp-arch")
	}

	if context.IsSet("seccomp-syscalls") {
		ss = context.StringSlice("seccomp-syscalls")
	}

	if sd == "" && len(sa) == 0 && len(ss) == 0 {
		return nil
	}

	// Set the DefaultAction of seccomp
	if context.IsSet("seccomp-default") {
		if err := g.SetLinuxSeccompDefault(sd); err != nil {
			return err
		}
	}

	// Add the additional architectures permitted to be used for system calls
	if context.IsSet("seccomp-arch") {
		for _, arch := range sa {
			if err := g.AddLinuxSeccompArch(arch); err != nil {
				return err
			}
		}
	}

	// Set syscall restrict in Seccomp
	if context.IsSet("seccomp-syscalls") {
		for _, syscall := range ss {
			if err := g.AddLinuxSeccompSyscall(syscall); err != nil {
				return err
			}
		}
	}

	if context.IsSet("seccomp-allow") {
		seccompAllows := context.StringSlice("seccomp-allow")
		for _, s := range seccompAllows {
			g.AddLinuxSeccompSyscallAllow(s)
		}
	}

	if context.IsSet("seccomp-errno") {
		seccompErrnos := context.StringSlice("seccomp-errno")
		for _, s := range seccompErrnos {
			g.AddLinuxSeccompSyscallErrno(s)
		}
	}

	return nil
}

func checkNs(nsMaps map[string]string, nsName string) bool {
	if _, ok := nsMaps[nsName]; !ok {
		return false
	}
	return true
}

func setupLinuxNamespaces(g generate.Generator, needsNewUser bool, nsMaps map[string]string) {
	for _, nsName := range generate.Namespaces {
		if !checkNs(nsMaps, nsName) && !(needsNewUser && nsName == "user") {
			continue
		}
		nsPath := nsMaps[nsName]
		if nsPath == "host" {
			g.RemoveLinuxNamespace(nsName)
			continue
		}
		g.AddOrReplaceLinuxNamespace(nsName, nsPath)
	}
}
