## 1. Task-authoring rule

- [x] 1.1 Replace the commented `rules:` placeholder in `openspec/config.yaml` with a real block carrying a `tasks:` entry: a task is work completable and tickable before the change's pull request merges; post-merge verification belongs in the delta spec as a requirement scenario; post-merge action belongs in `design.md`'s migration plan or a follow-up change; and post-merge tasks are what prevent a change from being archived in the pull request that completes it.
- [x] 1.2 Use folded block scalars (`>-`) for the entries rather than plain multi-line scalars — a `: ` inside a plain scalar makes the file invalid YAML, and OpenSpec responds by warning and silently delivering no rules at all.

## 2. Apply-time archive convention

- [x] 2.1 Replace the commented `operations:` placeholder in `openspec/config.yaml` with a real `apply.guidance` block: when the last task of a change is complete, archive the change on the same branch with `openspec archive <name> --yes` so the pull request carries implementation, spec sync and archival together; keep the archive as the final commit so a review-driven revision can drop it, change the spec, and re-archive.
- [x] 2.2 Keep the `schema: spec-driven` key, and leave the remaining commented sections (project context) intact so the file still documents what else it accepts.

## 3. Verification

- [x] 3.1 Confirm `openspec/config.yaml` parses as YAML with a real parser, and that no `could not parse` warning appears on any `openspec` invocation.
- [x] 3.2 Confirm the task rule is actually delivered: `openspec instructions tasks --change add-change-lifecycle-conventions --json` returns the entries under `rules`.
- [x] 3.3 Confirm the apply guidance is actually delivered: `openspec instructions apply --change add-change-lifecycle-conventions --json` returns it under `operationGuidance`.
- [x] 3.4 Confirm `openspec validate add-change-lifecycle-conventions --strict` passes.
- [ ] 3.5 Archive this change on this branch with `openspec archive add-change-lifecycle-conventions --yes`, as its own final commit, and confirm `openspec validate --specs --strict` passes afterwards with `spec-lifecycle` present among the main specs.
