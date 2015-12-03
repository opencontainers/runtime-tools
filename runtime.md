# Operations

A conformant runtime should provide an executable (called `funC` in the following examples).
The template for supported commands is:

```sh
$ funC [global-options] <COMMAND> [command-specific-options] <command-specific-arguments>
```

## Global options

None are required, but the runtime may support options that start with at least one hyphen.
Global options may take positional arguments (e.g. `--log-level debug`), but the option parsing must be such that `funC <COMMAND>` is unambiguously an invocation of `<COMMAND>` for any `<COMMAND>` that does not start with a hyphen (including commands not specified in this document).

## Character encodings

This API specification does not cover character encodings, but runtimes should conform to their native operating system.
For example, POSIX systems define [`LANG` and related environment variables][posix-lang] for [declaring][posix-locale-encoding] [locale-specific character encodings][posix-encoding], so a runtime in an `en_US.UTF-8` locale should write its [version](#version) to stdout in [UTF-8][].

## Commands

### version

Print the runtime version and exit.

* *Options* None are required, but the runtime may support options.
* *Standard streams*
  * *stdin:* The runtime may not attempt to read from its stdin.
  * *stdout:* The runtime must print its name, a space, and its version as the first line to its stdout.
    The name may contain any Unicode characters except a control codes and newlines.
    The runtime may print additional lines its stdout, and the format for those lines is not specified in this document.
  * *stderr:* The runtime may print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* The runtime must exit with zero.

Example:
```sh
$ funC version
funC 1.0.0
Built for x86_64-pc-linux-gnu
$ echo $?
0
```

### start

Starts a container from a bundle directory. 
It operates by default on the `config.json` and `runtime.json` in the current directory.

* *Options*
  * *`--id <ID>`* Set the container ID when creating or joining a container.
    If not set, the runtime is free to pick any ID that is not already in use.
  * *`--config <PATH>`* Override `config.json` with an alternative path.  The path may not support seeking (e.g. `/dev/fd/3`).
  * *`--runtime <PATH>`* Override `runtime.json` with an alternative path.  The path may not support seeking (e.g. `/dev/fd/3`).
* *Standard streams:* The runtime must attach its standard streams directly to the application process without inspection.
* *Environment variables*
  * *`LISTEN_FDS`:* The number of file descriptors passed.
    For example, `LISTEN_FDS=2` would mean passing 3 and 4 (in addition to the [standard streams][standard-streams]) to support [socket activation][systemd-listen-fds].
* *Exit code:* The runtime must exit with the application process's exit code.

Example:
```sh
# in a bundle directory with a process that echos "hello" and exits 42
$ funC start --id hello-1
hello
 
$ echo $?
42
```

### exec

Runs a secondary process in the given container.

* *Options*
  * *`--process <PATH>`* Override `process.json` with an alternative path.  The path may not support seeking (e.g. `/dev/fd/3`).
* *Arguments*
  * *`<ID>`* The container ID to join.
* *Standard streams:* The runtime must attach its standard streams directly to the application process without inspection.
* *Exit code:* The runtime must exit with the application process's exit code.

If the main application (launched by `start`) dies, all other processes in its container will be killed [TODO: link to lifecycle docs explaining this].

Example:
```sh
# in a directory with a process.json that echos "goodbye" and exits 43
$ funC exec hello-1
goodbye
$ echo $?
43
```

### pause

Pause all processes in a container.

* *Options*
  * *`--wait`* Block until the process is completely paused.
  Otherwise return immediately after initiating the pause, which may happen before the pause is complete.
* *Arguments*
  * *`<ID>`* The container ID to join.
* *Exit code:* 0 on success, non-zero on error.

Example:
```sh
$ funC pause --wait hello-1
$ echo $?
0
```

### resume

Unpause all processes in a container.

* *Options*
  * *`--wait`* Block until the process is completely unpaused.
  Otherwise return immediately after initiating the unpause, which may happen before the unpause is complete.
* *Arguments*
  * *`<ID>`* The container ID to join.
* *Exit code:* 0 on success, non-zero on error.

Example:
```sh
$ funC resume hello-1
$ echo $?
0
```

### signal

Sends a signal to the container.

* *Options*
  * *`--signal <SIGNAL>`* The signal to send.
    This must be one of the valid POSIX signals, although runtimes on non-POSIX systems must translate the POSIX name to their platorm's analogous signal.
    Defaults to TERM.
* *Arguments*
  * *`<ID>`* The container ID to join.
* *Exit code:* 0 on success, non-zero on error.
  A 0 exit status does not imply the process has exited (as it may have caught the signal).

Example:
```sh
$ funC signal --signal KILL hello-1
$ echo $?
0
```

[posix-encoding]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap06.html#tag_06_02
[posix-lang]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap08.html#tag_08_02
[posix-locale-encoding: http://www.unicode.org/reports/tr35/#Bundle_vs_Item_Lookup
[standard-streams]: https://github.com/opencontainers/specs/blob/v0.1.1/runtime-linux.md#file-descriptors
[systemd-listen-fds]: http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
[UTF-8]: http://www.unicode.org/versions/Unicode8.0.0/ch03.pdf
