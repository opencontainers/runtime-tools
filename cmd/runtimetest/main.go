package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/gocapability/capability"
	"github.com/urfave/cli"

	"github.com/opencontainers/runtime-tools/cmd/runtimetest/mount"
	rfc2119 "github.com/opencontainers/runtime-tools/error"
	"github.com/opencontainers/runtime-tools/specerror"

	"golang.org/x/sys/unix"
)

// gitCommit will be the hash that the binary was built from
// and will be populated by the Makefile
var gitCommit = ""

// version will be populated by the Makefile, read from
// VERSION file of the source code.
var version = ""

// PrGetNoNewPrivs isn't exposed in Golang so we define it ourselves copying the value from
// the kernel
const PrGetNoNewPrivs = 39

const specConfig = "config.json"

var (
	defaultFS = map[string]string{
		"/proc":    "proc",
		"/sys":     "sysfs",
		"/dev/pts": "devpts",
		"/dev/shm": "tmpfs",
	}

	defaultSymlinks = map[string]string{
		"/dev/fd":     "/proc/self/fd",
		"/dev/stdin":  "/proc/self/fd/0",
		"/dev/stdout": "/proc/self/fd/1",
		"/dev/stderr": "/proc/self/fd/2",
	}

	defaultDevices = []string{
		"/dev/null",
		"/dev/zero",
		"/dev/full",
		"/dev/random",
		"/dev/urandom",
		"/dev/tty",
		"/dev/ptmx",
	}
)

type validation struct {
	test        func(*rspec.Spec) error
	description string
}

func loadSpecConfig(path string) (spec *rspec.Spec, err error) {
	configPath := filepath.Join(path, specConfig)
	cf, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, specerror.NewError(specerror.ConfigInRootBundleDir, err, rspec.Version)
		}

		return nil, err
	}
	defer cf.Close()

	if err = json.NewDecoder(cf).Decode(&spec); err != nil {
		return
	}
	return spec, nil
}

func validatePosixUser(spec *rspec.Spec) error {
	if spec.Process == nil {
		return nil
	}

	uid := os.Getuid()
	if uint32(uid) != spec.Process.User.UID {
		return fmt.Errorf("UID expected: %v, actual: %v", spec.Process.User.UID, uid)
	}
	gid := os.Getgid()
	if uint32(gid) != spec.Process.User.GID {
		return fmt.Errorf("GID expected: %v, actual: %v", spec.Process.User.GID, gid)
	}

	groups, err := os.Getgroups()
	if err != nil {
		return err
	}

	groupsMap := make(map[int]bool)
	for _, g := range groups {
		groupsMap[g] = true
	}

	for _, g := range spec.Process.User.AdditionalGids {
		if !groupsMap[int(g)] {
			return fmt.Errorf("Groups expected: %v, actual (should be superset): %v", spec.Process.User.AdditionalGids, groups)
		}
	}

	return nil
}

func validateProcess(spec *rspec.Spec) error {
	if spec.Process == nil {
		return nil
	}

	if spec.Process.Cwd != "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if cwd != spec.Process.Cwd {
			return fmt.Errorf("Cwd expected: %v, actual: %v", spec.Process.Cwd, cwd)
		}
	}

	for _, env := range spec.Process.Env {
		parts := strings.Split(env, "=")
		key := parts[0]
		expectedValue := parts[1]
		actualValue := os.Getenv(key)
		if actualValue != expectedValue {
			return fmt.Errorf("Env %v expected: %v, actual: %v", key, expectedValue, actualValue)
		}
	}

	return nil
}

func validateLinuxProcess(spec *rspec.Spec) error {
	if spec.Process == nil {
		return nil
	}

	cmdlineBytes, err := ioutil.ReadFile("/proc/self/cmdline")
	if err != nil {
		return err
	}

	args := bytes.Split(bytes.Trim(cmdlineBytes, "\x00"), []byte("\x00"))
	if len(args) != len(spec.Process.Args) {
		return fmt.Errorf("Process arguments expected: %v, actual: %v", len(spec.Process.Args), len(args))
	}
	for i, a := range args {
		if string(a) != spec.Process.Args[i] {
			return fmt.Errorf("Process arguments expected: %v, actual: %v", string(a), spec.Process.Args[i])
		}
	}

	ret, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, PrGetNoNewPrivs, 0, 0, 0, 0, 0)
	if errno != 0 {
		return errno
	}
	if spec.Process.NoNewPrivileges && ret != 1 {
		return fmt.Errorf("NoNewPrivileges expected: true, actual: false")
	}
	if !spec.Process.NoNewPrivileges && ret != 0 {
		return fmt.Errorf("NoNewPrivileges expected: false, actual: true")
	}

	return nil
}

func validateCapabilities(spec *rspec.Spec) error {
	if spec.Process == nil || spec.Process.Capabilities == nil {
		return nil
	}

	last := capability.CAP_LAST_CAP
	// workaround for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}

	processCaps, err := capability.NewPid(0)
	if err != nil {
		return err
	}

	for _, capType := range []struct {
		capType capability.CapType
		config  []string
	}{
		{
			capType: capability.BOUNDING,
			config:  spec.Process.Capabilities.Bounding,
		},
		{
			capType: capability.EFFECTIVE,
			config:  spec.Process.Capabilities.Effective,
		},
		{
			capType: capability.INHERITABLE,
			config:  spec.Process.Capabilities.Inheritable,
		},
		{
			capType: capability.PERMITTED,
			config:  spec.Process.Capabilities.Permitted,
		},
		{
			capType: capability.AMBIENT,
			config:  spec.Process.Capabilities.Ambient,
		},
	} {
		expectedCaps := make(map[string]bool)
		for _, ec := range capType.config {
			expectedCaps[ec] = true
		}

		for _, cap := range capability.List() {
			if cap > last {
				continue
			}

			capKey := fmt.Sprintf("CAP_%s", strings.ToUpper(cap.String()))
			expectedSet := expectedCaps[capKey]
			actuallySet := processCaps.Get(capType.capType, cap)
			if expectedSet && !actuallySet {
				return fmt.Errorf("expected %s capability %v not set", capType.capType, capKey)
			} else if !expectedSet && actuallySet {
				return fmt.Errorf("unexpected %s capability %v set", capType.capType, capKey)
			}
		}
	}

	return nil
}

func validateHostname(spec *rspec.Spec) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	if spec.Hostname != "" && hostname != spec.Hostname {
		return fmt.Errorf("Hostname expected: %v, actual: %v", spec.Hostname, hostname)
	}
	return nil
}

func validateRlimits(spec *rspec.Spec) error {
	if spec.Process == nil {
		return nil
	}

	for _, r := range spec.Process.Rlimits {
		rl, err := strToRlimit(r.Type)
		if err != nil {
			return err
		}

		var rlimit syscall.Rlimit
		if err := syscall.Getrlimit(rl, &rlimit); err != nil {
			return err
		}

		if rlimit.Cur != r.Soft {
			return fmt.Errorf("%v rlimit soft expected: %v, actual: %v", r.Type, r.Soft, rlimit.Cur)
		}
		if rlimit.Max != r.Hard {
			return fmt.Errorf("%v rlimit hard expected: %v, actual: %v", r.Type, r.Hard, rlimit.Max)
		}
	}
	return nil
}

func validateSysctls(spec *rspec.Spec) error {
	if spec.Linux == nil {
		return nil
	}
	for k, v := range spec.Linux.Sysctl {
		keyPath := filepath.Join("/proc/sys", strings.Replace(k, ".", "/", -1))
		vBytes, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return err
		}
		value := strings.TrimSpace(string(bytes.Trim(vBytes, "\x00")))
		if value != v {
			return fmt.Errorf("Sysctl %v value expected: %v, actual: %v", k, v, value)
		}
	}
	return nil
}

func testWriteAccess(path string) error {
	tmpfile, err := ioutil.TempFile(path, "Test")
	if err != nil {
		return err
	}

	tmpfile.Close()
	os.RemoveAll(filepath.Join(path, tmpfile.Name()))

	return nil
}

func validateRootFS(spec *rspec.Spec) error {
	if spec.Root == nil {
		return nil
	}

	if spec.Root.Readonly {
		err := testWriteAccess("/")
		if err == nil {
			return specerror.NewError(specerror.RootReadonlyImplement, fmt.Errorf("rootfs must be readonly"), rspec.Version)
		}
	} else {
		err := testWriteAccess("/")
		if err != nil {
			return specerror.NewError(specerror.RootReadonlyImplement, fmt.Errorf("rootfs must not be readonly"), rspec.Version)
		}
	}

	return nil
}

func validateRootfsPropagation(spec *rspec.Spec) error {
	if spec.Linux == nil || spec.Linux.RootfsPropagation == "" {
		return nil
	}

	targetDir, err := ioutil.TempDir("/", "target")
	if err != nil {
		return err
	}
	defer os.RemoveAll(targetDir)

	switch spec.Linux.RootfsPropagation {
	case "shared", "slave", "private":
		mountDir, err := ioutil.TempDir("/", "mount")
		if err != nil {
			return err
		}
		defer os.RemoveAll(mountDir)

		testDir, err := ioutil.TempDir("/", "test")
		if err != nil {
			return err
		}
		defer os.RemoveAll(testDir)

		tmpfile, err := ioutil.TempFile(testDir, "example")
		if err != nil {
			return err
		}
		defer os.Remove(tmpfile.Name())

		if err := unix.Mount("/", targetDir, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			return err
		}
		defer unix.Unmount(targetDir, unix.MNT_DETACH)
		if err := unix.Mount(testDir, mountDir, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			return err
		}
		defer unix.Unmount(mountDir, unix.MNT_DETACH)
		if _, err := os.Stat(filepath.Join(targetDir, filepath.Join(mountDir, filepath.Base(tmpfile.Name())))); os.IsNotExist(err) {
			if spec.Linux.RootfsPropagation == "shared" {
				return fmt.Errorf("rootfs should be %s, but not", spec.Linux.RootfsPropagation)
			}
			return nil
		}
		if spec.Linux.RootfsPropagation == "shared" {
			return nil
		}
		return fmt.Errorf("rootfs should be %s, but not", spec.Linux.RootfsPropagation)
	case "unbindable":
		if err := unix.Mount("/", targetDir, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
			if err == syscall.EINVAL {
				return nil
			}
			return err
		}
		defer unix.Unmount(targetDir, unix.MNT_DETACH)
		return fmt.Errorf("rootfs expected to be unbindable, but not")
	default:
		logrus.Warnf("unrecognized linux.rootfsPropagation %s", spec.Linux.RootfsPropagation)
	}

	return nil
}

func validateDefaultFS(spec *rspec.Spec) error {
	mountInfos, err := mount.GetMounts()
	if err != nil {
		specerror.NewError(specerror.DefaultFilesystems, err, spec.Version)
	}

	mountsMap := make(map[string]string)
	for _, mountInfo := range mountInfos {
		mountsMap[mountInfo.Mountpoint] = mountInfo.Fstype
	}

	for fs, fstype := range defaultFS {
		if !(mountsMap[fs] == fstype) {
			return specerror.NewError(specerror.DefaultFilesystems, fmt.Errorf("%v SHOULD exist and expected type is %v", fs, fstype), rspec.Version)
		}
	}

	return nil
}

func validateLinuxDevices(spec *rspec.Spec) error {
	if spec.Linux == nil {
		return nil
	}
	for _, device := range spec.Linux.Devices {
		fi, err := os.Stat(device.Path)
		if err != nil {
			return err
		}
		fStat, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			return specerror.NewError(specerror.DevicesAvailable, fmt.Errorf("cannot determine state for device %s", device.Path), rspec.Version)
		}
		var devType string
		switch fStat.Mode & syscall.S_IFMT {
		case syscall.S_IFCHR:
			devType = "c"
		case syscall.S_IFBLK:
			devType = "b"
		case syscall.S_IFIFO:
			devType = "p"
		default:
			devType = "unmatched"
		}
		if devType != device.Type || (devType == "c" && device.Type == "u") {
			return fmt.Errorf("device %v expected type is %v, actual is %v", device.Path, device.Type, devType)
		}
		if devType != "p" {
			dev := fStat.Rdev
			major := (dev >> 8) & 0xfff
			minor := (dev & 0xff) | ((dev >> 12) & 0xfff00)
			if int64(major) != device.Major || int64(minor) != device.Minor {
				return fmt.Errorf("%v device number expected is %v:%v, actual is %v:%v", device.Path, device.Major, device.Minor, major, minor)
			}
		}
		if device.FileMode != nil {
			expectedPerm := *device.FileMode & os.ModePerm
			actualPerm := fi.Mode() & os.ModePerm
			if expectedPerm != actualPerm {
				return fmt.Errorf("%v filemode expected is %v, actual is %v", device.Path, expectedPerm, actualPerm)
			}
		}
		if device.UID != nil {
			if *device.UID != fStat.Uid {
				return fmt.Errorf("%v uid expected is %v, actual is %v", device.Path, *device.UID, fStat.Uid)
			}
		}
		if device.GID != nil {
			if *device.GID != fStat.Gid {
				return fmt.Errorf("%v uid expected is %v, actual is %v", device.Path, *device.GID, fStat.Gid)
			}
		}
	}

	return nil
}

func validateDefaultSymlinks(spec *rspec.Spec) error {
	for symlink, dest := range defaultSymlinks {
		fi, err := os.Lstat(symlink)
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
			return specerror.NewError(specerror.DefaultRuntimeLinuxSymlinks,
				fmt.Errorf("%v is not a symbolic link as expected", symlink),
				rspec.Version)
		}
		realDest, err := os.Readlink(symlink)
		if err != nil {
			return err
		}
		if realDest != dest {
			return specerror.NewError(specerror.DefaultRuntimeLinuxSymlinks,
				fmt.Errorf("link destation of %v expected is %v, actual is %v",
					symlink, dest, realDest),
				rspec.Version)
		}
	}

	return nil
}

func validateDefaultDevices(spec *rspec.Spec) error {
	if spec.Process != nil && spec.Process.Terminal {
		defaultDevices = append(defaultDevices, "/dev/console")
	}

	for _, device := range defaultDevices {
		fi, err := os.Stat(device)
		if err != nil {
			if os.IsNotExist(err) {
				return specerror.NewError(specerror.DefaultDevices,
					fmt.Errorf("device node %v not found", device),
					rspec.Version)
			}
			return err
		}
		if fi.Mode()&os.ModeDevice != os.ModeDevice {
			return specerror.NewError(specerror.DefaultDevices,
				fmt.Errorf("file %v is not a device as expected", device),
				rspec.Version)
		}
	}

	return nil
}

func validateMaskedPaths(spec *rspec.Spec) error {
	if spec.Linux == nil {
		return nil
	}
	for _, maskedPath := range spec.Linux.MaskedPaths {
		f, err := os.Open(maskedPath)
		if err != nil {
			return err
		}
		defer f.Close()
		b := make([]byte, 1)
		_, err = f.Read(b)
		if err != io.EOF {
			return fmt.Errorf("%v should not be readable", maskedPath)
		}
	}
	return nil
}

func validateSeccomp(spec *rspec.Spec) error {
	if spec.Linux == nil || spec.Linux.Seccomp == nil {
		return nil
	}
	t := tap.New()
	for _, sys := range spec.Linux.Seccomp.Syscalls {
		if sys.Action == "SCMP_ACT_ERRNO" {
			for _, name := range sys.Names {
				if name == "getcwd" {
					_, err := os.Getwd()
					if err == nil {
						t.Diagnostic("getcwd did not return an error")
					}
				} else {
					t.Skip(1, fmt.Sprintf("%s syscall returns errno", name))
				}
			}
		} else {
			t.Skip(1, fmt.Sprintf("syscall action %s", sys.Action))
		}
	}
	return nil
}

func validateROPaths(spec *rspec.Spec) error {
	if spec.Linux == nil {
		return nil
	}
	for _, v := range spec.Linux.ReadonlyPaths {
		err := testWriteAccess(v)
		if err == nil {
			return fmt.Errorf("%v should be readonly", v)
		}
	}

	return nil
}

func validateOOMScoreAdj(spec *rspec.Spec) error {
	if spec.Process != nil && spec.Process.OOMScoreAdj != nil {
		expected := *spec.Process.OOMScoreAdj
		f, err := os.Open("/proc/self/oom_score_adj")
		if err != nil {
			return err
		}
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			if err := s.Err(); err != nil {
				return err
			}
			text := strings.TrimSpace(s.Text())
			actual, err := strconv.Atoi(text)
			if err != nil {
				return err
			}
			if actual != expected {
				return specerror.NewError(specerror.LinuxProcOomScoreAdjSet, fmt.Errorf("oomScoreAdj expected: %v, actual: %v", expected, actual), rspec.Version)
			}
		}
	}

	return nil
}

func getIDMappings(path string) ([]rspec.LinuxIDMapping, error) {
	var idMaps []rspec.LinuxIDMapping
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}

		idMap := strings.Fields(strings.TrimSpace(s.Text()))
		if len(idMap) == 3 {
			hostID, err := strconv.ParseUint(idMap[0], 0, 32)
			if err != nil {
				return nil, err
			}
			containerID, err := strconv.ParseUint(idMap[1], 0, 32)
			if err != nil {
				return nil, err
			}
			mapSize, err := strconv.ParseUint(idMap[2], 0, 32)
			if err != nil {
				return nil, err
			}
			idMaps = append(idMaps, rspec.LinuxIDMapping{HostID: uint32(hostID), ContainerID: uint32(containerID), Size: uint32(mapSize)})
		} else {
			return nil, fmt.Errorf("invalid format in %v", path)
		}
	}

	return idMaps, nil
}

func validateIDMappings(mappings []rspec.LinuxIDMapping, path string, property string) error {
	idMaps, err := getIDMappings(path)
	if err != nil {
		return fmt.Errorf("can not get items: %v", err)
	}
	if len(mappings) != 0 && len(mappings) != len(idMaps) {
		return fmt.Errorf("expected %d entries in %v, but acutal is %d", len(mappings), path, len(idMaps))
	}
	for _, v := range mappings {
		exist := false
		for _, cv := range idMaps {
			if v.HostID == cv.HostID && v.ContainerID == cv.ContainerID && v.Size == cv.Size {
				exist = true
				break
			}
		}
		if !exist {
			return fmt.Errorf("%v is not applied as expected", property)
		}
	}

	return nil
}

func validateUIDMappings(spec *rspec.Spec) error {
	if spec.Linux == nil {
		return nil
	}
	return validateIDMappings(spec.Linux.UIDMappings, "/proc/self/uid_map", "linux.uidMappings")
}

func validateGIDMappings(spec *rspec.Spec) error {
	if spec.Linux == nil {
		return nil
	}
	return validateIDMappings(spec.Linux.GIDMappings, "/proc/self/gid_map", "linux.gidMappings")
}

func mountMatch(configMount rspec.Mount, sysMount *mount.Info) error {
	sys := rspec.Mount{
		Destination: sysMount.Mountpoint,
		Type:        sysMount.Fstype,
		Source:      sysMount.Source,
	}

	if filepath.Clean(configMount.Destination) != sys.Destination {
		return fmt.Errorf("mount destination expected: %v, actual: %v", configMount.Destination, sys.Destination)
	}

	if configMount.Type != sys.Type {
		return fmt.Errorf("mount %v type expected: %v, actual: %v", configMount.Destination, configMount.Type, sys.Type)
	}

	if filepath.Clean(configMount.Source) != sys.Source {
		return fmt.Errorf("mount %v source expected: %v, actual: %v", configMount.Destination, configMount.Source, sys.Source)
	}

	return nil
}

func validatePosixMounts(spec *rspec.Spec) error {
	mountInfos, err := mount.GetMounts()
	if err != nil {
		return err
	}

	var mountErrs error
	var consumedSys = make(map[int]bool)
	highestMatchedConfig := -1
	highestMatchedSystem := -1
	var j = 0
	for i, configMount := range spec.Mounts {
		if configMount.Type == "bind" || configMount.Type == "rbind" {
			// TODO: add bind or rbind check.
			continue
		}

		found := false
		for k, sysMount := range mountInfos[j:] {
			if err := mountMatch(configMount, sysMount); err == nil {
				found = true
				j += k + 1
				consumedSys[j-1] = true
				if j > highestMatchedSystem {
					highestMatchedSystem = j - 1
					highestMatchedConfig = i
				}
				break
			}
		}
		if !found {
			if j > 0 {
				for k, sysMount := range mountInfos[:j-1] {
					if _, ok := consumedSys[k]; ok {
						continue
					}
					if err := mountMatch(configMount, sysMount); err == nil {
						found = true
						break
					}
				}
			}
			if found {
				mountErrs = multierror.Append(
					mountErrs,
					specerror.NewError(specerror.MountsInOrder,
						fmt.Errorf(
							"mounts[%d] %v mounted before mounts[%d] %v",
							i,
							configMount,
							highestMatchedConfig,
							spec.Mounts[highestMatchedConfig]),
						rspec.Version))
			} else {
				mountErrs = multierror.Append(
					mountErrs,
					specerror.NewError(specerror.MountsInOrder, fmt.Errorf(
						"mounts[%d] %v does not exist",
						i,
						configMount), rspec.Version))
			}
		}
	}

	return mountErrs
}

func run(context *cli.Context) error {
	logLevelString := context.String("log-level")
	logLevel, err := logrus.ParseLevel(logLevelString)
	if err != nil {
		return err
	}
	logrus.SetLevel(logLevel)

	platform := runtime.GOOS
	if platform != "linux" && platform != "solaris" && platform != "windows" {
		return fmt.Errorf("runtime-tools has not implemented testing for your platform %q, because the spec has nothing to say about it", platform)
	}

	inputPath := context.String("path")
	spec, err := loadSpecConfig(inputPath)
	if err != nil {
		return err
	}

	defaultValidations := []validation{
		{
			test:        validateRootFS,
			description: "root filesystem",
		},
		{
			test:        validateHostname,
			description: "hostname",
		},
		{
			test:        validateProcess,
			description: "process",
		},
	}

	posixValidations := []validation{
		{
			test:        validatePosixMounts,
			description: "mounts",
		},
		{
			test:        validatePosixUser,
			description: "user",
		},
		{
			test:        validateRlimits,
			description: "rlimits",
		},
	}

	linuxValidations := []validation{
		{
			test:        validateCapabilities,
			description: "capabilities",
		},
		{
			test:        validateDefaultSymlinks,
			description: "default symlinks",
		},
		{
			test:        validateDefaultFS,
			description: "default file system",
		},
		{
			test:        validateDefaultDevices,
			description: "default devices",
		},
		{
			test:        validateLinuxDevices,
			description: "linux devices",
		},
		{
			test:        validateLinuxProcess,
			description: "linux process",
		},
		{
			test:        validateMaskedPaths,
			description: "masked paths",
		},
		{
			test:        validateOOMScoreAdj,
			description: "oom score adj",
		},
		{
			test:        validateSeccomp,
			description: "seccomp",
		},
		{
			test:        validateROPaths,
			description: "read only paths",
		},
		{
			test:        validateRootfsPropagation,
			description: "rootfs propagation",
		},
		{
			test:        validateSysctls,
			description: "sysctls",
		},
		{
			test:        validateUIDMappings,
			description: "uid mappings",
		},
		{
			test:        validateGIDMappings,
			description: "gid mappings",
		},
	}

	t := tap.New()
	t.Header(0)

	complianceLevelString := context.String("compliance-level")
	complianceLevel, err := rfc2119.ParseLevel(complianceLevelString)
	if err != nil {
		complianceLevel = rfc2119.Must
		logrus.Warningf("%s, using 'MUST' by default.", err.Error())
	}

	validations := defaultValidations
	if platform == "linux" {
		validations = append(validations, posixValidations...)
		validations = append(validations, linuxValidations...)
	} else if platform == "solaris" {
		validations = append(validations, posixValidations...)
	}

	for _, v := range validations {
		err := v.test(spec)
		if err == nil {
			t.Pass(v.description)
		} else {
			merr, ok := err.(*multierror.Error)
			if ok {
				for _, err = range merr.Errors {
					if e, ok := err.(*rfc2119.Error); ok {
						t.Ok(e.Level < complianceLevel, v.description)
					} else {
						t.Fail(v.description)
					}
					t.YAML(map[string]string{"error": err.Error()})
				}
			} else {
				if e, ok := err.(*rfc2119.Error); ok {
					t.Ok(e.Level < complianceLevel, v.description)
				} else {
					t.Fail(v.description)
				}
				t.YAML(map[string]string{"error": err.Error()})
			}
		}
	}
	t.AutoPlan()

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "runtimetest"
	if gitCommit != "" {
		app.Version = fmt.Sprintf("%s, commit: %s", version, gitCommit)
	} else {
		app.Version = version
	}
	app.Usage = "Compare the environment with an OCI configuration"
	app.Description = "runtimetest compares its current environment with an OCI runtime configuration read from config.json in its current working directory.  The tests are fairly generic and cover most configurations used by the runtime validation suite, but there are corner cases where a container launched by a valid runtime would not satisfy runtimetest."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "error",
			Usage: "Log level (panic, fatal, error, warn, info, or debug)",
		},
		cli.StringFlag{
			Name:  "path",
			Value: ".",
			Usage: "Path to the configuration",
		},
		cli.StringFlag{
			Name:  "compliance-level",
			Value: "must",
			Usage: "Compliance level (may, should or must)",
		},
	}

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
