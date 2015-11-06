# ocitools

This repo has a collection of utilities for OCI

```
NAME:
   oci - Utilities for OCI

USAGE:
   oci [global options] command [command options] [arguments...]
   
VERSION:
   0.0.1
   
COMMANDS:
   generate     generate a OCI spec file
   help, h      Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --help, -h           show help
   --version, -v        print the version

```

Generate

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
```
