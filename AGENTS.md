# AGENTS.md

This repo is the Adventure game (Go + HTMX). Follow these conventions.

## HTMX and frontend

- Server returns HTML fragments; prefer OOB swaps for sidebar updates.
- Keep JS minimal for behavior HTML cannot handle (e.g. dice animation).
- Scope HTMX event listeners to `#game`; avoid global side effects/timeouts.
- No inline scripts. Keep JS in `static/js/` and load from layout templates.

## Tests

- When changing game logic, handlers, or session behavior, update the matching
  `*_test.go` file.
- Scenery handler changes must keep path validation and update
  `internal/web/scenery_test.go`.
- If `static/js/app.js` changes, update `static/js/app.test.js` and run
  `make test-js` or `npm test`.
- Maintain at least 75% coverage (`make test`). Use `make coverage-check`
  for an existing `coverage.out`.

## Docs and tooling

- Update README when adding requirements, commands, Makefile targets, or CI steps.
- Add Makefile targets for new tools/scripts and include them in `make check`
  when they should run in CI.
- Add CI steps in `.github/workflows/test.yml` when introducing new lint/test steps.

## References

- HTMX and JS refactor plan (see internal design notes or team docs)
- `make install-tools`, `make install-js`, `make check`
