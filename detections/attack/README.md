# ARGUS ATT&CK coverage

English · [中文](README.md)

> 中文: 这个目录是 [`../`](../)的 ARGUS 检测规则(Sigma·Suricata)这种攻击技术是
> [MITRE ATT&CK](https://attack.mitre.org/) 战术·按技术组织 **覆盖层**是.
> 每种技术都有一组基于存储库源的规则。 `attack.*` 我刚从标签上得到它, 包含在估算中的项目
> 无. [ATT&CK Navigator](https://mitre-attack.github.io/attack-navigator/)按原样穿上
> 其中 ARGUS 您可以一目了然地看到哪些规则适用于您的操作. This layer is owned or written by you.
> 保护许可系统 **防御·仅用于检测目的** 写. 完整的中文字符文件是
> **[README.md](README.md)** 参见.

A [MITRE ATT&CK](https://attack.mitre.org/) Navigator layer that maps the detection rules in this
repository to the ATT&CK (Enterprise) techniques they tag. It is built by hand from the `attack.*` tags on
the [Sigma rules](../sigma/) — every technique is grounded in a rule whose indicator is a string or
behaviour verified in this repository's source, and the [consistency test](../tests/attack/run.sh)
keeps the layer and the rules from drifting apart.

- **`argus_navigator_layer.json`** — the layer, in ATT&CK Navigator v4.5 format.

## What the score means

Coverage here means "this repository ships a detection that tags this technique", not "this technique is
fully covered". The score is deliberately honest about detection strength:

- **100 — ARGUS-specific signature or behaviour.** A static indicator unique to ARGUS (the
  `argus-enrich/1.0` / `argus-selfupdate` User-Agents, the guard audit marker) or a behaviour rule built
  on one (enrichment velocity / fan-out, guard-block burst).
- **50–65 — generic hunting lead.** Destructive-command hunting mirrored from the ARGUS guard deny list.
  The same commands are run by legitimate administrators, so these fire on benign activity too; treat a
  hit as a lead, not an attribution. 65 marks the case where a correlation rule raises specificity by
  pairing the command with the ARGUS guard marker.

## Techniques covered

Eight techniques across six tactics. Each maps to the rule(s) that tag it:

- **Reconnaissance — T1595 (Active Scanning), T1592 (Gather Victim Host Information).**
  [`sigma/argus_enrich_user_agent.yml`](../sigma/argus_enrich_user_agent.yml),
  [`sigma/correlation/argus_enrich_scan_velocity.yml`](../sigma/correlation/argus_enrich_scan_velocity.yml),
  [`sigma/correlation/argus_enrich_fanout.yml`](../sigma/correlation/argus_enrich_fanout.yml), and the
  [Suricata rules](../suricata/argus.rules) (sid 1000001 / 1000002).
- **Command and Control — T1105 (Ingress Tool Transfer).**
  [`sigma/argus_selfupdate_egress.yml`](../sigma/argus_selfupdate_egress.yml).
- **Execution — T1059 (Command and Scripting Interpreter).**
  [`sigma/argus_guard_audit_framing.yml`](../sigma/argus_guard_audit_framing.yml),
  [`sigma/correlation/argus_guard_block_burst.yml`](../sigma/correlation/argus_guard_block_burst.yml),
  [`sigma/correlation/argus_guard_marker_then_destructive.yml`](../sigma/correlation/argus_guard_marker_then_destructive.yml).
- **Impact — T1485 (Data Destruction), T1561.002 (Disk Wipe: Disk Structure Wipe), T1489 (Service Stop).**
  [`sigma/destructive_command_hunting.yml`](../sigma/destructive_command_hunting.yml), with T1485 also
  reinforced by
  [`sigma/correlation/argus_guard_marker_then_destructive.yml`](../sigma/correlation/argus_guard_marker_then_destructive.yml).
- **Credential Access / Collection — T1557 (Adversary-in-the-Middle).**
  [`sigma/argus_recording_proxy_ca.yml`](../sigma/argus_recording_proxy_ca.yml) — the MITM root-CA artifact
  ARGUS's embedded traffic recorder installs (`traffic/traffic.go`) to decrypt and log the worker tools'
  traffic. A host/forensic hunting lead.

## How to use it

1. Open the [ATT&CK Navigator](https://mitre-attack.github.io/attack-navigator/).
2. Choose **Open Existing Layer → Upload from local**, and select `argus_navigator_layer.json` (or point
   it at the raw file URL from this repository).
3. The scored techniques appear colour-graded by detection strength, each with a comment naming the rule
   file(s) and the defense-guide section behind it.

## Scope and honesty

- **Coverage is not completeness.** A technique scored here means a rule tags it, not that every variant
  of the technique is detected. Only two ARGUS-unique User-Agents are visible on the wire — the enrichment
  prober (`argus-enrich/1.0`) in the reconnaissance phase and the norma SDK WebFetch tool (`norma/0.4`) in
  the attack phase — while the rest of the attack traffic follows tool-default fingerprints; the durable
  detection is behavioural
  (see the defense guide, [Chinese](../../docs/defense-en.md) · [English](../../docs/defense-en.md), sections 1–2 and 4.1–4.2). The pure web multi-stage
  case still needs base rules specific to your environment.
- **Static indicators can be changed.** An operator can set a different User-Agent, so the absence of a
  tagged indicator does not imply safety. This is the same caveat the rule files carry.

## Validate and contribute

Run the [consistency test](../tests/attack/run.sh) — it needs only Docker and asserts that the layer's
scored techniques and tactics are exactly the `attack.*` tags on the rules, with every technique grounded
in a rule file that exists:

```sh
detections/tests/attack/run.sh
```

When you add or retag a rule, update this layer to match — the test fails if a rule technique is missing
from the layer or a layer technique is absent from the rules. See [`../README.md`](../README.md) and
[`../../CONTRIBUTING.en.md`](../../CONTRIBUTING.en.md).
