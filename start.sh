#!/bin/sh
# Argus launcher for Linux, macOS command-line packages, and Docker.
#
# Usage:
#   ./start.sh                         Run in the foreground.
#   nohup ./start.sh >argus.log 2>&1 & Run in the background.
#   ./start.sh -addr :9000             Pass extra arguments to Argus.
#
# Exit code 75 means Argus requested a restart after an update or rollback.
# Other unexpected exits are retried with an exponential backoff.
set -u

cd "$(dirname "$0")" || exit 1

BIN=./argus
[ -x "$BIN" ] || { echo "[Argus] executable not found: $BIN" >&2; exit 1; }

RESTART_CODE=75
MAX_DELAY=60

child=0
stopping=0

forward() {
	stopping=1
	if [ "$child" -ne 0 ]; then
		kill -TERM "$child" 2>/dev/null || true
	fi
}
trap forward INT TERM

delay=1
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
			delay=1
			;;
		*)
			echo "[Argus] unexpected exit (code=$code); restarting in ${delay}s" >&2
			sleep "$delay"
			delay=$((delay * 2))
			[ "$delay" -gt "$MAX_DELAY" ] && delay=$MAX_DELAY
			;;
	esac
done
