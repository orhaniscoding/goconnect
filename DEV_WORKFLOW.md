# Development Workflow Policy

Agreed single-thread flow on `main` branch.

## Branching Rules
1. **STRICTLY NO BRANCHING.** All development occurs on `main`.
2. No feature branches. No merge commits.
3. Commit often, push when stable.

## Commit Strategy
1. Keep commits focused and logically grouped.
2. Use Conventional Commit prefixes (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).
3. Ensure code compiles and tests pass before pushing to `main`.

## Conflict Avoidance
1. Pull `main` frequently (`git pull --rebase`) to stay up to date.
2. If a conflict occurs, resolve it locally before pushing.

## Tooling Expectations
1. Run `go vet` and (when configured) `golangci-lint run` locally before pushing.
2. Update related docs (`docs/`) and OpenAPI when public endpoints change.

## Audit / Persistence Specific
1. Schema or storage changes require a note in `docs/AUDIT_NOTES.md`.
2. Adding new audit actions: update `internal/audit/actions.go` only.

---
Revision: v2 – Date: 2025-11-25
