# Operations

A conformant runtime should provide an executable (called `funC` in the following examples).
That executable should support each operation listed below as its first argument.
It operates by default on the 'config.json' in the current directory.

## Start

Starts a container from a bundle directory. 

* *Flags:* none.
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

## Stop

 ...
