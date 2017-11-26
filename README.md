# oci-runtime-tool [![Build Status](https://travis-ci.org/opencontainers/runtime-tools.svg?branch=master)](https://travis-ci.org/opencontainers/runtime-tools) [![Go Report Card](https://goreportcard.com/badge/github.com/opencontainers/runtime-tools)](https://goreportcard.com/report/github.com/opencontainers/runtime-tools)

oci-runtime-tool is a collection of tools for working with the [OCI runtime specification][runtime-spec].
To build from source code, runtime-tools requires Go 1.7.x or above.

## Generating an OCI runtime spec configuration files

[`oci-runtime-tool generate`][generate.1] generates [configuration JSON][config.json] for an [OCI bundle][bundle].
[OCI-compatible runtimes][runtime-spec] like [runC][] expect to read the configuration from `config.json`.

```console
$ oci-runtime-tool generate --output config.json
$ cat config.json
{
        "ociVersion": "0.5.0",
        …
}
```

## Validating an OCI bundle

[`oci-runtime-tool validate`][validate.1] validates an OCI bundle.
The error message will be printed if the OCI bundle failed the validation procedure.

```console
$ oci-runtime-tool generate
$ oci-runtime-tool validate
INFO[0000] Bundle validation succeeded.
```

## Testing OCI runtimes

The runtime validation suite uses [`prove`][prove], which is packaged for most distributions (for example, it is in [Debian's `perl` package][debian-perl] and [Gentoo's `dev-lang/perl` package][gentoo-perl]).
If you cannot install `prove`, you can probably run the test suite with another [TAP consumer][tap-consumers], although you'll have to edit the [`Makefile`](Makefile) to replace `prove`.

```console
$ make runtimetest validation-executables
$ sudo make RUNTIME=runc localvalidation
RUNTIME=runc prove validation/linux_rootfs_propagation_shared.t validation/create.t validation/default.t validation/linux_readonly_paths.t validation/linux_masked_paths.t validation/mounts.t validation/process.t validation/root_readonly_false.t validation/linux_sysctl.t validation/linux_devices.t validation/linux_gid_mappings.t validation/process_oom_score_adj.t validation/process_capabilities.t validation/process_rlimits.t validation/root_readonly_true.t validation/linux_rootfs_propagation_unbindable.t validation/hostname.t validation/linux_uid_mappings.t
validation/linux_rootfs_propagation_shared.t ...... Failed 1/19 subtests
validation/create.t ............................... ok
validation/default.t .............................. ok
validation/linux_readonly_paths.t ................. ok
validation/linux_masked_paths.t ................... Failed 1/19 subtests
validation/mounts.t ............................... ok
validation/process.t .............................. ok
validation/root_readonly_false.t .................. ok
validation/linux_sysctl.t ......................... ok
validation/linux_devices.t ........................ ok
validation/linux_gid_mappings.t ................... Failed 1/19 subtests
validation/process_oom_score_adj.t ................ ok
validation/process_capabilities.t ................. ok
validation/process_rlimits.t ...................... ok
validation/root_readonly_true.t ................... ok
validation/linux_rootfs_propagation_unbindable.t .. failed to create the container
rootfsPropagation=unbindable is not supported
exit status 1
validation/linux_rootfs_propagation_unbindable.t .. Dubious, test returned 1 (wstat 256, 0x100)
No subtests run
validation/hostname.t ............................. ok
validation/linux_uid_mappings.t ................... failed to create the container
User namespace mappings specified, but USER namespace isn't enabled in the config
exit status 1
validation/linux_uid_mappings.t ................... Dubious, test returned 1 (wstat 256, 0x100)
No subtests run

Test Summary Report
-------------------
validation/linux_rootfs_propagation_shared.t    (Wstat: 0 Tests: 19 Failed: 1)
  Failed test:  16
validation/linux_masked_paths.t                 (Wstat: 0 Tests: 19 Failed: 1)
  Failed test:  13
validation/linux_gid_mappings.t                 (Wstat: 0 Tests: 19 Failed: 1)
  Failed test:  19
validation/linux_rootfs_propagation_unbindable.t (Wstat: 256 Tests: 0 Failed: 0)
  Non-zero exit status: 1
  Parse errors: No plan found in TAP output
validation/linux_uid_mappings.t                 (Wstat: 256 Tests: 0 Failed: 0)
  Non-zero exit status: 1
  Parse errors: No plan found in TAP output
Files=18, Tests=271, 31 wallclock secs ( 0.06 usr  0.01 sys +  0.52 cusr  0.21 csys =  0.80 CPU)
Result: FAIL
make: *** [Makefile:42: localvalidation] Error 1
```

If you are confident that the `validation/*.t` are current, you can run `prove` (or your preferred TAP consumer) directly to avoid unnecessary rebuilds:

```console
$ sudo RUNTIME=runc prove -Q validation/*.t

Test Summary Report
-------------------
…
Files=18, Tests=271, 31 wallclock secs ( 0.07 usr  0.01 sys +  0.51 cusr  0.22 csys =  0.81 CPU)
Result: FAIL
```

And you can run an individual test executable directly:

```console
$ RUNTIME=runc validation/default.t
TAP version 13
ok 1 - root filesystem
ok 2 - hostname
ok 3 - process
ok 4 - mounts
ok 5 - user
ok 6 - rlimits
ok 7 - capabilities
ok 8 - default symlinks
ok 9 - default file system
ok 10 - default devices
ok 11 - linux devices
ok 12 - linux process
ok 13 - masked paths
ok 14 - oom score adj
ok 15 - read only paths
ok 16 - rootfs propagation
ok 17 - sysctls
ok 18 - uid mappings
ok 19 - gid mappings
1..19
```

[bundle]: https://github.com/opencontainers/runtime-spec/blob/master/bundle.md
[config.json]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[debian-perl]: https://packages.debian.org/stretch/perl
[gentoo-perl]: https://packages.gentoo.org/packages/dev-lang/perl
[prove]: http://search.cpan.org/~leont/Test-Harness-3.39/bin/prove
[runC]: https://github.com/opencontainers/runc
[runtime-spec]: https://github.com/opencontainers/runtime-spec
[tap-consumers]: https://testanything.org/consumers.html

[generate.1]: man/oci-runtime-tool-generate.1.md
[validate.1]: man/oci-runtime-tool-validate.1.md
