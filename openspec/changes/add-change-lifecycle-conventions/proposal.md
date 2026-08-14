## Why

Archiving a change is currently its own pull request, opened after the implementation has already merged. That leaves a window in which the default branch carries an implementation whose delta specs have not been folded into the main specs: `add-ci-release-pipeline` merged as PR #8 and sat that way until PR #9 landed. During that window the specs on `main` describe a system that is not the one on `main`.

The window exists only because the archive is separated from the work it describes. Archiving is not independent work — it moves a directory and merges the change's own deltas into the main specs, both of which are fully determined the moment the last task is ticked. Doing it in the same pull request closes the window by construction: abandon the pull request and nothing was archived; merge it and the specs moved with the code. The state of the specs on the default branch becomes a function of what actually merged, rather than of whether someone remembered to open a second pull request.

A second, smaller problem blocks this. Changes have been written with **post-merge tasks** — "after the pull request merges, confirm CI is green on `main`". A change carrying one cannot reach completion before it merges, so it cannot be archived in the pull request that completes it. Those tasks were never work: each restates a scenario that already exists in the same change's delta spec. Task 6.5 of `add-ci-release-pipeline` duplicates the `ci` scenario "Push to the default branch triggers verification"; tasks 5.2, 5.3, 5.5 and 5.6 of `add-branch-protection` duplicate four `repo-settings` scenarios. A ticked checkbox is weaker evidence than the scenario it copies — it is a one-time assertion in a file that is read once and then archived, while the scenario is durable, testable, and survives into the main spec.

## What Changes

- Establish that a completed change is archived **in the pull request that completes it**, so implementation, spec sync and archival land together or not at all.
- Record that convention in `openspec/config.yaml` under `operations.apply.guidance`, which the apply workflow reads, so it travels with the repository instead of depending on memory.
- Establish that a **task is work completable before the change merges**. Post-merge verification belongs in the delta spec as a requirement scenario; post-merge action belongs in the design document's migration plan or in a follow-up change. Recorded in `openspec/config.yaml` under `rules.tasks`, which is injected into the instructions for authoring a task list.
- Establish that the archive is produced by the OpenSpec tooling's own non-interactive archive operation rather than reproduced by a model, so the result is deterministic and a maintainer can regenerate it locally with one command.
- Explicitly **not** done: a merge-triggered workflow that archives after the fact. An earlier revision of this change proposed one — `workflow_run` on CI success, a two-signal candidate detector, a bot App token, `allow_auto_merge`, and guards for archiving the wrong thing. All of that machinery exists to reconstruct after the fact what the apply pull request already knows for certain, and it is unnecessary once the archive travels with the implementation. The reasoning is kept in `design.md` so it is not rediscovered.

## Capabilities

### New Capabilities

- `spec-lifecycle`: How a change moves from a plan to archived specification — what may be written as a task, when a completed change is archived, how that archival reaches the default branch, and what guarantees the resulting specs carry.

### Modified Capabilities

_None._ No existing capability's requirements change: this change adds conventions about how changes are authored and archived, and touches no repository setting, workflow, or application code.

## Impact

- **Modified files**: `openspec/config.yaml` only — its commented placeholders are replaced with a real `rules:` block and a real `operations:` block.
- **No workflow, settings, or code changes.** Nothing about CI, branch protection, or the released binary is affected.
- **Process impact**: an apply pull request grows by the archive — the change directory moves under `openspec/changes/archive/`, and the main specs gain that change's requirements. Reviewing the spec merge alongside the implementation it describes is the point, not a side effect.
- **Trade-off accepted**: once the archive is applied on a branch, the delta spec lives under the archive directory and the main spec is already updated, so a spec revision demanded in review means editing both. Mitigated by keeping the archive as the last commit on the branch, so it can be dropped, revised and re-applied. Named in `design.md`.
- **This change follows its own convention**: its task list contains no post-merge task, and its pull request carries its own archive.
