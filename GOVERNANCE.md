# OCI Project Release Approval Process v1.0

OCI projects need a standard process for making releases so the community of maintainers can consistently know when something can be tagged and released. As the OCI maintains three categories of projects: specifications, applications, and conformance/testing tools we will set different rules for each.

This approval process hopes to encourage early consistent consensus building during project and specification development. The mechanisms used are regular community communication on the mailing list about progress, scheduled meetings for issue resolution and release triage, and regularly paced and communicated releases. An anti-pattern that we want to avoid is heavy development or discussions "late cycle" around major releases. We want do build a community that is involved and communicates consistently through all releases instead of relying on "silent periods" as a judge of stability.

## Specifications

**Planning a release:** Every OCI specification project SHOULD hold a weekly meeting that involves maintainers reviewing pull requests, debating outstanding issues, and planning releases. This meeting MUST be advertised on the project README and MAY happen on a phone call, video conference, or on IRC. Maintainers MUST send updates to the dev@opencontainers.org with results of these meetings. Maintainers MAY change the meeting cadence once a specification has reached v1.0.0. The meeting cadence MUST NOT be greater than once every four weeks The release plans, corresponding milestones and estimated due dates MUST be published on GitHub (e.g. https://github.com/opencontainers/runtime-spec/milestones).

**Making a release:** OCI specification projects MUST announce intentions to release with two project maintainer sponsors (listed in the repo MAINTAINERS file) on the dev@opencontainers.org mailing list. After the announcement a two-thirds super majority of project maintainers MUST reply to the list with an LGTM within one week for the release to be approved. The maintainers MUST wait a full week for maintainers to reply but if all maintainers reply with an LGTM then the release MAY release earlier except in the case of a major releases.

**Rejecting a release:** A project maintainer MAY choose to reply with REJECT and MUST include a list of concerns filed as GitHub issues. The project maintainers SHOULD try to resolve the concerns and wait for the rejecting maintainer to change their opinion to LGTM. However, a release MAY continue with a single REJECT as long two-thirds of the project maintainers approved the release.

**Timelines:** Specifications have a variety of different timelines in their lifecycle. In early stages the spec SHOULD release often to garner feedback. In later stages there will be bug fix releases, security fix releases, and major breaking change releases. Each of these should have a different level of notification.

- Pre-v1.0.0 specifications SHOULD release on a regular cadence and MUST follow the normal release process. In practice this should be a release every week or two.
- Major specification releases that introduce new base or optional layers, break backwards compatibility, or add non-optional features MUST release at least two release candidates spaced a minimum of one week apart. In practice this means a major release like a v1.0.0 or v2.0.0 release will take 3 weeks at minimum: one week for v1.0.0-rc1, one week for v1.0.0-rc2, and one week for v1.0.0. Maintainers SHOULD strive to make zero breaking changes during this cycle of release candidates and SHOULD add an additional release candidate when a breaking change is introduced. For example if a breaking change is introduced in -rc2 then a -rc3 SHOULD be made following the normal release process.
- Minor releases that fix bugs, grammar, introduce optional features, tests, or tooling SHOULD be made on an as-needed basis and MUST follow the normal release process.
- Security fix releases MUST follow a special release process that substitutes the dev@opencontainers.org email for security@opencontainers.org.

## Conformance/Testing and Applications Releases

**Making a release:** OCI application projects MUST announce intentions to release with two project maintainer sponsors on the dev@opencontainers.org mailing list. After the announcement at least one more project maintainer MUST reply to the list with an LGTM within one week for the release to be approved. The maintainers SHOULD wait one week for maintainers to reply and review. If all maintainers reply with an LGTM then the release MAY release earlier.

**Rejecting a release:** A project maintainer MAY choose to reply with REJECT and MUST include a list of concerns filed as GitHub issues. The project maintainers SHOULD try to resolve the concerns and wait for the rejecting maintainer to change their opinion to LGTM. However, a release MAY continue with a single REJECT.

Security fix releases MUST follow a special release process that substitutes the dev@opencontainers.org email for security@opencontainers.org.
