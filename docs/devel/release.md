# runtime-tools release guide

## Release process

This section shows how to perform a release of runtime-tools.
The following example assumes we're going from version 0.5.0 (`v0.5.0`) to 0.6.0 (`v0.6.0`).

Let's get started:

- Start at the relevant milestone on GitHub (e.g. https://github.com/opencontainers/runtime-tools/milestones/v0.6.0): ensure all referenced issues are closed (or moved elsewhere, if they're not done). Close the milestone.
- runtime-tools does not use a [roadmap file](https://github.com/opencontainers/runtime-tools/issues/465) but GitHub milestones. Update the [other milestones](https://github.com/opencontainers/runtime-tools/milestones), if necessary.
- Branch from the latest master, make sure your `git status` is clean.
- Update the [VERSION](https://github.com/opencontainers/runtime-tools/blob/master/VERSION).
- Update the [release notes][changelog].
  Try to capture most of the salient changes since the last release, but don't go into unnecessary detail (better to link/reference the documentation wherever possible).

Ensure the branch is correct:

- Ensure the build is clean!
  - `git clean -ffdx && make && make test` should work.
- Integration tests on CI should be green.
- Check the version of the binaries:
  - Check `./oci-runtime-tool --version`.
  - Check `./runtimetest --version`.

Once everything is fine:

- File a pull request.
- Ensure the CI on the release PR is green.
- Send an email to the [mailing list][mailinglist] ([example for v0.5.0](https://groups.google.com/a/opencontainers.org/forum/#!topic/dev/iuWpWUai4_I)) and get reviews from other [maintainers][maintainers].
- Once the maintainers agree, merge the PR.

Sign a tagged release and push it to GitHub:

- Add a signed tag: `git tag -s v0.6.0 -m "release v0.6.0"`.
- Push the tag to GitHub: `git push origin v0.6.0`.

Now we switch to the GitHub web UI to conduct the release:

- Start a [new release][gh-new-release] on Github.
- Tag "v0.6.0", release title "v0.6.0".
- Copy-paste the release notes you added earlier in [CHANGELOG.md][changelog].
- Attach the release.
  This is a simple tarball:

```console
$ export VER="1.2.0"
$ export NAME="runtime-tools-v$VER"
$ mkdir -p $NAME/validation
$ cp oci-runtime-tool $NAME/
$ cp validation/*.t $NAME/validation/
$ sudo chown -R root:root $NAME/
$ tar czvf $NAME.tar.gz --numeric-owner $NAME/
```

- Publish the release!

- Clean your git tree: `sudo git clean -ffdx`.

[changelog]: ../../CHANGELOG.md
[maintainers]: ../../MAINTAINERS
[mailinglist]: https://groups.google.com/a/opencontainers.org/forum/#!forum/dev
[gh-new-release]: https://github.com/opencontainers/runtime-tools/releases/new
