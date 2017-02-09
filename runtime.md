# OCI Runtime Command Line Interface

This section defines the OCI Runtime Command Line Interface version 1.0.0.

## Versioning

The command line interface is versioned with [SemVer v2.0.0][semver].
The command line interface version is independent of the OCI Runtime Specification as a whole (which is tied to the [configuration format][runtime-spec-version].
For example, if a caller is compliant with version 1.1 of the command line interface, they are compatible with all runtimes that support any 1.1 or later release of the command line interface, but are not compatible with a runtime that supports 1.0 and not 1.1.

## Global usage

The runtime MUST provide an executable (called `funC` in the following examples).
That executable MUST support commands with the following template:

```
$ funC [global-options] <COMMAND> [command-specific-options] <command-specific-arguments>
```

## Global options

None are required, but the runtime MAY support options that start with at least one hyphen.
Global options MAY take positional arguments (e.g. `--log-level debug`).
Command names MUST NOT start with hyphens.
The option parsing MUST be such that `funC <COMMAND>` is unambiguously an invocation of `<COMMAND>` (even for commands not specified in this document).
If the runtime is invoked with an unrecognized command, it MUST exit with a nonzero exit code and MAY log a warning to stderr.
Beyond the above rules, the behavior of the runtime in the presence of commands and options not specified in this document is unspecified.

## Character encodings

This API specification does not cover character encodings, but runtimes SHOULD conform to their native operating system.
For example, POSIX systems define [`LANG` and related environment variables][posix-lang] for [declaring][posix-locale-encoding] [locale-specific character encodings][posix-encoding], so a runtime in an `en_US.UTF-8` locale SHOULD write its [state](#state) to stdout in [UTF-8][].

## Commands

### create

[Create][create] a container from a [bundle directory][bundle].

* *Arguments*
    * *`<ID>`* Set the container ID to create.
* *Options*
    * *`--bundle <PATH>`* Override the path to the [bundle directory][bundle] (defaults to the current working directory).
* *Standard streams:*
    * *stdin:* The runtime MUST NOT attempt to read from its stdin.
    * *stdout:* The handling of stdout is unspecified.
    * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Environment variables*
    * *`LISTEN_FDS`:* The number of file descriptors passed.
      For example, `LISTEN_FDS=2` would mean that the runtime MUST pass file descriptors 3 and 4 to the container process (in addition to the [standard streams][standard-streams]) to support [socket activation][systemd-listen-fds].
* *Exit code:* Zero if the container was successfully created and non-zero on errors.

Callers MAY block on this command's successful exit to trigger post-create activity.

#### Example

```
# in a bundle directory with a process that echos "hello" and exits 42
$ test -t 1 && echo 'stdout is a terminal'
stdout is a terminal
$ funC create hello-1 <&- >stdout 2>stderr
$ echo $?
0
$ wc stdout
0 0 0 stdout
$ funC start hello-1
$ echo $?
0
$ cat stdout
hello
$ block-on-exit-and-collect-exit-code hello-1
$ echo $?
42
$ funC delete hello-1
$ echo $?
0
```

#### Container process exit

The [example's](#example) `block-on-exit-and-collect-exit-code` is platform-specific and is not specified in this document.
On Linux, it might involve an ancestor process which had set [`PR_SET_CHILD_SUBREAPER`][prctl.2] and collected the container PID [from the state][state], or a process that was [ptracing][ptrace.2] the container process for [`exit_group`][exit_group.2], although both of those race against the container process exiting before the watcher is monitoring.

### start

[Start][start] the user-specified code from [`process`][process].

* *Arguments*
    * *`<ID>`* The container to start.
* *Standard streams:*
    * *stdin:* The runtime MUST NOT attempt to read from its stdin.
    * *stdout:* The handling of stdout is unspecified.
    * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* Zero if the container was successfully started and non-zero on errors.

Callers MAY block on this command's successful exit to trigger post-start activity.

See [create](#example) for an example.

### state

[Request][state-request] the container [state][state].

* *Arguments*
    * *`<ID>`* The container whose state is being requested.
* *Standard streams:*
    * *stdin:* The runtime MUST NOT attempt to read from its stdin.
    * *stdout:* The runtime MUST print the [state JSON][state] to its stdout.
    * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* Zero if the state was successfully written to stdout and non-zero on errors.

#### Example

```
$ funC create sleeper-1
$ funC state sleeper-1
{
  "ociVersion": "1.0.0-rc1",
  "id": "sleeper-1",
  "status": "created",
  "pid": 4422,
  "bundlePath": "/containers/sleeper",
  "annotations" {
    "myKey": "myValue"
  }
}
$ echo $?
0
```

### kill

[Send a signal][kill] to the container process.

* *Arguments*
    * *`<ID>`* The container being signaled.
* *Options*
    * *`--signal <SIGNAL>`* The signal to send (defaults to `TERM`).
      The runtime MUST support `TERM` and `KILL` signals with [the POSIX semantics][posix-signals].
      The runtime MAY support additional signal names.
      On platforms that support [POSIX signals][posix-signals], the runtime MUST implement this command using POSIX signals.
      On platforms that do not support POSIX signals, the runtime MAY implement this command with alternative technology as long as `TERM` and `KILL` retain their POSIX semantics.
      Runtime authors on non-POSIX platforms SHOULD submit documentation for their TERM implementation to this specificiation, so runtime callers can configure the container process to gracefully handle the signals.
* *Standard streams:*
    * *stdin:* The runtime MUST NOT attempt to read from its stdin.
    * *stdout:* The handling of stdout is unspecified.
    * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* Zero if the signal was successfully sent to the container process and non-zero on errors.
  Successfully sent does not mean that the signal was successfully received or handled by the container process.

#### Example

```
# in a bundle directory where the container process ignores TERM
$ funC create sleeper-1
$ funC start sleeper-1
$ funC kill sleeper-1
$ echo $?
0
$ funC kill --signal KILL sleeper-1
$ echo $?
0
```

### delete

[Release](#delete) container resources after the container process has exited.

* *Arguments*
    * *`<ID>`* The container to delete.
* *Standard streams:*
    * *stdin:* The runtime MUST NOT attempt to read from its stdin.
    * *stdout:* The handling of stdout is unspecified.
    * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* Zero if the container was successfully deleted and non-zero on errors.

See [create](#example) for an example.

[bundle]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/bundle.md
[create]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/runtime.md#create
[delete]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/runtime.md#delete
[exit_group.2]: http://man7.org/linux/man-pages/man2/exit_group.2.html
[kill]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/runtime.md#kill
[kill.2]: http://man7.org/linux/man-pages/man2/kill.2.html
[process]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/config.md#process
[posix-encoding]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap06.html#tag_06_02
[posix-lang]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap08.html#tag_08_02
[posix-locale-encoding]: http://www.unicode.org/reports/tr35/#Bundle_vs_Item_Lookup
[posix-signals]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/signal.h.html#tag_13_42_03
[prctl.2]: http://man7.org/linux/man-pages/man2/prctl.2.html
[ptrace.2]: http://man7.org/linux/man-pages/man2/ptrace.2.html
[semver]: http://semver.org/spec/v2.0.0.html
[standard-streams]: https://github.com/opencontainers/specs/blob/v0.1.1/runtime-linux.md#file-descriptors
[start]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/runtime.md#start
[state]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/runtime.md#state
[state-request]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/runtime.md#query-state
[systemd-listen-fds]: http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
[runtime-spec-version]: https://github.com/opencontainers/runtime-spec/blob/v1.0.0-rc4/config.md#specification-version
[UTF-8]: http://www.unicode.org/versions/Unicode8.0.0/ch03.pdf
