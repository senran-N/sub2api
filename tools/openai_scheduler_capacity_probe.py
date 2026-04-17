#!/usr/bin/env python3
"""Run sticky/non-sticky scheduler capacity scenarios against a live Sub2API instance."""

from __future__ import annotations

import argparse
import dataclasses
import http.client
import json
import queue
import signal
import sys
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
import uuid
from typing import Any


def parse_header(values: list[str]) -> dict[str, str]:
    headers: dict[str, str] = {}
    for raw in values:
        name, sep, value = raw.partition(":")
        if not sep or not name.strip():
            raise ValueError(f"invalid header format: {raw!r}")
        headers[name.strip()] = value.strip()
    return headers


def join_url(base: str, path: str) -> str:
    return urllib.parse.urljoin(base.rstrip("/") + "/", path.lstrip("/"))


def join_path(base_path: str, path: str) -> str:
    left = base_path.rstrip("/")
    right = "/" + path.lstrip("/")
    return (left + right) if left else right


def default_gateway_body(request_path: str, model: str, prompt: str, seq: int, scenario: str) -> bytes:
    text = f"{prompt} [{scenario} #{seq}]"
    if "chat/completions" in request_path:
        payload: dict[str, Any] = {
            "model": model,
            "stream": False,
            "messages": [{"role": "user", "content": text}],
        }
    else:
        payload = {
            "model": model,
            "stream": False,
            "input": [
                {
                    "role": "user",
                    "content": [{"type": "input_text", "text": text}],
                }
            ],
        }
    return json.dumps(payload, separators=(",", ":")).encode("utf-8")


class PersistentGatewayClient:
    def __init__(self, base_url: str, timeout_seconds: float) -> None:
        parsed = urllib.parse.urlsplit(base_url)
        if parsed.scheme not in {"http", "https"}:
            raise ValueError(f"unsupported gateway scheme: {parsed.scheme!r}")
        if not parsed.hostname:
            raise ValueError(f"gateway URL missing hostname: {base_url!r}")
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

    def post(self, request_path: str, body: bytes, headers: dict[str, str]) -> tuple[int, bytes]:
        path = join_path(self.base_path, request_path)
        try:
            conn = self._ensure_conn()
            conn.request("POST", path, body=body, headers=headers)
            response = conn.getresponse()
            payload = response.read()
            return response.status, payload
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
    latency_samples: int = 0

    def clone(self) -> "TrafficCounters":
        return dataclasses.replace(self)


class SharedTrafficCounters:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._counters = TrafficCounters()

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
            self._counters.latency_samples += 1

    def snapshot(self) -> TrafficCounters:
        with self._lock:
            return self._counters.clone()


@dataclasses.dataclass
class AdminSnapshot:
    fetched_at: float
    qps_current: float
    qps_peak: float
    qps_avg: float
    select_total: int
    non_sticky_intent_total: int
    sticky_miss_fallback_total: int
    indexed_load_balance_total: int
    sticky_miss_indexed_total: int
    sticky_hit_rate: float
    sticky_miss_rate: float
    non_sticky_share: float
    indexed_load_balance_share: float
    sticky_miss_indexed_share: float


@dataclasses.dataclass
class ScenarioSample:
    elapsed_seconds: float
    qps_current: float
    qps_avg: float
    non_sticky_rps: float
    sticky_miss_rps: float
    indexed_load_balance_rps: float
    sticky_miss_indexed_rps: float
    sticky_hit_rate: float
    sticky_miss_rate: float
    non_sticky_share: float
    indexed_load_balance_share: float


@dataclasses.dataclass
class ScenarioSummary:
    name: str
    submitted: int
    completed: int
    http_2xx: int
    http_4xx: int
    http_5xx: int
    transport_errors: int
    avg_latency_ms: float
    qps_current: float
    qps_avg: float
    non_sticky_rps: float
    sticky_miss_rps: float
    indexed_load_balance_rps: float
    sticky_miss_indexed_rps: float
    sticky_hit_rate: float
    sticky_miss_rate: float
    non_sticky_share: float
    indexed_load_balance_share: float


class ProbeError(RuntimeError):
    pass


def fetch_admin_snapshot(
    admin_url: str,
    headers: dict[str, str],
    timeout_seconds: float,
) -> AdminSnapshot:
    request = urllib.request.Request(admin_url, headers=headers, method="GET")
    try:
        with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        raise ProbeError(f"admin endpoint returned {exc.code}: {body}") from exc
    except urllib.error.URLError as exc:
        raise ProbeError(f"admin endpoint error: {exc}") from exc

    if payload.get("code") != 0:
        raise ProbeError(f"admin endpoint returned non-zero code: {payload!r}")
    data = payload.get("data") or {}
    if not data.get("enabled", True):
        raise ProbeError("admin realtime monitoring is disabled")

    summary = (data.get("summary") or {})
    qps = summary.get("qps") or {}
    runtime = data.get("runtime_observability") or {}
    scheduler = runtime.get("openai_account_scheduler") or {}
    scheduler_summary = ((runtime.get("summary") or {}).get("openai_account_scheduler") or {})

    return AdminSnapshot(
        fetched_at=time.monotonic(),
        qps_current=float(qps.get("current") or 0.0),
        qps_peak=float(qps.get("peak") or 0.0),
        qps_avg=float(qps.get("avg") or 0.0),
        select_total=int(scheduler.get("select_total") or 0),
        non_sticky_intent_total=int(scheduler.get("non_sticky_intent_total") or 0),
        sticky_miss_fallback_total=int(scheduler.get("sticky_miss_fallback_total") or 0),
        indexed_load_balance_total=int(scheduler.get("indexed_load_balance_select_total") or 0),
        sticky_miss_indexed_total=int(scheduler.get("sticky_miss_indexed_select_total") or 0),
        sticky_hit_rate=float(scheduler_summary.get("sticky_intent_hit_rate") or 0.0),
        sticky_miss_rate=float(scheduler_summary.get("sticky_intent_miss_rate") or 0.0),
        non_sticky_share=float(scheduler_summary.get("non_sticky_intent_share") or 0.0),
        indexed_load_balance_share=float(scheduler_summary.get("indexed_load_balance_share") or 0.0),
        sticky_miss_indexed_share=float(scheduler_summary.get("sticky_miss_indexed_share") or 0.0),
    )


def diff_rate(current: int, previous: int, elapsed_seconds: float) -> float:
    if elapsed_seconds <= 0:
        return 0.0
    return max(0, current - previous) / elapsed_seconds


def diff_counters(current: TrafficCounters, baseline: TrafficCounters) -> TrafficCounters:
    return TrafficCounters(
        submitted=max(0, current.submitted - baseline.submitted),
        completed=max(0, current.completed - baseline.completed),
        http_2xx=max(0, current.http_2xx - baseline.http_2xx),
        http_4xx=max(0, current.http_4xx - baseline.http_4xx),
        http_5xx=max(0, current.http_5xx - baseline.http_5xx),
        transport_errors=max(0, current.transport_errors - baseline.transport_errors),
        latency_ms_total=max(0.0, current.latency_ms_total - baseline.latency_ms_total),
        latency_samples=max(0, current.latency_samples - baseline.latency_samples),
    )


@dataclasses.dataclass
class ScenarioConfig:
    name: str
    sticky_pool_size: int
    warmup_seconds: float
    measure_seconds: float


def scenario_session_ids(run_id: str, scenario: ScenarioConfig, seq: int) -> tuple[str, str]:
    if scenario.name == "high-sticky":
        pool_index = seq % max(1, scenario.sticky_pool_size)
        seed = f"{run_id}-sticky-{pool_index}"
    else:
        seed = f"{run_id}-{scenario.name}-{seq}-{uuid.uuid4().hex[:8]}"
    return f"sess-{seed}", f"conv-{seed}"


def producer_loop(
    stop_event: threading.Event,
    tasks: queue.Queue[int | None],
    counters: SharedTrafficCounters,
    qps: float,
    total_seconds: float,
) -> None:
    deadline = time.monotonic() + total_seconds
    seq = 0
    next_emit = time.monotonic()
    interval = 1.0 / qps
    while not stop_event.is_set():
        now = time.monotonic()
        if now >= deadline:
            break
        seq += 1
        counters.submitted()
        tasks.put(seq)
        next_emit += interval
        sleep_seconds = next_emit - time.monotonic()
        if sleep_seconds > 0:
            time.sleep(sleep_seconds)


def worker_loop(
    tasks: queue.Queue[int | None],
    counters: SharedTrafficCounters,
    gateway_base_url: str,
    request_path: str,
    request_body_file: bytes | None,
    gateway_headers: dict[str, str],
    gateway_api_key: str | None,
    timeout_seconds: float,
    model: str,
    prompt: str,
    scenario: ScenarioConfig,
    run_id: str,
) -> None:
    client = PersistentGatewayClient(gateway_base_url, timeout_seconds)
    try:
        while True:
            seq = tasks.get()
            if seq is None:
                tasks.task_done()
                return
            session_id, conversation_id = scenario_session_ids(run_id, scenario, seq)
            headers = dict(gateway_headers)
            headers.setdefault("Content-Type", "application/json")
            headers.setdefault("Accept", "application/json")
            headers["session_id"] = session_id
            headers["conversation_id"] = conversation_id
            headers["x-client-request-id"] = f"{run_id}-{scenario.name}-{seq}"
            if gateway_api_key:
                headers.setdefault("Authorization", f"Bearer {gateway_api_key}")
            body = request_body_file or default_gateway_body(request_path, model, prompt, seq, scenario.name)
            started = time.monotonic()
            status = 0
            transport_error = False
            try:
                status, _ = client.post(request_path, body, headers)
            except Exception:
                transport_error = True
            latency_ms = (time.monotonic() - started) * 1000.0
            counters.completed(status, latency_ms, transport_error)
            tasks.task_done()
    finally:
        client.close()


def print_sample(name: str, sample: ScenarioSample) -> None:
    print(
        f"[{name:11s}] +{sample.elapsed_seconds:5.1f}s "
        f"qps={sample.qps_current:7.2f} "
        f"non_sticky_rps={sample.non_sticky_rps:7.2f} "
        f"sticky_miss_rps={sample.sticky_miss_rps:7.2f} "
        f"indexed_lb_rps={sample.indexed_load_balance_rps:7.2f} "
        f"sticky_hit={sample.sticky_hit_rate:6.2%} "
        f"sticky_miss={sample.sticky_miss_rate:6.2%} "
        f"non_sticky_share={sample.non_sticky_share:6.2%} "
        f"indexed_share={sample.indexed_load_balance_share:6.2%}"
    )


def run_scenario(
    args: argparse.Namespace,
    admin_url: str,
    admin_headers: dict[str, str],
    gateway_headers: dict[str, str],
    request_body_file: bytes | None,
    scenario: ScenarioConfig,
    run_id: str,
    stop_event: threading.Event,
) -> tuple[ScenarioSummary, list[ScenarioSample]]:
    total_seconds = scenario.warmup_seconds + scenario.measure_seconds
    tasks: queue.Queue[int | None] = queue.Queue(maxsize=max(args.concurrency * 4, 64))
    counters = SharedTrafficCounters()
    producer = threading.Thread(
        target=producer_loop,
        args=(stop_event, tasks, counters, args.qps, total_seconds),
        daemon=True,
    )
    workers = [
        threading.Thread(
            target=worker_loop,
            args=(
                tasks,
                counters,
                args.gateway_base_url,
                args.request_path,
                request_body_file,
                gateway_headers,
                args.gateway_api_key,
                args.timeout_seconds,
                args.model,
                args.prompt,
                scenario,
                run_id,
            ),
            daemon=True,
        )
        for _ in range(args.concurrency)
    ]

    print(
        f"==> scenario={scenario.name} "
        f"qps={args.qps:.2f} concurrency={args.concurrency} "
        f"warmup={scenario.warmup_seconds:.1f}s measure={scenario.measure_seconds:.1f}s"
    )
    producer.start()
    for worker in workers:
        worker.start()

    samples: list[ScenarioSample] = []
    try:
        time.sleep(scenario.warmup_seconds)
        baseline_admin = fetch_admin_snapshot(admin_url, admin_headers, args.timeout_seconds)
        baseline_traffic = counters.snapshot()
        previous_admin = baseline_admin
        measurement_start = baseline_admin.fetched_at
        measurement_deadline = measurement_start + scenario.measure_seconds

        while time.monotonic() < measurement_deadline:
            if stop_event.is_set():
                break
            time.sleep(args.sample_interval_seconds)
            current_admin = fetch_admin_snapshot(admin_url, admin_headers, args.timeout_seconds)
            elapsed = max(0.001, current_admin.fetched_at - previous_admin.fetched_at)
            sample = ScenarioSample(
                elapsed_seconds=max(0.0, current_admin.fetched_at - measurement_start),
                qps_current=current_admin.qps_current,
                qps_avg=current_admin.qps_avg,
                non_sticky_rps=diff_rate(
                    current_admin.non_sticky_intent_total,
                    previous_admin.non_sticky_intent_total,
                    elapsed,
                ),
                sticky_miss_rps=diff_rate(
                    current_admin.sticky_miss_fallback_total,
                    previous_admin.sticky_miss_fallback_total,
                    elapsed,
                ),
                indexed_load_balance_rps=diff_rate(
                    current_admin.indexed_load_balance_total,
                    previous_admin.indexed_load_balance_total,
                    elapsed,
                ),
                sticky_miss_indexed_rps=diff_rate(
                    current_admin.sticky_miss_indexed_total,
                    previous_admin.sticky_miss_indexed_total,
                    elapsed,
                ),
                sticky_hit_rate=current_admin.sticky_hit_rate,
                sticky_miss_rate=current_admin.sticky_miss_rate,
                non_sticky_share=current_admin.non_sticky_share,
                indexed_load_balance_share=current_admin.indexed_load_balance_share,
            )
            samples.append(sample)
            print_sample(scenario.name, sample)
            previous_admin = current_admin

        final_admin = fetch_admin_snapshot(admin_url, admin_headers, args.timeout_seconds)
        final_traffic = counters.snapshot()
    finally:
        stop_event.set()
        producer.join(timeout=5)
        for _ in workers:
            tasks.put(None)
        tasks.join()
        for worker in workers:
            worker.join(timeout=5)

    duration = max(0.001, final_admin.fetched_at - baseline_admin.fetched_at)
    traffic_delta = diff_counters(final_traffic, baseline_traffic)
    avg_latency_ms = (
        traffic_delta.latency_ms_total / traffic_delta.latency_samples
        if traffic_delta.latency_samples
        else 0.0
    )
    summary = ScenarioSummary(
        name=scenario.name,
        submitted=traffic_delta.submitted,
        completed=traffic_delta.completed,
        http_2xx=traffic_delta.http_2xx,
        http_4xx=traffic_delta.http_4xx,
        http_5xx=traffic_delta.http_5xx,
        transport_errors=traffic_delta.transport_errors,
        avg_latency_ms=avg_latency_ms,
        qps_current=final_admin.qps_current,
        qps_avg=final_admin.qps_avg,
        non_sticky_rps=diff_rate(
            final_admin.non_sticky_intent_total,
            baseline_admin.non_sticky_intent_total,
            duration,
        ),
        sticky_miss_rps=diff_rate(
            final_admin.sticky_miss_fallback_total,
            baseline_admin.sticky_miss_fallback_total,
            duration,
        ),
        indexed_load_balance_rps=diff_rate(
            final_admin.indexed_load_balance_total,
            baseline_admin.indexed_load_balance_total,
            duration,
        ),
        sticky_miss_indexed_rps=diff_rate(
            final_admin.sticky_miss_indexed_total,
            baseline_admin.sticky_miss_indexed_total,
            duration,
        ),
        sticky_hit_rate=final_admin.sticky_hit_rate,
        sticky_miss_rate=final_admin.sticky_miss_rate,
        non_sticky_share=final_admin.non_sticky_share,
        indexed_load_balance_share=final_admin.indexed_load_balance_share,
    )
    return summary, samples


def emit_summary(summary: ScenarioSummary) -> None:
    print(
        f"SUMMARY {summary.name}: "
        f"submitted={summary.submitted} completed={summary.completed} "
        f"http_2xx={summary.http_2xx} http_4xx={summary.http_4xx} http_5xx={summary.http_5xx} "
        f"transport_errors={summary.transport_errors} avg_latency_ms={summary.avg_latency_ms:.1f} "
        f"qps_current={summary.qps_current:.2f} qps_avg={summary.qps_avg:.2f} "
        f"non_sticky_rps={summary.non_sticky_rps:.2f} sticky_miss_rps={summary.sticky_miss_rps:.2f} "
        f"indexed_lb_rps={summary.indexed_load_balance_rps:.2f} "
        f"sticky_hit={summary.sticky_hit_rate:.2%} sticky_miss={summary.sticky_miss_rate:.2%} "
        f"non_sticky_share={summary.non_sticky_share:.2%} indexed_share={summary.indexed_load_balance_share:.2%}"
    )


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Drive high-sticky / low-sticky OpenAI scheduler scenarios and sample runtime observability counters.",
    )
    parser.add_argument("--gateway-base-url", required=True, help="Gateway base URL, for example http://127.0.0.1:8080")
    parser.add_argument("--admin-base-url", required=True, help="Admin API base URL, for example http://127.0.0.1:8080/api/v1")
    parser.add_argument("--request-path", default="/v1/responses", help="OpenAI-compatible request path")
    parser.add_argument("--gateway-api-key", help="Gateway API key; if omitted provide Authorization via --gateway-header")
    parser.add_argument("--admin-token", help="Admin bearer token; if omitted provide Authorization via --admin-header")
    parser.add_argument("--gateway-header", action="append", default=[], help="Extra gateway header in 'Name: Value' form")
    parser.add_argument("--admin-header", action="append", default=[], help="Extra admin header in 'Name: Value' form")
    parser.add_argument("--model", help="Model used when --request-body-file is not provided")
    parser.add_argument("--prompt", default="capacity probe", help="Prompt text used for generated request bodies")
    parser.add_argument("--request-body-file", help="Optional JSON request body file to replay as-is")
    parser.add_argument("--platform", help="Optional admin realtime platform filter")
    parser.add_argument("--group-id", type=int, help="Optional admin realtime group filter")
    parser.add_argument("--scenario", choices=["high-sticky", "low-sticky", "both"], default="both")
    parser.add_argument("--qps", type=float, default=20.0, help="Target requests per second")
    parser.add_argument("--concurrency", type=int, default=8, help="Worker count")
    parser.add_argument("--warmup-seconds", type=float, default=10.0)
    parser.add_argument("--measure-seconds", type=float, default=60.0)
    parser.add_argument("--sample-interval-seconds", type=float, default=5.0)
    parser.add_argument("--sticky-pool-size", type=int, default=64, help="Session pool size used by high-sticky scenario")
    parser.add_argument("--cooldown-seconds", type=float, default=3.0, help="Pause between scenarios when --scenario=both")
    parser.add_argument("--timeout-seconds", type=float, default=30.0)
    parser.add_argument("--output-json", help="Optional path for a JSON summary report")
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    if not args.request_body_file and not args.model:
        parser.error("--model is required when --request-body-file is not provided")
    if args.qps <= 0:
        parser.error("--qps must be > 0")
    if args.concurrency <= 0:
        parser.error("--concurrency must be > 0")
    if args.sample_interval_seconds <= 0:
        parser.error("--sample-interval-seconds must be > 0")
    if args.sticky_pool_size <= 0:
        parser.error("--sticky-pool-size must be > 0")

    gateway_headers = parse_header(args.gateway_header)
    admin_headers = parse_header(args.admin_header)
    if args.gateway_api_key:
        gateway_headers.setdefault("Authorization", f"Bearer {args.gateway_api_key}")
    if args.admin_token:
        admin_headers.setdefault("Authorization", f"Bearer {args.admin_token}")
    admin_headers.setdefault("Accept", "application/json")

    request_body_file = None
    if args.request_body_file:
        with open(args.request_body_file, "rb") as handle:
            request_body_file = handle.read()

    admin_url = join_url(args.admin_base_url, "/admin/ops/realtime-traffic")
    query: dict[str, str] = {"window": "1min"}
    if args.platform:
        query["platform"] = args.platform
    if args.group_id:
        query["group_id"] = str(args.group_id)
    admin_url = admin_url + "?" + urllib.parse.urlencode(query)

    run_id = uuid.uuid4().hex[:12]
    stop_event = threading.Event()

    def handle_signal(_signum: int, _frame: Any) -> None:
        stop_event.set()

    signal.signal(signal.SIGINT, handle_signal)
    signal.signal(signal.SIGTERM, handle_signal)

    scenarios = (
        [ScenarioConfig(name="high-sticky", sticky_pool_size=args.sticky_pool_size, warmup_seconds=args.warmup_seconds, measure_seconds=args.measure_seconds)]
        if args.scenario == "high-sticky"
        else [ScenarioConfig(name="low-sticky", sticky_pool_size=args.sticky_pool_size, warmup_seconds=args.warmup_seconds, measure_seconds=args.measure_seconds)]
        if args.scenario == "low-sticky"
        else [
            ScenarioConfig(name="high-sticky", sticky_pool_size=args.sticky_pool_size, warmup_seconds=args.warmup_seconds, measure_seconds=args.measure_seconds),
            ScenarioConfig(name="low-sticky", sticky_pool_size=args.sticky_pool_size, warmup_seconds=args.warmup_seconds, measure_seconds=args.measure_seconds),
        ]
    )

    summaries: list[ScenarioSummary] = []
    sample_log: dict[str, list[dict[str, Any]]] = {}
    try:
        initial_admin = fetch_admin_snapshot(admin_url, admin_headers, args.timeout_seconds)
        print(
            "admin snapshot ready: "
            f"qps_current={initial_admin.qps_current:.2f} "
            f"sticky_hit={initial_admin.sticky_hit_rate:.2%} "
            f"sticky_miss={initial_admin.sticky_miss_rate:.2%}"
        )
        for index, scenario in enumerate(scenarios):
            stop_event.clear()
            summary, samples = run_scenario(
                args,
                admin_url,
                admin_headers,
                gateway_headers,
                request_body_file,
                scenario,
                run_id,
                stop_event,
            )
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
            "gateway_base_url": args.gateway_base_url,
            "admin_url": admin_url,
            "request_path": args.request_path,
            "qps": args.qps,
            "concurrency": args.concurrency,
            "warmup_seconds": args.warmup_seconds,
            "measure_seconds": args.measure_seconds,
            "sticky_pool_size": args.sticky_pool_size,
            "summaries": [dataclasses.asdict(item) for item in summaries],
            "samples": sample_log,
        }
        with open(args.output_json, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, indent=2, ensure_ascii=True)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
