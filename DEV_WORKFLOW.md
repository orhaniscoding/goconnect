# Development Workflow Policy

Agreed single-thread feature flow to avoid merge conflicts and unreviewed parallel branches.

## Branching Rules
1. Only one active feature branch at a time.
2. New feature branch MUST be approved before creation.
3. Feature branch name: `feat/<short-topic>` or `chore/<topic>` consistent with Conventional Commits.
4. Do not open a second feature branch until the first is merged into `main`.

## Commit Strategy
1. Keep commits focused and logically grouped (avoid giant mixed changes).
2. Use Conventional Commit prefixes (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
3. No auto-pushing arbitrary intermediate WIP without prior sync if it touches shared core files (RBAC, audit, service layers).

## Merge Process
1. Open Pull Request early (draft) for visibility.
2. Ensure CI: `go test ./... -race -cover` and `npm run typecheck && npm run build` (web) are green before marking ready.
3. Resolve review comments before squash/merge (or rebase merge if preserving commit history is required).
4. After merge: delete branch remotely and locally to prevent stale divergence.

## Conflict Avoidance
1. High-churn files (`docs/AUDIT_NOTES.md`, `server/internal/service/*`, `server/internal/audit/*`) edited by only the active feature.
2. If urgent hotfix needed: branch from latest `main`, merge hotfix first, then rebase active feature branch.

## Emergency Patch
1. For production-impacting bug: create `fix/<issue>` from `main`.
2. Patch, test, PR, merge; then rebase active feature.

## Tooling Expectations
1. Run `go vet` and (when configured) `golangci-lint run` locally before pushing.
2. Race detector is mandatory for test runs affecting concurrency primitives.
3. Update related docs (`docs/`) and OpenAPI when public endpoints change.

## Audit / Persistence Specific
1. Schema or storage changes require a note in `docs/AUDIT_NOTES.md` in the same PR.
2. Adding new audit actions: update `internal/audit/actions.go` only; services reference constants.
3. Before introducing async buffering, add metrics placeholders & backpressure strategy notes.

## Deviation Protocol
If a deviation is necessary (time-critical security patch), explicitly document rationale in PR description and link follow-up cleanup issue.

---
Revision: v1 – Date: 2025-09-29