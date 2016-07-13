% OCI(1) OCITOOLS User Manuals
% OCI Community
% APRIL 2016
# NAME
ocitools-validate - Validate an OCI bundle

# SYNOPSIS
**ocitools validate**  *[OPTIONS]*

# DESCRIPTION

Validate an OCI bundle

# OPTIONS
**--help**
  Print usage statement

**--path=PATH
  Path to bundle

**--host-specific**
  Check host specific configs.
  By default, validation only tests for compatibility with a hypothetical host.
  With this flag, validation will also run more specific tests to see whether
  the current host is capable of launching a container from the configuration.
  For example, validating a compliant Windows configuration on a Linux machine
  will pass without this flag ("there may be a Windows host capable of
  launching this container"), but will fail with it ("this host is not capable
  of launching this container").

# SEE ALSO
**ocitools**(1)

# HISTORY
April 2016, Originally compiled by Dan Walsh (dwalsh at redhat dot com)
