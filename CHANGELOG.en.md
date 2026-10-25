# Argus Changelog

[中文](CHANGELOG.md) · [English](CHANGELOG.en.md)

## v0.1

Argus v0.1 is the first public desktop release. Simplified Chinese is the default, with a complete English interface and English documentation.

### Desktop applications

- Release packages for macOS Apple Silicon, macOS Intel, Windows x86-64, and Linux.
- DMG installation on macOS, an EXE installer on Windows, and portable archives on Linux.
- Desktop builds keep data in the current user's application-data directory and do not require Docker or manual database setup.
- Users configure their own OpenAI-compatible model service from the interface.

### Authorized security research

- Target/Test Range controls for protocols, ports, hosts, and paths.
- A single authorization check runs before every network-capable operation and fails closed for out-of-scope requests.
- Asset, traffic, browser-evidence, correlation, evidence-quality, and report-quality workflows.
- Markdown, HTML, and PDF reports with traceable evidence references.
- Unvalidated observations are never promoted automatically to confirmed vulnerabilities.

### Product and documentation

- Product, package, internal service, script, and release names are consistently Argus.
- User-facing languages are limited to Simplified Chinese and English, with Chinese as the default.
- README, contribution guide, security policy, and code of conduct are available in Chinese and English.
- Chat messages include copy controls, and model, scope, and runtime failures provide clear diagnostics.

### Security notice

Use Argus only against assets you own, explicitly authorized scopes, or isolated local labs. Unauthorized scanning, destructive testing, denial of service, data theft, and privilege abuse are prohibited.
