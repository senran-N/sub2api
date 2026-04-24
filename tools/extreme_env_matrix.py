#!/usr/bin/env python3
"""Emit reproducible environment override profiles for extreme local testing."""

from __future__ import annotations

import argparse
import json
import shlex
import sys
from dataclasses import dataclass


@dataclass(frozen=True)
class EnvironmentProfile:
    name: str
    description: str
    overrides: dict[str, str]


PROFILES: dict[str, EnvironmentProfile] = {
    "baseline": EnvironmentProfile(
        name="baseline",
        description="Control group. Uses deploy/.env and compose defaults without extra pressure knobs.",
        overrides={},
    ),
    "cpu-starved": EnvironmentProfile(
        name="cpu-starved",
        description="Constrain the app container so UI/API latency regressions surface under modest CPU and memory pressure.",
        overrides={
            "SUB2API_CPUS": "0.35",
            "SUB2API_MEMORY_LIMIT": "768m",
            "SUB2API_MEMORY_RESERVATION": "256m",
            "DATABASE_MAX_OPEN_CONNS": "48",
            "DATABASE_MAX_IDLE_CONNS": "16",
            "REDIS_POOL_SIZE": "128",
            "REDIS_MIN_IDLE_CONNS": "16",
        },
    ),
    "connection-pressure": EnvironmentProfile(
        name="connection-pressure",
        description="Squeeze PostgreSQL and Redis connection headroom to expose pool starvation and lock-step retries.",
        overrides={
            "POSTGRES_MAX_CONNECTIONS": "96",
            "POSTGRES_SHARED_BUFFERS": "256MB",
            "POSTGRES_EFFECTIVE_CACHE_SIZE": "512MB",
            "DATABASE_MAX_OPEN_CONNS": "72",
            "DATABASE_MAX_IDLE_CONNS": "8",
            "DATABASE_CONN_MAX_LIFETIME_MINUTES": "10",
            "DATABASE_CONN_MAX_IDLE_TIME_MINUTES": "1",
            "REDIS_MAXCLIENTS": "512",
            "REDIS_POOL_SIZE": "96",
            "REDIS_MIN_IDLE_CONNS": "8",
        },
    ),
    "redis-pressure": EnvironmentProfile(
        name="redis-pressure",
        description="Reduce Redis memory and client capacity to observe queue buildup, reconnect churn, and burst collapse.",
        overrides={
            "REDIS_MEMORY_LIMIT": "192m",
            "REDIS_MEMORY_RESERVATION": "96m",
            "REDIS_MAXCLIENTS": "256",
            "REDIS_POOL_SIZE": "64",
            "REDIS_MIN_IDLE_CONNS": "4",
            "SUB2API_PIDS_LIMIT": "256",
        },
    ),
    "gateway-buffer-pressure": EnvironmentProfile(
        name="gateway-buffer-pressure",
        description="Shrink request-body and h2c buffers to validate oversized payload handling and degraded streaming paths.",
        overrides={
            "SERVER_MAX_REQUEST_BODY_SIZE": "8388608",
            "GATEWAY_MAX_BODY_SIZE": "8388608",
            "SERVER_H2C_MAX_CONCURRENT_STREAMS": "16",
            "SERVER_H2C_MAX_READ_FRAME_SIZE": "262144",
            "SERVER_H2C_MAX_UPLOAD_BUFFER_PER_CONNECTION": "262144",
            "SERVER_H2C_MAX_UPLOAD_BUFFER_PER_STREAM": "65536",
        },
    ),
}


def emit_profiles() -> str:
    lines = []
    for profile in sorted(PROFILES.values(), key=lambda item: item.name):
        lines.append(f"{profile.name}")
        lines.append(f"  {profile.description}")
        if profile.overrides:
            for key, value in sorted(profile.overrides.items()):
                lines.append(f"  {key}={value}")
        else:
            lines.append("  (no extra overrides)")
    return "\n".join(lines) + "\n"


def render_profile(profile: EnvironmentProfile, output_format: str) -> str:
    overrides = dict(sorted(profile.overrides.items()))
    if output_format == "json":
        return json.dumps(overrides, indent=2, ensure_ascii=True) + "\n"
    if output_format == "shell":
        return "".join(
            f"export {key}={shlex.quote(value)}\n" for key, value in overrides.items()
        )
    return "".join(f"{key}={value}\n" for key, value in overrides.items())


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description=(
            "Print compose-friendly environment override profiles for extreme "
            "local simulation runs."
        ),
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    subparsers.add_parser("profiles", help="List built-in extreme-test profiles")

    env_parser = subparsers.add_parser(
        "env",
        help="Render a profile as shell exports, dotenv pairs, or JSON",
    )
    env_parser.add_argument("--profile", required=True, choices=sorted(PROFILES))
    env_parser.add_argument(
        "--format",
        choices=["dotenv", "shell", "json"],
        default="dotenv",
    )
    env_parser.add_argument(
        "--output",
        help="Optional file path. Omit to print to stdout.",
    )
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    if args.command == "profiles":
        sys.stdout.write(emit_profiles())
        return 0

    profile = PROFILES[args.profile]
    rendered = render_profile(profile, args.format)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as handle:
            handle.write(rendered)
        return 0

    sys.stdout.write(rendered)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
