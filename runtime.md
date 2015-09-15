# Operations: 

The command line should support each operation listed below as its first argument.
It operates by default on the 'config.json' in the current directory.

## Start

Starts a container from a bundle directory. 

* *Flags:* none.
* *Output:* The process output is printed to stdout and stderr, and the process exits with the delegate process exit code.

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
