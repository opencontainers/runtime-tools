% OCI(1) OCITOOLS User Manuals
% OCI Community
% APRIL 2016
# NAME
ocitools \- OCI (Open Container Initiative) tools

# SYNOPSIS
**ocitools** [OPTIONS] COMMAND [arg...]

**ocitools** [--help|-v|--version]

# DESCRIPTION
ocitools is a collection of tools for working with the [OCI runtime specification](https://github.com/opencontainers/runtime-spec).


# OPTIONS
**--help**
  Print usage statement.

**-v**, **--version**
  Print version information.

**--log-level**
  Log level (panic, fatal, error, warn, info, or debug) (default: "error").

**--host-specific**
  Generate host-specific configs or do host-specific validations.

  By default, generator generates configs without checking whether they are
  supported on the current host. With this flag, generator will first check
  whether each config is supported on the current host, and only add it into
  the config file if it passes the checking.

  By default, validation only tests for compatibility with a hypothetical host.
  With this flag, validation will also run more specific tests to see whether
  the current host is capable of launching a container from the configuration.

# COMMANDS
**validate**
  Validating OCI bundle
  See **ocitools-validate(1)** for full documentation on the **validate** command.

**generate**
  Generating OCI runtime spec configuration files
  See **ocitools-generate(1)** for full documentation on the **generate** command.

# SEE ALSO
**ocitools-validate**(1), **ocitools-generate**(1)

# HISTORY
April 2016, Originally compiled by Daniel Walsh (dwalsh at redhat dot com)
