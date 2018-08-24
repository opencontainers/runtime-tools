# oci-runtime-tool [![Build Status](https://travis-ci.org/opencontainers/runtime-tools.svg?branch=master)](https://travis-ci.org/opencontainers/runtime-tools) [![Go Report Card](https://goreportcard.com/badge/github.com/opencontainers/runtime-tools)](https://goreportcard.com/report/github.com/opencontainers/runtime-tools)

oci-runtime-tool is a collection of tools for working with the [OCI runtime specification][runtime-spec].
To build from source code, runtime-tools requires Go 1.7.x or above.

## Generating an OCI runtime spec configuration files

[`oci-runtime-tool generate`][generate.1] generates [configuration JSON][config.json] for an [OCI bundle][bundle].
[OCI-compatible runtimes][runtime-spec] like [runC][] expect to read the configuration from `config.json`.

```console
$ oci-runtime-tool generate --output config.json
$ cat config.json
{
        "ociVersion": "0.5.0",
        …
}
```

## Validating an OCI bundle

[`oci-runtime-tool validate`][validate.1] validates an OCI bundle.
The error message will be printed if the OCI bundle failed the validation procedure.

```console
$ oci-runtime-tool generate
$ oci-runtime-tool validate
INFO[0000] Bundle validation succeeded.
```

## Testing OCI runtimes

Use the [runtime validation suite](docs/runtime-compliance-testing.md).

[bundle]: https://github.com/opencontainers/runtime-spec/blob/master/bundle.md
[config.json]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[runC]: https://github.com/opencontainers/runc
[runtime-spec]: https://github.com/opencontainers/runtime-spec

[generate.1]: man/oci-runtime-tool-generate.1.md
[validate.1]: man/oci-runtime-tool-validate.1.md
