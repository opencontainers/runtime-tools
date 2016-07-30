package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/opencontainers/ocitools/generate"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

var generateFlags = []cli.Flag{
	cli.StringFlag{Name: "output", Usage: "output file (defaults to stdout)"},
	cli.StringFlag{Name: "rootfs", Value: "rootfs", Usage: "path to the rootfs"},
	cli.BoolFlag{Name: "read-only", Usage: "make the container's rootfs read-only"},
	cli.BoolFlag{Name: "privileged", Usage: "enable privileged container settings"},
	cli.BoolFlag{Name: "no-new-privileges", Usage: "set no new privileges bit for the container process"},
	cli.BoolFlag{Name: "tty", Usage: "allocate a new tty for the container process"},
	cli.StringFlag{Name: "hostname", Usage: "hostname value for the container"},
	cli.IntFlag{Name: "uid", Usage: "uid for the process"},
	cli.IntFlag{Name: "gid", Usage: "gid for the process"},
	cli.StringSliceFlag{Name: "groups", Usage: "supplementary groups for the process"},
	cli.StringSliceFlag{Name: "cap-add", Usage: "add Linux capabilities"},
	cli.StringSliceFlag{Name: "cap-drop", Usage: "drop Linux capabilities"},
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
	cli.BoolFlag{Name: "seccomp-only", Usage: "specifies to export just a seccomp configuration file"},
	cli.StringFlag{Name: "seccomp-arch", Usage: "specifies additional architectures permitted to be used for system calls"},
	cli.StringFlag{Name: "seccomp-default", Usage: "specifies default action to be used for system calls"},
	cli.StringFlag{Name: "seccomp-allow", Usage: "specifies syscalls to respond with allow"},
	cli.StringFlag{Name: "seccomp-trap", Usage: "specifies syscalls to respond with trap"},
	cli.StringFlag{Name: "seccomp-errno", Usage: "specifies syscalls to respond with errno"},
	cli.StringFlag{Name: "seccomp-trace", Usage: "specifies syscalls to respond with trace"},
	cli.StringFlag{Name: "seccomp-kill", Usage: "specifies syscalls to respond with kill"},
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

		err := setupSpec(&specgen, context)
		if err != nil {
			return err
		}
		var exportOpts generate.ExportOptions
		exportOpts.Seccomp = context.Bool("seccomp-only")

		if context.IsSet("output") {
			err = specgen.SaveToFile(context.String("output"), exportOpts)
		} else {
			err = specgen.Save(os.Stdout, exportOpts)
		}

		if err != nil {
			return err
		}
		return nil
	},
}

func setupSpec(g *generate.Generator, context *cli.Context) error {
	if context.GlobalBool("host-specific") {
		g.HostSpecific = true
	}

	spec := g.Spec()

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

	g, err := addSeccomp(g, context)
	if err != nil {
		return err
	}

	return nil
}

func addSeccomp(g generate.Generator, context *cli.Context) (generate.Generator, error) {
	seccompDefault := context.String("seccomp-default")
	seccompArch := context.String("seccomp-arch")
	seccompKill := context.String("seccomp-kill")
	seccompTrace := context.String("seccomp-trace")
	seccompErrno := context.String("seccomp-errno")
	seccompTrap := context.String("seccomp-trap")
	seccompAllow := context.String("seccomp-allow")

	// Set the DefaultAction of seccomp
	if seccompDefault == "" {
		seccompDefault = "errno"
	}

	err := g.SetDefaultSeccompAction(seccompDefault)
	if err != nil {
		return g, err
	}

	// Add the additional architectures permitted to be used for system calls
	if seccompArch == "" {
		seccompArch = "amd64,x86,x32" // Default Architectures
	}

	architectureArgs := strings.Split(seccompArch, ",")
	err = g.SetSeccompArchitectures(architectureArgs)
	if err != nil {
		return g, err
	}

	if seccompKill != "" {
		killArgs := strings.Split(seccompKill, ",")
		err = g.SetSyscallActions("kill", killArgs)
		if err != nil {
			return g, err
		}
	}

	if seccompTrace != "" {
		traceArgs := strings.Split(seccompTrace, ",")
		err = g.SetSyscallActions("trace", traceArgs)
		if err != nil {
			return g, err
		}
	}

	if seccompErrno != "" {
		errnoArgs := strings.Split(seccompErrno, ",")
		err = g.SetSyscallActions("errno", errnoArgs)
		if err != nil {
			return g, err
		}
	}

	if seccompTrap != "" {
		trapArgs := strings.Split(seccompTrap, ",")
		err = g.SetSyscallActions("trap", trapArgs)
		if err != nil {
			return g, err
		}
	}

	if seccompAllow != "" {
		allowArgs := strings.Split(seccompAllow, ",")
		err = g.SetSyscallActions("allow", allowArgs)
		if err != nil {
			return g, err
		}
	}

	return g, nil
}

func checkNs(nsMaps map[string]string, nsName string) bool {
	if _, ok := nsMaps[nsName]; !ok {
		return false
	}
	return true
}

func setupLinuxNamespaces(g *generate.Generator, needsNewUser bool, nsMaps map[string]string) {
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
