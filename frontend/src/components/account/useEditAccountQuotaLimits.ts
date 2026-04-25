import { ref } from "vue";
import type { Account } from "@/types";

type QuotaResetMode = "rolling" | "fixed" | null;
type QuotaThresholdType = "fixed" | "percentage" | null;

const getAccountExtra = (account: Account): Record<string, unknown> =>
  (account.extra as Record<string, unknown>) || {};

const readPositiveNumber = (value: unknown): number | null =>
  typeof value === "number" && value > 0 ? value : null;

const readNumber = (value: unknown): number | null =>
  typeof value === "number" ? value : null;

export function useEditAccountQuotaLimits() {
  const editQuotaLimit = ref<number | null>(null);
  const editQuotaDailyLimit = ref<number | null>(null);
  const editQuotaWeeklyLimit = ref<number | null>(null);
  const editDailyResetMode = ref<QuotaResetMode>(null);
  const editDailyResetHour = ref<number | null>(null);
  const editWeeklyResetMode = ref<QuotaResetMode>(null);
  const editWeeklyResetDay = ref<number | null>(null);
  const editWeeklyResetHour = ref<number | null>(null);
  const editResetTimezone = ref<string | null>(null);
  const editQuotaNotifyDailyEnabled = ref<boolean | null>(null);
  const editQuotaNotifyDailyThreshold = ref<number | null>(null);
  const editQuotaNotifyDailyThresholdType = ref<QuotaThresholdType>(null);
  const editQuotaNotifyWeeklyEnabled = ref<boolean | null>(null);
  const editQuotaNotifyWeeklyThreshold = ref<number | null>(null);
  const editQuotaNotifyWeeklyThresholdType = ref<QuotaThresholdType>(null);
  const editQuotaNotifyTotalEnabled = ref<boolean | null>(null);
  const editQuotaNotifyTotalThreshold = ref<number | null>(null);
  const editQuotaNotifyTotalThresholdType = ref<QuotaThresholdType>(null);

  const resetQuotaLimits = () => {
    editQuotaLimit.value = null;
    editQuotaDailyLimit.value = null;
    editQuotaWeeklyLimit.value = null;
    editDailyResetMode.value = null;
    editDailyResetHour.value = null;
    editWeeklyResetMode.value = null;
    editWeeklyResetDay.value = null;
    editWeeklyResetHour.value = null;
    editResetTimezone.value = null;
    editQuotaNotifyDailyEnabled.value = null;
    editQuotaNotifyDailyThreshold.value = null;
    editQuotaNotifyDailyThresholdType.value = null;
    editQuotaNotifyWeeklyEnabled.value = null;
    editQuotaNotifyWeeklyThreshold.value = null;
    editQuotaNotifyWeeklyThresholdType.value = null;
    editQuotaNotifyTotalEnabled.value = null;
    editQuotaNotifyTotalThreshold.value = null;
    editQuotaNotifyTotalThresholdType.value = null;
  };

  const hydrateQuotaLimitsFromAccount = (account: Account) => {
    resetQuotaLimits();

    if (
      account.type !== "apikey" &&
      account.type !== "upstream" &&
      account.type !== "bedrock"
    ) {
      return;
    }

    const extra = getAccountExtra(account);
    const readLimit =
      account.type === "bedrock" ? readNumber : readPositiveNumber;

    editQuotaLimit.value = readLimit(extra.quota_limit);
    editQuotaDailyLimit.value = readLimit(extra.quota_daily_limit);
    editQuotaWeeklyLimit.value = readLimit(extra.quota_weekly_limit);
    editDailyResetMode.value =
      (extra.quota_daily_reset_mode as QuotaResetMode) || null;
    editDailyResetHour.value = (extra.quota_daily_reset_hour as number) ?? null;
    editWeeklyResetMode.value =
      (extra.quota_weekly_reset_mode as QuotaResetMode) || null;
    editWeeklyResetDay.value = (extra.quota_weekly_reset_day as number) ?? null;
    editWeeklyResetHour.value =
      (extra.quota_weekly_reset_hour as number) ?? null;
    editResetTimezone.value = (extra.quota_reset_timezone as string) || null;
    editQuotaNotifyDailyEnabled.value =
      (extra.quota_notify_daily_enabled as boolean) || null;
    editQuotaNotifyDailyThreshold.value =
      (extra.quota_notify_daily_threshold as number) ?? null;
    editQuotaNotifyDailyThresholdType.value =
      (extra.quota_notify_daily_threshold_type as QuotaThresholdType) || null;
    editQuotaNotifyWeeklyEnabled.value =
      (extra.quota_notify_weekly_enabled as boolean) || null;
    editQuotaNotifyWeeklyThreshold.value =
      (extra.quota_notify_weekly_threshold as number) ?? null;
    editQuotaNotifyWeeklyThresholdType.value =
      (extra.quota_notify_weekly_threshold_type as QuotaThresholdType) || null;
    editQuotaNotifyTotalEnabled.value =
      (extra.quota_notify_total_enabled as boolean) || null;
    editQuotaNotifyTotalThreshold.value =
      (extra.quota_notify_total_threshold as number) ?? null;
    editQuotaNotifyTotalThresholdType.value =
      (extra.quota_notify_total_threshold_type as QuotaThresholdType) || null;
  };

  return {
    editDailyResetHour,
    editDailyResetMode,
    editQuotaDailyLimit,
    editQuotaLimit,
    editQuotaNotifyDailyEnabled,
    editQuotaNotifyDailyThreshold,
    editQuotaNotifyDailyThresholdType,
    editQuotaNotifyTotalEnabled,
    editQuotaNotifyTotalThreshold,
    editQuotaNotifyTotalThresholdType,
    editQuotaNotifyWeeklyEnabled,
    editQuotaNotifyWeeklyThreshold,
    editQuotaNotifyWeeklyThresholdType,
    editQuotaWeeklyLimit,
    editResetTimezone,
    editWeeklyResetDay,
    editWeeklyResetHour,
    editWeeklyResetMode,
    hydrateQuotaLimitsFromAccount,
    resetQuotaLimits,
  };
}
