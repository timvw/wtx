# spec-lifecycle Specification

## Purpose
Governs how a change moves from a plan to archived specification: what may be written as a task, when and where a completed change is archived, and what the specifications on the default branch are guaranteed to describe.
## Requirements
### Requirement: A task is work that can be completed before the change merges

A change's task list SHALL contain only work that alters the repository and can be finished before the change's pull request is merged. A task SHALL NOT require an event that can only happen after the merge.

Post-merge *verification* SHALL be expressed as a scenario in the change's delta spec rather than as a task. The scenario is stated as required behaviour, is testable, and survives into the main spec, whereas a ticked checkbox is a one-time assertion in a file that is read once and then archived.

Post-merge *action* — work a maintainer chooses to perform later, such as publishing a release — SHALL be recorded in the change's design document as part of its migration plan, or deferred to a follow-up change. It is not a task of the change that enables it.

This constraint SHALL be declared in the project's OpenSpec configuration so that it reaches the planning workflow automatically rather than depending on an author remembering it.

#### Scenario: A change's tasks can all be completed before merge

- **WHEN** a change's task list is written
- **THEN** every task can be completed and marked done in the pull request that introduces it
- **AND** no task is phrased as an instruction to be carried out after merging

#### Scenario: Post-merge verification is expressed as a scenario

- **WHEN** a change needs behaviour confirmed that can only be observed after it lands
- **THEN** that behaviour appears as a requirement scenario in the change's delta spec
- **AND** it does not also appear as a task

#### Scenario: Post-merge action is recorded in the design

- **WHEN** a change enables an action a maintainer will take later at a time of their choosing
- **THEN** that action is described in the change's design document or deferred to a follow-up change
- **AND** it is not written as a task of the current change

#### Scenario: The constraint is delivered to the planning workflow

- **WHEN** the planning workflow requests the instructions for authoring tasks
- **THEN** the project's OpenSpec configuration supplies this constraint as a rule

### Requirement: A change is archived in the pull request that completes it

When the last task of a change is completed, the change SHALL be archived on the same branch, in the same pull request: its delta specs merged into the main specs, and its change directory moved under the archive directory.

Archiving SHALL NOT be deferred to a separate pull request opened after the implementation has merged. Deferring it leaves the default branch carrying an implementation whose specifications have not been updated to describe it.

This convention SHALL be declared in the project's OpenSpec configuration so that it reaches the implementation workflow rather than depending on memory.

#### Scenario: Implementation and archive land together

- **WHEN** a pull request completes the last task of a change
- **THEN** that same pull request also contains the change's delta specs merged into the main specs
- **AND** it contains the change directory moved under the archive directory

#### Scenario: An unmerged pull request archives nothing

- **WHEN** a pull request that would complete a change is closed without merging
- **THEN** no change has been archived
- **AND** the main specs are unchanged

#### Scenario: The convention is delivered to the implementation workflow

- **WHEN** the implementation workflow requests its operation guidance
- **THEN** the project's OpenSpec configuration supplies this convention

### Requirement: The specifications on the default branch describe what the default branch contains

At every commit on the default branch, the main specifications SHALL describe the state of the repository at that commit: every merged change's requirements SHALL be present in the main specs, and no unmerged change's requirements SHALL be.

There SHALL NOT be an interval during which the default branch carries an implementation whose delta specs have not yet been merged into the main specs.

#### Scenario: A merged change's requirements are in the main specs

- **WHEN** a change's implementation is present on the default branch
- **THEN** that change's requirements are present in the main specs at the same commit

#### Scenario: No change is left half-applied

- **WHEN** any commit on the default branch is inspected
- **THEN** no change directory outside the archive has all of its tasks complete

### Requirement: The archive is the tooling's deterministic output

Archiving SHALL be performed by the OpenSpec tooling's own non-interactive archive operation, rather than by reproducing its behaviour by hand or by a model. A maintainer SHALL be able to regenerate an archive locally with a single command and obtain the same result.

Archiving SHALL be refused when the tooling reports that it cannot apply a delta safely, and the change SHALL NOT be archived until the reported problem is resolved.

#### Scenario: An archive is reproducible

- **WHEN** a maintainer runs the archive command locally for a change
- **THEN** it produces the same result that is committed on the branch

#### Scenario: An unsafe delta is not archived

- **WHEN** the tooling reports that applying a delta would drop content the main spec still has
- **THEN** the change is not archived
- **AND** the main specs are left unmodified

