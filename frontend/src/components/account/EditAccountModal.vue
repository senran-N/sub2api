<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.editAccount')"
    width="normal"
    @close="handleClose"
  >
    <form
      v-if="account"
      id="edit-account-form"
      @submit.prevent="handleSubmit"
      class="space-y-5"
    >
      <EditAccountCoreFieldsSection
        v-model:name="form.name"
        v-model:notes="form.notes"
        v-model:proxy-id="form.proxy_id"
        v-model:concurrency="form.concurrency"
        v-model:load-factor="form.load_factor"
        v-model:priority="form.priority"
        v-model:rate-multiplier="form.rate_multiplier"
        v-model:expires-at="expiresAtInput"
        v-model:status="form.status"
        v-model:allow-overages="allowOverages"
        v-model:group-ids="form.group_ids"
        :proxies="proxies"
        :platform="account.platform"
        :mixed-scheduling="mixedScheduling"
        :groups="groups"
        :simple-mode="authStore.isSimpleMode"
        :status-options="statusOptions"
      />

      <GrokRuntimeSummary :account="account" />

      <CompatibleCredentialsSection
        v-if="showCompatibleCredentialsForm"
        v-model:base-url="editBaseUrl"
        v-model:api-key-value="editApiKey"
        v-model:model-restriction-mode="modelRestrictionMode"
        v-model:allowed-models="allowedModels"
        v-model:pool-mode-enabled="poolModeEnabled"
        v-model:pool-mode-retry-count="poolModeRetryCount"
        v-model:custom-error-codes-enabled="customErrorCodesEnabled"
        v-model:custom-error-code-input="customErrorCodeInput"
        :platform="account.platform"
        :base-url-presets="compatibleBaseUrlPresets"
        :base-url-placeholder="baseUrlPlaceholder"
        :base-url-hint="baseUrlHint"
        :api-key-label="t('admin.accounts.apiKey')"
        :api-key-placeholder="apiKeyPlaceholder"
        :api-key-hint="t('admin.accounts.leaveEmptyToKeep')"
        api-key-autocomplete="new-password"
        :ignore-password-managers="true"
        :show-gemini-api-key-tier="false"
        :show-model-restriction="account.platform !== 'antigravity'"
        :mappings="modelMappings"
        :preset-mappings="presetMappings"
        :mapping-key="getModelMappingKey"
        :model-restriction-disabled="isOpenAIModelRestrictionDisabled"
        :selected-error-codes="selectedErrorCodes"
        @add-mapping="addModelMapping"
        @remove-mapping="removeModelMapping"
        @add-preset="addPresetMapping"
        @update-mapping="updateModelMapping"
        @toggle-code="toggleErrorCode"
        @add-code="addCustomErrorCode"
        @remove-code="removeErrorCode"
      />

      <ModelRestrictionSection
        v-if="account.platform === 'openai' && account.type === 'oauth'"
        v-model:mode="modelRestrictionMode"
        v-model:allowed-models="allowedModels"
        platform="openai"
        :mappings="modelMappings"
        :preset-mappings="presetMappings"
        :mapping-key="getModelMappingKey"
        :disabled="isOpenAIModelRestrictionDisabled"
        @add-mapping="addModelMapping"
        @remove-mapping="removeModelMapping"
        @add-preset="addPresetMapping"
        @update-mapping="updateModelMapping"
      />

      <EditGrokSessionCredentialsSection
        v-if="account.type === 'session'"
        v-model:session-token="editSessionToken"
      />

      <EditBedrockCredentialsSection
        v-if="account.type === 'bedrock'"
        :auth-mode="editBedrockAuthMode"
        v-model:access-key-id="editBedrockAccessKeyId"
        v-model:secret-access-key="editBedrockSecretAccessKey"
        v-model:session-token="editBedrockSessionToken"
        v-model:api-key-value="editBedrockApiKeyValue"
        v-model:region="editBedrockRegion"
        v-model:force-global="editBedrockForceGlobal"
        v-model:model-restriction-mode="modelRestrictionMode"
        v-model:allowed-models="allowedModels"
        v-model:pool-mode-enabled="poolModeEnabled"
        v-model:pool-mode-retry-count="poolModeRetryCount"
        :mappings="modelMappings"
        :preset-mappings="bedrockPresets"
        :mapping-key="getModelMappingKey"
        @add-mapping="addModelMapping"
        @remove-mapping="removeModelMapping"
        @add-preset="addPresetMapping"
        @update-mapping="updateModelMapping"
      />

      <AntigravityModelMappingSection
        v-if="account.platform === 'antigravity'"
        :mappings="antigravityModelMappings"
        :preset-mappings="antigravityPresetMappings"
        :mapping-key="getAntigravityModelMappingKey"
        @add="addAntigravityModelMapping"
        @remove="removeAntigravityModelMapping"
        @add-preset="addAntigravityPresetMapping"
        @update-mapping="updateAntigravityModelMapping"
      />

      <TempUnschedRulesSection
        v-model:enabled="tempUnschedEnabled"
        :presets="tempUnschedPresets"
        :rules="tempUnschedRules"
        :rule-key="getTempUnschedRuleKey"
        @add-rule="addTempUnschedRule"
        @remove-rule="removeTempUnschedRule"
        @move-rule="moveTempUnschedRule"
        @update-rule="updateTempUnschedRule"
      />

      <WarmupSection
        v-if="showWarmupSection"
        v-model:enabled="interceptWarmupRequests"
      />

      <OpenAIOptionsSection
        v-if="showOpenAIRuntimeSection"
        v-model:passthrough-enabled="openaiPassthroughEnabled"
        v-model:ws-mode="openaiResponsesWebSocketV2Mode"
        v-model:codex-cli-only-enabled="codexCLIOnlyEnabled"
        :account-category="openAIAccountCategory"
        :ws-mode-options="openAIWSModeOptions"
        :ws-mode-concurrency-hint-key="openAIWSModeConcurrencyHintKey"
      />

      <AnthropicOptionsSection
        v-if="showAnthropicAPIKeyRuntimeSection"
        v-model:api-key-passthrough-enabled="anthropicPassthroughEnabled"
        account-category="apikey"
      />

      <QuotaLimitSection
        v-if="showQuotaLimitSection"
        v-model:total-limit="editQuotaLimit"
        v-model:daily-limit="editQuotaDailyLimit"
        v-model:weekly-limit="editQuotaWeeklyLimit"
        v-model:daily-reset-mode="editDailyResetMode"
        v-model:daily-reset-hour="editDailyResetHour"
        v-model:weekly-reset-mode="editWeeklyResetMode"
        v-model:weekly-reset-day="editWeeklyResetDay"
        v-model:weekly-reset-hour="editWeeklyResetHour"
        v-model:reset-timezone="editResetTimezone"
        v-model:notify-daily-enabled="editQuotaNotifyDailyEnabled"
        v-model:notify-daily-threshold="editQuotaNotifyDailyThreshold"
        v-model:notify-daily-threshold-type="
          editQuotaNotifyDailyThresholdType
        "
        v-model:notify-weekly-enabled="editQuotaNotifyWeeklyEnabled"
        v-model:notify-weekly-threshold="editQuotaNotifyWeeklyThreshold"
        v-model:notify-weekly-threshold-type="
          editQuotaNotifyWeeklyThresholdType
        "
        v-model:notify-total-enabled="editQuotaNotifyTotalEnabled"
        v-model:notify-total-threshold="editQuotaNotifyTotalThreshold"
        v-model:notify-total-threshold-type="
          editQuotaNotifyTotalThresholdType
        "
      />

      <AutoPauseOnExpiredSection v-model:enabled="autoPauseOnExpired" />

      <AnthropicQuotaControlsSection
        v-if="showAnthropicQuotaControls"
        v-model:window-cost-enabled="windowCostEnabled"
        v-model:window-cost-limit="windowCostLimit"
        v-model:window-cost-sticky-reserve="windowCostStickyReserve"
        v-model:session-limit-enabled="sessionLimitEnabled"
        v-model:max-sessions="maxSessions"
        v-model:session-idle-timeout="sessionIdleTimeout"
        v-model:rpm-limit-enabled="rpmLimitEnabled"
        v-model:base-rpm="baseRpm"
        v-model:rpm-strategy="rpmStrategy"
        v-model:rpm-sticky-buffer="rpmStickyBuffer"
        v-model:user-msg-queue-mode="userMsgQueueMode"
        v-model:tls-fingerprint-enabled="tlsFingerprintEnabled"
        v-model:tls-fingerprint-profile-id="tlsFingerprintProfileId"
        v-model:session-id-masking-enabled="sessionIdMaskingEnabled"
        v-model:cache-ttl-override-enabled="cacheTTLOverrideEnabled"
        v-model:cache-ttl-override-target="cacheTTLOverrideTarget"
        v-model:custom-base-url-enabled="customBaseUrlEnabled"
        v-model:custom-base-url="customBaseUrl"
        :user-msg-queue-mode-options="umqModeOptions"
        :tls-fingerprint-profiles="tlsFingerprintProfiles"
      />

    </form>

    <template #footer>
      <div v-if="account" class="flex justify-end gap-3">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t("common.cancel") }}
        </button>
        <button
          type="submit"
          form="edit-account-form"
          :disabled="submitting"
          class="btn btn-primary"
          data-tour="account-form-submit"
        >
          <svg
            v-if="submitting"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            ></circle>
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {{ submitting ? t("admin.accounts.updating") : t("common.update") }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <!-- Mixed Channel Warning Dialog -->
  <ConfirmDialog
    :show="showMixedChannelWarning"
    :title="t('admin.accounts.mixedChannelWarningTitle')"
    :message="mixedChannelWarningMessageText"
    :confirm-text="t('common.confirm')"
    :cancel-text="t('common.cancel')"
    :danger="true"
    @confirm="handleMixedChannelConfirm"
    @cancel="handleMixedChannelCancel"
  />
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores/app";
import { useAuthStore } from "@/stores/auth";
import { adminAPI } from "@/api/admin";
import type {
  Account,
  AccountType,
  Proxy,
  AdminGroup,
  CheckMixedChannelResponse,
  UpdateAccountRequest,
} from "@/types";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import AntigravityModelMappingSection from "@/components/account/AntigravityModelMappingSection.vue";
import AnthropicOptionsSection from "@/components/account/AnthropicOptionsSection.vue";
import AnthropicQuotaControlsSection from "@/components/account/AnthropicQuotaControlsSection.vue";
import AutoPauseOnExpiredSection from "@/components/account/AutoPauseOnExpiredSection.vue";
import CompatibleCredentialsSection from "@/components/account/CompatibleCredentialsSection.vue";
import EditBedrockCredentialsSection from "@/components/account/EditBedrockCredentialsSection.vue";
import EditAccountCoreFieldsSection from "@/components/account/EditAccountCoreFieldsSection.vue";
import EditGrokSessionCredentialsSection from "@/components/account/EditGrokSessionCredentialsSection.vue";
import ModelRestrictionSection from "@/components/account/ModelRestrictionSection.vue";
import OpenAIOptionsSection from "@/components/account/OpenAIOptionsSection.vue";
import QuotaLimitSection from "@/components/account/QuotaLimitSection.vue";
import TempUnschedRulesSection from "@/components/account/TempUnschedRulesSection.vue";
import WarmupSection from "@/components/account/WarmupSection.vue";
import GrokRuntimeSummary from "@/components/account/GrokRuntimeSummary.vue";
import { useEditAccountQuotaControls } from "@/components/account/useEditAccountQuotaControls";
import {
  buildCompatibleBaseUrlPresets,
  buildAccountOpenAIWSModeOptions,
  buildAccountTempUnschedPresets,
  buildAccountUmqModeOptions,
  buildEditAccountBasePayload,
  buildMixedChannelDetails,
  createDefaultEditAccountForm,
  hydrateEditAccountForm,
  needsMixedChannelCheck,
  resolveAccountApiKeyPlaceholder,
  resolveAccountBaseUrlHint,
  resolveAccountBaseUrlPlaceholder,
  resolveMixedChannelWarningMessage,
  type EditAccountForm,
} from "@/components/account/accountModalShared";
import {
  createEmptyModelRestrictionState,
  deriveAntigravityModelMappings,
  deriveModelRestrictionStateFromMapping,
  deriveOpenAIExtraState,
} from "@/components/account/editAccountModalHelpers";
import type {
  BedrockAuthMode,
  CreateAccountCategory,
} from "@/components/account/createAccountModalHelpers";
import {
  buildEditAccountMutationPayload,
  resolveAccountMutationPayloadErrorKey,
} from "@/components/account/accountMutationPayload";
import {
  accountMutationProfileHasSection,
  resolveAccountMutationProfile,
} from "@/components/account/accountMutationProfiles";
import {
  createTempUnschedRule,
  DEFAULT_POOL_MODE_RETRY_COUNT,
  getDefaultBaseURL,
  loadTempUnschedRuleState,
  moveItemInPlace,
  normalizePoolModeRetryCount,
  type ModelMapping,
  type TempUnschedRuleForm,
} from "@/components/account/credentialsBuilder";
import {
  appendEmptyModelMapping,
  appendPresetModelMapping,
  confirmCustomErrorCodeSelection,
  removeModelMappingAt,
} from "@/components/account/accountModalInteractions";
import {
  formatDateTimeLocalInput,
  parseDateTimeLocalInput,
} from "@/utils/format";
import { createStableObjectKeyResolver } from "@/utils/stableObjectKey";
import {
  OPENAI_WS_MODE_OFF,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode,
} from "@/utils/openaiWsMode";
import { resolveRequestErrorMessage } from "@/utils/requestError";
import {
  ensureModelCatalogLoaded,
  getPresetMappingsByPlatform,
} from "@/composables/useModelWhitelist";

interface Props {
  show: boolean;
  account: Account | null;
  proxies: Proxy[];
  groups: AdminGroup[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
  updated: [account: Account];
}>();

const { t } = useI18n();
const appStore = useAppStore();
const authStore = useAuthStore();

// Platform-specific hint for Base URL
const baseUrlHint = computed(() => {
  return resolveAccountBaseUrlHint(props.account?.platform, t);
});

const baseUrlPlaceholder = computed(() => {
  return resolveAccountBaseUrlPlaceholder(props.account?.platform, t);
});

const apiKeyPlaceholder = computed(() => {
  return resolveAccountApiKeyPlaceholder(props.account?.platform, t);
});

const compatibleBaseUrlPresets = computed(() => {
  return buildCompatibleBaseUrlPresets(props.account?.platform, t);
});

const antigravityPresetMappings = computed(() =>
  getPresetMappingsByPlatform("antigravity"),
);
const bedrockPresets = computed(() => getPresetMappingsByPlatform("bedrock"));

// State
const submitting = ref(false);
const editBaseUrl = ref(getDefaultBaseURL("anthropic"));
const editApiKey = ref("");
const editSessionToken = ref("");
// Bedrock credentials
const editBedrockAccessKeyId = ref("");
const editBedrockSecretAccessKey = ref("");
const editBedrockSessionToken = ref("");
const editBedrockRegion = ref("");
const editBedrockForceGlobal = ref(false);
const editBedrockApiKeyValue = ref("");
const isBedrockAPIKeyMode = computed(
  () =>
    props.account?.type === "bedrock" &&
    (props.account?.credentials as Record<string, unknown>)?.auth_mode ===
      "apikey",
);
const editBedrockAuthMode = computed<BedrockAuthMode>(() =>
  isBedrockAPIKeyMode.value ? "apikey" : "sigv4",
);
const modelMappings = ref<ModelMapping[]>([]);
const modelRestrictionMode = ref<"whitelist" | "mapping">("whitelist");
const allowedModels = ref<string[]>([]);
const poolModeEnabled = ref(false);
const poolModeRetryCount = ref(DEFAULT_POOL_MODE_RETRY_COUNT);
const customErrorCodesEnabled = ref(false);
const selectedErrorCodes = ref<number[]>([]);
const customErrorCodeInput = ref<number | null>(null);
const interceptWarmupRequests = ref(false);
const autoPauseOnExpired = ref(false);
const mixedScheduling = ref(false); // For antigravity accounts: enable mixed scheduling
const allowOverages = ref(false); // For antigravity accounts: enable AI Credits overages
const antigravityModelRestrictionMode = ref<"whitelist" | "mapping">(
  "whitelist",
);
const antigravityWhitelistModels = ref<string[]>([]);
const antigravityModelMappings = ref<ModelMapping[]>([]);
const tempUnschedEnabled = ref(false);
const tempUnschedRules = ref<TempUnschedRuleForm[]>([]);
const getModelMappingKey =
  createStableObjectKeyResolver<ModelMapping>("edit-model-mapping");
const getAntigravityModelMappingKey =
  createStableObjectKeyResolver<ModelMapping>("edit-antigravity-model-mapping");
const getTempUnschedRuleKey =
  createStableObjectKeyResolver<TempUnschedRuleForm>("edit-temp-unsched-rule");

const showMixedChannelWarning = ref(false);
const mixedChannelWarningDetails = ref<{
  groupName: string;
  currentPlatform: string;
  otherPlatform: string;
} | null>(null);
const mixedChannelWarningRawMessage = ref("");
const mixedChannelWarningAction = ref<(() => Promise<void>) | null>(null);
const antigravityMixedChannelConfirmed = ref(false);

const {
  baseRpm,
  cacheTTLOverrideEnabled,
  cacheTTLOverrideTarget,
  customBaseUrl,
  customBaseUrlEnabled,
  hydrateQuotaControlsFromAccount,
  maxSessions,
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
} = useEditAccountQuotaControls();
const umqModeOptions = computed(() => buildAccountUmqModeOptions(t));

// OpenAI 自动透传开关（OAuth/API Key）
const openaiPassthroughEnabled = ref(false);
const openaiOAuthResponsesWebSocketV2Mode =
  ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
const openaiAPIKeyResponsesWebSocketV2Mode =
  ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
const codexCLIOnlyEnabled = ref(false);
const anthropicPassthroughEnabled = ref(false);
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
const editQuotaNotifyDailyThresholdType = ref<"fixed" | "percentage" | null>(
  null,
);
const editQuotaNotifyWeeklyEnabled = ref<boolean | null>(null);
const editQuotaNotifyWeeklyThreshold = ref<number | null>(null);
const editQuotaNotifyWeeklyThresholdType = ref<"fixed" | "percentage" | null>(
  null,
);
const editQuotaNotifyTotalEnabled = ref<boolean | null>(null);
const editQuotaNotifyTotalThreshold = ref<number | null>(null);
const editQuotaNotifyTotalThresholdType = ref<"fixed" | "percentage" | null>(
  null,
);
const openAIWSModeOptions = computed(() => buildAccountOpenAIWSModeOptions(t));
const openaiResponsesWebSocketV2Mode = computed({
  get: () => {
    if (props.account?.type === "apikey") {
      return openaiAPIKeyResponsesWebSocketV2Mode.value;
    }
    return openaiOAuthResponsesWebSocketV2Mode.value;
  },
  set: (mode: OpenAIWSMode) => {
    if (props.account?.type === "apikey") {
      openaiAPIKeyResponsesWebSocketV2Mode.value = mode;
      return;
    }
    openaiOAuthResponsesWebSocketV2Mode.value = mode;
  },
});
const openAIWSModeConcurrencyHintKey = computed(() =>
  resolveOpenAIWSModeConcurrencyHintKey(openaiResponsesWebSocketV2Mode.value),
);
const isOpenAIModelRestrictionDisabled = computed(
  () => props.account?.platform === "openai" && openaiPassthroughEnabled.value,
);

const mutationProfile = computed(() => {
  const account = props.account;
  return account
    ? resolveAccountMutationProfile(account.platform, account.type)
    : null;
});

const showCompatibleCredentialsForm = computed(() => {
  return accountMutationProfileHasSection(
    mutationProfile.value,
    "compatible-credentials",
  );
});

const showQuotaLimitSection = computed(() => {
  return accountMutationProfileHasSection(mutationProfile.value, "quota-limits");
});

const showWarmupSection = computed(() => {
  return accountMutationProfileHasSection(mutationProfile.value, "warmup");
});

const showOpenAIRuntimeSection = computed(() => {
  return accountMutationProfileHasSection(
    mutationProfile.value,
    "openai-runtime",
  );
});

const showAnthropicAPIKeyRuntimeSection = computed(() => {
  const account = props.account;
  return account?.platform === "anthropic" && account.type === "apikey";
});

const showAnthropicQuotaControls = computed(() => {
  return accountMutationProfileHasSection(
    mutationProfile.value,
    "anthropic-runtime",
  );
});

const openAIAccountCategory = computed<CreateAccountCategory>(() =>
  props.account?.type === "apikey" ? "apikey" : "oauth-based",
);

watch(
  () => props.account?.platform,
  (platform) => {
    if (platform === "grok") {
      void ensureModelCatalogLoaded(platform);
    }
  },
  { immediate: true },
);

// Computed: current preset mappings based on platform
const presetMappings = computed(() =>
  getPresetMappingsByPlatform(props.account?.platform || "anthropic"),
);
const tempUnschedPresets = computed(() => buildAccountTempUnschedPresets(t));

// Computed: default base URL based on platform
const defaultBaseUrl = computed(() => {
  return getDefaultBaseURL(props.account?.platform || "anthropic");
});

const mixedChannelWarningMessageText = computed(() => {
  return resolveMixedChannelWarningMessage({
    details: mixedChannelWarningDetails.value,
    rawMessage: mixedChannelWarningRawMessage.value,
    t,
  });
});

const resetMixedChannelDialogState = () => {
  showMixedChannelWarning.value = false;
  mixedChannelWarningDetails.value = null;
  mixedChannelWarningRawMessage.value = "";
  mixedChannelWarningAction.value = null;
};

const getAccountCredentials = () =>
  (props.account?.credentials as Record<string, unknown>) || {};

const getAccountExtra = () =>
  (props.account?.extra as Record<string, unknown>) || {};

const applyModelRestrictionState = (rawMapping: unknown) => {
  const nextState = deriveModelRestrictionStateFromMapping(rawMapping);
  modelRestrictionMode.value = nextState.mode;
  allowedModels.value = nextState.allowedModels;
  modelMappings.value = nextState.modelMappings;
};

const resetModelRestrictionState = () => {
  const nextState = createEmptyModelRestrictionState();
  modelRestrictionMode.value = nextState.mode;
  allowedModels.value = nextState.allowedModels;
  modelMappings.value = nextState.modelMappings;
};

const syncAntigravityModelRestrictionState = (
  credentials: Record<string, unknown> | undefined,
) => {
  antigravityModelRestrictionMode.value = "mapping";
  antigravityWhitelistModels.value = [];
  antigravityModelMappings.value = deriveAntigravityModelMappings(credentials);
};

const syncOpenAIExtraState = (
  accountType: AccountType,
  extra: Record<string, unknown> | undefined,
) => {
  const nextState = deriveOpenAIExtraState(accountType, extra);
  openaiPassthroughEnabled.value = nextState.openaiPassthroughEnabled;
  openaiOAuthResponsesWebSocketV2Mode.value =
    nextState.openaiOAuthResponsesWebSocketV2Mode;
  openaiAPIKeyResponsesWebSocketV2Mode.value =
    nextState.openaiAPIKeyResponsesWebSocketV2Mode;
  codexCLIOnlyEnabled.value = nextState.codexCLIOnlyEnabled;
};

const form = reactive<EditAccountForm>(createDefaultEditAccountForm());

const statusOptions = computed<Array<{ value: EditAccountForm["status"]; label: string }>>(() => {
  const options: Array<{ value: EditAccountForm["status"]; label: string }> = [
    { value: "active", label: t("common.active") },
    { value: "inactive", label: t("common.inactive") },
  ];
  if (form.status === "error") {
    options.push({ value: "error", label: t("admin.accounts.status.error") });
  }
  return options;
});

const expiresAtInput = computed({
  get: () => formatDateTimeLocal(form.expires_at),
  set: (value: string) => {
    form.expires_at = parseDateTimeLocal(value);
  },
});

// Watchers
const syncFormFromAccount = (newAccount: Account | null) => {
  if (!newAccount) {
    return;
  }
  antigravityMixedChannelConfirmed.value = false;
  resetMixedChannelDialogState();
  hydrateEditAccountForm(form, newAccount);

  // Load intercept warmup requests setting (applies to all account types)
  const credentials = newAccount.credentials as
    | Record<string, unknown>
    | undefined;
  const platformDefaultUrl = getDefaultBaseURL(newAccount.platform);
  interceptWarmupRequests.value =
    credentials?.intercept_warmup_requests === true;
  autoPauseOnExpired.value = newAccount.auto_pause_on_expired === true;
  editBaseUrl.value = platformDefaultUrl;
  editApiKey.value = "";
  editSessionToken.value = "";
  resetModelRestrictionState();
  poolModeEnabled.value = false;
  poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT;
  customErrorCodesEnabled.value = false;
  selectedErrorCodes.value = [];
  customErrorCodeInput.value = null;

  // Load mixed scheduling setting (only for antigravity accounts)
  mixedScheduling.value = false;
  allowOverages.value = false;
  const extra = newAccount.extra as Record<string, unknown> | undefined;
  mixedScheduling.value = extra?.mixed_scheduling === true;
  allowOverages.value = extra?.allow_overages === true;

  // Load OpenAI passthrough toggle (OpenAI OAuth/API Key)
  openaiPassthroughEnabled.value = false;
  openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
  openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
  codexCLIOnlyEnabled.value = false;
  anthropicPassthroughEnabled.value = false;
  if (
    newAccount.platform === "openai" &&
    (newAccount.type === "oauth" || newAccount.type === "apikey")
  ) {
    syncOpenAIExtraState(newAccount.type, extra);
  }
  if (newAccount.platform === "anthropic" && newAccount.type === "apikey") {
    anthropicPassthroughEnabled.value = extra?.anthropic_passthrough === true;
  }

  // Load quota limit for compatible key/upstream and bedrock accounts.
  if (
    newAccount.type === "apikey" ||
    newAccount.type === "upstream" ||
    newAccount.type === "bedrock"
  ) {
    const quotaVal = extra?.quota_limit as number | undefined;
    editQuotaLimit.value = quotaVal && quotaVal > 0 ? quotaVal : null;
    const dailyVal = extra?.quota_daily_limit as number | undefined;
    editQuotaDailyLimit.value = dailyVal && dailyVal > 0 ? dailyVal : null;
    const weeklyVal = extra?.quota_weekly_limit as number | undefined;
    editQuotaWeeklyLimit.value = weeklyVal && weeklyVal > 0 ? weeklyVal : null;
    // Load quota reset mode config
    editDailyResetMode.value =
      (extra?.quota_daily_reset_mode as "rolling" | "fixed") || null;
    editDailyResetHour.value =
      (extra?.quota_daily_reset_hour as number) ?? null;
    editWeeklyResetMode.value =
      (extra?.quota_weekly_reset_mode as "rolling" | "fixed") || null;
    editWeeklyResetDay.value =
      (extra?.quota_weekly_reset_day as number) ?? null;
    editWeeklyResetHour.value =
      (extra?.quota_weekly_reset_hour as number) ?? null;
    editResetTimezone.value = (extra?.quota_reset_timezone as string) || null;
    editQuotaNotifyDailyEnabled.value =
      (extra?.quota_notify_daily_enabled as boolean) || null;
    editQuotaNotifyDailyThreshold.value =
      (extra?.quota_notify_daily_threshold as number) ?? null;
    editQuotaNotifyDailyThresholdType.value =
      (extra?.quota_notify_daily_threshold_type as "fixed" | "percentage") ||
      null;
    editQuotaNotifyWeeklyEnabled.value =
      (extra?.quota_notify_weekly_enabled as boolean) || null;
    editQuotaNotifyWeeklyThreshold.value =
      (extra?.quota_notify_weekly_threshold as number) ?? null;
    editQuotaNotifyWeeklyThresholdType.value =
      (extra?.quota_notify_weekly_threshold_type as "fixed" | "percentage") ||
      null;
    editQuotaNotifyTotalEnabled.value =
      (extra?.quota_notify_total_enabled as boolean) || null;
    editQuotaNotifyTotalThreshold.value =
      (extra?.quota_notify_total_threshold as number) ?? null;
    editQuotaNotifyTotalThresholdType.value =
      (extra?.quota_notify_total_threshold_type as "fixed" | "percentage") ||
      null;
  } else {
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
  }

  // Load antigravity model mapping (Antigravity 只支持映射模式)
  if (newAccount.platform === "antigravity") {
    syncAntigravityModelRestrictionState(
      newAccount.credentials as Record<string, unknown> | undefined,
    );
  } else {
    antigravityModelRestrictionMode.value = "mapping";
    antigravityWhitelistModels.value = [];
    antigravityModelMappings.value = [];
  }

  hydrateQuotaControlsFromAccount(newAccount);

  const tempUnschedState = loadTempUnschedRuleState(credentials);
  tempUnschedEnabled.value = tempUnschedState.enabled;
  tempUnschedRules.value = tempUnschedState.rules;

  // Initialize compatible API key/upstream fields.
  if (
    (newAccount.type === "apikey" || newAccount.type === "upstream") &&
    newAccount.credentials
  ) {
    const credentials = newAccount.credentials as Record<string, unknown>;
    editBaseUrl.value = (credentials.base_url as string) || platformDefaultUrl;

    applyModelRestrictionState(credentials.model_mapping);

    // Load pool mode
    poolModeEnabled.value = credentials.pool_mode === true;
    poolModeRetryCount.value = normalizePoolModeRetryCount(
      Number(
        credentials.pool_mode_retry_count ?? DEFAULT_POOL_MODE_RETRY_COUNT,
      ),
    );

    // Load custom error codes
    customErrorCodesEnabled.value =
      credentials.custom_error_codes_enabled === true;
    const existingErrorCodes = credentials.custom_error_codes as
      | number[]
      | undefined;
    if (existingErrorCodes && Array.isArray(existingErrorCodes)) {
      selectedErrorCodes.value = [...existingErrorCodes];
    } else {
      selectedErrorCodes.value = [];
    }
  } else if (newAccount.type === "bedrock" && newAccount.credentials) {
    const bedrockCreds = newAccount.credentials as Record<string, unknown>;
    const authMode = (bedrockCreds.auth_mode as string) || "sigv4";
    editBedrockRegion.value = (bedrockCreds.aws_region as string) || "";
    editBedrockForceGlobal.value =
      (bedrockCreds.aws_force_global as string) === "true";

    if (authMode === "apikey") {
      editBedrockApiKeyValue.value = "";
    } else {
      editBedrockAccessKeyId.value =
        (bedrockCreds.aws_access_key_id as string) || "";
      editBedrockSecretAccessKey.value = "";
      editBedrockSessionToken.value = "";
    }

    // Load pool mode for bedrock
    poolModeEnabled.value = bedrockCreds.pool_mode === true;
    const retryCount = bedrockCreds.pool_mode_retry_count;
    poolModeRetryCount.value =
      typeof retryCount === "number" && retryCount >= 0
        ? retryCount
        : DEFAULT_POOL_MODE_RETRY_COUNT;

    // Load quota limits for bedrock
    const bedrockExtra = (newAccount.extra as Record<string, unknown>) || {};
    editQuotaLimit.value =
      typeof bedrockExtra.quota_limit === "number"
        ? bedrockExtra.quota_limit
        : null;
    editQuotaDailyLimit.value =
      typeof bedrockExtra.quota_daily_limit === "number"
        ? bedrockExtra.quota_daily_limit
        : null;
    editQuotaWeeklyLimit.value =
      typeof bedrockExtra.quota_weekly_limit === "number"
        ? bedrockExtra.quota_weekly_limit
        : null;
    editQuotaNotifyDailyEnabled.value =
      (bedrockExtra.quota_notify_daily_enabled as boolean) || null;
    editQuotaNotifyDailyThreshold.value =
      (bedrockExtra.quota_notify_daily_threshold as number) ?? null;
    editQuotaNotifyDailyThresholdType.value =
      (bedrockExtra.quota_notify_daily_threshold_type as
        | "fixed"
        | "percentage") || null;
    editQuotaNotifyWeeklyEnabled.value =
      (bedrockExtra.quota_notify_weekly_enabled as boolean) || null;
    editQuotaNotifyWeeklyThreshold.value =
      (bedrockExtra.quota_notify_weekly_threshold as number) ?? null;
    editQuotaNotifyWeeklyThresholdType.value =
      (bedrockExtra.quota_notify_weekly_threshold_type as
        | "fixed"
        | "percentage") || null;
    editQuotaNotifyTotalEnabled.value =
      (bedrockExtra.quota_notify_total_enabled as boolean) || null;
    editQuotaNotifyTotalThreshold.value =
      (bedrockExtra.quota_notify_total_threshold as number) ?? null;
    editQuotaNotifyTotalThresholdType.value =
      (bedrockExtra.quota_notify_total_threshold_type as
        | "fixed"
        | "percentage") || null;

    applyModelRestrictionState(bedrockCreds.model_mapping);
  } else if (newAccount.type === "session") {
    editSessionToken.value = "";
  } else {
    editBaseUrl.value = platformDefaultUrl;

    // Load model mappings for OpenAI OAuth accounts
    if (newAccount.platform === "openai" && newAccount.credentials) {
      const oauthCredentials = newAccount.credentials as Record<
        string,
        unknown
      >;
      applyModelRestrictionState(oauthCredentials.model_mapping);
    } else {
      resetModelRestrictionState();
    }
    poolModeEnabled.value = false;
    poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT;
    customErrorCodesEnabled.value = false;
    selectedErrorCodes.value = [];
  }
};

watch(
  [() => props.show, () => props.account],
  ([show, newAccount], [wasShow, previousAccount]) => {
    if (!show || !newAccount) {
      return;
    }
    if (!wasShow || newAccount !== previousAccount) {
      syncFormFromAccount(newAccount);
      loadTLSProfiles();
    }
  },
  { immediate: true },
);

async function loadTLSProfiles() {
  try {
    const profiles = await adminAPI.tlsFingerprintProfiles.list();
    setTlsFingerprintProfiles(
      profiles.map((p) => ({
        id: p.id,
        name: p.name,
      })),
    );
  } catch {
    setTlsFingerprintProfiles([]);
  }
}

// Model mapping helpers
const addModelMapping = () => {
  appendEmptyModelMapping(modelMappings.value);
};

const removeModelMapping = (index: number) => {
  removeModelMappingAt(modelMappings.value, index);
};

const updateModelMapping = (
  index: number,
  field: keyof ModelMapping,
  value: string,
) => {
  const mapping = modelMappings.value[index];
  if (!mapping) {
    return;
  }
  modelMappings.value[index] = {
    ...mapping,
    [field]: value,
  };
};

const addPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(modelMappings.value, from, to, (model) => {
    appStore.showInfo(t("admin.accounts.mappingExists", { model }));
  });
};

const addAntigravityModelMapping = () => {
  appendEmptyModelMapping(antigravityModelMappings.value);
};

const removeAntigravityModelMapping = (index: number) => {
  removeModelMappingAt(antigravityModelMappings.value, index);
};

const updateAntigravityModelMapping = (
  index: number,
  field: keyof ModelMapping,
  value: string,
) => {
  const mapping = antigravityModelMappings.value[index];
  if (!mapping) {
    return;
  }
  antigravityModelMappings.value[index] = {
    ...mapping,
    [field]: value,
  };
};

const addAntigravityPresetMapping = (from: string, to: string) => {
  appendPresetModelMapping(
    antigravityModelMappings.value,
    from,
    to,
    (model) => {
      appStore.showInfo(t("admin.accounts.mappingExists", { model }));
    },
  );
};

// Error code toggle helper
const toggleErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code);
  if (index === -1) {
    if (!confirmCustomErrorCodeSelection(code, confirm, t)) {
      return;
    }
    selectedErrorCodes.value.push(code);
  } else {
    selectedErrorCodes.value.splice(index, 1);
  }
};

// Add custom error code from input
const addCustomErrorCode = () => {
  const code = customErrorCodeInput.value;
  if (code === null || code < 100 || code > 599) {
    appStore.showError(t("admin.accounts.invalidErrorCode"));
    return;
  }
  if (selectedErrorCodes.value.includes(code)) {
    appStore.showInfo(t("admin.accounts.errorCodeExists"));
    return;
  }
  if (!confirmCustomErrorCodeSelection(code, confirm, t)) {
    return;
  }
  selectedErrorCodes.value.push(code);
  customErrorCodeInput.value = null;
};

// Remove error code
const removeErrorCode = (code: number) => {
  const index = selectedErrorCodes.value.indexOf(code);
  if (index !== -1) {
    selectedErrorCodes.value.splice(index, 1);
  }
};

const addTempUnschedRule = (preset?: TempUnschedRuleForm) => {
  tempUnschedRules.value.push(createTempUnschedRule(preset));
};

const removeTempUnschedRule = (index: number) => {
  tempUnschedRules.value.splice(index, 1);
};

const moveTempUnschedRule = (index: number, direction: number) => {
  moveItemInPlace(tempUnschedRules.value, index, direction);
};

const updateTempUnschedRule = (
  index: number,
  field: keyof TempUnschedRuleForm,
  value: TempUnschedRuleForm[keyof TempUnschedRuleForm],
) => {
  const rule = tempUnschedRules.value[index];
  if (!rule) {
    return;
  }

  const nextRule = { ...rule };
  if (field === "error_code" || field === "duration_minutes") {
    nextRule[field] = typeof value === "number" ? value : null;
  } else {
    nextRule[field] = typeof value === "string" ? value : "";
  }
  tempUnschedRules.value[index] = nextRule;
};

const clearMixedChannelDialog = () => {
  resetMixedChannelDialogState();
};

const openMixedChannelDialog = (opts: {
  response?: CheckMixedChannelResponse;
  message?: string;
  onConfirm: () => Promise<void>;
}) => {
  mixedChannelWarningDetails.value = buildMixedChannelDetails(opts.response);
  mixedChannelWarningRawMessage.value =
    opts.message ||
    opts.response?.message ||
    t("admin.accounts.failedToUpdate");
  mixedChannelWarningAction.value = opts.onConfirm;
  showMixedChannelWarning.value = true;
};

const withAntigravityConfirmFlag = (payload: UpdateAccountRequest) => {
  if (
    props.account?.platform &&
    needsMixedChannelCheck(props.account.platform) &&
    antigravityMixedChannelConfirmed.value
  ) {
    return {
      ...payload,
      confirm_mixed_channel_risk: true,
    };
  }
  const cloned = { ...payload };
  delete cloned.confirm_mixed_channel_risk;
  return cloned;
};

const ensureAntigravityMixedChannelConfirmed = async (
  onConfirm: () => Promise<void>,
): Promise<boolean> => {
  if (
    !props.account?.platform ||
    !needsMixedChannelCheck(props.account.platform)
  ) {
    return true;
  }
  if (antigravityMixedChannelConfirmed.value) {
    return true;
  }
  if (!props.account) {
    return false;
  }

  try {
    const result = await adminAPI.accounts.checkMixedChannelRisk({
      platform: props.account.platform,
      group_ids: form.group_ids,
      account_id: props.account.id,
    });
    if (!result.has_risk) {
      return true;
    }
    openMixedChannelDialog({
      response: result,
      onConfirm: async () => {
        antigravityMixedChannelConfirmed.value = true;
        await onConfirm();
      },
    });
    return false;
  } catch (error: any) {
    appStore.showError(
      resolveRequestErrorMessage(error, t("admin.accounts.failedToUpdate")),
    );
    return false;
  }
};

const formatDateTimeLocal = formatDateTimeLocalInput;
const parseDateTimeLocal = parseDateTimeLocalInput;

// Methods
const handleClose = () => {
  antigravityMixedChannelConfirmed.value = false;
  clearMixedChannelDialog();
  emit("close");
};

const submitUpdateAccount = async (
  accountID: number,
  updatePayload: UpdateAccountRequest,
) => {
  submitting.value = true;
  try {
    const updatedAccount = await adminAPI.accounts.update(
      accountID,
      withAntigravityConfirmFlag(updatePayload),
    );
    appStore.showSuccess(t("admin.accounts.accountUpdated"));
    emit("updated", updatedAccount);
    handleClose();
  } catch (error: any) {
    if (
      error.status === 409 &&
      error.error === "mixed_channel_warning" &&
      props.account?.platform &&
      needsMixedChannelCheck(props.account.platform)
    ) {
      openMixedChannelDialog({
        message: error.message,
        onConfirm: async () => {
          antigravityMixedChannelConfirmed.value = true;
          await submitUpdateAccount(accountID, updatePayload);
        },
      });
      return;
    }
    appStore.showError(
      resolveRequestErrorMessage(error, t("admin.accounts.failedToUpdate")),
    );
  } finally {
    submitting.value = false;
  }
};

const handleSubmit = async () => {
  if (!props.account) return;
  const accountID = props.account.id;

  if (
    form.status !== "active" &&
    form.status !== "inactive" &&
    form.status !== "error"
  ) {
    appStore.showError(t("admin.accounts.pleaseSelectStatus"));
    return;
  }

  const basePayload = buildEditAccountBasePayload(
    form,
    autoPauseOnExpired.value,
  );
  try {
    const payloadResult = buildEditAccountMutationPayload({
      account: props.account,
      anthropicAPIKeyExtra: {
        anthropicPassthroughEnabled: anthropicPassthroughEnabled.value,
      },
      anthropicQuotaExtra: {
        baseRpm: baseRpm.value,
        cacheTTLOverrideEnabled: cacheTTLOverrideEnabled.value,
        cacheTTLOverrideTarget: cacheTTLOverrideTarget.value,
        customBaseUrl: customBaseUrl.value,
        customBaseUrlEnabled: customBaseUrlEnabled.value,
        maxSessions: maxSessions.value,
        rpmLimitEnabled: rpmLimitEnabled.value,
        rpmStickyBuffer: rpmStickyBuffer.value,
        rpmStrategy: rpmStrategy.value,
        sessionIdMaskingEnabled: sessionIdMaskingEnabled.value,
        sessionIdleTimeout: sessionIdleTimeout.value,
        sessionLimitEnabled: sessionLimitEnabled.value,
        tlsFingerprintEnabled: tlsFingerprintEnabled.value,
        tlsFingerprintProfileId: tlsFingerprintProfileId.value,
        userMsgQueueMode: userMsgQueueMode.value,
        windowCostEnabled: windowCostEnabled.value,
        windowCostLimit: windowCostLimit.value,
        windowCostStickyReserve: windowCostStickyReserve.value,
      },
      antigravity: {
        allowOverages: allowOverages.value,
        mixedScheduling: mixedScheduling.value,
        modelMappings: antigravityModelMappings.value,
      },
      basePayload,
      bedrock: {
        accessKeyId: editBedrockAccessKeyId.value,
        allowedModels: allowedModels.value,
        apiKeyInput: editBedrockApiKeyValue.value,
        forceGlobal: editBedrockForceGlobal.value,
        isApiKeyMode: isBedrockAPIKeyMode.value,
        mode: modelRestrictionMode.value,
        modelMappings: modelMappings.value,
        poolModeEnabled: poolModeEnabled.value,
        poolModeRetryCount: poolModeRetryCount.value,
        region: editBedrockRegion.value,
        secretAccessKey: editBedrockSecretAccessKey.value,
        sessionToken: editBedrockSessionToken.value,
      },
      compatible: {
        allowedModels: allowedModels.value,
        apiKeyInput: editApiKey.value,
        baseUrlInput: editBaseUrl.value,
        customErrorCodesEnabled: customErrorCodesEnabled.value,
        defaultBaseUrl: defaultBaseUrl.value,
        isOpenAIModelRestrictionDisabled:
          isOpenAIModelRestrictionDisabled.value,
        mode: modelRestrictionMode.value,
        modelMappings: modelMappings.value,
        poolModeEnabled: poolModeEnabled.value,
        poolModeRetryCount: poolModeRetryCount.value,
        selectedErrorCodes: selectedErrorCodes.value,
      },
      currentCredentials: getAccountCredentials(),
      currentExtra: getAccountExtra(),
      openAIExtra: {
        accountType: props.account.type,
        codexCLIOnlyEnabled: codexCLIOnlyEnabled.value,
        openaiAPIKeyResponsesWebSocketV2Mode:
          openaiAPIKeyResponsesWebSocketV2Mode.value,
        openaiOAuthResponsesWebSocketV2Mode:
          openaiOAuthResponsesWebSocketV2Mode.value,
        openaiPassthroughEnabled: openaiPassthroughEnabled.value,
      },
      quota: {
        dailyResetHour: editDailyResetHour.value,
        dailyResetMode: editDailyResetMode.value,
        quotaDailyLimit: editQuotaDailyLimit.value,
        quotaLimit: editQuotaLimit.value,
        quotaWeeklyLimit: editQuotaWeeklyLimit.value,
        quotaNotifyDailyEnabled: editQuotaNotifyDailyEnabled.value,
        quotaNotifyDailyThreshold: editQuotaNotifyDailyThreshold.value,
        quotaNotifyDailyThresholdType: editQuotaNotifyDailyThresholdType.value,
        quotaNotifyWeeklyEnabled: editQuotaNotifyWeeklyEnabled.value,
        quotaNotifyWeeklyThreshold: editQuotaNotifyWeeklyThreshold.value,
        quotaNotifyWeeklyThresholdType:
          editQuotaNotifyWeeklyThresholdType.value,
        quotaNotifyTotalEnabled: editQuotaNotifyTotalEnabled.value,
        quotaNotifyTotalThreshold: editQuotaNotifyTotalThreshold.value,
        quotaNotifyTotalThresholdType: editQuotaNotifyTotalThresholdType.value,
        resetTimezone: editResetTimezone.value,
        weeklyResetDay: editWeeklyResetDay.value,
        weeklyResetHour: editWeeklyResetHour.value,
        weeklyResetMode: editWeeklyResetMode.value,
      },
      sessionTokenInput: editSessionToken.value,
      sharedCredentials: {
        interceptWarmupRequests: interceptWarmupRequests.value,
        tempUnschedEnabled: tempUnschedEnabled.value,
        tempUnschedRules: tempUnschedRules.value,
      },
    });
    if (payloadResult.error) {
      appStore.showError(
        t(resolveAccountMutationPayloadErrorKey(payloadResult.error)),
      );
      return;
    }
    if (!payloadResult.payload) {
      appStore.showError(t("admin.accounts.failedToUpdate"));
      return;
    }
    const updatePayload = payloadResult.payload;

    const canContinue = await ensureAntigravityMixedChannelConfirmed(
      async () => {
        await submitUpdateAccount(accountID, updatePayload);
      },
    );
    if (!canContinue) {
      return;
    }

    await submitUpdateAccount(accountID, updatePayload);
  } catch (error: any) {
    appStore.showError(
      resolveRequestErrorMessage(error, t("admin.accounts.failedToUpdate")),
    );
  }
};

// Handle mixed channel warning confirmation
const handleMixedChannelConfirm = async () => {
  const action = mixedChannelWarningAction.value;
  if (!action) {
    clearMixedChannelDialog();
    return;
  }
  clearMixedChannelDialog();
  submitting.value = true;
  try {
    await action();
  } finally {
    submitting.value = false;
  }
};

const handleMixedChannelCancel = () => {
  clearMixedChannelDialog();
};
</script>
