# Runtime compliance testing

## Supported APIs

In order to be tested for [compliance][], runtimes MUST support at least one of the following APIs:

* Version 1.0.1 of the [OCI Runtime Command Line Interface](command-line-interface.md).

## Running the runtime validation suite

The runtime validation suite uses [node-tap][], which is packaged for some distributions (for example, it is in [Debian's `node-tap` package][debian-node-tap]).
If your distribution does not package node-tap, you can install [npm][] (for example, from [Gentoo's `nodejs` package][gentoo-nodejs]) and use it:

```console
$ npm install tap
```
### From a release

Check if your release has pre-compiled tests on the [release page][releases] page.

```
$ tar xf runtime-tools-v0.6.0.tar.gz
$ RUNTIME=runc tap ./runtime-tools-v0.6.0/validation/*.t
```

### From source

Build the validation executables:

```console
$ make runtimetest validation-executables
```

Runtime validation currently [only supports](docs/runtime-compliance-testing.md) the [OCI Runtime Command Line Interface](doc/command-line-interface.md).
If we add support for alternative APIs in the future, runtime validation will gain an option to select the desired runtime API.
For the command line interface, the `RUNTIME` option selects the runtime command (`funC` in the [OCI Runtime Command Line Interface](doc/command-line-interface.md)).

```
$ sudo TAP="$(which tap)" RUNTIME=runc make localvalidation
RUNTIME=runc /home/alban/.nvm/versions/node/v9.7.1/bin/tap validation/pidfile.t validation/linux_cgroups_memory.t validation/linux_rootfs_propagation_shared.t validation/kill.t validation/linux_readonly_paths.t validation/hostname.t validation/hooks_stdin.t validation/create.t validation/poststart.t validation/linux_cgroups_network.t validation/poststop_fail.t validation/prestart_fail.t validation/linux_cgroups_relative_blkio.t validation/default.t validation/poststop.t validation/linux_seccomp.t validation/prestart.t validation/process_rlimits.t validation/linux_masked_paths.t validation/killsig.t validation/process.t validation/linux_cgroups_relative_pids.t validation/hooks.t validation/linux_rootfs_propagation_unbindable.t validation/linux_cgroups_relative_cpus.t validation/misc_props.t validation/linux_sysctl.t validation/process_oom_score_adj.t validation/linux_devices.t validation/process_capabilities_fail.t validation/start.t validation/linux_cgroups_pids.t validation/process_capabilities.t validation/poststart_fail.t validation/linux_cgroups_relative_hugetlb.t validation/mounts.t validation/linux_cgroups_hugetlb.t validation/linux_cgroups_relative_memory.t validation/state.t validation/root_readonly_true.t validation/linux_cgroups_blkio.t validation/delete.t validation/linux_cgroups_relative_network.t validation/process_rlimits_fail.t validation/linux_cgroups_cpus.t validation/linux_uid_mappings.t
validation/pidfile.t .................................. 1/1 455ms
validation/linux_cgroups_memory.t ..................... 9/9
validation/linux_rootfs_propagation_shared.t ........ 19/20
  not ok rootfs propagation
    error: 'rootfs should be shared, but not'

validation/kill.t ..................................... 5/5 13s
validation/linux_readonly_paths.t ................... 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/hostname.t ............................... 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/hooks_stdin.t .............................. 3/3 1s
validation/create.t ................................... 4/4
validation/poststart.t ................................ 0/1
  Skipped: 1
    validation/poststart.t

validation/linux_cgroups_network.t .................... 5/5
failed to create the container
container_linux.go:348: starting container process caused "process_linux.go:402: container init caused \"process_linux.go:385: running prestart hook 0 caused \\\"error running hook: exit status 1, stdout: , stderr: \\\"\""
validation/poststop_fail.t ............................ 0/1
  Skipped: 1
    validation/poststop_fail.t

validation/prestart_fail.t ............................ 0/1
  Skipped: 1
    validation/prestart_fail.t

validation/linux_cgroups_relative_blkio.t ........... 15/15
validation/default.t ................................ 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

failed to create the container
container_linux.go:348: starting container process caused "error adding seccomp rule for syscall personality: requested action matches default action of filter"
exit status 1
validation/poststop.t ................................. 0/1
  Skipped: 1
    validation/poststop.t

validation/linux_seccomp.t ............................ 0/1
  not ok validation/linux_seccomp.t
    error: >-
      Pre-start hooks MUST be called after the `start` operation is called
    
      Refer to:
      https://github.com/opencontainers/runtime-spec/blob/v1.0.0/config.md#prestart

validation/prestart.t ................................. 0/1
  Skipped: 1
    validation/prestart.t

validation/process_rlimits.t ........................ 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/linux_masked_paths.t ..................... 19/26
  not ok masked paths
    error: /masktest should not be readable

  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/killsig.t .................................. 1/1 1s
validation/process.t ................................ 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/linux_cgroups_relative_pids.t .............. 3/3
validation/hooks.t .................................... 0/1
  Skipped: 1
    validation/hooks.t

validation/linux_rootfs_propagation_unbindable.t .... 19/20
  not ok rootfs propagation
    error: 'rootfs expected to be unbindable, but not'

validation/linux_cgroups_relative_cpus.t .............. 7/7
validation/misc_props.t ............................... 2/3 4s
  not ok runtimes that are reading or processing this configuration file MUST generate an error when invalid or unsupported values are encountered
    reference: >-
      https://github.com/opencontainers/runtime-spec/blob/v1.0.0/config.md#valid-values

validation/linux_sysctl.t ........................... 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/process_oom_score_adj.t .................. 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/linux_devices.t .......................... 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/process_capabilities_fail.t ......Any value which cannot be mapped to a relevant kernel interface MUST cause an error
Refer to: https://github.com/opencontainers/runtime-spec/blob/v1.0.0/config.md#linux-process
validation/process_capabilities_fail.t .............. 20/27
  not ok validation/process_capabilities_fail.t
    timeout: 30000
    file: validation/process_capabilities_fail.t
    command: validation/process_capabilities_fail.t
    args: []
    stdio:
      - 0
      - pipe
      - 2
    cwd: /home/alban/go/src/github.com/opencontainers/runtime-tools
    exitCode: 1

  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/start.t .....exit status 2
validation/start.t .................................... 6/7
  not ok test count !== plan
    +++ found                                                           
    --- wanted                                                          
    -1                                                                  
    +6                                                                  
    results:
      ok: false
      count: 6
      pass: 6
      fail: 1
      bailout: false
      todo: 0
      skip: 0
      plan:
        start: null
        end: null
        skipAll: false
        skipReason: ''
        comment: ''
      failures:
        - tapError: no plan

validation/linux_cgroups_pids.t ....................... 3/3
validation/process_capabilities.t ....................failed to create the container
container_linux.go:376: running poststart hook 0 caused "error running hook: exit status 1, stdout: , stderr: "
validation/process_capabilities.t ................... 20/20
validation/poststart_fail.t ........................... 0/1
  Skipped: 1
    validation/poststart_fail.t

validation/linux_cgroups_relative_hugetlb.t ........... 4/4
validation/mounts.t ................................. 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/linux_cgroups_hugetlb.t .................... 0/1
  Skipped: 1
    validation/linux_cgroups_hugetlb.t no tests found

validation/linux_cgroups_relative_memory.t ............ 9/9
validation/state.t .................................... 3/3
validation/root_readonly_true.t ..................... 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

validation/linux_cgroups_blkio.t .................... 15/15
validation/delete.t ................................... 3/5 22s
  not ok attempting to `delete` a container that is not `stopped` MUST generate an error
    reference: 'https://github.com/opencontainers/runtime-spec/blob/v1.0.0/runtime.md#delete'
  
  not ok attempting to `delete` a container that is not `stopped` MUST have no effect on the container
    error: exit status 1
    reference: 'https://github.com/opencontainers/runtime-spec/blob/v1.0.0/runtime.md#delete'
    stderr: |
      container "dd7f1685-1ea8-468c-8e6a-37f5f9f9ca64" does not exist

validation/linux_cgroups_relative_network.t .....failed to create the container
wrong rlimit value: RLIMIT_TEST
validation/linux_cgroups_relative_network.t ........... 5/5
validation/process_rlimits_fail.t ..................... 0/1
  Skipped: 1
    validation/process_rlimits_fail.t no tests found

validation/linux_cgroups_cpus.t ....................... 0/1
  Skipped: 1
    validation/linux_cgroups_cpus.t no tests found

validation/linux_uid_mappings.t ..................... 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

total ............................................. 420/517
  

  420 passing (1m)
  88 pending
  9 failing

make: *** [Makefile:44: localvalidation] Error 1
```

You can also run an individual test executable directly:

```console
$ sudo RUNTIME=runc validation/default.t
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
ok 15 # SKIP syscall action SCMP_ACT_ALLOW
ok 16 # SKIP syscall action SCMP_ACT_ALLOW
ok 17 # SKIP syscall action SCMP_ACT_ALLOW
ok 18 # SKIP syscall action SCMP_ACT_ALLOW
ok 19 # SKIP syscall action SCMP_ACT_ALLOW
ok 20 # SKIP syscall action SCMP_ACT_ALLOW
ok 21 - seccomp
ok 22 - read only paths
ok 23 - rootfs propagation
ok 24 - sysctls
ok 25 - uid mappings
ok 26 - gid mappings
1..26
```

or with the environment variable `VALIDATION_TESTS`:

```console
$ sudo make TAP=$(which tap) RUNTIME=runc VALIDATION_TESTS=validation/default.t localvalidation
RUNTIME=runc /home/alban/.nvm/versions/node/v9.7.1/bin/tap validation/default.t
validation/default.t ................................ 20/26
  Skipped: 6
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW
     syscall action SCMP_ACT_ALLOW

total ............................................... 20/26

  20 passing (257.078ms)
  6 pending
```

If you cannot install node-tap, you can probably run the test suite with another [TAP consumer][tap-consumers].
For example, with [`prove`][prove]:

```console
$ sudo make TAP="prove -Q -j9" RUNTIME=runc VALIDATION_TESTS="validation/default.t validation/linux_cgroups_memory.t" localvalidation
RUNTIME=runc prove -Q -j9 validation/default.t validation/linux_cgroups_memory.t
All tests successful.
Files=2, Tests=35,  1 wallclock secs ( 0.03 usr  0.00 sys +  0.12 cusr  0.12 csys =  0.27 CPU)
Result: PASS
```


[compliance]: https://github.com/opencontainers/runtime-spec/blob/v1.0.1/spec.md
[debian-node-tap]: https://packages.debian.org/stretch/node-tap
[debian-nodejs]: https://packages.debian.org/stretch/nodejs
[gentoo-nodejs]: https://packages.gentoo.org/packages/net-libs/nodejs
[node-tap]: http://www.node-tap.org/
[npm]: https://www.npmjs.com/
[prove]: http://search.cpan.org/~leont/Test-Harness-3.39/bin/prove
[tap-consumers]: https://testanything.org/consumers.html
[releases]: https://github.com/opencontainers/runtime-tools/releases
