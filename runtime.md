# Operations

A conformant runtime MUST provide an executable (called `funC` in the following examples).
That executable MUST support commands with the following template:

```sh
$ funC [global-options] <COMMAND> [command-specific-options] <command-specific-arguments>
```

## Global options

None are required, but the runtime MAY support options that start with at least one hyphen.
Global options MAY take positional arguments (e.g. `--log-level debug`).
Command names MUST not start with hyphens.
The option parsing MUST be such that `funC <COMMAND>` is unambiguously an invocation of `<COMMAND>` (even for commands not specified in this document).
If the runtime is invoked with an unrecognized command, it MUST exit with a nonzero exit code and MAY log a warning to stderr.

## Character encodings

This API specification does not cover character encodings, but runtimes should conform to their native operating system.
For example, POSIX systems define [`LANG` and related environment variables][posix-lang] for [declaring][posix-locale-encoding] [locale-specific character encodings][posix-encoding], so a runtime in an `en_US.UTF-8` locale should write its [version](#version) to stdout in [UTF-8][].

## Commands

### version

Print the runtime version and exit.

* *Options* None are required, but the runtime MAY support options.
* *Standard streams*
  * *stdin:* The runtime MUST NOT attempt to read from its stdin.
  * *stdout:* The runtime MUST print its name, a space, and its version as the first line to its stdout.
    The name MAY contain any Unicode characters, but MUST NOT contain control codes or newlines.
    The runtime MAY print additional lines to its stdout, and the format for those lines is not specified in this document.
  * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* The runtime MUST exit with zero.

#### Example

```
$ funC version
funC 1.0.0
Built for x86_64-pc-linux-gnu
$ echo $?
0
```

### start

Start a container from a bundle directory.

* *Options*
  * *`--id <ID>`* Set the container ID when creating or joining a container.
    If not set, the runtime is free to pick any ID that is not already in use.
  * *`--bundle <PATH>`* Override the path to the bundle directory (defaults to the current working directory).
* *Standard streams:* The runtime MUST attach its standard streams directly to the application process without inspection.
* *Environment variables*
  * *`LISTEN_FDS`:* The number of file descriptors passed.
    For example, `LISTEN_FDS=2` would mean that the runtime MUST pass file descriptors 3 and 4 to the application process (in addition to the [standard streams][standard-streams]) to support [socket activation][systemd-listen-fds].
* *Exit code:* The runtime MUST exit with the application process's exit code.

#### Example

```
# in a bundle directory with a process that echos "hello" and exits 42
$ funC start --id hello-1
hello

$ echo $?
42
```

### state

Request the container state.

* *Arguments*
  * *`<ID>`* The container whose state is being requested.
* *Standard streams:*
  * *stdin:* The runtime MUST NOT attempt to read from its stdin.
  * *stdout:* The runtime MUST print the state JSON to its stdout.
  * *stderr:* The runtime MAY print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* Zero if the state was successfully written to stdout and non-zero on errors.

#### Example

```
# in a bundle directory with a process that sleeps for several seconds
$ funC start --id sleeper-1 &
$ funC state sleeper-1
{
  "ociVersion": "1.0.0-rc1",
  "id": "sleeper-1",
  "status": "running",
  "pid": 4422,
  "bundlePath": "/containers/sleeper",
  "annotations" {
    "myKey": "myValue"
  }
}
$ echo $?
0
```

[posix-encoding]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap06.html#tag_06_02
[posix-lang]: http://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap08.html#tag_08_02
[posix-locale-encoding]: http://www.unicode.org/reports/tr35/#Bundle_vs_Item_Lookup
[standard-streams]: https://github.com/opencontainers/specs/blob/v0.1.1/runtime-linux.md#file-descriptors
[systemd-listen-fds]: http://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
[UTF-8]: http://www.unicode.org/versions/Unicode8.0.0/ch03.pdf
