# ARTEX/DeepSeek SRC Agent Rebuild Plan

**Goal:** Use ARTEX Global as the Go security workflow core, DeepSeek Harness as the runtime and UI interaction model, and preserve the SRC Agent two-level rail/sidebar with mode-aware authorization.

**Architecture:** ARTEX remains the domain engine for tasks, assets, findings, evidence, graph state, reports, MCP, and multi-agent orchestration. A Harness adapter owns sessions, provider-native tool calls, public events, plugin registration, and run control. The Next.js shell keeps DeepSeek visual language while the existing two-column rail plus secondary sidebar remains the navigation contract.

**Upstream:** ARTEX Global `fdfb7a17b7da5e2de716c0e98475ec77a41c64c6`; DeepSeek Harness master `687f282e8625c6b8d11ff58211a90524be2e4317`; AOCI-CODE main `536f3cc42339d1cb5f554026f84d3f2ed0e12d2a`.

## Invariants

- `/Users/shu/Documents/src-agent` remains read-only.
- No global Scope product concept; Target + Test Range appear only in SRC.
- CTF binds only an explicit CTF target/context; ANALYSIS has no target UI; LOCAL_LAB binds an explicit local target.
- Text XML, markdown shell, and unsupported provider text never execute as tools.
- Scanner and intelligence results remain potential until validation, impact evidence, provenance, and authorization checks pass.
- The two-level rail/sidebar is retained and branded as SRC Agent.
- New AppData and ports are isolated from the existing workbench.

## Delivery slices

1. Vendor and verify ARTEX, Harness, and AOCI commits; map ARTEX packages and existing SRC policies.
2. Add Go runtime seams for ModeContext, TargetContext, AuthorizationContext, canonical ToolCall, public Event Stream, and provider adapters.
3. Wire Harness session/run lifecycle to ARTEX task graph and checkpoint/recovery services.
4. Port the two-level rail/sidebar into the DeepSeek/Next shell; preserve grouped recent sessions and mode-specific target controls.
5. Connect ARTEX findings/evidence/report graph to SRC Agent report profiles and quality gates.
6. Add AOCI whole-index context resolver and stale/aligned maintenance state.
7. Run Go tests, web build/lint/typecheck, policy golden tests, and isolated startup smoke tests.

## Review focus

- An unsupported provider text snippet must remain text and emit no tool event.
- A SRC call outside Test Range must fail closed before ARTEX executes it.
- CTF without an explicit target must not gain unrestricted network access.
- Approved reports must remain immutable while new evidence creates a new version.
- The rail/sidebar must remain two-level after routing and responsive layout changes.
