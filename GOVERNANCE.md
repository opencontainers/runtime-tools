# Project governance

The [OCI charter][charter] §5.b.viii tasks an OCI Project's maintainers (listed in the repository's MAINTAINERS file and sometimes referred to as "the TDC", [§5.e][charter]) with:

> Creating, maintaining and enforcing governance guidelines for the TDC, approved by the maintainers, and which shall be posted visibly for the TDC.

This section describes generic rules and procedures for fulfilling that mandate.

## Proposing a motion

A maintainer SHOULD propose a motion on the dev@opencontainers.org mailing list (except [security issues](#security-issues)) with another maintainer as a co-sponsor.

## Voting

Voting on a proposed motion SHOULD happen on the dev@opencontainers.org mailing list (except [security issues](#security-issues)) with maintainers posting LGTM or REJECT.
Maintainers MAY also explicitly not vote by posting ABSTAIN (which is useful to revert a previous vote).
Maintainers MAY post multiple times (e.g. as they revise their position based on feeback), but only their final post counts in the tally.
A proposed motion is adopted if two-thirds of votes cast, a quorum having voted, are in favor of the release.

Voting SHOULD remain open for a week to collect feedback from the wider community and allow the maintainers to digest the proposed motion.
Under exceptional conditions (e.g. non-major security fix releases) proposals which reach quorum with unanimous support MAY be adopted earlier.

A maintainer MAY choose to reply with REJECT.
A maintainer posting a REJECT MUST include a list of concerns or links to written documentation for those concerns (e.g. GitHub issues or mailing-list threads).
The maintainers SHOULD try to resolve the concerns and wait for the rejecting maintainer to change their opinion to LGTM.
However, a motion MAY be adopted with REJECTs, as outlined in the previous paragraphs.

## Quorum

A quorum is established when at least two-thirds of maintainers have voted.

For projects that are not specifications, a [motion to release](#release-approval) MAY be adopted if the tally is at least three LGTMs and no REJECTs, even if three votes does not meet the usual two-thirds quorum.

## Security issues

Motions with sensitive security implications MUST be proposed on the security@opencontainers.org mailing list instead of dev@opencontainers.org, but should otherwise follow the standard [proposal](#proposing-a-motion) process.
The security@opencontainers.org mailing list includes all members of the TOB.
The TOB will contact the project maintainers and provide a channel for discussing and voting on the motion, but voting will otherwise follow the standard [voting](#voting) and [quorum](#quorum) rules.
The TOB and project maintainers will work together to notify affected parties before making an adopted motion public.

## Amendments

The [project governance](#project-governance) rules and procedures MAY be ammended or replaced using the procedures themselves.
No additional quorum or voting restrictions apply to such motions.

## Subject templates

Maintainers are busy and get lots of email.
To make project proposals recognizable, proposed motions SHOULD use the following subject templates.

### Proposing a motion

> [{project} VOTE]: {motion description} (closes {end of voting window})

For example:

> [runtime-spec VOTE]: Tag 0647920 as 1.0.0-rc (closes 2016-06-03 20:00 UTC)

### Tallying results

After voting closes, a maintainer SHOULD post a tally to the motion thread with a subject template like:

> [{project} {status}]: {motion description} (+{LGTMs} -{REJECTs} #{ABSTAINs})

Where `{status}` is either `adopted` or `rejected`.
For example:

> [runtime-spec adopted]: Tag 0647920 as 1.0.0-rc (+6 -0 #3)

# Releases

The release process hopes to encourage early, consistent consensus-building during project development.
The mechanisms used are regular community communication on the mailing list about progress, scheduled meetings for issue resolution and release triage, and regularly paced and communicated releases.
Releases are proposed and adopted or rejected using the usual [project governance](#project-governance) rules and procedures.

An anti-pattern that we want to avoid is heavy development or discussions "late cycle" around major releases.
We want to build a community that is involved and communicates consistently through all releases instead of relying on "silent periods" as a judge of stability.

## Parallel releases

A single project MAY consider several motions to release in parallel.
However each motion to release after the initial 0.1.0 MUST be based on a previous release that has already landed.

For example, runtime-spec maintainers may propose a v1.0.0-rc2 on the 1st of the month and a v0.9.1 bugfix on the 2nd of the month.
They may not propose a v1.0.0-rc3 until the v1.0.0-rc2 is accepted (on the 7th if the vote initiated on the 1st passes).

## Specifications

The OCI maintains three categories of projects: specifications, applications, and conformance-testing tools.
However, specification releases have special restrictions in the [OCI charter][charter]:

* They are the target of backwards compatibility (§7.g), and
* They are subject to the OFWa patent grant (§8.d and e).

To avoid unfortunate side effects (onerous backwards compatibity requirements or Member resignations), the following additional procedures apply to specification releases:

### Planning a release

Every OCI specification project SHOULD hold meetings that involve maintainers reviewing pull requests, debating outstanding issues, and planning releases.
This meeting MUST be advertised on the project README and MAY happen on a phone call, video conference, or on IRC.
Maintainers MUST send updates to the dev@opencontainers.org with results of these meetings.

Before the specification reaches v1.0.0, the meetings SHOULD be weekly.
Once a specification has reached v1.0.0, the maintainers may alter the cadence, but a meeting MUST be held within four weeks of the previous meeting.

The release plans, corresponding milestones and estimated due dates MUST be published on GitHub (e.g. https://github.com/opencontainers/runtime-spec/milestones).
GitHub milestones and issues are only used for community organization and all releases MUST follow the [project governance](#project-governance) rules and procedures.

### Timelines

Specifications have a variety of different timelines in their lifecycle.

* Pre-v1.0.0 specifications SHOULD release on a monthly cadence to garner feedback.
* Major specification releases MUST release at least three release candidates spaced a minimum of one week apart.
  This means a major release like a v1.0.0 or v2.0.0 release will take 1 month at minimum: one week for rc1, one week for rc2, one week for rc3, and one week for the major release itself.
  Maintainers SHOULD strive to make zero breaking changes during this cycle of release candidates and SHOULD restart the three-candidate count when a breaking change is introduced.
  For example if a breaking change is introduced in v1.0.0-rc2 then the series would end with v1.0.0-rc4 and v1.0.0.
- Minor and patch releases SHOULD be made on an as-needed basis.

[charter]: https://www.opencontainers.org/about/governance
