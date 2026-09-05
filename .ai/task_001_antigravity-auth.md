# Task: T1 — Direct Antigravity auth and API projection

## Context

- Repository: `/home/nurul/agentusage`
- Role: `delegated-executor`
- Plan: [Usage simplification](../docs/superpowers/plans/2026-09-05-usage-simplification.md) — section “T1 — Direct Antigravity auth and API projection”.
- Dependency order: none.
- Coordination: [plan.yml](../.coord/plan.yml).

## Pre-task scope confirmation

Read the plan's evidence, collection/auth contracts, migration rules and your task section. Report the baseline commit, resolved allowed paths and dependency evidence before editing. File globs in the coordination manifest are ceilings; the task section narrows them. Unlisted files require coordinator scope reconciliation before another writer starts. All paths outside your scope are read-only.

## Goal and execution

Execute the checklist in the linked task section; it is the single source of truth for behavior and tests. Work in `.worktrees/usage-simplify-t1` after the coordinator preserves the current uncommitted baseline. This handoff does not itself launch agents or authorize external BXO edits, credential migration or deployment.

## Return contract

Return changed files, baseline/head commits, exact validation commands and results, acceptance-case evidence, and remaining limitations. Do not expose secrets. Return implementation evidence for independent review; do not self-approve shipping.
