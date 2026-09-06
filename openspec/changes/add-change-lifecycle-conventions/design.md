## Context

See `proposal.md` — Why. What shapes the approach:

- Four changes have gone through this repository's lifecycle. Three were archived in a pull request opened after the implementation had already merged (#4, #6, #9); the fourth (#11) followed the same pattern. Every one of them left the default branch, for a while, holding an implementation whose specs had not been updated.
- `openspec/config.yaml` exists but is an all-comment placeholder. It supports `rules:` per artifact, delivered through `openspec instructions <artifact> --json`, and `operations.<op>.guidance`, delivered to the apply and archive workflows as `operationGuidance`. Both are read by the workflows before they act.
- `openspec archive <name> --yes` is deterministic. It reproduced the agent-driven sync byte-for-byte, modulo blank lines, on all three previously archived changes; the only non-whitespace difference across all three was one checkbox a human had ticked. It was then used directly to produce #11, on the harder path — an existing capability with both ADDED and REMOVED requirements.
- `openspec archive` fails safe on the case that matters: a MODIFIED block that would drop scenarios the main spec still has aborts with exit 1 and an untouched working tree.
- `--yes` is required for a non-interactive run, and it suppresses the tool's own task-completeness check: it archived a 22/23 change with only a warning. Completeness has to be established before calling it, not by it.

## Goals / Non-Goals

**Goals:**

- Make the spec state on the default branch a consequence of what merged, not of a follow-up action someone has to remember.
- Keep the mechanism as small as the problem: a convention and two configuration blocks.
- Keep the archive reproducible by hand, so it can always be inspected, regenerated, or redone.

**Non-Goals:**

- Automating the propose or apply steps. Those are judgement.
- Folding the proposal into the same pull request as the implementation. Reviewing a plan and reviewing its execution are different acts, and the plan is worth landing before the work starts.
- Enforcing any of this mechanically. These are conventions delivered to the workflow that writes the artifacts; a check that fails a pull request for having a post-merge task would cost more than it saves at this size.

## Decisions

### The archive travels with the implementation, rather than being reconstructed afterwards

**Chosen:** when the last task is ticked, archive on the same branch, in the same pull request.

The property this buys is not convenience, it is atomicity. The main specs on the default branch stop being a thing that trails the code by however long it takes someone to open a second pull request, and become a function of what merged. An abandoned pull request archives nothing. A merged one moved the specs with the code.

**Alternative — a merge-triggered workflow that archives after the fact.** This was the previous revision of this change, and it was built and verified before being rejected. It triggered on `workflow_run` after CI succeeded on the default branch, identified the change to archive from two independent signals (`openspec list --json` reporting `status: "complete"`, *and* the built commit having touched that change's `tasks.md`), guarded with strict validation before and after, minted a `timvw-ci-bot` installation token because events created with `GITHUB_TOKEN` do not trigger the check runs that branch protection requires, and opened a pull request set to auto-merge — which in turn required enabling `allow_auto_merge` on the repository.

It worked. The detector was replayed against real history and classified every case correctly. It was rejected anyway, because all of that machinery exists to reconstruct after the fact something the apply pull request already knows for certain: which change it just completed. Every part of it — the trigger, the two-signal detector, the multiple-candidate failure mode, the bot credentials, the auto-merge setting, the recursion analysis — is a consequence of having thrown that knowledge away by deferring the archive. Removing the deferral removes the need for all of it, and leaves nothing to fail at three in the morning under a token nobody remembers granting.

Recorded here rather than discarded, because "why don't we automate the archive?" is a question that will be asked again, and the answer is not "we never thought of it".

### The conventions live in `openspec/config.yaml`, not in a document

`rules.tasks` is injected into the output of `openspec instructions tasks`, which the planning workflow reads *before* writing a task list. `operations.apply.guidance` is delivered to the apply workflow as `operationGuidance`. Both therefore reach the thing that would otherwise get it wrong, at the moment it would get it wrong.

A convention in a README reaches an author only if they remember it exists. That is precisely the failure mode being fixed — the current process already "requires" archiving after merge, and the requirement lives in nobody's way.

The distinction between the two blocks is deliberate: how to *author a task list* is a constraint on an artifact, so it is an artifact rule; *when to archive* is a constraint on an operation, so it is operation guidance. Putting either in the other's place would still be delivered, but to the wrong step.

### Guidance, not enforcement

Neither block is a check. `operationGuidance` is explicitly advisory in the workflow contract, and `rules` constrain the artifact being written rather than gating anything.

A CI check could reject a pull request whose `tasks.md` contains an unticked box while its implementation is present, or one that leaves a completed change outside the archive. It was considered and rejected for now: the failure it prevents is visible immediately in review, it costs a false-positive class of its own (a legitimately in-progress change pushed for early feedback), and the repository has one contributor. The `spec-lifecycle` requirements state the property, so a later change can add enforcement without reopening the question of what should be true.

### The archive is the last commit on the branch

Once the archive is applied, the delta spec lives under `openspec/changes/archive/…` and the main spec is already updated. If review then demands a spec change, both halves need editing in an awkward place.

Keeping the archive as its own final commit makes the recovery mechanical: drop the commit, make the change, re-run the archive. This is a working practice rather than a requirement — it constrains nothing observable — so it is recorded here and not in the spec.

### `openspec archive --yes`, with completeness established first

The tooling's own archive is used rather than a hand-merge or a model reproducing it, because it is deterministic, refuses the dangerous MODIFIED case outright, and gives a maintainer a one-command way to regenerate exactly what is on the branch.

The one sharp edge is that `--yes`, which a non-interactive run requires, also suppresses the tool's task-completeness check — it will archive an unfinished change with a warning and exit 0. Under this design that matters far less than it would have under the rejected workflow, because a human is running the command with the task list in front of them. It is still worth knowing, and it is why the `spec-lifecycle` requirement is phrased as archiving when the *last task is completed* rather than whenever the command happens to be run.

## Risks / Trade-offs

- **A spec revision demanded in review has to be made in two places** — the archived delta and the already-updated main spec. → Keep the archive as the final commit; drop it, revise, re-archive.
- **A larger apply pull request**, mixing implementation with a spec merge. → That is the review being put where it belongs. The archive portion is mechanically generated and reproducible with one command, so it can be verified rather than read line by line.
- **The conventions are advisory and can simply be ignored.** → They are delivered to the workflow at the point of authoring, which is where the mistake was being made. If they turn out to be ignored in practice, the `spec-lifecycle` requirements are already stated precisely enough to be enforced by a check later.
- **A change legitimately needs something confirmed post-merge.** → It becomes a scenario in the delta spec, which is where the four existing examples already were. If it needs *doing* rather than confirming, it goes in the migration plan or a follow-up change.
- **`--yes` will archive an incomplete change.** → Completeness is established before running it; under this design the person running it is looking at the task list.

## Migration Plan

1. Land this change. Its own pull request carries its own archive, which is the first exercise of the convention.
2. The next ordinary change follows it end to end: propose in one pull request, then implement and archive together in the second.
3. Nothing has to be undone for changes already archived under the old two-pull-request pattern. They are correct as they stand; only the interval before their archive landed was undesirable, and that interval is over.

**Rollback:** revert the `openspec/config.yaml` edit. Nothing else exists to undo — no workflow, no repository setting, no code.

## Open Questions

- Whether to eventually enforce these properties with a check rather than guidance. Deferred deliberately: the requirements are stated, so adding enforcement later is additive and does not reopen the design.
