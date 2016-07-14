package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Sirupsen/logrus"
	"github.com/blang/semver"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/urfave/cli"
)

type configCheck func(rspec.Spec, string, bool) []string

var bundleValidateFlags = []cli.Flag{
	cli.StringFlag{Name: "path", Value: ".", Usage: "path to a bundle"},
	cli.BoolFlag{Name: "host-specific", Usage: "Check host specific configs."},
}

var (
	defaultRlimits = []string{
		"RLIMIT_CPU",
		"RLIMIT_FSIZE",
		"RLIMIT_DATA",
		"RLIMIT_STACK",
		"RLIMIT_CORE",
		"RLIMIT_RSS",
		"RLIMIT_NPROC",
		"RLIMIT_NOFILE",
		"RLIMIT_MEMLOCK",
		"RLIMIT_AS",
		"RLIMIT_LOCKS",
		"RLIMIT_SIGPENDING",
		"RLIMIT_MSGQUEUE",
		"RLIMIT_NICE",
		"RLIMIT_RTPRIO",
		"RLIMIT_RTTIME",
	}
)

var bundleValidateCommand = cli.Command{
	Name:   "validate",
	Usage:  "validate a OCI bundle",
	Flags:  bundleValidateFlags,
	Before: before,
	Action: func(context *cli.Context) error {
		inputPath := context.String("path")
		if inputPath == "" {
			return fmt.Errorf("Bundle path shouldn't be empty")
		}

		if _, err := os.Stat(inputPath); err != nil {
			return err
		}

		configPath := path.Join(inputPath, "config.json")
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			return err
		}
		if !utf8.Valid(content) {
			return fmt.Errorf("%q is not encoded in UTF-8", configPath)
		}
		var spec rspec.Spec
		if err = json.Unmarshal(content, &spec); err != nil {
			return err
		}

		rootfsPath := path.Join(inputPath, spec.Root.Path)
		if fi, err := os.Stat(rootfsPath); err != nil {
			return fmt.Errorf("Cannot find the root path %q", rootfsPath)
		} else if !fi.IsDir() {
			return fmt.Errorf("The root path %q is not a directory.", rootfsPath)
		}

		hostCheck := context.Bool("host-specific")

		checks := []configCheck{
			checkMandatoryFields,
			checkSemVer,
			checkMounts,
			checkPlatform,
			checkProcess,
			checkLinux,
			checkHooks,
		}

		errMsg := ""
		i := 1
		for _, check := range checks {
			for _, msg := range check(spec, rootfsPath, hostCheck) {
				errMsg = fmt.Sprintf("%s  %d. %s\n", errMsg, i, msg)
				i++
			}
		}

		if errMsg != "" {
			errMsg = fmt.Sprintf("%d Errors detected:\n%s", i-1, errMsg)
			return errors.New(errMsg)

		} else {
			fmt.Println("Bundle validation succeeded.")
			return nil
		}
	},
}

func checkSemVer(spec rspec.Spec, rootfs string, hostCheck bool) (msgs []string) {
	logrus.Debugf("check semver")

	version := spec.Version
	_, err := semver.Parse(version)
	if err != nil {
		msgs = append(msgs, fmt.Sprintf("%q is not valid SemVer: %s", version, err.Error()))
	}
	if version != rspec.Version {
		msgs = append(msgs, fmt.Sprintf("internal error: validate currently only handles version %s, but the supplied configuration targets %s", rspec.Version, version))
	}

	return
}

func checkPlatform(spec rspec.Spec, rootfs string, hostCheck bool) (msgs []string) {
	logrus.Debugf("check platform")

	validCombins := map[string][]string{
		"darwin":    {"386", "amd64", "arm", "arm64"},
		"dragonfly": {"amd64"},
		"freebsd":   {"386", "amd64", "arm"},
		"linux":     {"386", "amd64", "arm", "arm64", "ppc64", "ppc64le", "mips64", "mips64le"},
		"netbsd":    {"386", "amd64", "arm"},
		"openbsd":   {"386", "amd64", "arm"},
		"plan9":     {"386", "amd64"},
		"solaris":   {"amd64"},
		"windows":   {"386", "amd64"}}
	platform := spec.Platform
	for os, archs := range validCombins {
		if os == platform.OS {
			for _, arch := range archs {
				if arch == platform.Arch {
					return nil
				}
			}
			msgs = append(msgs, fmt.Sprintf("Combination of %q and %q is invalid.", platform.OS, platform.Arch))
		}
	}
	msgs = append(msgs, fmt.Sprintf("Operation system %q of the bundle is not supported yet.", platform.OS))

	return
}

func checkHooks(spec rspec.Spec, rootfs string, hostCheck bool) (msgs []string) {
	logrus.Debugf("check hooks")

	msgs = append(msgs, checkEventHooks("pre-start", spec.Hooks.Prestart, hostCheck)...)
	msgs = append(msgs, checkEventHooks("post-start", spec.Hooks.Poststart, hostCheck)...)
	msgs = append(msgs, checkEventHooks("post-stop", spec.Hooks.Poststop, hostCheck)...)

	return
}

func checkEventHooks(hookType string, hooks []rspec.Hook, hostCheck bool) (msgs []string) {
	for _, hook := range hooks {
		if !filepath.IsAbs(hook.Path) {
			msgs = append(msgs, fmt.Sprintf("The %s hook %v: is not absolute path", hookType, hook.Path))
		}

		if hostCheck {
			fi, err := os.Stat(hook.Path)
			if err != nil {
				msgs = append(msgs, fmt.Sprintf("Cannot find %s hook: %v", hookType, hook.Path))
			}
			if fi.Mode()&0111 == 0 {
				msgs = append(msgs, fmt.Sprintf("The %s hook %v: is not executable", hookType, hook.Path))
			}
		}

		for _, env := range hook.Env {
			if !envValid(env) {
				msgs = append(msgs, fmt.Sprintf("Env %q for hook %v is in the invalid form.", env, hook.Path))
			}
		}
	}

	return
}

func checkProcess(spec rspec.Spec, rootfs string, hostCheck bool) (msgs []string) {
	logrus.Debugf("check process")

	process := spec.Process
	if !path.IsAbs(process.Cwd) {
		msgs = append(msgs, fmt.Sprintf("cwd %q is not an absolute path", process.Cwd))
	}

	for _, env := range process.Env {
		if !envValid(env) {
			msgs = append(msgs, fmt.Sprintf("env %q should be in the form of 'key=value'. The left hand side must consist solely of letters, digits, and underscores '_'.", env))
		}
	}

	for index := 0; index < len(process.Capabilities); index++ {
		capability := process.Capabilities[index]
		if !capValid(capability) {
			msgs = append(msgs, fmt.Sprintf("capability %q is not valid, man capabilities(7)", process.Capabilities[index]))
		}
	}

	for index := 0; index < len(process.Rlimits); index++ {
		if !rlimitValid(process.Rlimits[index].Type) {
			msgs = append(msgs, fmt.Sprintf("rlimit type %q is invalid.", process.Rlimits[index].Type))
		}
	}

	if len(process.ApparmorProfile) > 0 {
		profilePath := path.Join(rootfs, "/etc/apparmor.d", process.ApparmorProfile)
		_, err := os.Stat(profilePath)
		if err != nil {
			msgs = append(msgs, err.Error())
		}
	}

	return
}

func supportedMountTypes(OS string, hostCheck bool) (map[string]bool, error) {
	supportedTypes := make(map[string]bool)

	if OS != "linux" && OS != "windows" {
		logrus.Warnf("%v is not supported to check mount type", OS)
		return nil, nil
	} else if OS == "windows" {
		supportedTypes["ntfs"] = true
		return supportedTypes, nil
	}

	if hostCheck {
		f, err := os.Open("/proc/filesystems")
		if err != nil {
			return nil, err
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			if err := s.Err(); err != nil {
				return supportedTypes, err
			}

			text := s.Text()
			parts := strings.Split(text, "\t")
			if len(parts) > 1 {
				supportedTypes[parts[1]] = true
			} else {
				supportedTypes[parts[0]] = true
			}
		}

		supportedTypes["bind"] = true

		return supportedTypes, nil
	} else {
		logrus.Warn("Checking linux mount types without --host-specific is not supported yet")
		return nil, nil
	}
}

func checkMounts(spec rspec.Spec, rootfs string, hostCheck bool) (msgs []string) {
	logrus.Debugf("check mounts")

	supportedTypes, err := supportedMountTypes(spec.Platform.OS, hostCheck)
	if err != nil {
		msgs = append(msgs, err.Error())
		return
	}

	if supportedTypes != nil {
		for _, mount := range spec.Mounts {
			if !supportedTypes[mount.Type] {
				msgs = append(msgs, fmt.Sprintf("Unsupported mount type %q", mount.Type))
			}
		}
	}

	return
}

//Linux only
func checkLinux(spec rspec.Spec, rootfs string, hostCheck bool) (msgs []string) {
	logrus.Debugf("check linux")

	utsExists := false
	ipcExists := false
	mountExists := false
	netExists := false

	if len(spec.Linux.UIDMappings) > 5 {
		msgs = append(msgs, "Only 5 UID mappings are allowed (linux kernel restriction).")
	}
	if len(spec.Linux.GIDMappings) > 5 {
		msgs = append(msgs, "Only 5 GID mappings are allowed (linux kernel restriction).")
	}

	for index := 0; index < len(spec.Linux.Namespaces); index++ {
		if !namespaceValid(spec.Linux.Namespaces[index]) {
			msgs = append(msgs, fmt.Sprintf("namespace %v is invalid.", spec.Linux.Namespaces[index]))
		} else if len(spec.Linux.Namespaces[index].Path) == 0 {
			if spec.Linux.Namespaces[index].Type == rspec.UTSNamespace {
				utsExists = true
			} else if spec.Linux.Namespaces[index].Type == rspec.IPCNamespace {
				ipcExists = true
			} else if spec.Linux.Namespaces[index].Type == rspec.NetworkNamespace {
				netExists = true
			} else if spec.Linux.Namespaces[index].Type == rspec.MountNamespace {
				mountExists = true
			}
		}
	}

	for k := range spec.Linux.Sysctl {
		if strings.HasPrefix(k, "net.") && !netExists {
			msgs = append(msgs, fmt.Sprintf("Sysctl %v requires a new Network namespace to be specified as well", k))
		}
		if strings.HasPrefix(k, "fs.mqueue.") {
			if !mountExists || !ipcExists {
				msgs = append(msgs, fmt.Sprintf("Sysctl %v requires a new IPC namespace and Mount namespace to be specified as well", k))
			}
		}
	}

	if spec.Platform.OS == "linux" && !utsExists && spec.Hostname != "" {
		msgs = append(msgs, fmt.Sprintf("On Linux, hostname requires a new UTS namespace to be specified as well"))
	}

	for index := 0; index < len(spec.Linux.Devices); index++ {
		if !deviceValid(spec.Linux.Devices[index]) {
			msgs = append(msgs, fmt.Sprintf("device %v is invalid.", spec.Linux.Devices[index]))
		}
	}

	if spec.Linux.Seccomp != nil {
		ms := checkSeccomp(*spec.Linux.Seccomp)
		msgs = append(msgs, ms...)
	}

	switch spec.Linux.RootfsPropagation {
	case "":
	case "private":
	case "rprivate":
	case "slave":
	case "rslave":
	case "shared":
	case "rshared":
	default:
		msgs = append(msgs, "rootfsPropagation must be empty or one of \"private|rprivate|slave|rslave|shared|rshared\"")
	}

	return
}

func checkSeccomp(s rspec.Seccomp) (msgs []string) {
	logrus.Debugf("check seccomp")

	if !seccompActionValid(s.DefaultAction) {
		msgs = append(msgs, fmt.Sprintf("seccomp defaultAction %q is invalid.", s.DefaultAction))
	}
	for index := 0; index < len(s.Syscalls); index++ {
		if !syscallValid(s.Syscalls[index]) {
			msgs = append(msgs, fmt.Sprintf("syscall %v is invalid.", s.Syscalls[index]))
		}
	}
	for index := 0; index < len(s.Architectures); index++ {
		switch s.Architectures[index] {
		case rspec.ArchX86:
		case rspec.ArchX86_64:
		case rspec.ArchX32:
		case rspec.ArchARM:
		case rspec.ArchAARCH64:
		case rspec.ArchMIPS:
		case rspec.ArchMIPS64:
		case rspec.ArchMIPS64N32:
		case rspec.ArchMIPSEL:
		case rspec.ArchMIPSEL64:
		case rspec.ArchMIPSEL64N32:
		case rspec.ArchPPC:
		case rspec.ArchPPC64:
		case rspec.ArchPPC64LE:
		case rspec.ArchS390:
		case rspec.ArchS390X:
		default:
			msgs = append(msgs, fmt.Sprintf("seccomp architecture %q is invalid", s.Architectures[index]))
		}
	}

	return
}

func envValid(env string) bool {
	items := strings.Split(env, "=")
	if len(items) < 2 {
		return false
	}
	for i, ch := range strings.TrimSpace(items[0]) {
		if !unicode.IsDigit(ch) && !unicode.IsLetter(ch) && ch != '_' {
			return false
		}
		if i == 0 && unicode.IsDigit(ch) {
			logrus.Warnf("Env %v: variable name beginning with digit is not recommended.", env)
		}
	}
	return true
}

func capValid(capability string) bool {
	for _, val := range defaultCaps {
		if val == capability {
			return true
		}
	}
	return false
}

func rlimitValid(rlimit string) bool {
	for _, val := range defaultRlimits {
		if val == rlimit {
			return true
		}
	}
	return false
}

func namespaceValid(ns rspec.Namespace) bool {
	switch ns.Type {
	case rspec.PIDNamespace:
	case rspec.NetworkNamespace:
	case rspec.MountNamespace:
	case rspec.IPCNamespace:
	case rspec.UTSNamespace:
	case rspec.UserNamespace:
	case rspec.CgroupNamespace:
	default:
		return false
	}
	return true
}

func deviceValid(d rspec.Device) bool {
	switch d.Type {
	case "b":
	case "c":
	case "u":
		if d.Major <= 0 {
			return false
		}
		if d.Minor <= 0 {
			return false
		}
	case "p":
		if d.Major > 0 || d.Minor > 0 {
			return false
		}
	default:
		return false
	}
	return true
}

func seccompActionValid(secc rspec.Action) bool {
	switch secc {
	case "":
	case rspec.ActKill:
	case rspec.ActTrap:
	case rspec.ActErrno:
	case rspec.ActTrace:
	case rspec.ActAllow:
	default:
		return false
	}
	return true
}

func syscallValid(s rspec.Syscall) bool {
	if !seccompActionValid(s.Action) {
		return false
	}
	for index := 0; index < len(s.Args); index++ {
		arg := s.Args[index]
		switch arg.Op {
		case rspec.OpNotEqual:
		case rspec.OpLessThan:
		case rspec.OpLessEqual:
		case rspec.OpEqualTo:
		case rspec.OpGreaterEqual:
		case rspec.OpGreaterThan:
		case rspec.OpMaskedEqual:
		default:
			return false
		}
	}
	return true
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func checkMandatoryUnit(field reflect.Value, tagField reflect.StructField, parent string) (msgs []string) {
	mandatory := !strings.Contains(tagField.Tag.Get("json"), "omitempty")
	switch field.Kind() {
	case reflect.Ptr:
		if mandatory && field.IsNil() {
			msgs = append(msgs, fmt.Sprintf("'%s.%s' should not be empty.", parent, tagField.Name))
		}
	case reflect.String:
		if mandatory && (field.Len() == 0) {
			msgs = append(msgs, fmt.Sprintf("'%s.%s' should not be empty.", parent, tagField.Name))
		}
	case reflect.Slice:
		if mandatory && (field.IsNil() || field.Len() == 0) {
			msgs = append(msgs, fmt.Sprintf("'%s.%s' should not be empty.", parent, tagField.Name))
			return
		}
		for index := 0; index < field.Len(); index++ {
			mValue := field.Index(index)
			if mValue.CanInterface() {
				msgs = append(msgs, checkMandatory(mValue.Interface())...)
			}
		}
	case reflect.Map:
		if mandatory && (field.IsNil() || field.Len() == 0) {
			msgs = append(msgs, fmt.Sprintf("'%s.%s' should not be empty.", parent, tagField.Name))
			return msgs
		}
		keys := field.MapKeys()
		for index := 0; index < len(keys); index++ {
			mValue := field.MapIndex(keys[index])
			if mValue.CanInterface() {
				msgs = append(msgs, checkMandatory(mValue.Interface())...)
			}
		}
	default:
	}

	return
}

func checkMandatory(obj interface{}) (msgs []string) {
	objT := reflect.TypeOf(obj)
	objV := reflect.ValueOf(obj)
	if isStructPtr(objT) {
		objT = objT.Elem()
		objV = objV.Elem()
	} else if !isStruct(objT) {
		return
	}

	for i := 0; i < objT.NumField(); i++ {
		t := objT.Field(i).Type
		if isStructPtr(t) && objV.Field(i).IsNil() {
			if !strings.Contains(objT.Field(i).Tag.Get("json"), "omitempty") {
				msgs = append(msgs, fmt.Sprintf("'%s.%s' should not be empty", objT.Name(), objT.Field(i).Name))
			}
		} else if (isStruct(t) || isStructPtr(t)) && objV.Field(i).CanInterface() {
			msgs = append(msgs, checkMandatory(objV.Field(i).Interface())...)
		} else {
			msgs = append(msgs, checkMandatoryUnit(objV.Field(i), objT.Field(i), objT.Name())...)
		}

	}
	return
}

func checkMandatoryFields(spec rspec.Spec, rootfs string, hostCheck bool) []string {
	logrus.Debugf("check mandatory fields")

	return checkMandatory(spec)
}
