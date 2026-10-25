# Maintainer Notes

Argus's public documentation is available in Chinese and English only. New user-facing text must be added to both locale resources.

## Release checks

- Run `go test ./...`.
- Run the frontend TypeScript, i18n-key, and production build checks.
- Run `git diff --check`.
- Confirm release archives contain no `.env`, secrets, databases, or local data.
- Verify amd64 (x86-64) packages for Windows, macOS, and Linux; ARM64 packages are optional.

## Security and data

Never commit API keys, database files, user sessions, target data, or runtime logs. Security testing must enforce the authorized Target/Test Range boundary.

Chinese version: [MAINTAINING.md](MAINTAINING.md)
