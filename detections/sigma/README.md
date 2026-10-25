# ARGUS detection rules (Sigma / host · log · SIEM)

English · [中文](README.md)

> 中文: 这个目录是 [防御·检测指南(docs/defense-en.md)](../../docs/defense-en.md) 4弓
> "检测规则"实际可部署的伪规则 [Sigma](https://sigmahq.io) 这已被转化为规则.
> 网络层是 [`../suricata/`](../suricata/)负责. 所有规则都是您自己制定的或书面的
> 保护许可系统 **防御·仅用于检测目的** 请使用. 完整的中文字符文件是
> **[README.md](README.md)** 到, 完整的检测套件概述是 **[../README.md](../README.md)** 参见.

The host, log, and SIEM layer of the ARGUS detection set. These [Sigma](https://sigmahq.io) rules
formalize the pseudo-rules in the defense guide ([Chinese](../../docs/defense-en.md) ·
[English](../../docs/defense-en.md), section 4) into a vendor-neutral format you convert to your own
SIEM or EDR query language. Every indicator is grounded in a string or behaviour verified in this
repository's source, not inferred. The network layer lives under [`../suricata/`](../suricata/); the
full detection set — the ATT&CK coverage layer, the indicator CSV / MISP export, and the host-triage
script — is indexed in [`../README.md`](../README.md).

## Atomic rules

One rule, one observable fact. Convert them individually or as part of the whole tree.

- **[`argus_enrich_user_agent.yml`](argus_enrich_user_agent.yml)** — *ARGUS Asset Enrichment Probe
  User-Agent*. Inbound `argus-enrich/1.0` User-Agent from asset enrichment (`enrich/enrich.go`).
  Target-side, supporting indicator. `level: high`.
- **[`argus_selfupdate_egress.yml`](argus_selfupdate_egress.yml)** — *ARGUS Self-Update Egress
  User-Agent*. Outbound `argus-selfupdate` User-Agent from the self-update routine
  (`selfupdate/github.go`). Host/forensic egress indicator. `level: medium`.
- **[`argus_guard_audit_framing.yml`](argus_guard_audit_framing.yml)** — *ARGUS Platform Guard
  Audit-Log Framing*. The platform-guard control marker written to the audit log on a blocked tool
  call (`guard/guard.go`). Host/forensic indicator. `level: high`.
- **[`argus_recording_proxy_ca.yml`](argus_recording_proxy_ca.yml)** — *ARGUS Recording-Proxy MITM CA
  Certificate Artifact*. Creation of the recording proxy's MITM CA file under the
  `_ca/mitmproxy-ca-cert.pem` layout (`traffic/traffic.go`). Host/forensic artifact; the bare filename
  is shared with standalone mitmproxy, so it is a hunting lead. `level: medium`.
- **[`destructive_command_hunting.yml`](destructive_command_hunting.yml)** — *Destructive Command
  Execution (ARGUS Guard-List Hunting)*. Destructive shell/DB commands mirroring the ARGUS guard's
  built-in deny list (`db/db.go` seed). Generic hunting lead, **not** an ARGUS signature. `level: medium`.

## Correlation rules (behaviour) — [`correlation/`](correlation/)

Static strings can be changed; behaviour is harder to hide. These Sigma **correlation** rules encode the
behaviour-based layer of the defense guide (sections 4.1–4.2 and 4.4). Each references an atomic rule
above by its `id`, so **convert the whole `sigma/` tree, not a single correlation file**, or the
reference will not resolve (the [Sigma test](../tests/sigma/) asserts exactly this dependency).

- **[`correlation/argus_enrich_scan_velocity.yml`](correlation/argus_enrich_scan_velocity.yml)** —
  *Enrichment Scan Velocity*. A burst of `argus-enrich/1.0` probes from one source in a short window
  (enrichment runs at concurrency 4 with no rate limit) — the velocity the single-request rule misses.
  `event_count`, `level: high`.
- **[`correlation/argus_enrich_fanout.yml`](correlation/argus_enrich_fanout.yml)** — *Enrichment
  Fan-Out*. One source carrying the enrichment User-Agent to many *distinct* hosts: machine-speed breadth
  across an asset list, where the distinct-host count, not request volume, is the tell. `value_count`,
  `level: high`.
- **[`correlation/argus_guard_block_burst.yml`](correlation/argus_guard_block_burst.yml)** —
  *Guard-Block Burst*. Repeated platform-guard control markers on one host — an actively engaged ARGUS
  run tripping its own guard, not a document that merely quotes the marker. `event_count`, `level: high`.
- **[`correlation/argus_guard_marker_then_destructive.yml`](correlation/argus_guard_marker_then_destructive.yml)**
  — *Guard Marker With Destructive Command*. The guard marker and a destructive command co-occurring on
  one host within a window (defense guide §4.2, multi-stage): combining an ARGUS-specific marker with the
  otherwise-generic destructive-command signal raises specificity. `temporal`, `level: high`.

Thresholds and windows are conservative defaults — tune them to your baseline. The pure web multi-stage
case (enumerate → probe → authenticate) still needs base rules specific to your environment, because that
pattern does not reduce to a single ARGUS-unique User-Agent; a generic behavioural base template to start
from is in the [defense guide §4.2](../../docs/defense-en.md), kept out of this tested tree because it
cannot be grounded in ARGUS source.

## Scope and honesty — read before deploying

- **Static indicators can be changed.** An operator can set a different User-Agent or clean up the CA
  file, so the absence of an atomic indicator does **not** mean safety. The durable signal is the
  behaviour the `correlation/` rules key on — one source chaining recon → enumeration → probing →
  auth/injection attempts, adapting to responses, running without pause.
- **The destructive-command rule is generic hunting.** It mirrors ARGUS's guard deny list, but the same
  commands are run by legitimate administrators. Treat a hit as a lead, allow-list your environment, and
  do not attribute it to ARGUS on its own.
- **Ports and schema are host-forensic, not Sigma.** The server default `:8787` and recording proxy
  `127.0.0.1:8788` (`cmd/argus/main.go`), and the PostgreSQL exploration-graph schema, are best checked on
  a suspected host, so they ship in the [indicator CSV](../indicators/) and the
  [host-triage script](../triage/) rather than as noisy rules.
- **`logsource` and field names are generic.** The rules use generic `category`/`product` log sources and
  field names (`cs-user-agent`, `CommandLine`, `TargetFilename`). Map them to your product's schema with a
  pipeline (`-p`) at convert time; see the backend notes below.

## Validate and convert

Validated with [sigma-cli](https://github.com/SigmaHQ/sigma-cli) (pySigma). From the repository root:

```sh
python3 -m venv .venv && . .venv/bin/activate
pip install sigma-cli

# structural + best-practice validation (expect: 0 errors, 0 issues)
sigma check detections/sigma/

# full SigmaHQ convention set with this rule set's documented baseline (expect: 0 issues)
pip install pySigma-validators-sigmahq
sigma check --validation-config detections/tests/sigma_lint/validators.yml detections/sigma/

# compile the WHOLE tree so the correlation rules resolve the atomic rules they reference by id
sigma plugin install splunk
sigma convert -t splunk --without-pipeline detections/sigma/
```

Backends vary in correlation support, so the `-t` choice matters: Splunk, Elasticsearch EQL, and Grafana
Loki convert the whole tree, while Elasticsearch Lucene, OpenSearch, and the Microsoft `kusto` backend
convert the five atomic rules only (express the window natively in the product). The measured per-backend
matrix and the `--without-pipeline` / `-p` field-mapping notes are in [`../README.md`](../README.md), and
they are reproduced by [`../tests/sigma_backends/`](../tests/sigma_backends/).

## Tests

Four reproducible suites under [`../tests/`](../tests/) cover these rules, each needing only Docker:
[`sigma/`](../tests/sigma/) (validation, whole-tree compilation, and that a correlation rule fails to
convert alone), [`sigma_match/`](../tests/sigma_match/) (the rules actually fire on malicious samples and
stay quiet on benign ones), [`sigma_backends/`](../tests/sigma_backends/) (portability across five
backends), and [`sigma_lint/`](../tests/sigma_lint/) (the full SigmaHQ validator baseline, 0 issues). See
[`../tests/README.md`](../tests/README.md).

## Contributing

Detection contributions are welcome. New rules should keep every indicator grounded in an observable fact,
state limitations in the `description`, pass the SigmaHQ validator baseline cleanly
(`sigma check --validation-config ../tests/sigma_lint/validators.yml .`), and avoid any content that reads
as attack guidance. See [`../../CONTRIBUTING.en.md`](../../CONTRIBUTING.en.md) and the network layer in
[`../suricata/`](../suricata/) / [`../README.md`](../README.md).
