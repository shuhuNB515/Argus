#!/bin/sh
# Argus launcher for Linux, macOS command-line packages, and Docker.
#
# Usage:
#   ./start.sh                         Run in the foreground.
#   nohup ./start.sh >argus.log 2>&1 & Run in the background.
#   ./start.sh -addr :9000             Pass extra arguments to Argus.
#
# Exit code 75 means Argus requested a restart after an update or rollback.
# Startup errors stop with a nonzero status and a log location.
set -u

cd "$(dirname "$0")" || exit 1

BIN=./argus
[ -x "$BIN" ] || { echo "[Argus] executable not found: $BIN" >&2; exit 1; }

RESTART_CODE=75

child=0
stopping=0

forward() {
	stopping=1
	if [ "$child" -ne 0 ]; then
		kill -TERM "$child" 2>/dev/null || true
	fi
}
trap forward INT TERM

while :; do
	"$BIN" "$@" &
	child=$!

	wait "$child"
	code=$?
	if [ "$code" -gt 128 ]; then
		wait "$child"
		code=$?
	fi
	child=0

	if [ "$stopping" -eq 1 ]; then
		echo "[Argus] stopped"
		exit 0
	fi

	case "$code" in
		0)
			echo "[Argus] exited normally"
			exit 0
			;;
		"$RESTART_CODE")
			echo "[Argus] restart requested; applying the new version"
			;;
		*)
			echo "[Argus] startup failed (code=$code)" >&2
			echo "[Argus] See the startup.log in your Argus user data directory." >&2
			exit "$code"
			;;
	esac
done
