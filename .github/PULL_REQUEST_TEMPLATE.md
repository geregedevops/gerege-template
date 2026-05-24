<!-- Thanks for contributing to Gerege Template! -->

## What & why

<!-- What does this PR change, and why? Link related issues: Closes #123 -->

## Area

- [ ] backend (Go / Fiber v3)
- [ ] frontend (Next.js)
- [ ] docs / CI

## Checklist

- [ ] Clean Architecture boundaries preserved (business/domain do not import the web framework)
- [ ] Tests added/updated and passing (`make test` / `npm run build`)
- [ ] Lint passes (`make lint` / `npm run lint`)
- [ ] Docs updated (`backend/docs/*` and `_MN` counterpart, if applicable)
- [ ] `make swag` run if handler annotations changed (no `docs/` drift)
- [ ] No secrets committed; `.env*` stays gitignored

## Notes

<!-- Anything reviewers should know: trade-offs, follow-ups, screenshots. -->
