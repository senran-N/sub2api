#!/usr/bin/env python3
"""Run repeatable HTTP extreme scenarios against a local or remote Sub2API instance."""

from __future__ import annotations

import argparse
import dataclasses
import http.client
import json
import math
import queue
import signal
import sys
import threading
import time
import urllib.parse
import uuid
from typing import Any


DEFAULT_SUITE = [
    "public-settings",
    "auth-invalid-password",
    "admin-realtime",
    "gateway-auth-reject",
]


class ProbeError(RuntimeError):
    pass


def join_path(base_path: str, path: str) -> str:
    left = base_path.rstrip("/")
    right = "/" + path.lstrip("/")
    return left + right if left else right


def percentile(latencies_ms: list[float], ratio: float) -> float:
    if not latencies_ms:
        return 0.0
    if ratio <= 0:
        return min(latencies_ms)
    if ratio >= 1:
        return max(latencies_ms)
    ordered = sorted(latencies_ms)
    position = ratio * (len(ordered) - 1)
    lower = math.floor(position)
    upper = math.ceil(position)
    if lower == upper:
        return ordered[lower]
    lower_value = ordered[lower]
    upper_value = ordered[upper]
    return lower_value + (upper_value - lower_value) * (position - lower)


def unwrap_api_response(payload: bytes) -> Any:
    try:
        decoded = json.loads(payload.decode("utf-8"))
    except json.JSONDecodeError as exc:
        raise ProbeError(f"unexpected non-JSON API response: {payload[:200]!r}") from exc
    if decoded.get("code") != 0:
        raise ProbeError(
            f"API request failed with code={decoded.get('code')!r}: "
            f"{decoded.get('message')!r}"
        )
    return decoded.get("data")


class PersistentHTTPClient:
    def __init__(self, base_url: str, timeout_seconds: float) -> None:
        parsed = urllib.parse.urlsplit(base_url)
        if parsed.scheme not in {"http", "https"}:
            raise ValueError(f"unsupported scheme: {parsed.scheme!r}")
        if not parsed.hostname:
            raise ValueError(f"missing hostname in base URL: {base_url!r}")
        self.scheme = parsed.scheme
        self.hostname = parsed.hostname
        self.port = parsed.port or (443 if parsed.scheme == "https" else 80)
        self.base_path = parsed.path.rstrip("/")
        self.timeout_seconds = timeout_seconds
        self._conn: http.client.HTTPConnection | None = None

    def _ensure_conn(self) -> http.client.HTTPConnection:
        if self._conn is None:
            if self.scheme == "https":
                self._conn = http.client.HTTPSConnection(
                    self.hostname,
                    self.port,
                    timeout=self.timeout_seconds,
                )
            else:
                self._conn = http.client.HTTPConnection(
                    self.hostname,
                    self.port,
                    timeout=self.timeout_seconds,
                )
        return self._conn

    def close(self) -> None:
        if self._conn is not None:
            try:
                self._conn.close()
            finally:
                self._conn = None

    def request(
        self,
        method: str,
        path: str,
        headers: dict[str, str],
        body: bytes | None,
    ) -> tuple[int, bytes]:
        full_path = join_path(self.base_path, path)
        try:
            conn = self._ensure_conn()
            conn.request(method, full_path, body=body, headers=headers)
            response = conn.getresponse()
            return response.status, response.read()
        except Exception:
            self.close()
            raise


@dataclasses.dataclass
class TrafficCounters:
    submitted: int = 0
    completed: int = 0
    http_2xx: int = 0
    http_4xx: int = 0
    http_5xx: int = 0
    transport_errors: int = 0
    latency_ms_total: float = 0.0
    latency_ms_max: float = 0.0
    latency_samples: int = 0

    def clone(self) -> "TrafficCounters":
        return dataclasses.replace(self)


class SharedProbeStats:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._counters = TrafficCounters()
        self._latencies_ms: list[float] = []

    def submitted(self) -> None:
        with self._lock:
            self._counters.submitted += 1

    def completed(self, status: int, latency_ms: float, transport_error: bool) -> None:
        with self._lock:
            self._counters.completed += 1
            if transport_error:
                self._counters.transport_errors += 1
            elif 200 <= status < 300:
                self._counters.http_2xx += 1
            elif 400 <= status < 500:
                self._counters.http_4xx += 1
            elif status >= 500:
                self._counters.http_5xx += 1
            self._counters.latency_ms_total += latency_ms
            self._counters.latency_ms_max = max(self._counters.latency_ms_max, latency_ms)
            self._counters.latency_samples += 1
            self._latencies_ms.append(latency_ms)

    def snapshot(self) -> TrafficCounters:
        with self._lock:
            return self._counters.clone()

    def latencies(self) -> list[float]:
        with self._lock:
            return list(self._latencies_ms)


def diff_counters(current: TrafficCounters, baseline: TrafficCounters) -> TrafficCounters:
    return TrafficCounters(
        submitted=max(0, current.submitted - baseline.submitted),
        completed=max(0, current.completed - baseline.completed),
        http_2xx=max(0, current.http_2xx - baseline.http_2xx),
        http_4xx=max(0, current.http_4xx - baseline.http_4xx),
        http_5xx=max(0, current.http_5xx - baseline.http_5xx),
        transport_errors=max(0, current.transport_errors - baseline.transport_errors),
        latency_ms_total=max(0.0, current.latency_ms_total - baseline.latency_ms_total),
        latency_ms_max=max(0.0, current.latency_ms_max),
        latency_samples=max(0, current.latency_samples - baseline.latency_samples),
    )


def diff_rate(current: int, previous: int, elapsed_seconds: float) -> float:
    if elapsed_seconds <= 0:
        return 0.0
    return max(0, current - previous) / elapsed_seconds


@dataclasses.dataclass(frozen=True)
class RequestSpec:
    method: str
    path: str
    headers: dict[str, str]
    body: bytes | None = None


@dataclasses.dataclass(frozen=True)
class ScenarioDefinition:
    name: str
    description: str
    requires_admin_token: bool = False


SCENARIOS: dict[str, ScenarioDefinition] = {
    "public-settings": ScenarioDefinition(
        name="public-settings",
        description="Hot-loop the unauthenticated public settings endpoint to expose cache misses and response jitter.",
    ),
    "auth-invalid-password": ScenarioDefinition(
        name="auth-invalid-password",
        description="Burst invalid login attempts to exercise auth, hashing, DB access, and rate limiting.",
    ),
    "admin-realtime": ScenarioDefinition(
        name="admin-realtime",
        description="Reload the admin realtime metrics endpoint to stress authenticated monitoring reads.",
        requires_admin_token=True,
    ),
    "gateway-auth-reject": ScenarioDefinition(
        name="gateway-auth-reject",
        description="Hammer the OpenAI-compatible gateway with rejected credentials to observe auth-path saturation.",
    ),
    "gateway-large-body-reject": ScenarioDefinition(
        name="gateway-large-body-reject",
        description="Send oversized OpenAI-compatible payloads to validate early body-limit rejection under pressure.",
    ),
}


def build_scenario_request(
    scenario_name: str,
    seq: int,
    run_id: str,
    admin_token: str | None,
    gateway_api_key: str | None,
    large_body_bytes: int,
) -> RequestSpec:
    if scenario_name == "public-settings":
        return RequestSpec(
            method="GET",
            path="/api/v1/settings/public",
            headers={"Accept": "application/json"},
        )

    if scenario_name == "auth-invalid-password":
        body = json.dumps(
            {
                "email": f"extreme-{run_id}-{seq}@example.invalid",
                "password": "definitely-invalid-password",
            },
            separators=(",", ":"),
        ).encode("utf-8")
        return RequestSpec(
            method="POST",
            path="/api/v1/auth/login",
            headers={
                "Accept": "application/json",
                "Content-Type": "application/json",
            },
            body=body,
        )

    if scenario_name == "admin-realtime":
        if not admin_token:
            raise ProbeError("admin token required for admin-realtime scenario")
        return RequestSpec(
            method="GET",
            path="/api/v1/admin/ops/realtime-traffic?window=1min",
            headers={
                "Accept": "application/json",
                "Authorization": f"Bearer {admin_token}",
            },
        )

    base_headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
        "Authorization": f"Bearer {gateway_api_key or 'sk-extreme-invalid'}",
        "session_id": f"sess-{run_id}-{seq}",
        "conversation_id": f"conv-{run_id}-{seq}",
        "x-client-request-id": f"{run_id}-{scenario_name}-{seq}",
    }
    if scenario_name == "gateway-auth-reject":
        body = json.dumps(
            {
                "model": "gpt-5",
                "stream": False,
                "input": [
                    {
                        "role": "user",
                        "content": [{"type": "input_text", "text": f"extreme auth reject {seq}"}],
                    }
                ],
            },
            separators=(",", ":"),
        ).encode("utf-8")
        return RequestSpec(
            method="POST",
            path="/v1/responses",
            headers=base_headers,
            body=body,
        )

    if scenario_name == "gateway-large-body-reject":
        body_bytes = max(256, large_body_bytes)
        payload = json.dumps(
            {
                "model": "gpt-5",
                "stream": False,
                "input": [
                    {
                        "role": "user",
                        "content": [
                            {
                                "type": "input_text",
                                "text": "X" * body_bytes,
                            }
                        ],
                    }
                ],
            },
            separators=(",", ":"),
        ).encode("utf-8")
        return RequestSpec(
            method="POST",
            path="/v1/responses",
            headers=base_headers,
            body=payload,
        )

    raise ProbeError(f"unknown scenario: {scenario_name}")


def login_admin(base_url: str, email: str, password: str, timeout_seconds: float) -> str:
    client = PersistentHTTPClient(base_url, timeout_seconds)
    body = json.dumps(
        {"email": email, "password": password},
        separators=(",", ":"),
    ).encode("utf-8")
    try:
        status, payload = client.request(
            "POST",
            "/api/v1/auth/login",
            {
                "Accept": "application/json",
                "Content-Type": "application/json",
            },
            body,
        )
    finally:
        client.close()

    if status != 200:
        raise ProbeError(f"admin login failed with status={status}: {payload[:300]!r}")
    data = unwrap_api_response(payload)
    if not isinstance(data, dict):
        raise ProbeError(f"unexpected admin login payload: {data!r}")
    if data.get("requires_2fa") is True:
        raise ProbeError("admin login requires 2FA; provide --admin-token instead")
    token = data.get("access_token")
    if not isinstance(token, str) or not token.strip():
        raise ProbeError(f"admin login missing access_token: {data!r}")
    return token


def producer_loop(
    stop_event: threading.Event,
    tasks: queue.Queue[int | None],
    stats: SharedProbeStats,
    qps: float,
    total_seconds: float,
) -> None:
    deadline = time.monotonic() + total_seconds
    next_emit = time.monotonic()
    interval = 1.0 / qps
    seq = 0
    while not stop_event.is_set():
        now = time.monotonic()
        if now >= deadline:
            return
        seq += 1
        stats.submitted()
        tasks.put(seq)
        next_emit += interval
        sleep_seconds = next_emit - time.monotonic()
        if sleep_seconds > 0:
            time.sleep(sleep_seconds)


def worker_loop(
    stop_event: threading.Event,
    tasks: queue.Queue[int | None],
    stats: SharedProbeStats,
    base_url: str,
    scenario_name: str,
    run_id: str,
    admin_token: str | None,
    gateway_api_key: str | None,
    large_body_bytes: int,
    timeout_seconds: float,
) -> None:
    client = PersistentHTTPClient(base_url, timeout_seconds)
    try:
        while True:
            seq = tasks.get()
            if seq is None:
                tasks.task_done()
                return
            status = 0
            transport_error = False
            started = time.monotonic()
            try:
                request = build_scenario_request(
                    scenario_name,
                    seq,
                    run_id,
                    admin_token,
                    gateway_api_key,
                    large_body_bytes,
                )
                status, _ = client.request(
                    request.method,
                    request.path,
                    request.headers,
                    request.body,
                )
            except Exception:
                transport_error = True
            latency_ms = (time.monotonic() - started) * 1000.0
            stats.completed(status, latency_ms, transport_error)
            tasks.task_done()
            if stop_event.is_set():
                continue
    finally:
        client.close()


@dataclasses.dataclass
class Sample:
    elapsed_seconds: float
    submitted_rps: float
    completed_rps: float
    http_2xx_rps: float
    http_4xx_rps: float
    http_5xx_rps: float
    transport_error_rps: float
    avg_latency_ms: float


@dataclasses.dataclass
class Summary:
    name: str
    description: str
    submitted: int
    completed: int
    http_2xx: int
    http_4xx: int
    http_5xx: int
    transport_errors: int
    avg_latency_ms: float
    p50_latency_ms: float
    p95_latency_ms: float
    max_latency_ms: float
    effective_rps: float


def print_sample(name: str, sample: Sample) -> None:
    print(
        f"[{name:22s}] +{sample.elapsed_seconds:5.1f}s "
        f"submitted_rps={sample.submitted_rps:7.2f} "
        f"completed_rps={sample.completed_rps:7.2f} "
        f"2xx_rps={sample.http_2xx_rps:7.2f} "
        f"4xx_rps={sample.http_4xx_rps:7.2f} "
        f"5xx_rps={sample.http_5xx_rps:7.2f} "
        f"transport_rps={sample.transport_error_rps:7.2f} "
        f"avg_ms={sample.avg_latency_ms:7.1f}",
        flush=True,
    )


def emit_summary(summary: Summary) -> None:
    print(
        f"SUMMARY {summary.name}: "
        f"submitted={summary.submitted} completed={summary.completed} "
        f"http_2xx={summary.http_2xx} http_4xx={summary.http_4xx} http_5xx={summary.http_5xx} "
        f"transport_errors={summary.transport_errors} "
        f"avg_latency_ms={summary.avg_latency_ms:.1f} "
        f"p50_latency_ms={summary.p50_latency_ms:.1f} "
        f"p95_latency_ms={summary.p95_latency_ms:.1f} "
        f"max_latency_ms={summary.max_latency_ms:.1f} "
        f"effective_rps={summary.effective_rps:.2f}",
        flush=True,
    )


def run_scenario(
    definition: ScenarioDefinition,
    args: argparse.Namespace,
    admin_token: str | None,
    run_id: str,
    stop_event: threading.Event,
) -> tuple[Summary, list[Sample]]:
    total_seconds = args.warmup_seconds + args.measure_seconds
    tasks: queue.Queue[int | None] = queue.Queue(maxsize=max(args.concurrency * 4, 64))
    stats = SharedProbeStats()
    producer = threading.Thread(
        target=producer_loop,
        args=(stop_event, tasks, stats, args.qps, total_seconds),
        daemon=True,
    )
    workers = [
        threading.Thread(
            target=worker_loop,
            args=(
                stop_event,
                tasks,
                stats,
                args.base_url,
                definition.name,
                run_id,
                admin_token,
                args.gateway_api_key,
                args.large_body_bytes,
                args.timeout_seconds,
            ),
            daemon=True,
        )
        for _ in range(args.concurrency)
    ]

    print(
        f"==> scenario={definition.name} qps={args.qps:.2f} "
        f"concurrency={args.concurrency} warmup={args.warmup_seconds:.1f}s "
        f"measure={args.measure_seconds:.1f}s",
        flush=True,
    )
    producer.start()
    for worker in workers:
        worker.start()

    samples: list[Sample] = []
    try:
        time.sleep(args.warmup_seconds)
        baseline = stats.snapshot()
        previous = baseline
        measurement_start = time.monotonic()
        previous_at = measurement_start
        deadline = measurement_start + args.measure_seconds

        while time.monotonic() < deadline:
            if stop_event.is_set():
                break
            time.sleep(args.sample_interval_seconds)
            now = time.monotonic()
            current = stats.snapshot()
            elapsed = max(0.001, now - previous_at)
            delta = diff_counters(current, previous)
            sample = Sample(
                elapsed_seconds=max(0.0, now - measurement_start),
                submitted_rps=diff_rate(current.submitted, previous.submitted, elapsed),
                completed_rps=diff_rate(current.completed, previous.completed, elapsed),
                http_2xx_rps=diff_rate(current.http_2xx, previous.http_2xx, elapsed),
                http_4xx_rps=diff_rate(current.http_4xx, previous.http_4xx, elapsed),
                http_5xx_rps=diff_rate(current.http_5xx, previous.http_5xx, elapsed),
                transport_error_rps=diff_rate(
                    current.transport_errors,
                    previous.transport_errors,
                    elapsed,
                ),
                avg_latency_ms=(
                    delta.latency_ms_total / delta.latency_samples
                    if delta.latency_samples
                    else 0.0
                ),
            )
            samples.append(sample)
            print_sample(definition.name, sample)
            previous = current
            previous_at = now

        final_snapshot = stats.snapshot()
        final_latencies = stats.latencies()
    finally:
        stop_event.set()
        producer.join(timeout=5)
        for _ in workers:
            tasks.put(None)
        tasks.join()
        for worker in workers:
            worker.join(timeout=5)

    measurement_duration = max(0.001, time.monotonic() - measurement_start)
    totals = diff_counters(final_snapshot, baseline)
    latencies = final_latencies[baseline.latency_samples:]
    avg_latency_ms = totals.latency_ms_total / totals.latency_samples if totals.latency_samples else 0.0
    summary = Summary(
        name=definition.name,
        description=definition.description,
        submitted=totals.submitted,
        completed=totals.completed,
        http_2xx=totals.http_2xx,
        http_4xx=totals.http_4xx,
        http_5xx=totals.http_5xx,
        transport_errors=totals.transport_errors,
        avg_latency_ms=avg_latency_ms,
        p50_latency_ms=percentile(latencies, 0.50),
        p95_latency_ms=percentile(latencies, 0.95),
        max_latency_ms=max(latencies) if latencies else 0.0,
        effective_rps=totals.completed / measurement_duration,
    )
    return summary, samples


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Run extreme HTTP scenarios against Sub2API without external dependencies.",
    )
    parser.add_argument("--base-url", required=True, help="Base URL, for example http://127.0.0.1:8080")
    parser.add_argument(
        "--scenario",
        action="append",
        choices=["default-suite", *sorted(SCENARIOS)],
        help="Scenario to run. Repeat the flag to run multiple scenarios. Defaults to default-suite.",
    )
    parser.add_argument("--admin-token", help="Existing admin bearer token")
    parser.add_argument("--admin-email", help="Admin email used to auto-login when --admin-token is omitted")
    parser.add_argument("--admin-password", help="Admin password used to auto-login when --admin-token is omitted")
    parser.add_argument(
        "--gateway-api-key",
        help="Gateway API key. Omit to intentionally exercise rejected-key scenarios.",
    )
    parser.add_argument("--qps", type=float, default=20.0, help="Target requests per second")
    parser.add_argument("--concurrency", type=int, default=8, help="Worker count")
    parser.add_argument("--warmup-seconds", type=float, default=3.0)
    parser.add_argument("--measure-seconds", type=float, default=20.0)
    parser.add_argument("--sample-interval-seconds", type=float, default=5.0)
    parser.add_argument("--cooldown-seconds", type=float, default=2.0)
    parser.add_argument("--timeout-seconds", type=float, default=30.0)
    parser.add_argument(
        "--large-body-bytes",
        type=int,
        default=2 * 1024 * 1024,
        help="Target input_text payload size for gateway-large-body-reject",
    )
    parser.add_argument("--output-json", help="Optional summary JSON output path")
    return parser


def resolve_scenarios(raw_scenarios: list[str] | None) -> list[ScenarioDefinition]:
    scenario_names = raw_scenarios or ["default-suite"]
    expanded: list[str] = []
    for name in scenario_names:
        if name == "default-suite":
            expanded.extend(DEFAULT_SUITE)
        else:
            expanded.append(name)

    ordered: list[ScenarioDefinition] = []
    seen: set[str] = set()
    for name in expanded:
        if name in seen:
            continue
        seen.add(name)
        ordered.append(SCENARIOS[name])
    return ordered


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    if args.qps <= 0:
        parser.error("--qps must be > 0")
    if args.concurrency <= 0:
        parser.error("--concurrency must be > 0")
    if args.sample_interval_seconds <= 0:
        parser.error("--sample-interval-seconds must be > 0")
    if args.large_body_bytes <= 0:
        parser.error("--large-body-bytes must be > 0")

    scenarios = resolve_scenarios(args.scenario)
    requires_admin = any(item.requires_admin_token for item in scenarios)
    admin_token = args.admin_token
    if requires_admin and not admin_token:
        if not args.admin_email or not args.admin_password:
            parser.error(
                "--admin-token or both --admin-email and --admin-password are required "
                "when running admin-realtime",
            )
        admin_token = login_admin(
            args.base_url,
            args.admin_email,
            args.admin_password,
            args.timeout_seconds,
        )
        print("admin token acquired via /api/v1/auth/login", flush=True)

    run_id = uuid.uuid4().hex[:12]
    stop_event = threading.Event()

    def handle_signal(_signum: int, _frame: Any) -> None:
        stop_event.set()

    signal.signal(signal.SIGINT, handle_signal)
    signal.signal(signal.SIGTERM, handle_signal)

    summaries: list[Summary] = []
    sample_log: dict[str, list[dict[str, Any]]] = {}

    try:
        for index, scenario in enumerate(scenarios):
            stop_event.clear()
            summary, samples = run_scenario(scenario, args, admin_token, run_id, stop_event)
            summaries.append(summary)
            sample_log[scenario.name] = [dataclasses.asdict(sample) for sample in samples]
            emit_summary(summary)
            if index < len(scenarios) - 1 and args.cooldown_seconds > 0:
                time.sleep(args.cooldown_seconds)
    except ProbeError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    if args.output_json:
        payload = {
            "base_url": args.base_url,
            "qps": args.qps,
            "concurrency": args.concurrency,
            "warmup_seconds": args.warmup_seconds,
            "measure_seconds": args.measure_seconds,
            "summaries": [dataclasses.asdict(item) for item in summaries],
            "samples": sample_log,
        }
        with open(args.output_json, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, indent=2, ensure_ascii=True)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
