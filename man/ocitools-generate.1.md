% OCI(1) Oci User Manuals
% OCI Community
% APRIL 2016
# NAME
ocitools-generate - Generate a OCI spec file

# SYNOPSIS
**ocitools generate**
[**--arch**[=*[]*]
[**--apparmor**[=*[]*]]
[**--args**[=*[]*]]
[**--bind**[=*[]*]]
[**--cap-add**[=*[]*]]
[**--cap-drop**[=*[]*]]
[**--cwd**[=*[]*]]
[**--env**[=*[]*]]
[**--gid**[=*GID*]]
[**--gidmappings**[=*[]*]]]
[**--groups**[=*[]*]]
[**--hostname**[=*[]*]]
[**--help**]
[**--ipc**]
[**--network**]
[**--no-new-privileges**]
[**--mount**]
[**--mount-cgroups**]
[**--os**[=*[]*]]
[**--pid**]
[**--poststart**[=*[]*]]
[**--poststop**[=*[]*]]
[**--prestart**[=*[]*]]
[**--privileged**]
[**--read-only**]
[**--root-propagation**[=*[]*]]
[**--rootfs**[=*[]*]]
[**--seccomp-default**[=*[]*]]
[**--seccomp-arch**[=*[]*]]
[**--seccomp-syscalls**[=*[]*]]
[**--selinux-label**[=*[]*]]
[**--sysctl**[=*[]*]]
[**--tmpfs**[=*[]*]]
[**--uid**[=*[]*]]
[**--uidmappings**[=*[]*]]
[**--uts**]
[ARG...]

# DESCRIPTION

Generate a process in a new container. **ocitools generate** starts a process with its own
file system, its own networking, and its own isolated process tree. The IMAGE
which starts the process may define defaults related to the process that will be
generate in the container, the networking to expose, and more, but **ocitools generate**
gives final control to the operator or administrator who starts the container
from the image. For that reason **ocitools generate** has more options than any other
Oci command.

If the IMAGE is not already loaded then **ocitools generate** will pull the IMAGE, and
all image dependencies, from the repository in the same way generatening **ocitools
pull** IMAGE, before it starts the container from that image.

# OPTIONS
**--apparmor**="PROFILE"
   Specifies the the apparmor profile for the container

**--arch**="ARCH"
   Architecture used within the container.
   "amd64"

**--args**=OPTION
   Command to run in the container

**--bind**=[=*[[HOST-DIR:CONTAINER-DIR][:OPTIONS]]*] Bind mount
   directories src:dest:(rw,ro) If you specify, ` --bind
   /HOST-DIR:/CONTAINER-DIR`, runc bind mounts `/HOST-DIR` in the host
   to `/CONTAINER-DIR` in the oci container. The `OPTIONS` are a comma
   delimited list and can be: [rw|ro] The `HOST_DIR` and
   `CONTAINER-DIR` must be absolute paths such as `/src/docs`.  You
   can add `:ro` or `:rw` suffix to a volume to mount it read-only or
   read-write mode, respectively. By default, the volumes are mounted
   read-write.

**--cap-add**=[]
   Add Linux capabilities

**--cap-drop**=[]
   Drop Linux capabilities

**--cwd**=PATH
   Current working directory for the process

**-e**, **--env**=[]
   Set environment variables

   This option allows you to specify arbitrary
environment variables that are available for the process that will be launched
inside of the container.

**--hostname**=""
   Container host name

   Sets the container host name that is available inside the container.

**--help**
  Print usage statement

**--gid**=GID
  Gid for the process inside of container

**--groups**=GROUP
  Supplementary groups for the processes inside of container

**--gidmappings**=GIDMAPPINGS
  Add GIDMappings e.g HostID:ContainerID:Size for use with User Namespace

**--ipc**
  Use ipc namespace

**--network**
  Use network namespace

**--no-new-privileges**
  Set no new privileges bit for the container process.  Setting this flag
  will block the container processes from gaining any additonal privileges
  using tools like setuid apps.  It is a good idea to run unprivileged
  contaiers with this flag.

**--mount**
  Use a mount namespace

**--mount-cgroups**
  Mount cgroups (rw,ro,no)

**--os**=OS
  Operating used within the container

**--pid**
  Use a pid namespace

**--poststart**=CMD
  Path to command to run in poststart hooks. This command will be run before
  the container process gets launched but after the container environment and
  main process has been created.

**--poststop**=CMD
  Path to command to run in poststop hooks. This command will be run after the
  container completes but before the container process is destroyed

**--prestart**
  Path to command to run in prestart hooks. This command will be run before
  the container process gets launched but after the container environment.

**--privileged**=*true*|*false*
  Give extended privileges to this container. The default is *false*.

  By default, Oci containers are
“unprivileged” (=false) and cannot, for example, generate a Oci daemon inside the
Oci container. This is because by default a container is not allowed to
access any devices. A “privileged” container is given access to all devices.

  When the operator executes **ocitools generate --privileged**, Oci will enable access
to all devices on the host as well as set some configuration in AppArmor to
allow the container nearly all the same access to the host as processes generatening
outside of a container on the host.

**--read-only**=*true*|*false*
  Mount the container's root filesystem as read only.

  By default a container will have its root filesystem writable allowing processes
to write files anywhere.  By specifying the `--read-only` flag the container will have
its root filesystem mounted as read only prohibiting any writes.

**--root-propagation**=PROPOGATIONMODE
  Mount propagation for root system.
  Values are "SHARED, RSHARED, PRIVATE, RPRIVATE, SLAVE, RSLAVE"

**--rootfs**="*ROOTFSPATH*"
  Path to the rootfs

**--sysctl**=SYSCTLSETTING
  Add sysctl settings e.g net.ipv4.forward=1, only allowed if the syctl is
  namespaced.

**--seccomp-default**=ACTION
  Specifies the the defaultaction of Seccomp syscall restrictions
  Values: KILL,ERRNO,TRACE,ALLOW

**--seccomp-arch**=ARCH
  Specifies Additional architectures permitted to be used for system calls.
  By default if you turn on seccomp, only the host architecture will be allowed.

**--seccomp-syscalls**=SYSCALLS
  Specifies Additional syscalls permitted to be used for system calls,
  e.g Name:Action:Arg1_index/Arg1_value/Arg1_valuetwo/Arg1_op, Arg2_index/Arg2_value/Arg2_valuetwo/Arg2_op

**--selinux-label**=[=*SELINUXLABEL*]]
  SELinux Label
  Depending on your SELinux policy, you would specify a label that looks like
  this:
  "system_u:system_r:svirt_lxc_net_t:s0:c1,c2"

  Note you would want your ROOTFS directory to be labeled with a context that
  this process type can use.

  "system_u:object_r:usr_t:s0" might be a good label for a readonly container,
  "system_u:system_r:svirt_sandbox_file_t:s0:c1,c2" for a read/write container.

**--tmpfs**=[] Create a tmpfs mount

  Mount a temporary filesystem (`tmpfs`) mount into a container, for example:

  $ ocitools generate -d --tmpfs /tmp:rw,size=787448k,mode=1777 my_image

  This command mounts a `tmpfs` at `/tmp` within the container.  The supported mount
options are the same as the Linux default `mount` flags. If you do not specify
any options, the systems uses the following options:
`rw,noexec,nosuid,nodev,size=65536k`.

**--uid**=UID									  Sets the UID used within the container.

**--uidmappings**
  Add UIDMappings e.g HostUID:ContainerID:Size for use with User Namespace

**--uts**
  Use the uts namespace

# EXAMPLES

## Generatening container in read-only mode

During container image development, containers often need to write to the image
content.  Installing packages into /usr, for example.  In production,
applications seldom need to write to the image.  Container applications write
to volumes if they need to write to file systems at all.  Applications can be
made more secure by generatening them in read-only mode using the --read-only switch.
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

## Attaching to one or more from STDIN, STDOUT, STDERR

If you do not specify -a then Oci will attach everything (stdin,stdout,stderr)
. You can specify to which of the three standard streams (stdin, stdout, stderr)
you’d like to connect instead, as in:

    # ocitools generate -a stdin -a stdout --rootfs /var/lib/containers/fedora /bin/bash

## Sharing IPC between containers

Using shm_server.c available here: https://www.cs.cf.ac.uk/Dave/C/node27.html

Testing `--ipc=host` mode:

Host shows a shared memory segment with 7 pids attached, happens to be from httpd:

```
 $ sudo ipcs -m

 ------ Shared Memory Segments --------
 key        shmid      owner      perms      bytes      nattch     status
 0x01128e25 0          root       600        1000       7
```

Now generate a regular container, and it correctly does NOT see the shared memory segment from the host:

```
 $ ocitools generate -it shm ipcs -m

 ------ Shared Memory Segments --------
 key        shmid      owner      perms      bytes      nattch     status
```

Generate a container with the new `--ipc=host` option, and it now sees the shared memory segment from the host httpd:

 ```
 $ ocitools generate -it --ipc=host shm ipcs -m

 ------ Shared Memory Segments --------
 key        shmid      owner      perms      bytes      nattch     status
 0x01128e25 0          root       600        1000       7
```
Testing `--ipc=container:CONTAINERID` mode:

Start a container with a program to create a shared memory segment:
```
 $ ocitools generate -it shm bash
 $ sudo shm/shm_server &
 $ sudo ipcs -m

 ------ Shared Memory Segments --------
 key        shmid      owner      perms      bytes      nattch     status
 0x0000162e 0          root       666        27         1
```
Create a 2nd container correctly shows no shared memory segment from 1st container:
```
 $ ocitools generate shm ipcs -m

 ------ Shared Memory Segments --------
 key        shmid      owner      perms      bytes      nattch     status
```

Create a 3rd container using the new --ipc=container:CONTAINERID option, now it shows the shared memory segment from the first:

```
 $ ocitools generate -it --ipc=container:ed735b2264ac shm ipcs -m
 $ sudo ipcs -m

 ------ Shared Memory Segments --------
 key        shmid      owner      perms      bytes      nattch     status
 0x0000162e 0          root       666        27         1
```

## Mapping Ports for External Usage

The exposed port of an application can be mapped to a host port using the **-p**
flag. For example, a httpd port 80 can be mapped to the host port 8080 using the
following:

    # ocitools generate -p 8080:80  --rootfs /var/lib/containers/fedorahttpd

## Mounting External Volumes

To mount a host directory as a container volume, specify the absolute path to
the directory and the absolute path for the container directory separated by a
colon:

    # ocitools generate --bind /var/db:/data1  --rootfs /var/lib/containers/fedora --args bash

When using SELinux, be aware that the host has no knowledge of container SELinux
policy. Therefore, in the above example, if SELinux policy is enforced, the
`/var/db` directory is not writable to the container. A "Permission Denied"
message will occur and an avc: message in the host's syslog.


To work around this, at time of writing this man page, the following command
needs to be generate in order for the proper SELinux policy type label to be attached
to the host directory:

    # chcon -Rt svirt_sandbox_file_t /var/db

Now, writing to the /data1 volume in the container will be allowed and the
changes will also be reflected on the host in /var/db.

# HISTORY
April 2016, Originally compiled by Dan Walsh (dwalsh at redhat dot com)
