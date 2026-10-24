# Argus RC Readiness

## Current decision

**Not RC-ready yet. Do not package.** The core policy, public-event privacy, AOCI runtime, browser evidence, current-task report rendering, and regression tests are in place. The real SRC Golden E2E is blocked by the configured model provider's HTTP 402 insufficient-balance response.

## Gates

| Gate | Status | Evidence |
|---|---|---|
| Unified Target/Test Range authorization | PASS | Persisted task context for `shu560.cn`; exact host and ports 80/443; fail-closed network wrappers |
| Public stream privacy | PASS | Regression tests cover conversation history/detail and task stream; private reasoning/usage are omitted |
| AOCI runtime alignment | PASS | Runtime status `aligned`, 798 files, fingerprint `be88ae99a9b5d5af79e93f9e1c7da31e8ac35230e2afc3cf7d1dd5628cbad427` |
| Browser evidence provenance | PASS | Task 4 browser page-state/screenshot evidence persisted with provenance and scope |
| Finding/evidence quality gate | PASS for zero-result handling | No unverified result was promoted; confirmed findings remain 0 |
| Current-task report bundle | PASS | Markdown, escaped HTML, PDF header, metadata and evidence refs validated |
| Full live Agent Golden E2E | BLOCKED | Provider returned HTTP 402 before any model tool call |
| Go and frontend build verification | PASS | `go test ./...`, TypeScript check, and production build passed |
| Repository-wide lint | BASELINE FAIL | Biome reports 30 errors/124 warnings/666 infos across existing files |

## Required before RC

1. Restore/configure a funded model provider and rerun task 4 (or a new equivalent task) from a clean SRC session.
2. Confirm at least five consecutive native tool calls, failure continuation, asset/traffic/event persistence, validation/impact/correlation gates, and current-task report generation in that live run.
3. Recheck the final report and UI pages after the successful run. Keep confirmed findings at zero unless evidence and quality gates genuinely support one.

The old project `/Users/shu/Documents/src-agent` is outside this work and was not modified.

## Local distribution authentication

For a GitHub download used as a single-user desktop install, authentication now defaults to `auto`: a backend bound only to `127.0.0.1` or another loopback address does not show a login gate. Remote binds still require authentication. Set `ARTEX_AUTH_MODE=required` to force login, or use `ARTEX_AUTH_MODE=disabled` only on a trusted local machine.

The desktop SQLite migration is not yet claimed as complete. The current data layer contains PostgreSQL-specific schema, types, locking, array/JSON operators, and transaction queries across the `db` package. A real SQLite desktop mode requires a dialect/storage abstraction and migration tests; until that work is finished, Docker/PostgreSQL remains a development/runtime dependency for this branch.
