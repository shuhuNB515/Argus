# ARTEX

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

- Windows Intel/AMD PCs: `artex-*-windows-amd64.zip`
- macOS Intel: `artex-*-darwin-amd64.zip`
- macOS Apple Silicon: `artex-*-darwin-arm64.zip`
- Linux x86-64: `artex-*-linux-amd64.zip`
- Linux ARM64: `artex-*-linux-arm64.zip`

`amd64` means x86-64. Windows x86-64 computers should use the Windows amd64 package; ARM64 packages are only for ARM devices.

## Run locally

The current release uses PostgreSQL for task, conversation, asset, evidence, and report data. The simplest setup is Docker Desktop for PostgreSQL plus the local runtime:

```bash
cp .env.example .env
# Edit .env and set POSTGRES_PASSWORD and the model API configuration.
./dev.sh
```

Open <http://127.0.0.1:8788> in a browser.

Docker Compose provides the local PostgreSQL service. A fully self-contained desktop database mode is not yet released as a stable feature.

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
