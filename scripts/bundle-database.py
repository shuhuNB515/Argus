#!/usr/bin/env python3
"""Download and checksum-verify the database runtime while building a release."""
import argparse
import hashlib
import io
import os
from pathlib import Path
import urllib.request
import zipfile

VERSION = "18.3.0"  # embedded-postgres v1.34.0 / V18

def bundle(goos, arch, destination):
    if arch == "arm64" and goos in ("linux", "darwin"):
        arch = "arm64v8"
    artifact = f"embedded-postgres-binaries-{goos}-{arch}"
    url = f"https://repo.maven.apache.org/maven2/io/zonky/test/postgres/{artifact}/{VERSION}/{artifact}-{VERSION}.jar"
    with urllib.request.urlopen(url, timeout=120) as response:
        jar = response.read()
    with urllib.request.urlopen(url + ".sha256", timeout=30) as response:
        expected = response.read().decode("ascii").strip().lower()
    if hashlib.sha256(jar).hexdigest() != expected:
        raise RuntimeError("Database archive SHA256 mismatch")
    with zipfile.ZipFile(io.BytesIO(jar)) as archive:
        entries = [entry for entry in archive.infolist() if not entry.is_dir() and entry.filename.endswith(".txz")]
        if len(entries) != 1:
            raise RuntimeError("Expected one database runtime archive")
        content = archive.read(entries[0])
    destination = Path(destination)
    destination.mkdir(parents=True, exist_ok=True)
    target = destination / f"{artifact}-{VERSION}.txz"
    temporary = target.with_suffix(".tmp")
    temporary.write_bytes(content)
    os.replace(temporary, target)
    print(f"Bundled offline database runtime: {target}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--os", required=True, choices=("linux", "darwin"))
    parser.add_argument("--arch", required=True, choices=("amd64", "arm64"))
    parser.add_argument("--dest", required=True)
    args = parser.parse_args()
    bundle(args.os, args.arch, args.dest)
