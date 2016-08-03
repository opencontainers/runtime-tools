# oci-runtime-tool integration tests

This project uses the [Sharness][] test harness, installed as a [Git
submodule][submodule].  To setup the test installation after a clone,
run:

    $ git submodule update --init

which will checkout the appropriate Sharness commit in the `sharness`
directory (after which the `Makefile`, `sharness.sh`, etc. symlinks
will resolve successfully).

Run the tests with:

    $ make

And read the `Makefile` source to find other useful targets
(e.g. [`prove`][prove]).

## Naming

0 - Global `oci-runtime-tool` options.

1 - `oci-runtime-tool generate`.

## Dependencies

* [GNU Core Utilities][coreutils] for [`cat`][cat.1],
  [`echo`][echo.1], [`head`][head.1], and [`sed][sed.1].
* [jq] for [`jq`][jq.1].

[coreutils]: http://www.gnu.org/software/coreutils/coreutils.html
[jq]: https://stedolan.github.io/jq/
[prove]: http://perldoc.perl.org/prove.html
[Sharness]: http://mlafeldt.github.io/sharness/
[submodule]: http://git-scm.com/docs/git-submodule

[cat.1]: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/cat.html
[echo.1]: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/echo.html
[head.1]: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/head.html
[jq.1]: https://stedolan.github.io/jq/manual/
[sed.1]: http://pubs.opengroup.org/onlinepubs/9699919799/utilities/head.html
