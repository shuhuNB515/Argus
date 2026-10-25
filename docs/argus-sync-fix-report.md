# Argus Synchronization Fix Report

Date: 2026-10-24

This report records the current RC synchronization fixes in `/Users/shu/Documents/src-agent-argus-harness-v2`. The legacy project `/Users/shu/Documents/src-agent` was left untouched.

## Fixed synchronization paths

- SRC structured target, range, and ports are saved in `task_target_context` and reloaded by the executor rather than inferred from display text.
- Task, conversation, execution intent, tool events, evidence, findings, and reports use the current task association.
- The canonical target authorization check is applied to all registered network-capability families. Out-of-range host, subdomain, protocol, port, and URL inputs fail closed and are sanitized into operational events.
- Public task/conversation/detail streams no longer transport private reasoning or usage details.
- AOCI runtime context refuses stale indexes and records the aligned runtime root/fingerprint, including an external active project path.
- Browser evidence is stored as a task-scoped deliverable with screenshot/page-state provenance.
- Report generation separates confirmed findings from candidates and observations and writes Markdown/HTML/PDF/metadata from current task data.
- Side-question snapshots normalize mixed parent identifiers before persistence.
- Local desktop authentication now defaults to loopback-only no-login mode; remote binds remain protected, with `ARGUS_AUTH_MODE=required|disabled` available for explicit deployments.
- The desktop UI normalizes stale Chinese locale preferences to Chinese, migrates generated `新对话` titles to `新对话`, and localizes the selected-LLM/API-key validation error.
- On startup, an environment LLM configuration is promoted into the active database profile when the seeded profile has no key, so a fresh local install can use the configured provider without manually recreating the profile.

## Real verification result

The authorized `shu560.cn` browser read-only check completed and produced task-scoped evidence. The current report has 0 confirmed findings. The orchestration request itself was blocked before tool execution by the model provider's HTTP 402 insufficient-balance response; this is the only remaining functional Golden E2E blocker identified in this run.

The service is left running at `http://127.0.0.1:8788`.
