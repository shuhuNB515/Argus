#!/usr/bin/env python3
"""Reject legacy branding and unsupported Korean text in publishable text files."""

from pathlib import Path
import subprocess
import sys


root = Path(__file__).resolve().parents[1]
tracked = subprocess.check_output(
    ["git", "ls-files", "-co", "--exclude-standard", "-z"], cwd=root
).split(b"\0")
paths = [root / item.decode() for item in tracked if item]
ignored_suffixes = {".png", ".jpg", ".jpeg", ".gif", ".ico", ".zip", ".dmg", ".exe", ".pdf"}
legacy = []
korean = []

for path in paths:
    if path.suffix.lower() in ignored_suffixes or not path.is_file():
        continue
    try:
        text = path.read_text(encoding="utf-8")
    except (UnicodeDecodeError, OSError):
        continue
    legacy_token = "art" + "ex"
    if legacy_token in text.lower():
        legacy.append(str(path.relative_to(root)))
    if any("\uac00" <= char <= "\ud7a3" for char in text):
        korean.append(str(path.relative_to(root)))

if legacy or korean:
    if legacy:
        print("legacy branding found:")
        print("\n".join(f"  {item}" for item in legacy))
    if korean:
        print("Korean text found:")
        print("\n".join(f"  {item}" for item in korean))
    sys.exit(1)

print(f"brand/language check passed ({len(paths)} files scanned)")
