#!/usr/bin/env bash
set -euo pipefail

# Cloud deployment entrypoint. The platform supplies PORT at runtime.
: "${PORT:?PORT must be set by the hosting platform}"
echo "[*] Starting Argus on port :$PORT"
exec ./argus -addr ":$PORT"
