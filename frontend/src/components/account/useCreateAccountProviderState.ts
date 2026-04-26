import { ref } from "vue";
import type { GrokSessionBatchImportResult } from "@/api/admin/accounts";
import {
  claudeModels,
} from "@/composables/useModelWhitelist";
import {
  OPENAI_WS_MODE_OFF,
  type OpenAIWSMode,
} from "@/utils/openaiWsMode";
import {
  type BedrockAuthMode,
  type GeminiAIStudioTier,
  type GeminiGcpTier,
  type GeminiGoogleOneTier,
  type GeminiOAuthType,
} from "@/components/account/createAccountModalHelpers";
import {
  DEFAULT_POOL_MODE_RETRY_COUNT,
  type ModelMapping,
  type TempUnschedRuleForm,
} from "@/components/account/credentialsBuilder";

export type GrokSessionInputMode = "single" | "batch";

export function useCreateAccountProviderState() {
  const grokSessionInputMode = ref<GrokSessionInputMode>("single");
  const grokSessionToken = ref("");
  const grokSessionBatchInput = ref("");
  const grokSessionBatchDryRun = ref(false);
  const grokSessionBatchTestAfterCreate = ref(true);
  const grokSessionBatchResult = ref<GrokSessionBatchImportResult | null>(null);

  const editQuotaLimit = ref<number | null>(null);
  const editQuotaDailyLimit = ref<number | null>(null);
  const editQuotaWeeklyLimit = ref<number | null>(null);
  const editDailyResetMode = ref<"rolling" | "fixed" | null>(null);
  const editDailyResetHour = ref<number | null>(null);
  const editWeeklyResetMode = ref<"rolling" | "fixed" | null>(null);
  const editWeeklyResetDay = ref<number | null>(null);
  const editWeeklyResetHour = ref<number | null>(null);
  const editResetTimezone = ref<string | null>(null);
  const editQuotaNotifyDailyEnabled = ref<boolean | null>(null);
  const editQuotaNotifyDailyThreshold = ref<number | null>(null);
  const editQuotaNotifyDailyThresholdType = ref<"fixed" | "percentage" | null>(null);
  const editQuotaNotifyWeeklyEnabled = ref<boolean | null>(null);
  const editQuotaNotifyWeeklyThreshold = ref<number | null>(null);
  const editQuotaNotifyWeeklyThresholdType = ref<"fixed" | "percentage" | null>(null);
  const editQuotaNotifyTotalEnabled = ref<boolean | null>(null);
  const editQuotaNotifyTotalThreshold = ref<number | null>(null);
  const editQuotaNotifyTotalThresholdType = ref<"fixed" | "percentage" | null>(null);

  const modelMappings = ref<ModelMapping[]>([]);
  const modelRestrictionMode = ref<"whitelist" | "mapping">("whitelist");
  const allowedModels = ref<string[]>([]);
  const poolModeEnabled = ref(false);
  const poolModeRetryCount = ref(DEFAULT_POOL_MODE_RETRY_COUNT);
  const customErrorCodesEnabled = ref(false);
  const selectedErrorCodes = ref<number[]>([]);
  const customErrorCodeInput = ref<number | null>(null);
  const interceptWarmupRequests = ref(false);
  const autoPauseOnExpired = ref(true);

  const openaiPassthroughEnabled = ref(false);
  const openaiOAuthResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
  const openaiAPIKeyResponsesWebSocketV2Mode = ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
  const codexCLIOnlyEnabled = ref(false);
  const anthropicPassthroughEnabled = ref(false);

  const mixedScheduling = ref(false);
  const allowOverages = ref(false);
  const antigravityAccountType = ref<"oauth" | "upstream">("oauth");
  const upstreamBaseUrl = ref("");
  const upstreamApiKey = ref("");
  const antigravityModelRestrictionMode = ref<"whitelist" | "mapping">("whitelist");
  const antigravityWhitelistModels = ref<string[]>([]);
  const antigravityModelMappings = ref<ModelMapping[]>([]);

  const bedrockAuthMode = ref<BedrockAuthMode>("sigv4");
  const bedrockAccessKeyId = ref("");
  const bedrockSecretAccessKey = ref("");
  const bedrockSessionToken = ref("");
  const bedrockRegion = ref("us-east-1");
  const bedrockForceGlobal = ref(false);
  const bedrockApiKeyValue = ref("");

  const tempUnschedEnabled = ref(false);
  const tempUnschedRules = ref<TempUnschedRuleForm[]>([]);

  const geminiOAuthType = ref<GeminiOAuthType>("google_one");
  const geminiAIStudioOAuthEnabled = ref(false);
  const geminiTierGoogleOne = ref<GeminiGoogleOneTier>("google_one_free");
  const geminiTierGcp = ref<GeminiGcpTier>("gcp_standard");
  const geminiTierAIStudio = ref<GeminiAIStudioTier>("aistudio_free");

  const windowCostEnabled = ref(false);
  const windowCostLimit = ref<number | null>(null);
  const windowCostStickyReserve = ref<number | null>(null);
  const sessionLimitEnabled = ref(false);
  const maxSessions = ref<number | null>(null);
  const sessionIdleTimeout = ref<number | null>(null);
  const rpmLimitEnabled = ref(false);
  const baseRpm = ref<number | null>(null);
  const rpmStrategy = ref<"tiered" | "sticky_exempt">("tiered");
  const rpmStickyBuffer = ref<number | null>(null);
  const userMsgQueueMode = ref("");
  const tlsFingerprintEnabled = ref(false);
  const tlsFingerprintProfileId = ref<number | null>(null);
  const tlsFingerprintProfiles = ref<{ id: number; name: string }[]>([]);
  const sessionIdMaskingEnabled = ref(false);
  const cacheTTLOverrideEnabled = ref(false);
  const cacheTTLOverrideTarget = ref<string>("5m");
  const customBaseUrlEnabled = ref(false);
  const customBaseUrl = ref("");

  const clearGrokSessionBatchResult = () => {
    grokSessionBatchResult.value = null;
  };

  const resetGrokSessionImportState = (resetMode = true) => {
    if (resetMode) {
      grokSessionInputMode.value = "single";
    }
    grokSessionToken.value = "";
    grokSessionBatchInput.value = "";
    grokSessionBatchDryRun.value = false;
    grokSessionBatchTestAfterCreate.value = true;
    clearGrokSessionBatchResult();
  };

  const resetBedrockCredentialState = () => {
    bedrockAccessKeyId.value = "";
    bedrockSecretAccessKey.value = "";
    bedrockSessionToken.value = "";
    bedrockRegion.value = "us-east-1";
    bedrockForceGlobal.value = false;
    bedrockAuthMode.value = "sigv4";
    bedrockApiKeyValue.value = "";
  };

  const resetOpenAICreateState = () => {
    openaiPassthroughEnabled.value = false;
    openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
    openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
    codexCLIOnlyEnabled.value = false;
  };

  const resetAnthropicQuotaControlState = () => {
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

  const resetAntigravityModelState = () => {
    antigravityModelRestrictionMode.value = "mapping";
    antigravityWhitelistModels.value = [];
    antigravityModelMappings.value = [];
  };

  const resetAntigravityCreateState = () => {
    allowOverages.value = false;
    antigravityAccountType.value = "oauth";
    upstreamBaseUrl.value = "";
    upstreamApiKey.value = "";
    resetAntigravityModelState();
  };

  const resetGeminiSelectionState = () => {
    geminiOAuthType.value = "code_assist";
    geminiTierGoogleOne.value = "google_one_free";
    geminiTierGcp.value = "gcp_standard";
    geminiTierAIStudio.value = "aistudio_free";
  };

  const resetCustomErrorCodeState = () => {
    customErrorCodesEnabled.value = false;
    selectedErrorCodes.value = [];
    customErrorCodeInput.value = null;
  };

  const resetQuotaResetState = () => {
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

  const resetCreateProviderState = () => {
    resetGrokSessionImportState();
    resetQuotaResetState();
    modelMappings.value = [];
    modelRestrictionMode.value = "whitelist";
    allowedModels.value = [...claudeModels];
    poolModeEnabled.value = false;
    poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT;
    resetCustomErrorCodeState();
    interceptWarmupRequests.value = false;
    autoPauseOnExpired.value = true;
    resetOpenAICreateState();
    anthropicPassthroughEnabled.value = false;
    resetAnthropicQuotaControlState();
    resetAntigravityCreateState();
    tempUnschedEnabled.value = false;
    tempUnschedRules.value = [];
    resetGeminiSelectionState();
    resetBedrockCredentialState();
  };

  return {
    allowOverages,
    allowedModels,
    anthropicPassthroughEnabled,
    antigravityAccountType,
    antigravityModelMappings,
    antigravityModelRestrictionMode,
    antigravityWhitelistModels,
    apiState: {
      modelMappings,
      modelRestrictionMode,
      allowedModels,
      poolModeEnabled,
      poolModeRetryCount,
      customErrorCodesEnabled,
      selectedErrorCodes,
      customErrorCodeInput,
    },
    autoPauseOnExpired,
    baseRpm,
    bedrockAccessKeyId,
    bedrockApiKeyValue,
    bedrockAuthMode,
    bedrockForceGlobal,
    bedrockRegion,
    bedrockSecretAccessKey,
    bedrockSessionToken,
    cacheTTLOverrideEnabled,
    cacheTTLOverrideTarget,
    clearGrokSessionBatchResult,
    codexCLIOnlyEnabled,
    customBaseUrl,
    customBaseUrlEnabled,
    customErrorCodeInput,
    customErrorCodesEnabled,
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
    geminiAIStudioOAuthEnabled,
    geminiOAuthType,
    geminiTierAIStudio,
    geminiTierGcp,
    geminiTierGoogleOne,
    grokSessionBatchDryRun,
    grokSessionBatchInput,
    grokSessionBatchResult,
    grokSessionBatchTestAfterCreate,
    grokSessionInputMode,
    grokSessionToken,
    interceptWarmupRequests,
    maxSessions,
    mixedScheduling,
    modelMappings,
    modelRestrictionMode,
    openaiAPIKeyResponsesWebSocketV2Mode,
    openaiOAuthResponsesWebSocketV2Mode,
    openaiPassthroughEnabled,
    poolModeEnabled,
    poolModeRetryCount,
    resetAntigravityCreateState,
    resetAntigravityModelState,
    resetAnthropicQuotaControlState,
    resetBedrockCredentialState,
    resetCreateProviderState,
    resetCustomErrorCodeState,
    resetGeminiSelectionState,
    resetGrokSessionImportState,
    resetOpenAICreateState,
    resetQuotaResetState,
    rpmLimitEnabled,
    rpmStickyBuffer,
    rpmStrategy,
    selectedErrorCodes,
    sessionIdMaskingEnabled,
    sessionIdleTimeout,
    sessionLimitEnabled,
    tempUnschedEnabled,
    tempUnschedRules,
    tlsFingerprintEnabled,
    tlsFingerprintProfileId,
    tlsFingerprintProfiles,
    upstreamApiKey,
    upstreamBaseUrl,
    userMsgQueueMode,
    windowCostEnabled,
    windowCostLimit,
    windowCostStickyReserve,
  };
}
