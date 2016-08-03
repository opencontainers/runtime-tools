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

[prove]: http://perldoc.perl.org/prove.html
[Sharness]: http://mlafeldt.github.io/sharness/
[submodule]: http://git-scm.com/docs/git-submodule
