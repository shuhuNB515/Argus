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

- Windows Intel/AMD PCs: `argus-*-windows-amd64-setup.exe`
- macOS Intel: `argus-*-darwin-amd64.dmg`
- macOS Apple Silicon (including M5): `argus-*-darwin-arm64.dmg`
- Linux x86-64: `argus-*-linux-amd64.zip`
- Linux ARM64: `argus-*-linux-arm64.zip`

`amd64` means x86-64. Windows x86-64 computers should use the Windows amd64 package; ARM64 packages are only for ARM devices.

## Run locally

Install the package and launch Argus from its application icon. Linux packages also include `start.sh`. All release packages include their database runtime; even the first startup works without downloading a database. No Docker, database account, password, path flags, or configuration file is required. The workbench opens at <http://127.0.0.1:8787>.

Data and startup logs are stored in a writable directory for the current user:

- Windows: `%LOCALAPPDATA%\\Argus`
- macOS: `~/Library/Application Support/Argus`
- Linux: `$XDG_DATA_HOME/Argus`, or `~/.local/share/Argus`

Configure your own model provider and API key in the application before starting Agent tasks. Credentials are never bundled with installers.

For the optional Docker deployment, run this in the repository directory:

```bash
docker compose up -d
```

No `.env` file or separate database service is needed. The image includes its database runtime, an ordinary system user, and directory permissions. Compose persists data in a named volume and exposes the workbench on localhost.

Source development uses `./dev.sh` and opens <http://127.0.0.1:8788>. Advanced deployments may still select an existing PostgreSQL server with `ARGUS_PG_DSN`.

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
