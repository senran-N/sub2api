import { describe, expect, it } from "vitest";
import { useEditAccountQuotaControls } from "../useEditAccountQuotaControls";
import type { Account } from "@/types";

function buildAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: "Claude",
    notes: "",
    platform: "anthropic",
    type: "oauth",
    credentials: {},
    extra: {},
    proxy_id: null,
    concurrency: 1,
    load_factor: null,
    priority: 1,
    rate_multiplier: 1,
    status: "active",
    group_ids: [],
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: "",
    updated_at: "",
    ...overrides,
  } as Account;
}

describe("useEditAccountQuotaControls", () => {
  it("hydrates Anthropic quota control state from account fields", () => {
    const controls = useEditAccountQuotaControls();

    controls.hydrateQuotaControlsFromAccount(
      buildAccount({
        window_cost_limit: 100,
        max_sessions: 4,
        base_rpm: 60,
        rpm_strategy: "sticky_exempt",
        rpm_sticky_buffer: 5,
        user_msg_queue_mode: "serialize",
        enable_tls_fingerprint: true,
        tls_fingerprint_profile_id: 9,
        session_id_masking_enabled: true,
        cache_ttl_override_enabled: true,
        cache_ttl_override_target: "1h",
        custom_base_url_enabled: true,
        custom_base_url: "https://relay.example.com",
      }),
    );

    expect(controls.windowCostEnabled.value).toBe(true);
    expect(controls.windowCostLimit.value).toBe(100);
    expect(controls.windowCostStickyReserve.value).toBe(10);
    expect(controls.sessionLimitEnabled.value).toBe(true);
    expect(controls.maxSessions.value).toBe(4);
    expect(controls.rpmLimitEnabled.value).toBe(true);
    expect(controls.rpmStrategy.value).toBe("sticky_exempt");
    expect(controls.rpmStickyBuffer.value).toBe(5);
    expect(controls.userMsgQueueMode.value).toBe("serialize");
    expect(controls.tlsFingerprintEnabled.value).toBe(true);
    expect(controls.tlsFingerprintProfileId.value).toBe(9);
    expect(controls.sessionIdMaskingEnabled.value).toBe(true);
    expect(controls.cacheTTLOverrideEnabled.value).toBe(true);
    expect(controls.cacheTTLOverrideTarget.value).toBe("1h");
    expect(controls.customBaseUrlEnabled.value).toBe(true);
    expect(controls.customBaseUrl.value).toBe("https://relay.example.com");
  });

  it("resets quota controls for unsupported account types", () => {
    const controls = useEditAccountQuotaControls();

    controls.hydrateQuotaControlsFromAccount(
      buildAccount({ window_cost_limit: 100 }),
    );
    controls.hydrateQuotaControlsFromAccount(
      buildAccount({ platform: "openai", type: "apikey" }),
    );

    expect(controls.windowCostEnabled.value).toBe(false);
    expect(controls.windowCostLimit.value).toBeNull();
    expect(controls.rpmStrategy.value).toBe("tiered");
    expect(controls.cacheTTLOverrideTarget.value).toBe("5m");
  });

  it("stores TLS fingerprint profile options", () => {
    const controls = useEditAccountQuotaControls();

    controls.setTlsFingerprintProfiles([{ id: 3, name: "Chrome" }]);

    expect(controls.tlsFingerprintProfiles.value).toEqual([
      { id: 3, name: "Chrome" },
    ]);
  });
});
