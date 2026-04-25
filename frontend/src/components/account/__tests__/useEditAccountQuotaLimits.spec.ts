import { describe, expect, it } from "vitest";
import { useEditAccountQuotaLimits } from "../useEditAccountQuotaLimits";
import type { Account } from "@/types";

function buildAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: "Account",
    notes: "",
    platform: "openai",
    type: "apikey",
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

describe("useEditAccountQuotaLimits", () => {
  it("hydrates compatible quota limits and notification settings", () => {
    const limits = useEditAccountQuotaLimits();

    limits.hydrateQuotaLimitsFromAccount(
      buildAccount({
        extra: {
          quota_limit: 100,
          quota_daily_limit: 10,
          quota_weekly_limit: 50,
          quota_daily_reset_mode: "fixed",
          quota_daily_reset_hour: 8,
          quota_weekly_reset_mode: "fixed",
          quota_weekly_reset_day: 1,
          quota_weekly_reset_hour: 9,
          quota_reset_timezone: "Asia/Shanghai",
          quota_notify_daily_enabled: true,
          quota_notify_daily_threshold: 2,
          quota_notify_daily_threshold_type: "fixed",
          quota_notify_weekly_enabled: true,
          quota_notify_weekly_threshold: 20,
          quota_notify_weekly_threshold_type: "percentage",
          quota_notify_total_enabled: true,
          quota_notify_total_threshold: 30,
          quota_notify_total_threshold_type: "percentage",
        },
      }),
    );

    expect(limits.editQuotaLimit.value).toBe(100);
    expect(limits.editQuotaDailyLimit.value).toBe(10);
    expect(limits.editQuotaWeeklyLimit.value).toBe(50);
    expect(limits.editDailyResetMode.value).toBe("fixed");
    expect(limits.editDailyResetHour.value).toBe(8);
    expect(limits.editWeeklyResetMode.value).toBe("fixed");
    expect(limits.editWeeklyResetDay.value).toBe(1);
    expect(limits.editWeeklyResetHour.value).toBe(9);
    expect(limits.editResetTimezone.value).toBe("Asia/Shanghai");
    expect(limits.editQuotaNotifyDailyEnabled.value).toBe(true);
    expect(limits.editQuotaNotifyDailyThreshold.value).toBe(2);
    expect(limits.editQuotaNotifyDailyThresholdType.value).toBe("fixed");
    expect(limits.editQuotaNotifyWeeklyEnabled.value).toBe(true);
    expect(limits.editQuotaNotifyWeeklyThreshold.value).toBe(20);
    expect(limits.editQuotaNotifyWeeklyThresholdType.value).toBe("percentage");
    expect(limits.editQuotaNotifyTotalEnabled.value).toBe(true);
    expect(limits.editQuotaNotifyTotalThreshold.value).toBe(30);
    expect(limits.editQuotaNotifyTotalThresholdType.value).toBe("percentage");
  });

  it("preserves numeric zero limits for Bedrock accounts", () => {
    const limits = useEditAccountQuotaLimits();

    limits.hydrateQuotaLimitsFromAccount(
      buildAccount({
        type: "bedrock",
        extra: {
          quota_limit: 0,
          quota_daily_limit: 0,
          quota_weekly_limit: 0,
        },
      }),
    );

    expect(limits.editQuotaLimit.value).toBe(0);
    expect(limits.editQuotaDailyLimit.value).toBe(0);
    expect(limits.editQuotaWeeklyLimit.value).toBe(0);
  });

  it("resets quota limits for unsupported account types", () => {
    const limits = useEditAccountQuotaLimits();

    limits.hydrateQuotaLimitsFromAccount(
      buildAccount({ extra: { quota_limit: 100 } }),
    );
    limits.hydrateQuotaLimitsFromAccount(
      buildAccount({ platform: "anthropic", type: "oauth", extra: {} }),
    );

    expect(limits.editQuotaLimit.value).toBeNull();
    expect(limits.editDailyResetMode.value).toBeNull();
    expect(limits.editQuotaNotifyDailyEnabled.value).toBeNull();
  });
});
