#!/usr/bin/env python3
"""Scan repository text files for high-confidence secret patterns."""

from __future__ import annotations

import os
import re
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MAX_FILE_BYTES = 1_000_000

SECRET_PATTERNS: tuple[tuple[str, re.Pattern[str]], ...] = (
    (
        "private-key",
        re.compile(r"-----BEGIN (?:RSA |OPENSSH |EC |DSA )?PRIVATE KEY-----"),
    ),
    ("aws-access-key-id", re.compile(r"\b(?:A3T|AKIA|ASIA)[0-9A-Z]{16}\b")),
    ("google-api-key", re.compile(r"\bAIza[0-9A-Za-z_-]{35}\b")),
    ("github-token", re.compile(r"\bgh[pousr]_[A-Za-z0-9_]{30,}\b")),
    ("slack-token", re.compile(r"\bxox[baprs]-[A-Za-z0-9-]{20,}\b")),
    ("openai-project-key", re.compile(r"\bsk-proj-[A-Za-z0-9_-]{40,}\b")),
    ("openai-api-key", re.compile(r"\bsk-[A-Za-z0-9]{32,}\b")),
)


def list_candidate_files() -> list[Path]:
    result = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        cwd=ROOT,
        check=True,
        stdout=subprocess.PIPE,
    )
    files: list[Path] = []
    for raw_path in result.stdout.split(b"\0"):
        if not raw_path:
            continue
        path = ROOT / raw_path.decode("utf-8", errors="surrogateescape")
        if path.is_file():
            files.append(path)
    return files


def read_text(path: Path) -> str | None:
    try:
        if path.stat().st_size > MAX_FILE_BYTES:
            return None
        data = path.read_bytes()
    except OSError:
        return None

    if b"\0" in data:
        return None

    try:
        return data.decode("utf-8")
    except UnicodeDecodeError:
        return data.decode("utf-8", errors="ignore")


def iter_findings(path: Path, text: str):
    for line_number, line in enumerate(text.splitlines(), start=1):
        if is_placeholder_secret_line(line):
            continue
        for name, pattern in SECRET_PATTERNS:
            if pattern.search(line):
                yield name, line_number


def is_placeholder_secret_line(line: str) -> bool:
    return "..." in line or "\\ndata\\n" in line


def main() -> int:
    findings: list[str] = []
    for path in list_candidate_files():
        text = read_text(path)
        if text is None:
            continue
        relative_path = os.fspath(path.relative_to(ROOT))
        for name, line_number in iter_findings(path, text):
            findings.append(f"{relative_path}:{line_number}: {name}")

    if findings:
        print("Potential secrets detected:", file=sys.stderr)
        for finding in findings:
            print(f"  {finding}", file=sys.stderr)
        return 1

    print("No high-confidence secrets detected.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
