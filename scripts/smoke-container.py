#!/usr/bin/env python3
"""Verify a production image starts offline using only its baked-in defaults."""
import json
import subprocess
import sys
import time
import uuid

name = "argus-default-" + uuid.uuid4().hex[:12]

def docker(*args):
    return subprocess.check_output(["docker", *args], text=True, stderr=subprocess.STDOUT).strip()

probe = """
import json, os, pathlib, urllib.request
assert os.getuid() != 0
assert os.environ['ARGUS_HOME'] == '/data'
assert pathlib.Path('/data/data/database/data/PG_VERSION').is_file()
assert pathlib.Path('/data/startup.log').is_file()
with urllib.request.urlopen('http://127.0.0.1:8787/api/health', timeout=2) as r:
    assert json.load(r)['ok']
with urllib.request.urlopen('http://127.0.0.1:8787/', timeout=2) as r:
    assert b'<html' in r.read().lower()
"""

def ready():
    deadline = time.monotonic() + 100
    while time.monotonic() < deadline:
        status = docker("inspect", "--format", "{{.State.Status}}", name)
        if status != "running":
            raise RuntimeError("Default image stopped before becoming ready: " + status)
        try:
            docker("exec", name, "python", "-c", probe)
            return
        except subprocess.CalledProcessError:
            time.sleep(1)
    raise RuntimeError("Default image did not become ready")

try:
    # No application flags, application environment, host paths, or user overrides.
    docker("run", "--network", "none", "--name", name, "-d", sys.argv[1])
    ready()
    docker("exec", name, "python", "-c", "from pathlib import Path; Path('/data/persistence-check').write_text('中文数据')")
    docker("restart", name)
    ready()
    docker("exec", name, "python", "-c", "from pathlib import Path; assert Path('/data/persistence-check').read_text() == '中文数据'")
    print(json.dumps({"result": "PASS", "image": sys.argv[1], "application_arguments": [],
                      "application_environment_overrides": {}, "network": "none",
                      "non_root": True, "database_and_html": True, "restart_persistence": True}))
finally:
    subprocess.run(["docker", "rm", "-f", "-v", name], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
