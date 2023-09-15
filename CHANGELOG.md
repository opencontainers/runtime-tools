# v0.9.1

This release is a hotfix based on v0.9.0 with runtime-spec 1.0.2 support and a
switch to Go modules, to allow `go install` to use the correct version of the
spec.

# Additions
* Switch to go modules.
* generate: add support for --process-umask (backport of #706).

# v0.9.0

## Additions
* Windows: Typos and incorrect defaults (#683).
* validation: Add apparmor profile test(#684).
* generate: add oci-version option (#681).
* validation: Add SELinux Check (#682).
* generate: add process-cap-add and process-cap-drop option (#675).
* generate: Add generate option (#672).
* Initialize Config Windows Network for Windows Namespace (#666).

## Minor fixes and documentation
* validation-tests: fix several tests (#687).
* adding security and CoC links (#686).
* Simplified code (#685).
* Godeps: update hashicorp/go-multierror (#678).
* fix up vm parameters (#676).
* generate: fix capabilities add/drop option (#674).
* update to golang 1.11 (#670).


# v0.8.0

## Additions
* generate: Add generate.New support for Windows (#667).
* validation: add resource validation after delete (#654).
* mountinfo: parse empty strings in source (#652).

## Minor fixes and documentation
* readme: fix wrong filepath (#665).
* Makefile: add generate to gotest (#656).
* MAINTAINERS: remove philips (#659).
* Vendor in windows runtime-spec changes (#663).
* /proc should be mounted with nosuid, noexec, nodev to match the host (#664).
* validation: mounts: fix condition of source & type check (#660).
* Fix TAP output with multiple RuntimeInsideValidate (#658).
* fix some misspells (#649).


# v0.7.0
## Additions

* validation: use t.Fail when checking for main test errors (#645).
* travis: add go 1.10 (#647).
* validation: add more test cases for read-only paths tests (#644).
* validation: add more test cases for masked paths tests (#643).
* validation: test cgroups with different input values (#637).
* validation: add more test cases for private & slave propagations (#650).
* runtimetest: correctly check for a readable directory (#625).
* validation: add minor checks for ptmx and kill signal (#642).
* validation: add a new test for NSPathMatchTypeError (#636).
* validation: add test for NSProcInPath (#628).
* validation: test validation test with an empty hostname (#640).
* validation: add cgroup devices validation (#633).
* check the status of the state passed to hooks over stdin (#608).
* validation: fix nil deferences in cpu & blkio cgroups tests (#638).

## Minor fixes and documentation

* validation: fix nil dereference when handling multierror in hooks_stdin (#641).
* fix generate test in calling generate.New (#648).
* README: fix broken links to documentation (#646).
* validation/kill_no_effect: fix bug(#635).


# v0.6.0

## Additions

* add test case for KillNonCreateRunHaveNoEffect (#607).
* Add cgroupsPath validation (#631).
* validation: create: don't skip errors on state (#626).
* validation: add tests for NSNewNSWithoutPath & NSInheritWithoutType (#620).
* specerror: Add NewRFCError and NewRFCErrorOrPanic (#627).
* implement specerror (#604, #602, #591, #587, #580, #583, #584, #586).
* generate: Move Generator.spec to Generator.Config (#266).
* Respect the host platform (#194).
* runtimetest: Make TAP output more granular (#308).
* generate: add process-username option and fix it's validation (#614).
* validation: add process_user validation (#611).
* add hooks stdin test (#589).
* runtimetest: count correctly TAP tests (#594).
* contrib/rootfs-builder: Support timestamps and xz compression (#598).
* Add system validation  (#592).
* validation: run CLI with correct argument order (#600).
* validation: Add system validation (#590).
* validate: CheckLinux is platform dependent (#560).
* validation: Add error judgment to SetConfig (#585).
* validate: allow non-linux compatibility (#588).

## Minor fixes and documentation

* cgroups_v1: Correction parameters (#629).
* travis: fix fetch issue of golint (#630).
* validation: add more values for rlimits test (#623).
* doc: add developer guidelines (#621).
* bash: add os (#622).
* docs/command-line-interface: Require complete runtime coverage (#615).
* validation/test-yaml: Drop this local experiment (#616).
* validation: LinuxUIDMapping: fix tests (#597).
* Fix error messages in validation cgroup tests (#605).
* contrib/rootfs-builder: Use $(cat rootfs-files) (#606).
* validate: mv deviceValid to validate_linux (#603).
* Validate_linux: Modify the returned error (#601).
* runtimetest: fix root readonly check (#599).
* runtimetest: fix uid_map parsing (#596).
* Fix condition in BlockIO test (#595).
* generate/seccomp: platform independent values (#561).

# v0.5.0
## Additions

* validation: add tests when prestart/poststart/poststop hooks fail (#569).
* validate_test: add TestCheckMandatoryFields (#554).
* validation: add lifecycle validation (#558).
* validation: add 'state' test; using WaitingForStatus in insideValidation (#562).
* Relax LGTM requirement (#559, #566).
* validation: Fixes #556 (#557).

## Minor fixes and documentation

* validate_test: Complement test (#568).
* man: Modify the legal value of the rootfs-propagation (#548).
* generate: don't overwrite hook which has a same path (#571).
* validation: nil config support in lifecycle validate (#567).
* runtimetest: cmd/runtimetest/main: Run validateDefaultDevices even with process unset (#553).
* validation: Remove runc 'create' exit timing crutches (#563).
* validation/util/container: Use ExitError for stderr (#564).

# v0.4.0

## Additions

* specerror: Redefine error code as int64 (#501).
* validate: Improve the test of the configuration file (#504, #534, #537, #541).
* runtimetest: Add rootfs propagation test (#511).
* runtimetest: Add posixValidations (#510).
* runtimetest: Add host platform validation (#507).
* Makefile: Add version file (#417).
* validation: Complete Container Inside Test (#521).
* generate: Support json value for hooks (#525).
* generate: Support adding additional mounts for container (#279).
* generate: Support blkio related options (#235).
* cmd/runtimetest/main: Use TAP diagnostics for errors (#439).
* generate: Add linux-intelRdt-l3CacheSchema option (#529).
* filepath/clean: Add Windows support (#539).
* validate: Add validation when host-specific is set (#495).
* runtimetest: Add validation of cgroups (#93).
* generate: Generator solaris application container configuration (#532).
* generate: Add interface to remove mounts. (#544).
* validation/linux_cgroups_*: Generate TAP output (and outside-validation cleanup) (#542).
* generate: Windows-specific container configuration generate (#528).
* runtimetest: Add validateSeccomp (#514).
* validation: Add mount validation (#547).
* ...: Transition from tap Diagnostic(...) to YAML(...) (#533).

## Minor fixes and documentation

* runtimetest: Fix error return (#505).
* runtimetest: Move validateRlimits to defaultValidations (#506).
* runtimetest: Make validateRlimits silent on Windows (#509).
* runtimetest: Raise ConfigInRootBundleDir for missing config.json (#508).
* generate: Change process-tty to process-terminal (#517).
* generate: Fixed seccompSet (#513).
* runtimetest: Remove debug info (#518).
* generate: Fix error return (#520).
* validate: Fix nil deference (#522).
* generate: Fix DropProcessCapability... (#519).
* runtimetest: Fix nil dereference (#523).
* man: Small fixs (#526).
* validation: Fix idmappings test (#530).
* generate: Solve conflicting options problem (#441).
* generate: Use non-null validation instead of initialization (#540).
* validate: Modify the non-conforming validation (#538).
* validate: Fix id mappings (#531).
* validate: Remove duplicate verification (#535).
* generate: AddMounts should be AddMount you are only adding a single Mount (#545).
* generate: Recursive propagation flags should be legal to use (#543).
* generate: Modify the function return value (#546).
* generate: Hooks should be passed in as rspec.Hook, not as a string. (#549).

# v0.3.0

## Additions

* cmd/runtimetest: Adopt `DevicesAvailable` RFC code (#502).
* cmd/runtimetest: Adopt `DefaultRuntimeLinuxSymlinks`, `DefaultDevices`,
    `LinuxProcOomScoreAdjSet`, `MountsInOrder`, `SpecVersionInSemVer`,
    `PosixHooksPathAbs`, `ProcCwdAbs`, `ProcArgsOneEntryRequired`,
    `PosixProcRlimitsErrorOnDup`, `MountsDestAbs`, `MountsDestOnWindowsNotNested`,
    `PlatformSpecConfOnWindowsSet`, `MaskedPathsAbs`, `ReadonlyPathsAbs`
    RFC codes (#500).
* specerror: Turn all the RFC 2119 key words described in runtime-spec
    to RFC codes (#498, #497, #481, #458).
* specerror:  Add SplitLevel helper, Implement `--compliance-level` (#492).
* generate: generate smoke test (#491).
* travis: Add go 1.9 version (#487).
* rootfs-{arch}.tar.gz: Add per-arch tarballs (#479).
* generate: Add `--linux-device-cgroup-add` and
    `--linux-device-cgroup-remove` (#446).
* filepath: Add a stand-alone package for explicit-OS path logic (#445).

## Minor fixes and documentation

* cmd/runtimetest: Fix nil reference (#494).
* man: Fix typo (#493).
* generate: Correct rootfs default, allow unset "type" fields
    in resource devices whitelist (#491).
* validate: Fix compile issue (#490).
* bash: Fix command (#489).
* validate: Fix cap valiadtion (#488).
* generate: Fix rootfs-propagation (#484).

# v0.2.0

## Additions

* cmd/oci-runtime-tool/generate: Add specific cap-add and -drop
    commands (#358).
* validate: Ensure `root.path` is a GUID on non-Hyper-V Windows
    (#472).
* validate: Check `process.rlimits[].type` on Solaris (#461, #480).
* validate: Check configuration against JSON Schema (#197, #473, #474,
    #475, #476).

## Minor fixes and documentation

* validate: Avoid "0 errors occurred" failure (#462).
* validate: Remove empty string from valid seccomp actions (#468).
* validate: Require 0 or unset `major`/`minor` when
    `linux.devices[].type` is `p` (#460).
* generate: Fix cap add/drop and initialize in privileged mode (#464).
* generate: Do not validate caps when being dropped (#466, #469,
    #472).
* completions/bash/oci-runtime-tool: Fix broken cap completion (#467).
* rootfs.tar.gz: Bump to BusyBox 1.25.1 (#478)
