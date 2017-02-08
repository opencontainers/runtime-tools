# OCI Runtime Command Line Interface

The [OCI Runtime Specification][runtime-spec] currently focuses on the bundle-author ↔ runtime interface, but there is renewed interest in specifying a command-line API for the runtime-caller ↔ runtime interface.
A common command-line API would make it easier to build higher-level tools that are runtime-agnostic (e.g. compliance testers like [runtime-tools][runtime-tools-compliance]).
This repository contains initial work on that API, with more detailed discussion in [this thread][thread].
The usual [development rules][rules] apply, and the legal stuff is spelled out [here](CONTRIBUTING.md).
The target for the inital design will be to match [the specified lifecycle][lifecycle], keeping as much similarity with the existing [runC][] command-line as possible.

[runtime-spec]: https://github.com/opencontainers/runtime-spec
[runtime-tools-compliance]: https://github.com/opencontainers/runtime-tools#testing-oci-runtimes
[thread]: https://groups.google.com/a/opencontainers.org/forum/#!topic/dev/BIxya5eSNLo
[rules]: https://github.com/opencontainers/runtime-spec#contributing
[lifecycle]: https://github.com/opencontainers/runtime-spec/blob/master/runtime.md#lifecycle
[runC]: https://github.com/opencontainers/runc
