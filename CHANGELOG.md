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
