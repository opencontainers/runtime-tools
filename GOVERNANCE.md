# OCI Project Release Approval Process v1.0

OCI projects need a standard process for making releases so the community of maintainers can consistently know when something can be tagged and released. This approval process hopes to encourage early consistent consensus building during project and specification development. The mechanisms used are regular community communication on the mailing list about progress, scheduled meetings for issue resolution and release triage, and regularly paced and communicated releases. An anti-pattern that we want to avoid is heavy development or discussions "late cycle" around major releases. We want do build a community that is involved and communicates consistently through all releases instead of relying on "silent periods" as a judge of stability.

## List-based voting

**Making a release:** Maintainers (listed in the repository's MAINTAINERS file) MUST announce intentions to release on the dev@opencontainers.org mailing list with another maintainer as a co-sponsor. Voting on proposed releases SHOULD happen on the dev@opencontainers.org mailing list (except [security fixes](#security-fixes)) with maintainers posting LGTM or REJECT. Maintainers may also explicitly not vote by posting ABSTAIN (which is useful to revert a previous vote). Maintainers may post multiple times (e.g. as they revise their position based on feeback), but only their final post counts in the final tally. A proposed release passes if two-thirds of votes cast, a quorum having voted, are in favor of the release. A quorum is established when at least two-thirds of maintainers have voted. Voting SHOULD remain open for a week, although under exceptional conditions (e.g. security fixes) non-major releases which reach quorum with unanimous support MAY be released earlier.  For projects that are not specifications, a proposed release also passes if the final tally is at least three LGTMs and no REJECTs, even if three votes does not meet the usual two-thirds quorum.

**Rejecting a release:** A project maintainer MAY choose to reply with REJECT. A project maintainer posting a REJECT MUST include a list of concerns or links to written documentation for those concerns (e.g. GitHub issues or mailing-list threads). The project maintainers SHOULD try to resolve the concerns and wait for the rejecting maintainer to change their opinion to LGTM. However, a release MAY pass with REJECTs, as outlined in the previous paragraph.

## Security fixes

Security fix releases MUST use security@opencontainers.org instead of dev@opencontainers.org, but should otherwise follow the standard [list-based voting process](#list-based-voting).

## Parallel proposals

A single repository MAY have several release proposals in parallel. However each proposed release after the first MUST be based on a previous release that has already landed.

For example, runtime-spec maintainers may propose a v1.0.0-rc2 on the 1st of the month and a v0.9.1 bugfix on the 2nd of the month. They may not propose a v1.0.0-rc3 until the v1.0.0-rc2 is accepted (on the 7th if the vote initiated on the 1st passes).

## Specifications

The OCI maintains three categories of projects: specifications, applications, and conformance-testing tools. However, specification releases have special restrictions in the [OCI charter][charter]:

* They are the target of backwards compatibility (§7.g), and
* They are subject to the OFWa patent grant (§8.d and e).

To avoid unfortunate side effects (onerous backwards compatibity requirements or Member resignations), the following additional procedures apply to specification releases:

**Planning a release:** Every OCI specification project SHOULD hold meetings that involves maintainers reviewing pull requests, debating outstanding issues, and planning releases. This meeting MUST be advertised on the project README and MAY happen on a phone call, video conference, or on IRC. Maintainers MUST send updates to the dev@opencontainers.org with results of these meetings. Before the specification reaches v1.0.0, the meetings SHOULD be weekly.  Once a specification has reached v1.0.0, the maintainers may alter the cadence, but the meeting cadence MUST NOT be greater than once every four weeks. The release plans, corresponding milestones and estimated due dates MUST be published on GitHub (e.g. https://github.com/opencontainers/runtime-spec/milestones). GitHub milestones and issues are only used for community organization and all releases MUST follow the [list-based voting process](#list-based-voting).

**Timelines:** Specifications have a variety of different timelines in their lifecycle.

- Pre-v1.0.0 specifications SHOULD release on a monthly cadence to garner feedback.
- Major specification releases MUST release at least three release candidates spaced a minimum of one week apart. This means a major release like a v1.0.0 or v2.0.0 release will take 1 month at minimum: one week for rc1, one week for rc2, one week for rc3, and one week for the major release itself. Maintainers SHOULD strive to make zero breaking changes during this cycle of release candidates and SHOULD add restart the three-candidate count when a breaking change is introduced. For example if a breaking change is introduced in v1.0.0-rc2 then the series would end with v1.0.0-rc4 and v1.0.0.
- Minor and patch releases SHOULD be made on an as-needed basis.

[charter]: https://www.opencontainers.org/about/governance
