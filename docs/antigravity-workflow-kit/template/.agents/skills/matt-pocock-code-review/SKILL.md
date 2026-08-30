---
name: matt-pocock-code-review
description: Two-axis code review adapted from Matt Pocock's code-review skill (MIT). Reviews diff for Standards (repo conventions + code smells) and Spec (matches requested work, no scope creep). Use in verifier agent or after implementation.
disable-slash-command: true
---

# Two-axis review (Matt Pocock–style)

Adapted from [mattpocock/skills — code-review](https://github.com/mattpocock/skills) (MIT).

Review the diff against a fixed point (branch, commit, or working tree).

## Two axes (report separately)

### Standards — does code follow conventions?

- Documented rules in AGENTS.md and `.agents/rules/`
- Existing patterns in touched packages
- Code smells (judgement calls, not auto-fail):
  - **Speculative Generality** — abstraction not required by task
  - **Shotgun Surgery** — one logical change scattered across many unrelated files
  - **Duplicated Code** — same logic copied in the diff
  - **Drive-by refactor** — behaviour change outside spec

Repo standards override smell heuristics.

### Spec — does code match the request?

- Requirements missing or partial
- Behaviour added that was not asked for (**scope creep**)
- Files changed outside the approved file plan

A change can pass one axis and fail the other. Report both.

## Process

1. Pin diff: `git diff` or `git diff main...HEAD`
2. Read spec/task description and **file plan**
3. Review each hunk — quote file:line for findings
4. Output under `## Standards` and `## Spec` headings

## Install upstream

For full Matt Pocock workflow (`/implement`, `/tdd`, parallel subagents):

```bash
npx skills add mattpocock/skills --skill tdd,code-review,implement
```

## Completion criterion

Both axes reported. Spec axis must show no unplanned files or unrequested behaviour for PASS.
