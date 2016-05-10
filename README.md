# ocitools

ocitools is a collection of tools for working with the [OCI runtime specification][runtime-spec].

## Generating an OCI runtime spec configuration files

[`ocitools generate`][generate.1] generates a [`config.json`][config.json] for an [OCI bundle][bundle].
This `config.json` file can be placed into a directory and used by an [OCI compatible runtime][runtime-spec] like [runC][] to run a container.

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

# ocitools runtimetest -r runc
INFO[0000] Start to test runtime lifecircle...
INFO[0001] Runtime lifecircle test succeeded.
INFO[0001] Start to test runtime state...
INFO[0006] Runtime state test succeeded.
INFO[0006] Start to test runtime main config...
INFO[0006] validating container process
validating capabilities
```

[bundle]: https://github.com/opencontainers/runtime-spec/blob/master/bundle.md
[config.json]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[runC]: https://github.com/opencontainers/runc
[runtime-spec]: https://github.com/opencontainers/runtime-spec

[generate.1]: man/ocitools-generate.1.md
[validate.1]: man/ocitools-validate.1.md
