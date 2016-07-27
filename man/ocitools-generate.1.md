% OCI(1) OCITOOLS User Manuals
% OCI Community
% APRIL 2016
# NAME
ocitools-generate - Generate a config.json for an OCI container

# SYNOPSIS
**ocitools generate** *[OPTIONS]*

# DESCRIPTION

`ocitools generate` generates configuration JSON for an OCI bundle.
By default, it writes the JSON to stdout, but you can use **--output**
to direct it to a file.  OCI-compatible runtimes like runC expect to
read the configuration from `config.json`.

# OPTIONS
**--apparmor**=PROFILE
  Specifies the apparmor profile for the container

**--arch**=ARCH
  Architecture used within the container.
  "amd64"

**--args**=OPTION
  Arguments to run within the container.  Can be specified multiple times.
  If you were going to run a command with multiple options, you would need
  to specify the command and each argument in order.

  --args "/usr/bin/httpd" --args "-D" --args "FOREGROUND"

**--bind**=*[[HOST-DIR:CONTAINER-DIR][:OPTIONS]]*
  Bind mount directories src:dest:(rw,ro) If you specify, ` --bind
  /HOST-DIR:/CONTAINER-DIR`, runc bind mounts `/HOST-DIR` in the host
  to `/CONTAINER-DIR` in the OCI container. The `OPTIONS` are a comma
  delimited list and can be: [rw|ro] The `HOST_DIR` and
  `CONTAINER-DIR` must be absolute paths such as `/src/docs`.  You
  can add `:ro` or `:rw` suffix to a volume to mount it read-only or
  read-write mode, respectively. By default, the volumes are mounted
  read-write.

**--cap-add**=[]
  Add Linux capabilities

**--cap-drop**=[]
  Drop Linux capabilities

**--cgroup**=[*PATH*]
  Use a Cgroup namespace.  If *PATH* is set, join that namespace.  If it
  is unset, create a new namespace.  The special *PATH* `host` removes
  any existing Cgroup namespace from the configuration.

**--cgroups-path**=""
  Specifies the path to the cgroups relative to the cgroups mount point.

**--cwd**=PATH
  Current working directory for the process

**--env**=[]
  Set environment variables
  This option allows you to specify arbitrary
environment variables that are available for the process that will be launched
inside of the container.

**--gid**=GID
  Gid for the process inside of container

**--gidmappings**=GIDMAPPINGS
  Add GIDMappings e.g HostID:ContainerID:Size.  Implies **-user=**.

**--groups**=GROUP
  Supplementary groups for the processes inside of container

**--help**
  Print usage statement

**--hostname**=""
  Set the container host name that is available inside the container.

**--ipc**=[*PATH*]
  Use an IPC namespace.  If *PATH* is set, join that namespace.  If it
  is unset, create a new namespace.  The special *PATH* `host` removes
  any existing IPC namespace from the configuration.

**--mount**=[*PATH*]
  Use a mount namespace.  If *PATH* is set, join that namespace.  If
  it is unset, create a new namespace.  The special *PATH* `host`
  removes any existing mount namespace from the configuration.

**--mount-cgroups**=[rw|ro|no]
  Mount cgroups.  The default is `no`.

**--mount-label**=MOUNTLABEL
  Mount Label
  Depending on your SELinux policy, you would specify a label that looks like
  this:
  "system_u:object_r:svirt_sandbox_file_t:s0:c1,c2"

    Note you would want your ROOTFS directory to be labeled with a context that
    this process type can use.

      "system_u:object_r:usr_t:s0" might be a good label for a readonly container,
      "system_u:system_r:svirt_sandbox_file_t:s0:c1,c2" for a read/write container.

**--network**=[*PATH*]
  Use a network namespace.  If *PATH* is set, join that namespace.  If
  it is unset, create a new namespace.  The special *PATH* `host`
  removes any existing network namespace from the configuration.

**--no-new-privileges**
  Set no new privileges bit for the container process.  Setting this flag
  will block the container processes from gaining any additional privileges
  using tools like setuid apps.  It is a good idea to run unprivileged
  containers with this flag.

**--output**=PATH
  Instead of writing the configuration JSON to stdout, write it to a
  file at *PATH* (overwriting the existing content if a file already
  exists at *PATH*).

**--os**=OS
  Operating system used within the container

**--pid**=[*PATH*]
  Use a PID namespace.  If *PATH* is set, join that namespace.  If it
  is unset, create a new namespace.  The special *PATH* `host` removes
  any existing PID namespace from the configuration.

**--poststart**=CMD
  Path to command to run in poststart hooks. This command will be run before
  the container process gets launched but after the container environment and
  main process has been created.

**--poststop**=CMD
  Path to command to run in poststop hooks. The command will be run
  after the container process is stopped.

**--prestart**=CMD
  Path to command to run in prestart hooks. This command will be run before
  the container process gets launched but after the container environment.

**--privileged**=true|false
  Give extended privileges to this container. The default is *false*.

  By default, OCI containers are
“unprivileged” (=false) and cannot do some of the things a normal root process can do.

  When the operator executes **ocitools generate --privileged**, OCI will enable access to all devices on the host as well as disable some of the confinement mechanisms like AppArmor, SELinux, and seccomp from blocking access to privileged processes.  This gives the container processes nearly all the same access to the host as processes generating outside of a container on the host.

**--read-only**=true|false
  Mount the container's root filesystem as read only.

  By default a container will have its root filesystem writable allowing processes to write files anywhere.  By specifying the `--read-only` flag the container will have its root filesystem mounted as read only prohibiting any writes.

**--root-propagation**=PROPOGATIONMODE
  Mount propagation for root filesystem.
  Values are "shared, rshared, private, rprivate, slave, rslave"

**--rootfs**=ROOTFSPATH
  Path to the rootfs

**--seccomp-only**
  Specifies that only the seccomp configuration should be exported to a file
  named config.seccomp.

**--seccomp-arch**=ARCH
  Specifies Additional architectures permitted to be used for system calls.
  By default if you turn on seccomp, only the host architecture will be allowed.

**--seccomp-default**=ACTION
  Specifies the the defaultaction of Seccomp syscall restrictions
  Values: kill,errno,trace,trap,allow

**--seccomp-allow**=SYSCALL,SYSCALL:INDEX:ARG1:ARG2:OP,...
  Specifies syscalls to be added to the ALLOW list.
  You can specify just the name of the syscall or you can specify arguments by
  using a `:` seperated list. You can specify as many as you want by using ','
  e.g Syscall:index:arg1:arg2:Op,Syscall,Syscall,...

**--seccomp-errno**=SYSCALL
  Specifies syscalls to be added to the ERRNO list.
  You can specify just the name of the syscall or you can specify arguments by
  using a `:` seperated list. You can specify as many as you want by using ','
  e.g Syscall:index:arg1:arg2:Op,Syscall,Syscall,...

**--seccomp-trace**=SYSCALL
  Specifies syscalls to be added to the TRACE list.
  You can specify just the name of the syscall or you can specify arguments by
  using a `:` seperated list. You can specify as many as you want by using ','
  e.g Syscall:index:arg1:arg2:Op,Syscall,Syscall,...

**--seccomp-trap**=SYSCALL
  Specifies syscalls to be added to the TRAP list.
  You can specify just the name of the syscall or you can specify arguments by
  using a `:` seperated list. You can specify as many as you want by using ','
  e.g Syscall:index:arg1:arg2:Op,Syscall,Syscall,...

**--seccomp-kill**=SYSCALL
  Specifies syscalls to be added to the KILL list.
  You can specify just the name of the syscall or you can specify arguments by
  using a `:` seperated list. You can specify as many as you want by using ','
  e.g Syscall:index:arg1:arg2:Op,Syscall,Syscall,...

**--selinux-label**=PROCESSLABEL
  SELinux Label
  Depending on your SELinux policy, you would specify a label that looks like
  this:
  "system_u:system_r:svirt_lxc_net_t:s0:c1,c2"

    Note you would want your ROOTFS directory to be labeled with a context that
    this process type can use.

      "system_u:object_r:usr_t:s0" might be a good label for a readonly container,
      "system_u:object_r:svirt_sandbox_file_t:s0:c1,c2" for a read/write container.

**--sysctl**=SYSCTLSETTING
  Add sysctl settings e.g net.ipv4.forward=1, only allowed if the syctl is
  namespaced.

**--template**=PATH
  Override the default template with your own.
  Additional options will only adjust the relevant portions of your template.

**--tmpfs**=[] Create a tmpfs mount
  Mount a temporary filesystem (`tmpfs`) mount into a container, for example:

    $ ocitools generate -d --tmpfs /tmp:rw,size=787448k,mode=1777 my_image

    This command mounts a `tmpfs` at `/tmp` within the container.  The supported mount options are the same as the Linux default `mount` flags. If you do not specify any options, the systems uses the following options:
    `rw,noexec,nosuid,nodev,size=65536k`.

**--uid**=UID
  Sets the UID used within the container.

**--uidmappings**
  Add UIDMappings e.g HostUID:ContainerID:Size.  Implies **--user=**.

**--user**=[*PATH*]
  Use a user namespace.  If *PATH* is set, join that namespace.  If it
  is unset, create a new namespace.  The special *PATH* `host` removes
  any existing user namespace from the configuration.

**--uts**=[*PATH*]
  Use a UTS namespace.  If *PATH* is set, join that namespace.  If it
  is unset, create a new namespace.  The special *PATH* `host` removes
  any existing UTS namespace from the configuration.

# EXAMPLES

## Generating container in read-only mode

During container image development, containers often need to write to the image
content.  Installing packages into /usr, for example.  In production,
applications seldom need to write to the image.  Container applications write
to volumes if they need to write to file systems at all.  Applications can be
made more secure by generating them in read-only mode using the --read-only switch.
This protects the containers image from modification. Read only containers may
still need to write temporary data.  The best way to handle this is to mount
tmpfs directories on /generate and /tmp.

    # ocitools generate --read-only --tmpfs /generate --tmpfs /tmp --tmpfs /run  --rootfs /var/lib/containers/fedora /bin/bash

## Exposing log messages from the container to the host's log

If you want messages that are logged in your container to show up in the host's
syslog/journal then you should bind mount the /dev/log directory as follows.

    # ocitools generate --bind /dev/log:/dev/log  --rootfs /var/lib/containers/fedora /bin/bash

From inside the container you can test this by sending a message to the log.

    (bash)# logger "Hello from my container"

Then exit and check the journal.

    # exit

    # journalctl -b | grep Hello

This should list the message sent to logger.

## Bind Mounting External Volumes

To mount a host directory as a container volume, specify the absolute path to
the directory and the absolute path for the container directory separated by a
colon:

    # ocitools generate --bind /var/db:/data1  --rootfs /var/lib/containers/fedora --args bash

## Using SELinux

You can use SELinux to add security to the container.  You must specify the process label to run the init process inside of the container using the --selinux-label.

    # ocitools generate --bind /var/db:/data1  --selinux-label system_u:system_r:svirt_lxc_net_t:s0:c1,c2 --mount-label system_u:object_r:svirt_sandbox_file_t:s0:c1,c2 --rootfs /var/lib/containers/fedora --args bash

Not in the above example we used a type of svirt_lxc_net_t and an MCS Label of s0:c1,c2.  If you want to guarantee separation between containers, you need to make sure that each container gets launched with a different MCS Label pair.

Also the underlying rootfs must be labeled with a matching label.  For the example above, you would execute a command like:

    # chcon -R system_u:object_r:svirt_sandbox_file_t:s0:c1,c2  /var/lib/containers/fedora

This will set up the labeling of the rootfs so that the process launched would be able to write to the container.  If you wanted to only allow it to read/execute the content in rootfs, you could execute:

    # chcon -R system_u:object_r:usr_t:s0  /var/lib/containers/fedora

When using SELinux, be aware that the host has no knowledge of container SELinux
policy. Therefore, in the above example, if SELinux policy is enforced, the
`/var/db` directory is not writable to the container. A "Permission Denied"
message will occur and an avc: message in the host's syslog.

To work around this, the following command needs to be generate in order for the proper SELinux policy type label to be attached to the host directory:

    # chcon -Rt svirt_sandbox_file_t -l s0:c1,c2 /var/db

Now, writing to the /data1 volume in the container will be allowed and the
changes will also be reflected on the host in /var/db.

# SEE ALSO
**runc**(1), **ocitools**(1)

# HISTORY
April 2016, Originally compiled by Dan Walsh (dwalsh at redhat dot com)
