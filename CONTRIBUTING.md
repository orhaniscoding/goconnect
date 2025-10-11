# Contributing to GoConnect

## Workflow
1. Create feature branch: `feat/<scope>-<concise>` or `fix/<scope>-<issue>`
2. Implement changes (handler → service → repository pattern for server).
3. Run local checks:
   - `go test ./server/... -race -cover`
   - `go vet ./server/...` & golangci-lint (run `golangci-lint run ./server/...`)
   - `npm run typecheck && npm run build` in `web-ui/`
4. Update / sync OpenAPI & `docs/TECH_SPEC.md` if contract changed.
5. Add / update tests for new behavior.
6. Conventional Commit message (signed): `feat(server): add exporter interface`.
7. Open PR; ensure checklist passes.

## PR Checklist
- [ ] Tests (race) pass
- [ ] Lint & vet clean
- [ ] Coverage not reduced meaningfully
- [ ] OpenAPI + TECH_SPEC in sync
- [ ] Audit events added (if mutating action)
- [ ] No raw PII in logs / audit
- [ ] Idempotency-Key enforced on new mutating endpoints
- [ ] RBAC outward errors unify to `ERR_NOT_AUTHORIZED`

## Commit Message Format
`<type>(<scope>): <subject>`
Types: feat, fix, refactor, docs, chore, test, perf, build, ci, security.
Use present tense, no trailing period.

## Testing Guidance
- Add at least one happy path + one edge case.
- For new concurrency-sensitive code, add `-race` run.
- For hashing / redaction logic, assert stable hash length (24 chars) and no raw IDs.

## Security
- Never commit real secrets or private keys.
- Token validation is a stub; do not rely on it for production.
- If adding external deps: evaluate license + supply chain risk.

## Release Flow
- Merges to main trigger release-please; version PR merges create tag.
- Tag publication triggers goreleaser (binaries, archives).

## Style
- Go: idiomatic, short receiver names, prefer explicit error wrapping.
- TS/JS: strict TS where possible, no implicit any.

## Need Help?
Open a GitHub Discussion or create an issue labeled `question`.
