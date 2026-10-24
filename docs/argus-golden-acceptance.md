# Argus SRC Golden Acceptance

Date: 2026-10-24

## Scope

- Project: `/Users/shu/Documents/src-agent-artex-harness-v2`
- Workbench: `http://127.0.0.1:8788`
- Backend: `127.0.0.1:8787`
- Harness HEAD: `ca6d869` (`replace legacy icons with Argus branding`), with the working tree intentionally containing the RC fixes listed below.
- Authorized target: `https://shu560.cn/`
- Test Range: `shu560.cn`
- Ports: `80,443`
- Subdomains, Cloudflare shared IPs, third-party redirects, state-changing requests, destructive actions, and high-concurrency probing were excluded.

## RC fixes verified in code and runtime

- The SRC `TargetContext` is persisted per task and reloaded by the executor. URL, host, port, scheme, and range checks use one canonical authorization path. Exact-host mode rejects subdomains; an explicit `*.domain` range is required to allow them. Network-capability wrappers fail closed, including Bash/WebFetch/browser/HTTP/recon/crawler/probe families. SRC disables the unconstrained remote search path.
- Conversation history, task streams, detail responses, and public runtime events omit private reasoning and usage details. Public operational text, tool events, file/test/browser events, and aggregate metrics remain available where applicable.
- AOCI Whole-Index is injected into the runtime only when aligned. Current runtime status is `aligned`, root `/Users/shu/Documents/src-agent-artex-harness-v2`, 798 files, and fingerprint `be88ae99a9b5d5af79e93f9e1c7da31e8ac35230e2afc3cf7d1dd5628cbad427`.
- Browser evidence is task-scoped and records task/run/node/URL/timestamp, page state, provenance, and screenshot data. The evidence endpoint rejects URLs outside the persisted SRC context.
- Reports are rendered from the current task's findings/evidence, not a cached demonstration report. Non-confirmed observations remain candidates or informational observations.

## Real target evidence

Task `4`, conversation `7`, title `SRC Golden E2E shu560.cn` was created with the structured target context. Read-only Chrome/Playwright research reached the exact authorized host and returned HTTP 200. No form submission, login, mutation, third-party navigation, or redirect interaction was performed. A page-state plus screenshot evidence record was persisted with browser provenance.

The current task report contains zero confirmed findings and one browser evidence reference. This is a real zero, not a test fixture finding.

## Golden E2E result

**FAIL / BLOCKED, not PASS.** The orchestration request was accepted, but the configured model provider returned HTTP 402 (`Insufficient account balance`) before the first model tool call. Therefore the live Agent stages Recon through Report Quality Gate did not complete in this run, and no asset/finding/tool-call success was fabricated. The independent safe browser evidence and report rendering paths did complete.

After the provider balance/configuration is restored, rerun the same task context and verify the complete agent chain before calling this acceptance PASS.

## Artifacts

- Markdown: `data/tasks/4/reports/src-report.md`
- HTML: `data/tasks/4/reports/src-report.html`
- PDF: `data/tasks/4/reports/src-report.pdf` (validated `%PDF-1.4` header)
- Metadata: `data/tasks/4/reports/src-report.metadata.json`
- Browser evidence: `data/tasks/4/evidence/browser/20261024T062122.745000000Z.json` and its `.png`

## Verification

- `go test ./...`: PASS
- Frontend `npx tsc --noEmit`: PASS
- Frontend `npm run build`: PASS
- `git diff --check` excluding generated `aoci.code.txt`: PASS
- Full `git diff --check`: reports trailing whitespace in generated AOCI code content.
- Frontend `npm run check`: existing repository-wide Biome baseline reports 30 errors, 124 warnings, and 666 infos; no bulk formatting rewrite was applied.

The service remains running on `127.0.0.1:8788` for manual follow-up.
