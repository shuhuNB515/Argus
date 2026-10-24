# ARTEX + DeepSeek Harness + SRC Agent integration

This worktree is based on ARTEX Global `fdfb7a17b7da5e2de716c0e98475ec77a41c64c6`.
The runtime and UI contract follows DeepSeek Harness master
`687f282e8625c6b8d11ff58211a90524be2e4317`. AOCI-CODE main is
`536f3cc42339d1cb5f554026f84d3f2ed0e12d2a`.

ARTEX remains responsible for task graphs, assets, findings, evidence, reports,
MCP, skills, and agent orchestration. `srcagentpolicy` is the policy seam that
binds each ARTEX run to a mode-specific TargetContext. Provider adapters and
public runtime events will be attached at the ARTEX agent boundary in the next
slice; text-only tool syntax is never interpreted as an executable call.

The web shell keeps the ARTEX routes and adds the SRC Agent two-level layout:
the narrow left rail is the stable top-level navigation, and ARTEX's existing
sidebar becomes the secondary workspace/session navigation. Target + Test Range
controls are reserved for SRC; CTF shows only its explicit challenge target;
ANALYSIS has no target control.

The project uses independent AppData and ports. No data is imported from the
old `/Users/shu/Documents/src-agent` project.
