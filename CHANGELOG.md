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
