#!/usr/bin/env python3
"""Start the unmodified release launcher with no arguments or Argus settings."""
import json
import os
from pathlib import Path
import re
import socket
import subprocess
import sys
import time
import urllib.error
import urllib.request

root = Path(sys.argv[1]).resolve()
macos = root.suffix == ".app"
launcher = root / "Contents/MacOS/Argus" if macos else root / "start.sh"
user_root = Path.home() / "Library/Application Support/Argus" if macos else Path.home() / ".local/share/Argus"
console = Path.home() / "argus-launcher-test.log"
title = "原始启动器：中文数据持久化"
process = None

def request(path, data=None):
    payload = json.dumps(data, ensure_ascii=False).encode() if data is not None else None
    req = urllib.request.Request("http://127.0.0.1:8787" + path, data=payload, headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=3) as response:
        assert response.status == 200, (path, response.status)
        return response.read()

def start(log):
    global process
    try:
        connection = socket.create_connection(("127.0.0.1", 8787), timeout=1)
    except OSError:
        pass
    else:
        connection.close()
        raise RuntimeError("Test port 8787 is already in use; refusing to touch another service")
    # No ARGUS_HOME, DSN, auth override, or command-line flags are supplied.
    process = subprocess.Popen([str(launcher)], cwd=root, stdout=log, stderr=subprocess.STDOUT)
    deadline = time.monotonic() + 90
    while time.monotonic() < deadline:
        if process.poll() is not None:
            raise RuntimeError(f"Launcher exited: {process.returncode}")
        try:
            health = json.loads(request("/api/health"))
            if health.get("ok") and health.get("service") == "Argus":
                return
        except (OSError, urllib.error.URLError):
            pass
        time.sleep(0.5)
    raise RuntimeError("Default launcher did not become ready")

def stop():
    if process is not None and process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=30)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
            raise RuntimeError("Launcher did not stop gracefully")

try:
    with console.open("wb") as log:
        start(log)
        assert (user_root / "startup.log").is_file(), user_root
        assert (user_root / "data/database/data/PG_VERSION").is_file(), user_root
        assert not (root / "config.json").exists(), "Configuration written into the installation"
        assert not (root / "jwt.key").exists(), "JWT key written into the installation"
        html = request("/").decode()
        assert "<html" in html.lower()
        scripts = re.findall(r'src="([^\"]+\.js[^\"]*)"', html)
        assert scripts
        for script in scripts[:5]:
            assert len(request(script)) > 100
        agents = json.loads(request("/api/agents"))["agents"]
        agent = next((a for a in agents if a["key"] in ("main", "mainagent")), agents[0])
        created = json.loads(request("/api/conversations", {"agent_key": agent["key"], "title": title}))
        assert created["title"] == title
        stop()
        start(log)
        conversations = json.loads(request("/api/conversations"))
        if isinstance(conversations, dict):
            conversations = conversations["conversations"]
        assert any(c["title"] == title for c in conversations)
        stop()
    print(json.dumps({"result": "PASS", "launcher": str(launcher), "arguments": [],
                      "default_user_data_directory": str(user_root), "html_and_js": True,
                      "chinese_data_after_restart": True, "external_database": False}, ensure_ascii=False))
except Exception:
    stop()
    if console.exists():
        print(re.sub(r"(设置令牌: ).*", r"\1[REDACTED]", console.read_text(errors="replace")[-12000:]), file=sys.stderr)
    raise
