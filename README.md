# ocitools

ocitools is a collection of tools for working with the [OCI runtime specification][runtime-spec].

## Generating an OCI runtime spec configuration files

[`ocitools generate`][generate.1] is used to generate a `config.json` ([OCI spec][runtime-spec] file) to be used to instantiate an OCI container.
This `config.json` file can be placed into a directory and used by an OCI compatable runtime like [**runc**][runC] to run a container.

```sh
$ ocitools generate
$ cat config.json
{
        "ociVersion": "0.5.0",
        …
}
```

## Validating an OCI bundle

[`ocitools validate`][validate.1] validates an OCI bundle.

```sh
$ ocitools generate
$ ocitools validate
FATA[0000] Bundle path shouldn't be empty
```

## Testing OCI runtimes

```sh
$ make
$ sudo make install
$ sudo ./test_runtime.sh -r runc
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

[runC]: https://github.com/opencontainers/runc
[runtime-spec]: https://github.com/opencontainers/runtime-spec

[generate.1]: man/ocitools-generate.1.md
[validate.1]: man/ocitools-validate.1.md
