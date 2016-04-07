# ocitools

ocitools is a collection of tools for working with the [OCI runtime specification](https://github.com/opencontainers/runtime-spec).

Generating OCI runtime spec configuration files
------------------------------------------

```
# ocitools generate --help
NAME:
   generate - generate a OCI spec file

USAGE:
   command generate [command options] [arguments...]

OPTIONS:
   --rootfs                                             path to the rootfs
   --read-only                                          make the container's rootfs read-only
   --privileged                                         enabled privileged container settings
   --hostname "acme"                                    hostname value for the container
   --uid "0"                                            uid for the process
   --gid "0"                                            gid for the process
   --groups [--groups option --groups option]           supplementary groups for the process
   --cap-add [--cap-add option --cap-add option]        add capabilities
   --cap-drop [--cap-drop option --cap-drop option]     drop capabilities
   --network                                            network namespace
   --mount                                              mount namespace
   --pid                                                pid namespace
   --ipc                                                ipc namespace
   --uts                                                uts namespace
   --selinux-label                                      process selinux label
   --tmpfs [--tmpfs option --tmpfs option]              mount tmpfs
   --args                                               command to run in the container
   --env [--env option --env option]                    add environment variable
   --mount-cgroups "ro"                                 mount cgroups (rw,ro,no)
   --bind [--bind option --bind option]                 bind mount directories src:dest:(rw,ro)
   --prestart [--prestart option --prestart option]     path to prestart hooks
   --poststop [--poststop option --poststop option]     path to poststop hooks
   --root-propagation                                   mount propagation for root
   --os "linux"                                         operating system the container is created for
   --arch "amd64"                                       architecture the container is created for
   --cwd "/"                                            current working directory for the process
   --uidmappings [--uidmappings option ]                add UIDMappings  e.g HostID:ContainerID:Size
   --gidmappings [--gidmappings option ]                add GIDMappings  e.g HostID:ContainerID:Size
   --apparmor                                           specify the the apparmor profile for the container
   --seccomp-default                                    specify the the defaultaction of Seccomp syscall restrictions
   --seccomp-arch [--seccomp-arch option ]              specify Additional architectures permitted to be used 
                                                         for system calls
   --seccomp-syscalls [--seccomp-syscalls option]       specify syscalls used in Seccomp
                                                        e.g Name:Action:Arg1_index/Arg1_value/Arg1_valuetwo/Arg1_op, 
                                                            Arg2_index/Arg2_value/Arg2_valuetwo/Arg2_op
```

Validating OCI bundle
------------------------------------------

```
# ocitools bvalidate --help
NAME:
   bvalidate - validate a OCI bundle

USAGE:
   command bvalidate [command options] [arguments...]

OPTIONS:
   --path       path to a bundle

```

Testing OCI runtimes
------------------------------------------

```
$ make
$ sudo make install
$ sudo ./test_runtime.sh -r runc
-----------------------------------------------------------------------------------
VALIDATING RUNTIME: runc
-----------------------------------------------------------------------------------
validating container process
validating capabilities
validating hostname
validating rlimits
validating sysctls
Runtime runc passed validation
```
