# Operations

A conformant runtime should provide an executable (called `funC` in the following examples).
That executable should support each operation listed below as its first argument.
It operates by default on the `config.json` and `runtime.json` in the current directory.
The template for supported commands is:

```sh
$ funC [global-options] <COMMAND> [command-specific-options] <command-specific-arguments>
```

## Global options

None are required, but the runtime may support options that start with at least one hyphen.
Global options may take positional arguments (e.g. `--log-level debug`), but the option parsing must be such that `funC <COMMAND>` is unambiguously an invocation of `<COMMAND>` for any `<COMMAND>` that does not start with a hyphen (including commands not specified in this document).

## Commands

### version

* *Options* None are required, but the runtime may support options.
* *Standard streams*
  * *stdin:* The runtime may not attempt to read from its stdin.
  * *stdout:* The runtime must print its name, a space, and its version as the first line to its stdout.
    The name may contain any Unicode characters except a control codes and newlines.
    The runtime may print additional lines its stdout, and the format for those lines is not specified in this document.
  * *stderr:* The runtime may print diagnostic messages to stderr, and the format for those lines is not specified in this document.
* *Exit code:* The runtime must exit with zero.

Print the runtime version and exit.
The version

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

* *Options*
  * *`--config <PATH>`* Override `config.json` with an alternative path.  The path may not support seeking (e.g. `/dev/fd/3`).
  * *`--runtime <PATH>`* Override `runtime.json` with an alternative path.  The path may not support seeking (e.g. `/dev/fd/3`).
* *Standard streams:* The runtime must attach its standard streams directly to the application process without inspection.
* *Exit code:* The runtime must exit with the application process's exit code.

Example:
```sh
# in a bundle directory with a process that echos "hello" and exits 42
$ funC start
hello
 
$ echo $?
42
```

### stop

 ...
