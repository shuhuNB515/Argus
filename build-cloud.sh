#!/usr/bin/env bash
set -euo pipefail

echo "[*] Building the frontend..."
(cd web && npm ci && npm run build:static)

echo "[*] Embedding frontend assets..."
rm -rf server/webui/dist
mkdir -p server/webui/dist
cp -R web/out/. server/webui/dist/

echo "[*] Compiling the Argus backend..."
CGO_ENABLED=0 go build -tags embedui -trimpath -o argus ./cmd/argus
echo "[+] Argus build complete."
