# ocitools

ocitools is a collection of tools for working with the [OCI runtime specification][runtime-spec].

## Generating an OCI runtime spec configuration files

[`ocitools generate`][generate.1] generates [configuration JSON][config.json] for an [OCI bundle][bundle].
[OCI-compatible runtimes][runtime-spec] like [runC][] expect to read the configuration from `config.json`.

```sh
$ ocitools generate --output config.json
$ cat config.json
{
        "ociVersion": "0.5.0",
        …
}
```

## Validating an OCI bundle

[`ocitools validate`][validate.1] validates an OCI bundle.
The error message will be printed if the OCI bundle failed the validation procedure.

```sh
$ ocitools generate
$ ocitools validate
INFO[0000] Bundle validation succeeded.
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

[bundle]: https://github.com/opencontainers/runtime-spec/blob/master/bundle.md
[config.json]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[runC]: https://github.com/opencontainers/runc
[runtime-spec]: https://github.com/opencontainers/runtime-spec

[generate.1]: man/ocitools-generate.1.md
[validate.1]: man/ocitools-validate.1.md
