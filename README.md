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

Testing OCI runtimes
------------------------------------------

```
make
sudo make install
./test_runtime.sh -r runc 
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

Building `rootfs.tar.gz`
------------------------

The root filesystem tarball is based on [Gentoo][]'s [amd64
stage3][stage3-amd64] (which we check for a valid GnuPG
signature][gentoo-signatures]), copying a [minimal
subset](rootfs-files) to the root filesytem, and adding symlinks for
all BusyBox commands.  To rebuild the tarball based on a newer stage3,
just run:

```
$ touch get-stage3.sh
$ make rootfs.tar.gz
```

### Getting Gentoo's Release Engineering public key

If `make rootfs.tar.gz` gives an error like:

```
gpg --verify downloads/stage3-amd64-current.tar.bz2.DIGESTS.asc
gpg: Signature made Thu 14 Jan 2016 09:00:11 PM EST using RSA key ID 2D182910
gpg: Can't check signature: public key not found
```

you will need to [add the missing public key to your
keystore][gentoo-signatures].  One way to do that is by [asking a
keyserver][recv-keys]:

```
$ gpg --keyserver pool.sks-keyservers.net --recv-keys 2D182910
```

[Gentoo]: https://www.gentoo.org/
[stage3-amd64]: http://distfiles.gentoo.org/releases/amd64/autobuilds/
[gentoo-signatures]: https://www.gentoo.org/downloads/signatures/
[recv-keys]: https://www.gnupg.org/documentation/manuals/gnupg/Operational-GPG-Commands.html
