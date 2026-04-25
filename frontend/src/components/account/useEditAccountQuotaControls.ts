import { ref } from "vue";
import type { Account } from "@/types";

type RpmStrategy = "tiered" | "sticky_exempt";
type TlsFingerprintProfileOption = { id: number; name: string };

export function useEditAccountQuotaControls() {
  const windowCostEnabled = ref(false);
  const windowCostLimit = ref<number | null>(null);
  const windowCostStickyReserve = ref<number | null>(null);
  const sessionLimitEnabled = ref(false);
  const maxSessions = ref<number | null>(null);
  const sessionIdleTimeout = ref<number | null>(null);
  const rpmLimitEnabled = ref(false);
  const baseRpm = ref<number | null>(null);
  const rpmStrategy = ref<RpmStrategy>("tiered");
  const rpmStickyBuffer = ref<number | null>(null);
  const userMsgQueueMode = ref("");
  const tlsFingerprintEnabled = ref(false);
  const tlsFingerprintProfileId = ref<number | null>(null);
  const tlsFingerprintProfiles = ref<TlsFingerprintProfileOption[]>([]);
  const sessionIdMaskingEnabled = ref(false);
  const cacheTTLOverrideEnabled = ref(false);
  const cacheTTLOverrideTarget = ref("5m");
  const customBaseUrlEnabled = ref(false);
  const customBaseUrl = ref("");

  const resetQuotaControls = () => {
    windowCostEnabled.value = false;
    windowCostLimit.value = null;
    windowCostStickyReserve.value = null;
    sessionLimitEnabled.value = false;
    maxSessions.value = null;
    sessionIdleTimeout.value = null;
    rpmLimitEnabled.value = false;
    baseRpm.value = null;
    rpmStrategy.value = "tiered";
    rpmStickyBuffer.value = null;
    userMsgQueueMode.value = "";
    tlsFingerprintEnabled.value = false;
    tlsFingerprintProfileId.value = null;
    sessionIdMaskingEnabled.value = false;
    cacheTTLOverrideEnabled.value = false;
    cacheTTLOverrideTarget.value = "5m";
    customBaseUrlEnabled.value = false;
    customBaseUrl.value = "";
  };

  const hydrateQuotaControlsFromAccount = (account: Account) => {
    resetQuotaControls();

    if (
      account.platform !== "anthropic" ||
      (account.type !== "oauth" && account.type !== "setup-token")
    ) {
      return;
    }

    if (account.window_cost_limit != null && account.window_cost_limit > 0) {
      windowCostEnabled.value = true;
      windowCostLimit.value = account.window_cost_limit;
      windowCostStickyReserve.value = account.window_cost_sticky_reserve ?? 10;
    }

    if (account.max_sessions != null && account.max_sessions > 0) {
      sessionLimitEnabled.value = true;
      maxSessions.value = account.max_sessions;
      sessionIdleTimeout.value = account.session_idle_timeout_minutes ?? 5;
    }

    if (account.base_rpm != null && account.base_rpm > 0) {
      rpmLimitEnabled.value = true;
      baseRpm.value = account.base_rpm;
      rpmStrategy.value = (account.rpm_strategy as RpmStrategy) || "tiered";
      rpmStickyBuffer.value = account.rpm_sticky_buffer ?? null;
    }

    userMsgQueueMode.value = account.user_msg_queue_mode ?? "";

    if (account.enable_tls_fingerprint === true) {
      tlsFingerprintEnabled.value = true;
    }
    tlsFingerprintProfileId.value = account.tls_fingerprint_profile_id ?? null;

    if (account.session_id_masking_enabled === true) {
      sessionIdMaskingEnabled.value = true;
    }

    if (account.cache_ttl_override_enabled === true) {
      cacheTTLOverrideEnabled.value = true;
      cacheTTLOverrideTarget.value = account.cache_ttl_override_target || "5m";
    }

    if (account.custom_base_url_enabled === true) {
      customBaseUrlEnabled.value = true;
      customBaseUrl.value = account.custom_base_url || "";
    }
  };

  const setTlsFingerprintProfiles = (
    profiles: TlsFingerprintProfileOption[],
  ) => {
    tlsFingerprintProfiles.value = profiles;
  };

  return {
    baseRpm,
    cacheTTLOverrideEnabled,
    cacheTTLOverrideTarget,
    customBaseUrl,
    customBaseUrlEnabled,
    hydrateQuotaControlsFromAccount,
    maxSessions,
    resetQuotaControls,
    rpmLimitEnabled,
    rpmStickyBuffer,
    rpmStrategy,
    sessionIdMaskingEnabled,
    sessionIdleTimeout,
    sessionLimitEnabled,
    setTlsFingerprintProfiles,
    tlsFingerprintEnabled,
    tlsFingerprintProfileId,
    tlsFingerprintProfiles,
    userMsgQueueMode,
    windowCostEnabled,
    windowCostLimit,
    windowCostStickyReserve,
  };
}
