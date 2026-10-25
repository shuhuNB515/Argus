# Argus

A local Agent workbench for authorized security research and SRC vulnerability validation.

> Use only against systems you own or are explicitly authorized to test. Do not use it for unauthorized scanning, destructive testing, denial of service, payment actions, or account/permission changes.

## Features

- SRC sessions with persisted Target, Test Range, and port authorization.
- Asset registration, passive reconnaissance, fingerprinting, and scope analysis.
- Evidence-first findings: potential issues require validation, impact evidence, provenance, and quality gates before confirmation.
- Read-only browser research with page state, screenshots, traffic observations, and evidence correlation.
- Evidence quality, deduplication, conflict handling, and Markdown/HTML/PDF reports.
- Agent tool calls, failure recovery, execution events, task graphs, and live work streams.
- AOCI runtime indexing for fast architecture context.
- OpenAI-compatible model profiles, connection tests, fallback profiles, and usage records.
- Chinese and English UI support.

## Downloads

Open [Releases](https://github.com/shuhuNB515/Argus/releases) and download the package for your system.

- Windows Intel/AMD PCs: `argus-*-windows-amd64.zip`
- macOS Intel: `argus-*-darwin-amd64.zip`
- macOS Apple Silicon: `argus-*-darwin-arm64.zip`
- Linux x86-64: `argus-*-linux-amd64.zip`
- Linux ARM64: `argus-*-linux-arm64.zip`

`amd64` means x86-64. Windows x86-64 computers should use the Windows amd64 package; ARM64 packages are only for ARM devices.

## Run locally

The desktop package starts a private embedded PostgreSQL instance automatically for task, conversation, asset, evidence, and report data. Users do not need Docker, a database server, or a DSN configuration. On first launch Argus downloads and caches the platform database runtime, then reuses the persistent local data directory:

```bash
cp .env.example .env
# Only edit .env when running the developer Docker environment.
./dev.sh
```

Open <http://127.0.0.1:8788> in a browser.

Docker Compose remains available for development. A desktop package does not require Docker. Existing PostgreSQL deployments may still be selected with `ARGUS_PG_DSN`.

## Build from source

```bash
cd web
npm ci
npm run build:static
cd ..
./build.sh --release --no-compress
```

The build script creates Windows, macOS, and Linux packages for amd64 and arm64, plus `SHA256SUMS`.

## Verification

```bash
go test ./...
cd web && npx tsc --noEmit && npm run check:i18n && npm run build:static
```

## License

This project remains licensed under AGPL-3.0. See [LICENSE](LICENSE) and [SECURITY.md](SECURITY.md).
