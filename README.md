# OCI command-line API

The [OCI Specifications][specs] currently focus on the bundle-author ↔ runtime interface, but there is renewed interest in specifying a command-line API for the runtime-caller ↔ runtime interface.
A common command-line API would make it easier to build higher-level tools that are runtime-agnostic (e.g. conformance testers like [ocitools][ocitools-test]).
This repository contains initial work on that API, with more detailed discussion in [this thread][thread].
The usual [development rules][rules] apply, and the legal stuff is spelled out [here](CONTRIBUTING.md).
The target for the inital design will be to match [the lifecycle pull request][lifecycle], keeping as much similarity with the existing [runC][] command-line as possible.

[specs]: https://github.com/opencontainers/specs
[ocitools-test]: https://github.com/mrunalp/ocitools#testing-oci-runtimes
[thread]: https://groups.google.com/a/opencontainers.org/forum/#!topic/dev/BIxya5eSNLo
[rules]: https://github.com/opencontainers/specs#contributing
[lifecycle]: https://github.com/opencontainers/specs/pull/231
[runC]: https://github.com/opencontainers/runc
