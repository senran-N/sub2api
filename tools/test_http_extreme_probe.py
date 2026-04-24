#!/usr/bin/env python3

from __future__ import annotations

import json
import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import http_extreme_probe as probe  # noqa: E402


class HttpExtremeProbeScenarioTest(unittest.TestCase):
    def setUp(self) -> None:
        self.context = probe.ScenarioRequestContext(
            run_id="run123",
            admin_token="admin-token",
            gateway_api_key="sk-test",
            gateway_model="gpt-test",
            gateway_max_output_tokens=7,
            large_body_bytes=1024,
        )

    def test_resolve_suite_preserves_order_and_deduplicates(self) -> None:
        scenarios = probe.resolve_scenarios(["admin-read-suite", "admin-realtime"])
        self.assertEqual(
            [
                "admin-realtime",
                "admin-dashboard-stats",
                "admin-accounts-list",
                "admin-usage-list",
            ],
            [scenario.name for scenario in scenarios],
        )

    def test_gateway_success_suite_requires_gateway_api_key(self) -> None:
        scenarios = probe.resolve_scenarios(["gateway-success-suite"])
        self.assertTrue(all(scenario.requires_gateway_api_key for scenario in scenarios))

    def test_builds_streaming_responses_success_request(self) -> None:
        request = probe.build_scenario_request("gateway-responses-stream", 3, self.context)

        self.assertEqual("POST", request.method)
        self.assertEqual("/v1/responses", request.path)
        self.assertEqual("Bearer sk-test", request.headers["Authorization"])

        payload = json.loads(request.body.decode("utf-8"))
        self.assertEqual("gpt-test", payload["model"])
        self.assertTrue(payload["stream"])
        self.assertEqual(7, payload["max_output_tokens"])
        self.assertIn("gateway-responses-stream", payload["input"][0]["content"][0]["text"])

    def test_builds_streaming_chat_success_request(self) -> None:
        request = probe.build_scenario_request("gateway-chat-stream", 4, self.context)

        self.assertEqual("POST", request.method)
        self.assertEqual("/v1/chat/completions", request.path)

        payload = json.loads(request.body.decode("utf-8"))
        self.assertEqual("gpt-test", payload["model"])
        self.assertTrue(payload["stream"])
        self.assertEqual(7, payload["max_tokens"])
        self.assertIn("gateway-chat-stream", payload["messages"][0]["content"])

    def test_builds_admin_usage_list_request(self) -> None:
        request = probe.build_scenario_request("admin-usage-list", 1, self.context)

        self.assertEqual("GET", request.method)
        self.assertIn("/api/v1/admin/usage?", request.path)
        self.assertIn("exact_total=false", request.path)
        self.assertEqual("Bearer admin-token", request.headers["Authorization"])


if __name__ == "__main__":
    unittest.main()
