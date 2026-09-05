# Task: T6 — Independent acceptance review

## Context

- Repository: `/home/nurul/agentusage`
- Role: `reviewer`
- Plan: [Usage simplification](../docs/superpowers/plans/2026-09-05-usage-simplification.md) — section “T6 — Independent acceptance review”.
- Dependency order: T5.
- Coordination: [plan.yml](../.coord/plan.yml).

## Pre-task scope confirmation

Read the plan's evidence, collection/auth contracts, migration rules and your task section. Report the baseline commit, resolved allowed paths and dependency evidence before editing. File globs in the coordination manifest are ceilings; the task section narrows them. Unlisted files require coordinator scope reconciliation before another writer starts. All paths outside your scope are read-only.

## Goal and execution

Execute the checklist in the linked task section; it is the single source of truth for behavior and tests. Work in `.worktrees/usage-simplify-t6` after the coordinator preserves the current uncommitted baseline. This handoff does not itself launch agents or authorize external BXO edits, credential migration or deployment.

## Return contract

Return changed files, baseline/head commits, exact validation commands and results, acceptance-case evidence, and remaining limitations. Do not expose secrets. Return an evidence-based PASS / CONDITIONAL PASS / FAIL; do not edit the candidate.
