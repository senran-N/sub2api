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
import { ref, computed, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores/app";
import { useAuthStore } from "@/stores/auth";
import { adminAPI } from "@/api/admin";
import type {
  Account,
  Proxy,
  AdminGroup,
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
import { useEditBedrockCredentials } from "@/components/account/useEditBedrockCredentials";
import { useEditAccountModelRestrictions } from "@/components/account/useEditAccountModelRestrictions";
import { useEditAccountQuotaLimits } from "@/components/account/useEditAccountQuotaLimits";
import { useEditAccountQuotaControls } from "@/components/account/useEditAccountQuotaControls";
import { useEditAccountRuntimeOptions } from "@/components/account/useEditAccountRuntimeOptions";
import { useEditAccountTempUnschedRules } from "@/components/account/useEditAccountTempUnschedRules";
import { useEditAccountFormState } from "@/components/account/useEditAccountFormState";
import { useEditCustomErrorCodes } from "@/components/account/useEditCustomErrorCodes";
import { useEditCredentialFields } from "@/components/account/useEditCredentialFields";
import { useAccountMutationSections } from "@/components/account/useAccountMutationSections";
import { useEditMixedChannelWarning } from "@/components/account/useEditMixedChannelWarning";
import { useEditAccountMutationPayload } from "@/components/account/useEditAccountMutationPayload";
import {
  buildCompatibleBaseUrlPresets,
  buildAccountOpenAIWSModeOptions,
  buildAccountTempUnschedPresets,
  buildAccountUmqModeOptions,
  buildEditAccountBasePayload,
  resolveAccountApiKeyPlaceholder,
  resolveAccountBaseUrlHint,
  resolveAccountBaseUrlPlaceholder,
} from "@/components/account/accountModalShared";
import { resolveAccountMutationPayloadErrorKey } from "@/components/account/accountMutationPayload";
import { getDefaultBaseURL } from "@/components/account/credentialsBuilder";
import { confirmCustomErrorCodeSelection } from "@/components/account/accountModalInteractions";
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
const credentialFields = useEditCredentialFields();
const {
  editApiKey,
  editBaseUrl,
  editSessionToken,
  hydrateBedrockPoolMode,
  hydrateCompatibleCredentialFields,
  poolModeEnabled,
  poolModeRetryCount,
  resetCredentialFields,
} = credentialFields;
const bedrockCredentials = useEditBedrockCredentials(() => props.account);
const {
  editBedrockAccessKeyId,
  editBedrockApiKeyValue,
  editBedrockAuthMode,
  editBedrockForceGlobal,
  editBedrockRegion,
  editBedrockSecretAccessKey,
  editBedrockSessionToken,
  hydrateBedrockCredentialsFromAccount,
} = bedrockCredentials;
const modelRestrictions = useEditAccountModelRestrictions({
  onMappingExists: (model) => {
    appStore.showInfo(t("admin.accounts.mappingExists", { model }));
  },
});
const {
  addAntigravityModelMapping,
  addAntigravityPresetMapping,
  addModelMapping,
  addPresetMapping,
  allowedModels,
  antigravityModelMappings,
  applyModelRestrictionState,
  getAntigravityModelMappingKey,
  getModelMappingKey,
  modelMappings,
  modelRestrictionMode,
  removeAntigravityModelMapping,
  removeModelMapping,
  resetAntigravityModelRestrictionState,
  resetModelRestrictionState,
  syncAntigravityModelRestrictionState,
  updateAntigravityModelMapping,
  updateModelMapping,
} = modelRestrictions;
const customErrorCodes = useEditCustomErrorCodes({
  confirmSelection: (code) => confirmCustomErrorCodeSelection(code, confirm, t),
  showDuplicate: () => appStore.showInfo(t("admin.accounts.errorCodeExists")),
  showInvalid: () => appStore.showError(t("admin.accounts.invalidErrorCode")),
});
const {
  addCustomErrorCode,
  customErrorCodeInput,
  customErrorCodesEnabled,
  hydrateCustomErrorCodesFromCredentials,
  removeErrorCode,
  resetCustomErrorCodes,
  selectedErrorCodes,
  toggleErrorCode,
} = customErrorCodes;
const interceptWarmupRequests = ref(false);
const formState = useEditAccountFormState(t);
const {
  allowOverages,
  autoPauseOnExpired,
  expiresAtInput,
  form,
  hydrateFormStateFromAccount,
  mixedScheduling,
  statusOptions,
} = formState;
const tempUnschedRulesState = useEditAccountTempUnschedRules();
const {
  addTempUnschedRule,
  getTempUnschedRuleKey,
  hydrateTempUnschedRulesFromCredentials,
  moveTempUnschedRule,
  removeTempUnschedRule,
  tempUnschedEnabled,
  tempUnschedRules,
  updateTempUnschedRule,
} = tempUnschedRulesState;

const quotaControls = useEditAccountQuotaControls();
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
} = quotaControls;
const umqModeOptions = computed(() => buildAccountUmqModeOptions(t));

const runtimeOptions = useEditAccountRuntimeOptions(() => props.account);
const {
  anthropicPassthroughEnabled,
  codexCLIOnlyEnabled,
  hydrateRuntimeOptionsFromAccount,
  isOpenAIModelRestrictionDisabled,
  openAIWSModeConcurrencyHintKey,
  openaiPassthroughEnabled,
  openaiResponsesWebSocketV2Mode,
} = runtimeOptions;
const quotaLimits = useEditAccountQuotaLimits();
const {
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
} = quotaLimits;
const openAIWSModeOptions = computed(() => buildAccountOpenAIWSModeOptions(t));
const {
  openAIAccountCategory,
  showAnthropicAPIKeyRuntimeSection,
  showAnthropicQuotaControls,
  showCompatibleCredentialsForm,
  showOpenAIRuntimeSection,
  showQuotaLimitSection,
  showWarmupSection,
} = useAccountMutationSections(() => props.account);

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

const {
  ensureMixedChannelConfirmed,
  mixedChannelWarningMessageText,
  openMixedChannelConflictDialog,
  resetMixedChannelDialog,
  resetMixedChannelState,
  showMixedChannelWarning,
  takeMixedChannelWarningAction,
  withMixedChannelConfirmFlag,
} = useEditMixedChannelWarning({
  getAccount: () => props.account,
  getGroupIds: () => form.group_ids,
  showError: (message) => appStore.showError(message),
  t,
});

const { buildEditMutationPayload } = useEditAccountMutationPayload({
  bedrockCredentials,
  credentialFields,
  customErrorCodes,
  defaultBaseUrl,
  formState,
  interceptWarmupRequests,
  modelRestrictions,
  quotaControls,
  quotaLimits,
  runtimeOptions,
  tempUnschedRules: tempUnschedRulesState,
});

// Watchers
const syncFormFromAccount = (newAccount: Account | null) => {
  if (!newAccount) {
    return;
  }
  resetMixedChannelState();
  hydrateFormStateFromAccount(newAccount);

  // Load intercept warmup requests setting (applies to all account types)
  const credentials = newAccount.credentials as
    | Record<string, unknown>
    | undefined;
  const platformDefaultUrl = getDefaultBaseURL(newAccount.platform);
  interceptWarmupRequests.value =
    credentials?.intercept_warmup_requests === true;
  resetCredentialFields(platformDefaultUrl);
  resetModelRestrictionState();
  resetCustomErrorCodes();

  hydrateRuntimeOptionsFromAccount(newAccount);
  hydrateBedrockCredentialsFromAccount(newAccount);
  hydrateQuotaLimitsFromAccount(newAccount);

  // Load antigravity model mapping (Antigravity 只支持映射模式)
  if (newAccount.platform === "antigravity") {
    syncAntigravityModelRestrictionState(
      newAccount.credentials as Record<string, unknown> | undefined,
    );
  } else {
    resetAntigravityModelRestrictionState();
  }

  hydrateQuotaControlsFromAccount(newAccount);

  hydrateTempUnschedRulesFromCredentials(credentials);

  // Initialize compatible API key/upstream fields.
  if (
    (newAccount.type === "apikey" || newAccount.type === "upstream") &&
    newAccount.credentials
  ) {
    const credentials = newAccount.credentials as Record<string, unknown>;
    hydrateCompatibleCredentialFields(credentials, platformDefaultUrl);
    applyModelRestrictionState(credentials.model_mapping);
    hydrateCustomErrorCodesFromCredentials(credentials);
  } else if (newAccount.type === "bedrock" && newAccount.credentials) {
    const bedrockCreds = newAccount.credentials as Record<string, unknown>;
    hydrateBedrockPoolMode(bedrockCreds);
    applyModelRestrictionState(bedrockCreds.model_mapping);
  } else {
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
    resetCustomErrorCodes();
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

// Methods
const handleClose = () => {
  resetMixedChannelState();
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
      withMixedChannelConfirmFlag(updatePayload),
    );
    appStore.showSuccess(t("admin.accounts.accountUpdated"));
    emit("updated", updatedAccount);
    handleClose();
  } catch (error: any) {
    if (
      openMixedChannelConflictDialog(error, async () => {
        await submitUpdateAccount(accountID, updatePayload);
      })
    ) {
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
    const payloadResult = buildEditMutationPayload({
      account: props.account,
      basePayload,
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

    const canContinue = await ensureMixedChannelConfirmed(async () => {
      await submitUpdateAccount(accountID, updatePayload);
    });
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
  const action = takeMixedChannelWarningAction();
  if (!action) {
    return;
  }
  submitting.value = true;
  try {
    await action();
  } finally {
    submitting.value = false;
  }
};

const handleMixedChannelCancel = () => {
  resetMixedChannelDialog();
};
</script>
