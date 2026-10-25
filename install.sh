#!/usr/bin/env bash
set -euo pipefail
cd "$(cd "$(dirname "$0")" && pwd)"

echo "Argus source installer"
echo "This path builds a local development copy. Release users should open the packaged app."

command -v go >/dev/null 2>&1 || { echo "Go 1.26 or newer is required." >&2; exit 1; }
if command -v npm >/dev/null 2>&1; then
  (cd web && npm ci && npm run build:static)
  rm -rf server/webui/dist
  mkdir -p server/webui/dist
  cp -R web/out/. server/webui/dist/
  CGO_ENABLED=0 go build -tags embedui -trimpath -o argus ./cmd/argus
else
  echo "npm is not installed; building the backend without embedded UI assets." >&2
  CGO_ENABLED=0 go build -trimpath -o argus ./cmd/argus
fi

chmod +x argus
echo "Argus is ready. Start it with ./argus and open http://127.0.0.1:8787."
echo "The desktop runtime uses its built-in local database; Docker and PostgreSQL are not required."
