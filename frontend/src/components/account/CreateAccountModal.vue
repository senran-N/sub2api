<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.createAccount')"
    width="wide"
    @close="handleClose"
  >
    <!-- Step Indicator for OAuth accounts -->
    <div v-if="isOAuthFlow" class="mb-6 flex items-center justify-center">
      <div class="create-account-modal__stepper flex items-center space-x-4">
        <div class="create-account-modal__step-group flex items-center">
          <div
            :class="[
              'create-account-modal__step-node flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold',
              step >= 1
                ? 'create-account-modal__step-node--active'
                : 'create-account-modal__step-node--idle',
            ]"
          >
            1
          </div>
          <span
            class="create-account-modal__step-label ml-2 text-sm font-medium"
            >{{ t("admin.accounts.oauth.authMethod") }}</span
          >
        </div>
        <div class="create-account-modal__step-connector h-0.5 w-8" />
        <div class="create-account-modal__step-group flex items-center">
          <div
            :class="[
              'create-account-modal__step-node flex h-8 w-8 items-center justify-center rounded-full text-sm font-semibold',
              step >= 2
                ? 'create-account-modal__step-node--active'
                : 'create-account-modal__step-node--idle',
            ]"
          >
            2
          </div>
          <span
            class="create-account-modal__step-label ml-2 text-sm font-medium"
            >{{ oauthStepTitle }}</span
          >
        </div>
      </div>
    </div>

    <!-- Step 1: Basic Info -->
    <form
      v-if="step === 1"
      id="create-account-form"
      @submit.prevent="handleSubmit"
      class="space-y-5"
    >
      <CreateAccountBasicInfoSection
        v-model:name="form.name"
        v-model:notes="form.notes"
        :name-label="accountNameLabel"
        :name-placeholder="accountNamePlaceholder"
        :name-hint="accountNameHint"
        :name-required="isAccountNameRequired"
      />

      <CreateAccountPlatformSelector v-model="form.platform" />

      <AnthropicAccountTypeSection
        v-if="form.platform === 'anthropic'"
        v-model:account-category="accountCategory"
      />

      <OpenAIAccountTypeSection
        v-if="form.platform === 'openai'"
        v-model:account-category="accountCategory"
      />

      <GrokAccountTypeSection
        v-if="form.platform === 'grok'"
        v-model:account-category="accountCategory"
      />

      <!-- Account Type Selection (Gemini) -->
      <div v-if="form.platform === 'gemini'">
        <GeminiAccountTypeSection
          v-model:account-category="accountCategory"
          :api-key-help-link="geminiHelpLinks.apiKey"
          @open-help="showGeminiHelpDialog = true"
        />

        <GeminiOAuthOptionsSection
          v-if="accountCategory === 'oauth-based'"
          v-model:oauth-type="geminiOAuthType"
          v-model:tier-google-one="geminiTierGoogleOne"
          v-model:tier-gcp="geminiTierGcp"
          v-model:tier-ai-studio="geminiTierAIStudio"
          :ai-studio-oauth-enabled="geminiAIStudioOAuthEnabled"
          :gcp-project-help-link="geminiHelpLinks.gcpProject"
        />
      </div>

      <AntigravityAccountTypeSection
        v-if="form.platform === 'antigravity'"
        v-model:account-type="antigravityAccountType"
      />

      <AntigravityUpstreamCredentialsSection
        v-if="
          form.platform === 'antigravity' &&
          antigravityAccountType === 'upstream'
        "
        v-model:base-url="upstreamBaseUrl"
        v-model:api-key="upstreamApiKey"
      />

      <AntigravityModelMappingSection
        v-if="form.platform === 'antigravity'"
        :mappings="antigravityModelMappings"
        :preset-mappings="antigravityPresetMappings"
        :mapping-key="getAntigravityModelMappingKey"
        @add="addAntigravityModelMapping"
        @remove="removeAntigravityModelMapping"
        @add-preset="addAntigravityPresetMapping"
        @update-mapping="updateAntigravityModelMapping"
      />

      <AnthropicAddMethodSection
        v-if="form.platform === 'anthropic' && isOAuthFlow"
        v-model="addMethod"
      />

      <CompatibleCredentialsSection
        v-if="showCompatibleCredentialsForm"
        v-model:base-url="apiKeyBaseUrl"
        v-model:api-key-value="apiKeyValue"
        v-model:tier-ai-studio="geminiTierAIStudio"
        v-model:model-restriction-mode="modelRestrictionMode"
        v-model:allowed-models="allowedModels"
        v-model:pool-mode-enabled="poolModeEnabled"
        v-model:pool-mode-retry-count="poolModeRetryCount"
        v-model:custom-error-codes-enabled="customErrorCodesEnabled"
        v-model:custom-error-code-input="customErrorCodeInput"
        :platform="form.platform"
        :base-url-presets="compatibleBaseUrlPresets"
        :base-url-placeholder="baseUrlPlaceholder"
        :base-url-hint="baseUrlHint"
        :api-key-label="t('admin.accounts.apiKeyRequired')"
        :api-key-required="true"
        :api-key-placeholder="apiKeyPlaceholder"
        :api-key-hint="apiKeyHint"
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

      <!-- Bedrock credentials (only for Anthropic Bedrock type) -->
      <div
        v-if="form.platform === 'anthropic' && accountCategory === 'bedrock'"
        class="space-y-4"
      >
        <BedrockCredentialsSection
          v-model:auth-mode="bedrockAuthMode"
          v-model:access-key-id="bedrockAccessKeyId"
          v-model:secret-access-key="bedrockSecretAccessKey"
          v-model:session-token="bedrockSessionToken"
          v-model:api-key-value="bedrockApiKeyValue"
          v-model:region="bedrockRegion"
          v-model:force-global="bedrockForceGlobal"
        />

        <ModelRestrictionSection
          v-model:mode="modelRestrictionMode"
          v-model:allowed-models="allowedModels"
          platform="anthropic"
          :mappings="modelMappings"
          :preset-mappings="bedrockPresets"
          :mapping-key="getModelMappingKey"
          from-placeholder-key="admin.accounts.fromModel"
          to-placeholder-key="admin.accounts.toModel"
          :show-mapping-notice="false"
          @add-mapping="addModelMapping"
          @remove-mapping="removeModelMapping"
          @add-preset="addPresetMapping"
          @update-mapping="updateModelMapping"
        />

        <PoolModeSection
          v-model:enabled="poolModeEnabled"
          v-model:retry-count="poolModeRetryCount"
        />
      </div>

      <GrokSessionCredentialsSection
        v-if="form.platform === 'grok' && form.type === 'session'"
        v-model:mode="grokSessionInputMode"
        v-model:single-token="grokSessionToken"
        v-model:batch-input="grokSessionBatchInput"
        v-model:dry-run="grokSessionBatchDryRun"
        v-model:test-after-create="grokSessionBatchTestAfterCreate"
        :result="grokSessionBatchResult"
        :submitting="submitting"
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

      <ModelRestrictionSection
        v-if="form.platform === 'openai' && accountCategory === 'oauth-based'"
        v-model:mode="modelRestrictionMode"
        v-model:allowed-models="allowedModels"
        :platform="form.platform"
        :mappings="modelMappings"
        :preset-mappings="presetMappings"
        :mapping-key="getModelMappingKey"
        :disabled="isOpenAIModelRestrictionDisabled"
        @add-mapping="addModelMapping"
        @remove-mapping="removeModelMapping"
        @add-preset="addPresetMapping"
        @update-mapping="updateModelMapping"
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

      <!-- Quota Control Section (Anthropic OAuth/SetupToken only) -->
      <div
        v-if="
          form.platform === 'anthropic' && accountCategory === 'oauth-based'
        "
        class="form-section space-y-4"
      >
        <div class="mb-3">
          <h3 class="input-label mb-0 text-base font-semibold">
            {{ t("admin.accounts.quotaControl.title") }}
          </h3>
          <p class="create-account-modal__choice-description mt-1 text-xs">
            {{ t("admin.accounts.quotaControl.hint") }}
          </p>
        </div>

        <WindowCostControlSection
          v-model:enabled="windowCostEnabled"
          v-model:limit="windowCostLimit"
          v-model:sticky-reserve="windowCostStickyReserve"
        />

        <SessionLimitControlSection
          v-model:enabled="sessionLimitEnabled"
          v-model:max-sessions="maxSessions"
          v-model:idle-timeout="sessionIdleTimeout"
        />

        <RpmLimitControlSection
          v-model:enabled="rpmLimitEnabled"
          v-model:base-rpm="baseRpm"
          v-model:strategy="rpmStrategy"
          v-model:sticky-buffer="rpmStickyBuffer"
          v-model:user-msg-queue-mode="userMsgQueueMode"
          :user-msg-queue-mode-options="umqModeOptions"
        />

        <TlsFingerprintControlSection
          v-model:enabled="tlsFingerprintEnabled"
          v-model:profile-id="tlsFingerprintProfileId"
          :profiles="tlsFingerprintProfiles"
        />

        <SessionIdMaskingControlSection
          v-model:enabled="sessionIdMaskingEnabled"
        />

        <CacheTtlOverrideSection
          v-model:enabled="cacheTTLOverrideEnabled"
          v-model:target="cacheTTLOverrideTarget"
        />

        <CustomBaseUrlControlSection
          v-model:enabled="customBaseUrlEnabled"
          v-model:base-url="customBaseUrl"
        />
      </div>

      <CreateAccountSchedulingSection
        v-model:proxy-id="form.proxy_id"
        v-model:concurrency="form.concurrency"
        v-model:load-factor="form.load_factor"
        v-model:priority="form.priority"
        v-model:rate-multiplier="form.rate_multiplier"
        v-model:expires-at="expiresAtInput"
        v-model:mixed-scheduling="mixedScheduling"
        v-model:allow-overages="allowOverages"
        v-model:group-ids="form.group_ids"
        :proxies="proxies"
        :platform="form.platform"
        :groups="groups"
        :simple-mode="authStore.isSimpleMode"
      />

      <OpenAIOptionsSection
        v-if="form.platform === 'openai'"
        v-model:passthrough-enabled="openaiPassthroughEnabled"
        v-model:ws-mode="openaiResponsesWebSocketV2Mode"
        v-model:codex-cli-only-enabled="codexCLIOnlyEnabled"
        :account-category="accountCategory"
        :ws-mode-options="openAIWSModeOptions"
        :ws-mode-concurrency-hint-key="openAIWSModeConcurrencyHintKey"
      />

      <AnthropicOptionsSection
        v-if="form.platform === 'anthropic'"
        v-model:api-key-passthrough-enabled="anthropicPassthroughEnabled"
        :account-category="accountCategory"
      />

      <AutoPauseOnExpiredSection v-model:enabled="autoPauseOnExpired" />

    </form>

    <!-- Step 2: OAuth Authorization -->
    <div v-else class="space-y-5">
      <OAuthAuthorizationFlow
        ref="oauthFlowRef"
        :add-method="form.platform === 'anthropic' ? addMethod : 'oauth'"
        :auth-url="currentOAuthState.authUrl"
        :session-id="currentOAuthState.sessionId"
        :loading="currentOAuthState.loading"
        :error="currentOAuthState.error"
        :show-help="form.platform === 'anthropic'"
        :show-proxy-warning="form.platform !== 'openai' && !!form.proxy_id"
        :allow-multiple="form.platform === 'anthropic'"
        :show-cookie-option="form.platform === 'anthropic'"
        :show-refresh-token-option="
          form.platform === 'openai' || form.platform === 'antigravity'
        "
        :show-mobile-refresh-token-option="form.platform === 'openai'"
        :platform="form.platform"
        :show-project-id="geminiOAuthType === 'code_assist'"
        @generate-url="handleGenerateUrl"
        @cookie-auth="handleCookieAuth"
        @validate-refresh-token="handleValidateRefreshToken"
        @validate-mobile-refresh-token="handleOpenAIValidateMobileRT"
      />
    </div>

    <template #footer>
      <div v-if="step === 1" class="flex justify-end gap-3">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t("common.cancel") }}
        </button>
        <button
          type="submit"
          form="create-account-form"
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
          {{ primarySubmitButtonLabel }}
        </button>
      </div>
      <div v-else class="flex justify-between gap-3">
        <button
          type="button"
          class="btn btn-secondary"
          @click="goBackToBasicInfo"
        >
          {{ t("common.back") }}
        </button>
        <button
          v-if="isManualInputMethod"
          type="button"
          :disabled="!canExchangeCode"
          class="btn btn-primary"
          @click="handleExchangeCode"
        >
          <svg
            v-if="currentOAuthState.loading"
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
          {{
            currentOAuthState.loading
              ? t("admin.accounts.oauth.verifying")
              : t("admin.accounts.oauth.completeAuth")
          }}
        </button>
      </div>
    </template>
  </BaseDialog>

  <GeminiHelpDialog
    :show="showGeminiHelpDialog"
    :help-links="geminiHelpLinks"
    :quota-docs="geminiQuotaDocs"
    @close="showGeminiHelpDialog = false"
  />

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
import {
  claudeModels,
  ensureModelCatalogLoaded,
  getPresetMappingsByPlatform,
  getModelsByPlatform,
  fetchAntigravityDefaultMappings,
} from "@/composables/useModelWhitelist";
import { useAuthStore } from "@/stores/auth";
import { adminAPI } from "@/api/admin";
import {
  useAccountOAuth,
  type AddMethod,
  type AuthInputMethod,
} from "@/composables/useAccountOAuth";
import { useOpenAIOAuth } from "@/composables/useOpenAIOAuth";
import { useGeminiOAuth } from "@/composables/useGeminiOAuth";
import { useAntigravityOAuth } from "@/composables/useAntigravityOAuth";
import type {
  Proxy,
  AdminGroup,
  AccountPlatform,
  AccountType,
  CheckMixedChannelResponse,
  CreateAccountRequest,
} from "@/types";
import type { GrokSessionBatchImportResult } from "@/api/admin/accounts";
import BaseDialog from "@/components/common/BaseDialog.vue";
import ConfirmDialog from "@/components/common/ConfirmDialog.vue";
import BedrockCredentialsSection from "@/components/account/BedrockCredentialsSection.vue";
import CompatibleCredentialsSection from "@/components/account/CompatibleCredentialsSection.vue";
import ModelRestrictionSection from "@/components/account/ModelRestrictionSection.vue";
import QuotaLimitSection from "@/components/account/QuotaLimitSection.vue";
import AnthropicAccountTypeSection from "@/components/account/AnthropicAccountTypeSection.vue";
import AnthropicAddMethodSection from "@/components/account/AnthropicAddMethodSection.vue";
import AnthropicOptionsSection from "@/components/account/AnthropicOptionsSection.vue";
import AntigravityAccountTypeSection from "@/components/account/AntigravityAccountTypeSection.vue";
import AntigravityModelMappingSection from "@/components/account/AntigravityModelMappingSection.vue";
import AntigravityUpstreamCredentialsSection from "@/components/account/AntigravityUpstreamCredentialsSection.vue";
import AutoPauseOnExpiredSection from "@/components/account/AutoPauseOnExpiredSection.vue";
import CacheTtlOverrideSection from "@/components/account/CacheTtlOverrideSection.vue";
import WarmupSection from "@/components/account/WarmupSection.vue";
import CreateAccountBasicInfoSection from "@/components/account/CreateAccountBasicInfoSection.vue";
import CreateAccountPlatformSelector from "@/components/account/CreateAccountPlatformSelector.vue";
import CreateAccountSchedulingSection from "@/components/account/CreateAccountSchedulingSection.vue";
import CustomBaseUrlControlSection from "@/components/account/CustomBaseUrlControlSection.vue";
import GeminiAccountTypeSection from "@/components/account/GeminiAccountTypeSection.vue";
import GeminiHelpDialog from "@/components/account/GeminiHelpDialog.vue";
import GeminiOAuthOptionsSection from "@/components/account/GeminiOAuthOptionsSection.vue";
import GrokAccountTypeSection from "@/components/account/GrokAccountTypeSection.vue";
import GrokSessionCredentialsSection from "@/components/account/GrokSessionCredentialsSection.vue";
import OpenAIAccountTypeSection from "@/components/account/OpenAIAccountTypeSection.vue";
import OpenAIOptionsSection from "@/components/account/OpenAIOptionsSection.vue";
import PoolModeSection from "@/components/account/PoolModeSection.vue";
import RpmLimitControlSection from "@/components/account/RpmLimitControlSection.vue";
import SessionLimitControlSection from "@/components/account/SessionLimitControlSection.vue";
import SessionIdMaskingControlSection from "@/components/account/SessionIdMaskingControlSection.vue";
import TempUnschedRulesSection from "@/components/account/TempUnschedRulesSection.vue";
import TlsFingerprintControlSection from "@/components/account/TlsFingerprintControlSection.vue";
import WindowCostControlSection from "@/components/account/WindowCostControlSection.vue";
import {
  buildCompatibleBaseUrlPresets,
  buildAccountOpenAIWSModeOptions,
  buildAccountTempUnschedPresets,
  buildAccountUmqModeOptions,
  buildMixedChannelDetails,
  createDefaultCreateAccountForm,
  geminiHelpLinks,
  geminiQuotaDocs,
  needsMixedChannelCheck,
  resetCreateAccountForm,
  resolveAccountApiKeyHint,
  resolveAccountApiKeyPlaceholder,
  resolveAccountBaseUrlHint,
  resolveAccountBaseUrlPlaceholder,
  resolveCreateAccountOAuthStepTitle,
  resolveMixedChannelWarningMessage,
  type CreateAccountForm,
} from "@/components/account/accountModalShared";
import { buildCreateAccountMutationPayload } from "@/components/account/accountMutationPayload";
import {
  createPlatformRequestGuard,
  createSequenceRequestGuard,
  type AccountModalPlatformRequestContext,
} from "@/components/account/accountModalRequestGuard";
import {
  type BedrockAuthMode,
  type CreateAccountCategory,
  type GeminiAIStudioTier,
  type GeminiGcpTier,
  type GeminiGoogleOneTier,
  type GeminiOAuthType,
  buildCreateAccountSharedPayload,
  buildCreateApiKeyCredentials,
  buildCreateAnthropicOAuthAccountPayload,
  buildCreateBatchAccountName,
  buildCreateOpenAICompatOAuthTarget,
  buildCreateAnthropicExtra,
  buildCreateAnthropicQuotaControlExtra,
  buildCreateAntigravityOAuthCredentials,
  buildCreateAntigravityUpstreamCredentials,
  buildCreateAntigravityExtra,
  buildCreateBedrockCredentials,
  buildCreateOpenAIExtra,
  resolveBatchCreateOutcome,
  resolveCreateAccountGeminiSelectedTier,
  resolveCreateAccountOAuthFlow,
} from "@/components/account/createAccountModalHelpers";
import {
  accountMutationProfileHasSection,
  resolveAccountMutationProfile,
} from "@/components/account/accountMutationProfiles";
import {
  appendEmptyModelMapping,
  appendPresetModelMapping,
  applyTempUnschedCredentialsState,
  confirmCustomErrorCodeSelection,
  removeModelMappingAt,
} from "@/components/account/accountModalInteractions";
import {
  assignBuiltModelMapping,
  buildTempUnschedRules,
  createTempUnschedRule,
  DEFAULT_POOL_MODE_RETRY_COUNT,
  getDefaultBaseURL,
  moveItemInPlace,
  type ModelMapping,
  type TempUnschedRuleForm,
} from "@/components/account/credentialsBuilder";
import {
  formatDateTimeLocalInput,
  parseDateTimeLocalInput,
} from "@/utils/format";
import {
  findInvalidGrokSessionBatchImportLine,
  normalizeGrokSessionToken,
} from "@/utils/grokSessionToken";
import { createStableObjectKeyResolver } from "@/utils/stableObjectKey";
import {
  OPENAI_WS_MODE_OFF,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode,
} from "@/utils/openaiWsMode";
import {
  consumeValidationFailureMessage,
  resolveAnthropicExchangeEndpoint,
  resolveOAuthExchangeState,
  runBatchCreateFlow,
  runOAuthExchangeFlow,
} from "@/components/account/oauthAuthorizationFlowHelpers";
import OAuthAuthorizationFlow from "./OAuthAuthorizationFlow.vue";

// Type for exposed OAuthAuthorizationFlow component
// Note: defineExpose automatically unwraps refs, so we use the unwrapped types
interface OAuthFlowExposed {
  authCode: string;
  oauthState: string;
  projectId: string;
  sessionKey: string;
  refreshToken: string;
  sessionToken: string;
  inputMethod: AuthInputMethod;
  reset: () => void;
}

const { t } = useI18n();
const authStore = useAuthStore();

const oauthStepTitle = computed(() => {
  return resolveCreateAccountOAuthStepTitle(form.platform, t);
});

const isGrokSessionBatchMode = computed(
  () =>
    form.platform === "grok" &&
    accountCategory.value === "session" &&
    grokSessionInputMode.value === "batch",
);

const accountNameLabel = computed(() =>
  isGrokSessionBatchMode.value
    ? t("admin.accounts.grok.batchNamePrefix")
    : t("admin.accounts.accountName"),
);

const accountNamePlaceholder = computed(() =>
  isGrokSessionBatchMode.value
    ? t("admin.accounts.grok.batchNamePrefixPlaceholder")
    : t("admin.accounts.enterAccountName"),
);

const accountNameHint = computed(() =>
  isGrokSessionBatchMode.value
    ? t("admin.accounts.grok.batchNamePrefixHint")
    : "",
);

const isAccountNameRequired = computed(() => !isGrokSessionBatchMode.value);

const primarySubmitButtonLabel = computed(() => {
  if (isOAuthFlow.value) {
    return t("common.next");
  }
  if (submitting.value) {
    return t("admin.accounts.creating");
  }
  if (isGrokSessionBatchMode.value) {
    return grokSessionBatchDryRun.value
      ? t("admin.accounts.grok.previewBatchImport")
      : t("admin.accounts.grok.batchImportAction");
  }
  return t("common.create");
});

// Platform-specific hints for API Key type
const baseUrlHint = computed(() => {
  return resolveAccountBaseUrlHint(form.platform, t);
});

const baseUrlPlaceholder = computed(() => {
  return resolveAccountBaseUrlPlaceholder(form.platform, t);
});

const apiKeyHint = computed(() => {
  return resolveAccountApiKeyHint(form.platform, t);
});

const apiKeyPlaceholder = computed(() => {
  return resolveAccountApiKeyPlaceholder(form.platform, t);
});

const compatibleBaseUrlPresets = computed(() => {
  return buildCompatibleBaseUrlPresets(form.platform, t);
});

interface Props {
  show: boolean;
  proxies: Proxy[];
  groups: AdminGroup[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  close: [];
  created: [];
}>();

type CreateRequestContext =
  AccountModalPlatformRequestContext<AccountPlatform>;

type GrokSessionInputMode = "single" | "batch";

const appStore = useAppStore();

// OAuth composables
const oauth = useAccountOAuth(); // For Anthropic OAuth
const openaiOAuth = useOpenAIOAuth(); // For OpenAI OAuth
const geminiOAuth = useGeminiOAuth(); // For Gemini OAuth
const antigravityOAuth = useAntigravityOAuth(); // For Antigravity OAuth

const currentOAuthState = computed(() => {
  if (form.platform === "openai") {
    return {
      authUrl: openaiOAuth.authUrl.value,
      sessionId: openaiOAuth.sessionId.value,
      loading: openaiOAuth.loading.value,
      error: openaiOAuth.error.value,
    };
  }
  if (form.platform === "gemini") {
    return {
      authUrl: geminiOAuth.authUrl.value,
      sessionId: geminiOAuth.sessionId.value,
      loading: geminiOAuth.loading.value,
      error: geminiOAuth.error.value,
    };
  }
  if (form.platform === "antigravity") {
    return {
      authUrl: antigravityOAuth.authUrl.value,
      sessionId: antigravityOAuth.sessionId.value,
      loading: antigravityOAuth.loading.value,
      error: antigravityOAuth.error.value,
    };
  }
  return {
    authUrl: oauth.authUrl.value,
    sessionId: oauth.sessionId.value,
    loading: oauth.loading.value,
    error: oauth.error.value,
  };
});

// Refs
const oauthFlowRef = ref<OAuthFlowExposed | null>(null);

const getDefaultAccountCategoryForPlatform = (
  platform: AccountPlatform,
): CreateAccountCategory => {
  switch (platform) {
    case "grok":
      return "apikey";
    case "anthropic":
    case "openai":
    case "gemini":
    case "antigravity":
    default:
      return "oauth-based";
  }
};

// State
const step = ref(1);
const submitting = ref(false);
const accountCategory = ref<CreateAccountCategory>(
  getDefaultAccountCategoryForPlatform("anthropic"),
); // UI selection for account category
const addMethod = ref<AddMethod>("oauth"); // For oauth-based: 'oauth' or 'setup-token'
const apiKeyBaseUrl = ref(getDefaultBaseURL("anthropic"));
const apiKeyValue = ref("");
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
const openaiOAuthResponsesWebSocketV2Mode =
  ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
const openaiAPIKeyResponsesWebSocketV2Mode =
  ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
const codexCLIOnlyEnabled = ref(false);
const anthropicPassthroughEnabled = ref(false);
const mixedScheduling = ref(false); // For antigravity accounts: enable mixed scheduling
const allowOverages = ref(false); // For antigravity accounts: enable AI Credits overages
const antigravityAccountType = ref<"oauth" | "upstream">("oauth"); // For antigravity: oauth or upstream
const upstreamBaseUrl = ref(""); // For upstream type: base URL
const upstreamApiKey = ref(""); // For upstream type: API key
const antigravityModelRestrictionMode = ref<"whitelist" | "mapping">(
  "whitelist",
);
const antigravityWhitelistModels = ref<string[]>([]);
const antigravityModelMappings = ref<ModelMapping[]>([]);
const antigravityPresetMappings = computed(() =>
  getPresetMappingsByPlatform("antigravity"),
);
const bedrockPresets = computed(() => getPresetMappingsByPlatform("bedrock"));

// Bedrock credentials
const bedrockAuthMode = ref<BedrockAuthMode>("sigv4");
const bedrockAccessKeyId = ref("");
const bedrockSecretAccessKey = ref("");
const bedrockSessionToken = ref("");
const bedrockRegion = ref("us-east-1");
const bedrockForceGlobal = ref(false);
const bedrockApiKeyValue = ref("");
const tempUnschedEnabled = ref(false);
const tempUnschedRules = ref<TempUnschedRuleForm[]>([]);
const getModelMappingKey = createStableObjectKeyResolver<ModelMapping>(
  "create-model-mapping",
);
const getAntigravityModelMappingKey =
  createStableObjectKeyResolver<ModelMapping>(
    "create-antigravity-model-mapping",
  );
const getTempUnschedRuleKey =
  createStableObjectKeyResolver<TempUnschedRuleForm>(
    "create-temp-unsched-rule",
  );
const geminiOAuthType = ref<GeminiOAuthType>("google_one");
const geminiAIStudioOAuthEnabled = ref(false);

const showMixedChannelWarning = ref(false);
const mixedChannelWarningDetails = ref<{
  groupName: string;
  currentPlatform: string;
  otherPlatform: string;
} | null>(null);
const mixedChannelWarningRawMessage = ref("");
const mixedChannelWarningAction = ref<(() => Promise<void>) | null>(null);
const antigravityMixedChannelConfirmed = ref(false);
const showGeminiHelpDialog = ref(false);

// Quota control state (Anthropic OAuth/SetupToken only)
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
const umqModeOptions = computed(() => buildAccountUmqModeOptions(t));
const tlsFingerprintEnabled = ref(false);
const tlsFingerprintProfileId = ref<number | null>(null);
const tlsFingerprintProfiles = ref<{ id: number; name: string }[]>([]);
const sessionIdMaskingEnabled = ref(false);
const cacheTTLOverrideEnabled = ref(false);
const cacheTTLOverrideTarget = ref<string>("5m");
const customBaseUrlEnabled = ref(false);
const customBaseUrl = ref("");

// Gemini tier selection (used as fallback when auto-detection is unavailable/fails)
const geminiTierGoogleOne = ref<GeminiGoogleOneTier>("google_one_free");
const geminiTierGcp = ref<GeminiGcpTier>("gcp_standard");
const geminiTierAIStudio = ref<GeminiAIStudioTier>("aistudio_free");

const geminiSelectedTier = computed(() => {
  return resolveCreateAccountGeminiSelectedTier({
    accountCategory: accountCategory.value,
    geminiOAuthType: geminiOAuthType.value,
    geminiTierAIStudio: geminiTierAIStudio.value,
    geminiTierGcp: geminiTierGcp.value,
    geminiTierGoogleOne: geminiTierGoogleOne.value,
    platform: form.platform,
  });
});

const openAIWSModeOptions = computed(() => buildAccountOpenAIWSModeOptions(t));

const openaiResponsesWebSocketV2Mode = computed({
  get: () => {
    if (form.platform === "openai" && accountCategory.value === "apikey") {
      return openaiAPIKeyResponsesWebSocketV2Mode.value;
    }
    return openaiOAuthResponsesWebSocketV2Mode.value;
  },
  set: (mode: OpenAIWSMode) => {
    if (form.platform === "openai" && accountCategory.value === "apikey") {
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
  () => form.platform === "openai" && openaiPassthroughEnabled.value,
);

const mutationProfile = computed(() =>
  resolveAccountMutationProfile(form.platform, form.type),
);

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

const mixedChannelWarningMessageText = computed(() => {
  return resolveMixedChannelWarningMessage({
    details: mixedChannelWarningDetails.value,
    rawMessage: mixedChannelWarningRawMessage.value,
    t,
  });
});

// Computed: current preset mappings based on platform
const presetMappings = computed(() =>
  getPresetMappingsByPlatform(form.platform),
);
const tempUnschedPresets = computed(() => buildAccountTempUnschedPresets(t));

const form = reactive<CreateAccountForm>(createDefaultCreateAccountForm());
let allowedModelsSyncSequence = 0;

const createRequestGuard = createPlatformRequestGuard<AccountPlatform>(
  (platform) => props.show && form.platform === platform,
);
const tlsFingerprintProfilesRequestGuard = createSequenceRequestGuard(
  () => props.show,
);
const antigravityDefaultMappingsRequestGuard = createSequenceRequestGuard(
  () => props.show && form.platform === "antigravity",
);
const geminiCapabilitiesRequestGuard = createSequenceRequestGuard(
  () =>
    props.show &&
    form.platform === "gemini" &&
    accountCategory.value === "oauth-based",
);

const beginCreateRequestContext = (
  platform: AccountPlatform = form.platform,
): CreateRequestContext => createRequestGuard.begin(platform);

const invalidateCreateRequests = () => {
  createRequestGuard.invalidate();
  submitting.value = false;
};

const isActiveCreateRequest = (requestContext: CreateRequestContext) =>
  createRequestGuard.isActive(requestContext);

const isCurrentCreateRequestSequence = (requestContext: CreateRequestContext) =>
  createRequestGuard.isCurrentSequence(requestContext);

const getCurrentCreateRequestSequence = () =>
  createRequestGuard.currentSequence();

const beginTlsFingerprintProfilesRequest = () =>
  tlsFingerprintProfilesRequestGuard.begin();

const invalidateTlsFingerprintProfilesRequests = () => {
  tlsFingerprintProfilesRequestGuard.invalidate();
};

const isActiveTlsFingerprintProfilesRequest = (requestSequence: number) =>
  tlsFingerprintProfilesRequestGuard.isActive(requestSequence);

const beginAntigravityDefaultMappingsRequest = () =>
  antigravityDefaultMappingsRequestGuard.begin();

const invalidateAntigravityDefaultMappingsRequests = () => {
  antigravityDefaultMappingsRequestGuard.invalidate();
};

const isActiveAntigravityDefaultMappingsRequest = (requestSequence: number) =>
  antigravityDefaultMappingsRequestGuard.isActive(requestSequence);

const beginGeminiCapabilitiesRequest = () =>
  geminiCapabilitiesRequestGuard.begin();

const invalidateGeminiCapabilitiesRequests = () => {
  geminiCapabilitiesRequestGuard.invalidate();
};

const isActiveGeminiCapabilitiesRequest = (requestSequence: number) =>
  geminiCapabilitiesRequestGuard.isActive(requestSequence);

const invalidateCreateModalAsyncLoads = () => {
  invalidateTlsFingerprintProfilesRequests();
  invalidateAntigravityDefaultMappingsRequests();
  invalidateGeminiCapabilitiesRequests();
  allowedModelsSyncSequence += 1;
};

const syncAllowedModelsForPlatform = async (
  platform: AccountPlatform = form.platform,
) => {
  const requestSequence = ++allowedModelsSyncSequence;
  if (platform === "grok") {
    await ensureModelCatalogLoaded(platform);
  }
  if (
    requestSequence !== allowedModelsSyncSequence ||
    !props.show ||
    form.platform !== platform
  ) {
    return;
  }
  allowedModels.value = [...getModelsByPlatform(platform)];
};

// Helper to check if current type needs OAuth flow
const isOAuthFlow = computed(() => {
  return resolveCreateAccountOAuthFlow({
    accountCategory: accountCategory.value,
    antigravityAccountType: antigravityAccountType.value,
    platform: form.platform,
  });
});

const isManualInputMethod = computed(() => {
  return oauthFlowRef.value?.inputMethod === "manual";
});

const expiresAtInput = computed({
  get: () => formatDateTimeLocal(form.expires_at),
  set: (value: string) => {
    form.expires_at = parseDateTimeLocal(value);
  },
});

const canExchangeCode = computed(() => {
  const authCode = oauthFlowRef.value?.authCode || "";
  return Boolean(
    authCode.trim() &&
    currentOAuthState.value.sessionId &&
    !currentOAuthState.value.loading,
  );
});

const loadTlsFingerprintProfiles = async () => {
  const requestSequence = beginTlsFingerprintProfilesRequest();
  tlsFingerprintProfiles.value = [];

  try {
    const profiles = await adminAPI.tlsFingerprintProfiles.list();
    if (!isActiveTlsFingerprintProfilesRequest(requestSequence)) {
      return;
    }
    tlsFingerprintProfiles.value = profiles.map((profile) => ({
      id: profile.id,
      name: profile.name,
    }));
  } catch {
    if (!isActiveTlsFingerprintProfilesRequest(requestSequence)) {
      return;
    }
    tlsFingerprintProfiles.value = [];
  }
};

const loadAntigravityDefaultMappings = async () => {
  const requestSequence = beginAntigravityDefaultMappingsRequest();
  const mappings = await fetchAntigravityDefaultMappings();
  if (!isActiveAntigravityDefaultMappingsRequest(requestSequence)) {
    return;
  }
  antigravityModelMappings.value = [...mappings];
};

const applyAntigravityModelDefaults = () => {
  antigravityModelRestrictionMode.value = "mapping";
  antigravityWhitelistModels.value = [];
  antigravityModelMappings.value = [];
  void loadAntigravityDefaultMappings();
};

const clearAntigravityModelState = () => {
  invalidateAntigravityDefaultMappingsRequests();
  antigravityModelRestrictionMode.value = "mapping";
  antigravityWhitelistModels.value = [];
  antigravityModelMappings.value = [];
};

const resetOAuthClientsState = (includeFlowState = false) => {
  oauth.resetState();
  openaiOAuth.resetState();
  geminiOAuth.resetState();
  antigravityOAuth.resetState();
  if (includeFlowState) {
    oauthFlowRef.value?.reset();
  }
};

// Watchers
watch(
  () => props.show,
  (newVal) => {
    if (newVal) {
      void loadTlsFingerprintProfiles();
      // Modal opened - fill related models
      void syncAllowedModelsForPlatform(form.platform);
      if (form.platform === "antigravity") {
        applyAntigravityModelDefaults();
      } else {
        clearAntigravityModelState();
      }
    } else {
      invalidateCreateRequests();
      invalidateCreateModalAsyncLoads();
      resetForm();
    }
  },
);

// Sync form.type based on accountCategory, addMethod, and platform-specific type
watch(
  [accountCategory, addMethod, antigravityAccountType, () => form.platform],
  ([category, method, agType, platform]) => {
    // Antigravity upstream 类型（实际创建为 apikey）
    if (platform === "antigravity" && agType === "upstream") {
      form.type = "apikey";
      return;
    }
    // Bedrock 类型
    if (platform === "anthropic" && category === "bedrock") {
      form.type = "bedrock" as AccountType;
      return;
    }
    if (platform === "grok") {
      if (category === "session") {
        form.type = "session";
        return;
      }
      if (category === "upstream") {
        form.type = "upstream";
        return;
      }
    }
    if (category === "oauth-based") {
      form.type = method as AccountType; // 'oauth' or 'setup-token'
    } else {
      form.type = "apikey";
    }
  },
  { immediate: true },
);

// Reset platform-specific settings when platform changes
watch(
  () => form.platform,
  (newPlatform) => {
    invalidateCreateRequests();
    resetMixedChannelState();
    // Reset base URL based on platform
    apiKeyBaseUrl.value = getDefaultBaseURL(newPlatform);
    apiKeyValue.value = "";
    resetGrokSessionImportState();
    // Clear model-related settings
    allowedModels.value = [];
    modelMappings.value = [];
    if (newPlatform === "antigravity") {
      applyAntigravityModelDefaults();
      antigravityAccountType.value = "oauth";
    } else {
      clearAntigravityModelState();
      allowOverages.value = false;
    }
    accountCategory.value = getDefaultAccountCategoryForPlatform(newPlatform);
    resetBedrockCredentialState();
    // Reset Anthropic/Antigravity-specific settings when switching to other platforms
    if (newPlatform !== "anthropic" && newPlatform !== "antigravity") {
      interceptWarmupRequests.value = false;
    }
    if (newPlatform !== "openai") {
      resetOpenAICreateState();
    }
    if (newPlatform !== "anthropic") {
      anthropicPassthroughEnabled.value = false;
    }
    resetOAuthClientsState();
  },
);

// Gemini AI Studio OAuth availability (requires operator-configured OAuth client)
watch([accountCategory, () => form.platform], ([category, platform]) => {
  if (platform === "openai" && category !== "oauth-based") {
    codexCLIOnlyEnabled.value = false;
  }
  if (platform !== "anthropic" || category !== "apikey") {
    anthropicPassthroughEnabled.value = false;
  }
  if (platform !== "grok" || category !== "session") {
    resetGrokSessionImportState();
  }
});

watch(
  [
    grokSessionInputMode,
    grokSessionBatchInput,
    grokSessionBatchDryRun,
    grokSessionBatchTestAfterCreate,
    grokSessionToken,
  ],
  () => {
    clearGrokSessionBatchResult();
  },
);

watch(
  [() => props.show, () => form.platform, accountCategory],
  ([show, platform, category]) => {
    void syncGeminiAIStudioOAuthAvailability(show, platform, category);
  },
  { immediate: true },
);

// Auto-fill related models when switching to whitelist mode or changing platform
watch([modelRestrictionMode, () => form.platform], ([newMode]) => {
  if (newMode === "whitelist") {
    void syncAllowedModelsForPlatform(form.platform);
  }
});

watch(
  [antigravityModelRestrictionMode, () => form.platform],
  ([, platform]) => {
    if (platform !== "antigravity") return;
    // Antigravity 默认不做限制：白名单留空表示允许所有（包含未来新增模型）。
    // 如果需要快速填充常用模型，可在组件内点“填充相关模型”。
  },
);

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
  showMixedChannelWarning.value = false;
  mixedChannelWarningDetails.value = null;
  mixedChannelWarningRawMessage.value = "";
  mixedChannelWarningAction.value = null;
};

const resetMixedChannelState = () => {
  antigravityMixedChannelConfirmed.value = false;
  clearMixedChannelDialog();
};

const resolveCreateAccountErrorMessage = (error: any) =>
  error.response?.data?.message ||
  error.response?.data?.detail ||
  t("admin.accounts.failedToCreate");

const resolveOAuthAuthErrorMessage = (error: any) =>
  error.response?.data?.detail || t("admin.accounts.oauth.authFailed");

const getCurrentProxyConfig = () =>
  form.proxy_id ? { proxy_id: form.proxy_id } : {};

const buildValidatedTempUnschedPayload = () => {
  if (!tempUnschedEnabled.value) {
    return [];
  }

  const payload = buildTempUnschedRules(tempUnschedRules.value);
  if (payload.length > 0) {
    return payload;
  }

  appStore.showError(t("admin.accounts.tempUnschedulable.rulesInvalid"));
  return null;
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
    t("admin.accounts.failedToCreate");
  mixedChannelWarningAction.value = opts.onConfirm;
  showMixedChannelWarning.value = true;
};

const withAntigravityConfirmFlag = (
  payload: CreateAccountRequest,
): CreateAccountRequest => {
  if (
    needsMixedChannelCheck(payload.platform) &&
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
  requestContext: CreateRequestContext,
): Promise<boolean> => {
  if (!needsMixedChannelCheck(form.platform)) {
    return true;
  }
  if (antigravityMixedChannelConfirmed.value) {
    return true;
  }

  try {
    const result = await adminAPI.accounts.checkMixedChannelRisk({
      platform: form.platform,
      group_ids: form.group_ids,
    });
    if (!isActiveCreateRequest(requestContext)) {
      return false;
    }
    if (!result.has_risk) {
      return true;
    }
    openMixedChannelDialog({
      response: result,
      onConfirm: async () => {
        if (!isActiveCreateRequest(requestContext)) {
          return;
        }
        antigravityMixedChannelConfirmed.value = true;
        await onConfirm();
      },
    });
    return false;
  } catch (error: any) {
    if (!isActiveCreateRequest(requestContext)) {
      return false;
    }
    appStore.showError(resolveCreateAccountErrorMessage(error));
    return false;
  }
};

const submitCreateAccount = async (
  payload: CreateAccountRequest,
  requestContext: CreateRequestContext,
) => {
  if (!isActiveCreateRequest(requestContext)) {
    return;
  }
  submitting.value = true;
  try {
    await adminAPI.accounts.create(withAntigravityConfirmFlag(payload));
    if (!isActiveCreateRequest(requestContext)) {
      return;
    }
    notifyAccountCreated();
    finalizeCreatedAndClose();
  } catch (error: any) {
    if (!isActiveCreateRequest(requestContext)) {
      return;
    }
    if (
      error.response?.status === 409 &&
      error.response?.data?.error === "mixed_channel_warning" &&
      needsMixedChannelCheck(form.platform)
    ) {
      openMixedChannelDialog({
        message: error.response?.data?.message,
        onConfirm: async () => {
          if (!isActiveCreateRequest(requestContext)) {
            return;
          }
          antigravityMixedChannelConfirmed.value = true;
          await submitCreateAccount(payload, requestContext);
        },
      });
      return;
    }
    appStore.showError(resolveCreateAccountErrorMessage(error));
  } finally {
    if (isCurrentCreateRequestSequence(requestContext)) {
      submitting.value = false;
    }
  }
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

const resetAntigravityCreateState = () => {
  allowOverages.value = false;
  antigravityAccountType.value = "oauth";
  upstreamBaseUrl.value = "";
  upstreamApiKey.value = "";
  clearAntigravityModelState();
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

async function syncGeminiAIStudioOAuthAvailability(
  show: boolean,
  platform: AccountPlatform,
  category: typeof accountCategory.value,
) {
  if (!show || platform !== "gemini" || category !== "oauth-based") {
    invalidateGeminiCapabilitiesRequests();
    geminiAIStudioOAuthEnabled.value = false;
    return;
  }

  const requestSequence = beginGeminiCapabilitiesRequest();
  const capabilities = await geminiOAuth.getCapabilities();
  if (!isActiveGeminiCapabilitiesRequest(requestSequence)) {
    return;
  }
  geminiAIStudioOAuthEnabled.value = !!capabilities?.ai_studio_oauth_enabled;
  if (
    !geminiAIStudioOAuthEnabled.value &&
    geminiOAuthType.value === "ai_studio"
  ) {
    geminiOAuthType.value = "code_assist";
  }
}

// Methods
const resetForm = () => {
  step.value = 1;
  resetCreateAccountForm(form);
  accountCategory.value = getDefaultAccountCategoryForPlatform(form.platform);
  addMethod.value = "oauth";
  apiKeyBaseUrl.value = getDefaultBaseURL("anthropic");
  apiKeyValue.value = "";
  resetGrokSessionImportState();
  resetQuotaResetState();
  modelMappings.value = [];
  modelRestrictionMode.value = "whitelist";
  allowedModels.value = [...claudeModels]; // Default fill related models
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
  resetOAuthClientsState(true);
  resetMixedChannelState();
};

const handleClose = () => {
  invalidateCreateRequests();
  invalidateCreateModalAsyncLoads();
  resetMixedChannelState();
  emit("close");
};

// Helper function to create account with mixed channel warning handling
const doCreateAccount = async (
  payload: CreateAccountRequest,
  requestContext: CreateRequestContext,
) => {
  if (!isActiveCreateRequest(requestContext)) {
    return;
  }
  const canContinue = await ensureAntigravityMixedChannelConfirmed(async () => {
    await submitCreateAccount(payload, requestContext);
  }, requestContext);
  if (!canContinue || !isActiveCreateRequest(requestContext)) {
    return;
  }
  await submitCreateAccount(payload, requestContext);
};

// Handle mixed channel warning confirmation
const handleMixedChannelConfirm = async () => {
  const action = mixedChannelWarningAction.value;
  if (!action) {
    clearMixedChannelDialog();
    return;
  }
  clearMixedChannelDialog();
  const confirmRequestSequence = getCurrentCreateRequestSequence();
  submitting.value = true;
  try {
    await action();
  } finally {
    if (confirmRequestSequence === getCurrentCreateRequestSequence()) {
      submitting.value = false;
    }
  }
};

const handleMixedChannelCancel = () => {
  clearMixedChannelDialog();
};

const finalizeCreatedAndClose = () => {
  emit("created");
  handleClose();
};

const notifyAccountCreated = () => {
  appStore.showSuccess(t("admin.accounts.accountCreated"));
};

const buildCurrentCreateSharedPayload = () =>
  buildCreateAccountSharedPayload({
    autoPauseOnExpired: autoPauseOnExpired.value,
    concurrency: form.concurrency,
    expiresAt: form.expires_at,
    groupIds: form.group_ids,
    loadFactor: form.load_factor,
    notes: form.notes,
    priority: form.priority,
    proxyId: form.proxy_id,
    rateMultiplier: form.rate_multiplier,
  });

const buildCurrentOpenAIExtra = (base?: Record<string, unknown>) =>
  buildCreateOpenAIExtra({
    accountCategory: accountCategory.value,
    base,
    codexCLIOnlyEnabled: codexCLIOnlyEnabled.value,
    openaiAPIKeyResponsesWebSocketV2Mode:
      openaiAPIKeyResponsesWebSocketV2Mode.value,
    openaiOAuthResponsesWebSocketV2Mode:
      openaiOAuthResponsesWebSocketV2Mode.value,
    openaiPassthroughEnabled: openaiPassthroughEnabled.value,
    platform: form.platform,
  });

const buildCurrentAnthropicQuotaExtra = (baseExtra?: Record<string, unknown>) =>
  buildCreateAnthropicQuotaControlExtra({
    baseExtra,
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
  });

const buildCurrentAntigravityExtra = () =>
  buildCreateAntigravityExtra({
    allowOverages: allowOverages.value,
    mixedScheduling: mixedScheduling.value,
  });

const applyOpenAIModelRestrictionIfNeeded = (
  credentials: Record<string, unknown>,
  shouldApply: boolean,
) => {
  if (!shouldApply) {
    return;
  }

  assignBuiltModelMapping(
    credentials,
    modelRestrictionMode.value,
    allowedModels.value,
    modelMappings.value,
  );
};

const handleBatchCreateOutcome = (
  options: {
    failedCount: number;
    successCount: number;
    errors: string[];
    setError: (message: string) => void;
  },
  requestContext: CreateRequestContext,
) => {
  if (!isActiveCreateRequest(requestContext)) {
    return;
  }
  const outcome = resolveBatchCreateOutcome({
    failedCount: options.failedCount,
    successCount: options.successCount,
    t,
  });

  if (outcome.type === "success") {
    appStore.showSuccess(outcome.message);
  } else if (outcome.type === "warning") {
    appStore.showWarning(outcome.message);
    options.setError(options.errors.join("\n"));
  } else {
    options.setError(options.errors.join("\n"));
    appStore.showError(outcome.message);
  }

  if (outcome.shouldEmitCreated) {
    emit("created");
  }
  if (outcome.shouldClose) {
    handleClose();
  }
};

const resolveBatchCreateUnexpectedError = (error: any) =>
  error?.response?.data?.detail || error?.message || "Unknown error";

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

const normalizeOptionalText = (value: string) => {
  const trimmed = value.trim();
  return trimmed ? trimmed : undefined;
};

const notifyGrokSessionBatchImportOutcome = (
  result: GrokSessionBatchImportResult,
) => {
  if (result.dry_run) {
    appStore.showInfo(
      t("admin.accounts.grok.batchImportDryRunSummary", {
        created: result.created,
        invalid: result.invalid,
        skipped: result.skipped,
      }),
    );
    return;
  }

  if (result.created > 0 && result.skipped === 0 && result.invalid === 0) {
    appStore.showSuccess(
      t("admin.accounts.grok.batchImportSuccess", { count: result.created }),
    );
    return;
  }

  if (result.created > 0) {
    appStore.showWarning(
      t("admin.accounts.grok.batchImportPartial", {
        created: result.created,
        skipped: result.skipped,
        invalid: result.invalid,
      }),
    );
    return;
  }

  appStore.showError(
    t("admin.accounts.grok.batchImportFailedSummary", {
      skipped: result.skipped,
      invalid: result.invalid,
    }),
  );
};

const submitGrokSessionBatchImport = async (
  requestContext: CreateRequestContext,
) => {
  const rawInput = grokSessionBatchInput.value.trim();
  if (!rawInput) {
    appStore.showError(t("admin.accounts.grok.batchImportInputRequired"));
    return;
  }
  const invalidLine = findInvalidGrokSessionBatchImportLine(rawInput);
  if (invalidLine !== null) {
    appStore.showError(
      t("admin.accounts.grok.batchImportInvalidFormat", { line: invalidLine }),
    );
    return;
  }
  if (!isActiveCreateRequest(requestContext)) {
    return;
  }

  submitting.value = true;
  clearGrokSessionBatchResult();

  try {
    const result = await adminAPI.accounts.batchImportGrokSession({
      raw_input: rawInput,
      name_prefix: normalizeOptionalText(form.name),
      group_ids: form.group_ids,
      proxy_id: form.proxy_id,
      priority: form.priority,
      concurrency: form.concurrency,
      rate_multiplier: form.rate_multiplier,
      load_factor: form.load_factor,
      notes: normalizeOptionalText(form.notes) ?? null,
      dedupe_strategy: "skip_existing",
      dry_run: grokSessionBatchDryRun.value,
      test_after_create: grokSessionBatchTestAfterCreate.value,
    });

    if (!isActiveCreateRequest(requestContext)) {
      return;
    }

    grokSessionBatchResult.value = result;
    notifyGrokSessionBatchImportOutcome(result);
    if (!result.dry_run && result.created > 0) {
      emit("created");
    }
  } catch (error: any) {
    if (!isActiveCreateRequest(requestContext)) {
      return;
    }
    appStore.showError(resolveBatchCreateUnexpectedError(error));
  } finally {
    if (isCurrentCreateRequestSequence(requestContext)) {
      submitting.value = false;
    }
  }
};

const createOAuthAccount = async (
  options: {
    commonPayload: ReturnType<typeof buildCurrentCreateSharedPayload>;
    name: string;
    platform: AccountPlatform;
    type: AccountType;
    credentials: Record<string, unknown>;
    extra?: Record<string, unknown>;
  },
  requestContext: CreateRequestContext,
) => {
  if (!isActiveCreateRequest(requestContext)) {
    return;
  }

  await adminAPI.accounts.create(
    buildCreateAccountMutationPayload({
      common: options.commonPayload,
      name: options.name,
      platform: options.platform,
      type: options.type,
      credentials: options.credentials,
      extra: options.extra,
    }),
  );
};

const createBatchCompletionHandler =
  (errorRef: { value: string }, requestContext: CreateRequestContext) =>
  (result: { failedCount: number; successCount: number; errors: string[] }) => {
    handleBatchCreateOutcome(
      {
        failedCount: result.failedCount,
        successCount: result.successCount,
        errors: result.errors,
        setError: (message) => {
          errorRef.value = message;
        },
      },
      requestContext,
    );
  };

const resolveCurrentOAuthState = (
  fallbackState: string | undefined,
  errorRef: { value: string },
) =>
  resolveOAuthExchangeState({
    fallbackState,
    inputState: oauthFlowRef.value?.oauthState,
    onMissingState: (message) => {
      errorRef.value = message;
      appStore.showError(message);
    },
    authFailedMessage: t("admin.accounts.oauth.authFailed"),
  });

const ensureCreateAccountName = () => {
  if (isGrokSessionBatchMode.value) {
    return true;
  }
  if (form.name.trim()) {
    return true;
  }
  appStore.showError(t("admin.accounts.pleaseEnterAccountName"));
  return false;
};

const buildBedrockCreateCredentials = () => {
  const result = buildCreateBedrockCredentials({
    accessKeyId: bedrockAccessKeyId.value,
    allowedModels: allowedModels.value,
    apiKey: bedrockApiKeyValue.value,
    authMode: bedrockAuthMode.value,
    forceGlobal: bedrockForceGlobal.value,
    interceptWarmupRequests: interceptWarmupRequests.value,
    mode: modelRestrictionMode.value,
    modelMappings: modelMappings.value,
    poolModeEnabled: poolModeEnabled.value,
    poolModeRetryCount: poolModeRetryCount.value,
    region: bedrockRegion.value,
    secretAccessKey: bedrockSecretAccessKey.value,
    sessionToken: bedrockSessionToken.value,
  });

  if (result.errorMessageKey) {
    appStore.showError(t(result.errorMessageKey));
    return null;
  }

  return result.credentials || null;
};

const buildAntigravityUpstreamCreateCredentials = () => {
  const result = buildCreateAntigravityUpstreamCredentials({
    apiKey: upstreamApiKey.value,
    baseUrl: upstreamBaseUrl.value,
    interceptWarmupRequests: interceptWarmupRequests.value,
    modelMappings: antigravityModelMappings.value,
  });

  if (result.errorMessageKey) {
    appStore.showError(t(result.errorMessageKey));
    return null;
  }

  return result.credentials || null;
};

const buildApiKeyCreateCredentials = () => {
  const result = buildCreateApiKeyCredentials({
    allowedModels: allowedModels.value,
    apiKey: apiKeyValue.value,
    baseUrl: apiKeyBaseUrl.value,
    customErrorCodesEnabled: customErrorCodesEnabled.value,
    geminiTierId: geminiTierAIStudio.value,
    interceptWarmupRequests: interceptWarmupRequests.value,
    isOpenAIModelRestrictionDisabled: isOpenAIModelRestrictionDisabled.value,
    mode: modelRestrictionMode.value,
    modelMappings: modelMappings.value,
    platform: form.platform,
    poolModeEnabled: poolModeEnabled.value,
    poolModeRetryCount: poolModeRetryCount.value,
    selectedErrorCodes: selectedErrorCodes.value,
  });

  if (result.errorMessageKey) {
    appStore.showError(t(result.errorMessageKey));
    return null;
  }

  const credentials = result.credentials;
  if (!credentials) {
    return null;
  }

  if (
    !applyTempUnschedCredentialsState(credentials, {
      tempUnschedEnabled: tempUnschedEnabled.value,
      tempUnschedRules: tempUnschedRules.value,
      showError: appStore.showError,
      t,
    })
  ) {
    return null;
  }

  return credentials;
};

const handleSubmit = async () => {
  // For OAuth-based type, handle OAuth flow (goes to step 2)
  if (isOAuthFlow.value) {
    if (!ensureCreateAccountName()) {
      return;
    }
    const requestContext = beginCreateRequestContext();
    const canContinue = await ensureAntigravityMixedChannelConfirmed(
      async () => {
        if (!isActiveCreateRequest(requestContext)) {
          return;
        }
        step.value = 2;
      },
      requestContext,
    );
    if (!canContinue || !isActiveCreateRequest(requestContext)) {
      return;
    }
    step.value = 2;
    return;
  }

  if (!ensureCreateAccountName()) {
    return;
  }

  // For Bedrock type, create directly
  if (form.platform === "anthropic" && accountCategory.value === "bedrock") {
    const requestContext = beginCreateRequestContext();
    const credentials = buildBedrockCreateCredentials();
    if (!credentials) {
      return;
    }
    await createAccountAndFinish(
      "anthropic",
      "bedrock" as AccountType,
      credentials,
      undefined,
      requestContext,
    );
    return;
  }

  // For Antigravity upstream type, create directly
  if (
    form.platform === "antigravity" &&
    antigravityAccountType.value === "upstream"
  ) {
    const requestContext = beginCreateRequestContext();
    const credentials = buildAntigravityUpstreamCreateCredentials();
    if (!credentials) {
      return;
    }
    await createAccountAndFinish(
      form.platform,
      "apikey",
      credentials,
      buildCurrentAntigravityExtra(),
      requestContext,
    );
    return;
  }

  if (form.platform === "grok" && accountCategory.value === "session") {
    const requestContext = beginCreateRequestContext();
    if (grokSessionInputMode.value === "batch") {
      await submitGrokSessionBatchImport(requestContext);
      return;
    }

    const sessionToken = grokSessionToken.value.trim();
    if (!sessionToken) {
      appStore.showError(t("admin.accounts.grok.sessionTokenRequired"));
      return;
    }
    const normalizedSessionToken = normalizeGrokSessionToken(sessionToken);
    if (!normalizedSessionToken) {
      appStore.showError(t("admin.accounts.grok.sessionTokenInvalidFormat"));
      return;
    }

    await createAccountAndFinish(
      form.platform,
      "session",
      { session_token: normalizedSessionToken },
      undefined,
      requestContext,
    );
    return;
  }

  // For apikey type, create directly
  const credentials = buildApiKeyCreateCredentials();
  if (!credentials) {
    return;
  }

  form.credentials = credentials;
  const extra = buildCreateAnthropicExtra({
    accountCategory: accountCategory.value,
    anthropicPassthroughEnabled: anthropicPassthroughEnabled.value,
    base: buildCurrentOpenAIExtra(),
    platform: form.platform,
  });

  const requestContext = beginCreateRequestContext();
  await createAccountAndFinish(
    form.platform,
    form.type,
    credentials,
    extra,
    requestContext,
  );
};

const goBackToBasicInfo = () => {
  invalidateCreateRequests();
  step.value = 1;
  resetMixedChannelState();
  resetOAuthClientsState(true);
};

const runPlatformOAuthGenerateUrl = async () => {
  switch (form.platform) {
    case "openai":
      await openaiOAuth.generateAuthUrl(form.proxy_id);
      return;
    case "gemini":
      await geminiOAuth.generateAuthUrl(
        form.proxy_id,
        oauthFlowRef.value?.projectId,
        geminiOAuthType.value,
        geminiSelectedTier.value,
      );
      return;
    case "antigravity":
      await antigravityOAuth.generateAuthUrl(form.proxy_id);
      return;
    default:
      await oauth.generateAuthUrl(addMethod.value, form.proxy_id);
  }
};

const handleGenerateUrl = async () => {
  await runPlatformOAuthGenerateUrl();
};

const runPlatformRefreshTokenValidation = (refreshToken: string) => {
  if (form.platform === "openai") {
    handleOpenAIValidateRT(refreshToken);
    return;
  }
  if (form.platform === "antigravity") {
    handleAntigravityValidateRT(refreshToken);
  }
};

const handleValidateRefreshToken = (rt: string) => {
  runPlatformRefreshTokenValidation(rt);
};

const formatDateTimeLocal = formatDateTimeLocalInput;
const parseDateTimeLocal = parseDateTimeLocalInput;

const buildCurrentQuotaMutationOptions = () => ({
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
  quotaNotifyWeeklyThresholdType: editQuotaNotifyWeeklyThresholdType.value,
  quotaNotifyTotalEnabled: editQuotaNotifyTotalEnabled.value,
  quotaNotifyTotalThreshold: editQuotaNotifyTotalThreshold.value,
  quotaNotifyTotalThresholdType: editQuotaNotifyTotalThresholdType.value,
  resetTimezone: editResetTimezone.value,
  weeklyResetDay: editWeeklyResetDay.value,
  weeklyResetHour: editWeeklyResetHour.value,
  weeklyResetMode: editWeeklyResetMode.value,
});

// Create account and handle success/failure
const createAccountAndFinish = async (
  platform: AccountPlatform,
  type: AccountType,
  credentials: Record<string, unknown>,
  extra: Record<string, unknown> | undefined,
  requestContext: CreateRequestContext,
) => {
  if (!isActiveCreateRequest(requestContext)) {
    return;
  }
  if (
    !applyTempUnschedCredentialsState(credentials, {
      tempUnschedEnabled: tempUnschedEnabled.value,
      tempUnschedRules: tempUnschedRules.value,
      showError: appStore.showError,
      t,
    })
  ) {
    return;
  }

  await doCreateAccount(
    buildCreateAccountMutationPayload({
      common: buildCurrentCreateSharedPayload(),
      name: form.name,
      platform,
      type,
      credentials,
      extra,
      quota: buildCurrentQuotaMutationOptions(),
    }),
    requestContext,
  );
};

const buildAnthropicOAuthExtra = (tokenInfo: Record<string, unknown>) =>
  buildCurrentAnthropicQuotaExtra(oauth.buildExtraInfo(tokenInfo) || {});

const createAnthropicOAuthAccountFromTokenInfo = async (options: {
  commonPayload: ReturnType<typeof buildCurrentCreateSharedPayload>;
  index?: number;
  requestContext: CreateRequestContext;
  tempUnschedPayload?: ReturnType<typeof buildTempUnschedRules>;
  tokenInfo: Record<string, unknown>;
  total?: number;
}) => {
  if (!isActiveCreateRequest(options.requestContext)) {
    return;
  }

  await adminAPI.accounts.create(
    buildCreateAnthropicOAuthAccountPayload({
      common: options.commonPayload,
      name: buildCreateBatchAccountName(
        form.name,
        options.index ?? 0,
        options.total ?? 1,
      ),
      platform: form.platform,
      type: addMethod.value as AccountType,
      interceptWarmupRequests: interceptWarmupRequests.value,
      tempUnschedPayload: options.tempUnschedPayload,
      tokenInfo: options.tokenInfo,
      extra: buildAnthropicOAuthExtra(options.tokenInfo),
    }),
  );
};

// OpenAI OAuth 授权码兑换
const handleOpenAIExchange = async (authCode: string) => {
  const oauthClient = openaiOAuth;
  if (!authCode.trim() || !oauthClient.sessionId.value) return;
  const requestContext = beginCreateRequestContext();

  await runOAuthExchangeFlow(
    oauthClient,
    async () => {
      const stateToUse = resolveCurrentOAuthState(
        oauthClient.oauthState.value,
        oauthClient.error,
      );
      if (!stateToUse) {
        return;
      }

      const tokenInfo = await oauthClient.exchangeAuthCode(
        authCode.trim(),
        oauthClient.sessionId.value,
        stateToUse,
        form.proxy_id,
      );
      if (!tokenInfo) return;
      if (!isActiveCreateRequest(requestContext)) return;

      const credentials = oauthClient.buildCredentials(tokenInfo);
      const oauthExtra = oauthClient.buildExtraInfo(tokenInfo) as
        | Record<string, unknown>
        | undefined;
      const extra = buildCurrentOpenAIExtra(oauthExtra);

      applyOpenAIModelRestrictionIfNeeded(
        credentials,
        form.platform === "openai" && !isOpenAIModelRestrictionDisabled.value,
      );

      // 应用临时不可调度配置
      if (
        !applyTempUnschedCredentialsState(credentials, {
          tempUnschedEnabled: tempUnschedEnabled.value,
          tempUnschedRules: tempUnschedRules.value,
          showError: appStore.showError,
          t,
        })
      ) {
        return;
      }

      const commonPayload = buildCurrentCreateSharedPayload();
      const target = buildCreateOpenAICompatOAuthTarget({
        baseName: form.name,
        credentials,
        extra,
        platform: "openai",
      });

      await createOAuthAccount(
        {
          commonPayload,
          ...target,
        },
        requestContext,
      );
      if (!isActiveCreateRequest(requestContext)) return;
      notifyAccountCreated();

      finalizeCreatedAndClose();
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError,
    {
      isActive: () => isActiveCreateRequest(requestContext),
    },
  );
};

// OpenAI 手动 RT 批量验证和创建
// OpenAI Mobile RT 使用的 client_id
const OPENAI_MOBILE_RT_CLIENT_ID = "app_LlGpXReQgckcGGUo2JrYvtJK";

// OpenAI RT 批量验证和创建
const handleOpenAIBatchRT = async (
  refreshTokenInput: string,
  clientId?: string,
) => {
  const oauthClient = openaiOAuth;
  const commonPayload = buildCurrentCreateSharedPayload();
  const requestContext = beginCreateRequestContext();
  await runBatchCreateFlow({
    rawInput: refreshTokenInput,
    emptyInputMessage: t("admin.accounts.oauth.openai.pleaseEnterRefreshToken"),
    loadingRef: oauthClient.loading,
    errorRef: oauthClient.error,
    isActive: () => isActiveCreateRequest(requestContext),
    onComplete: createBatchCompletionHandler(oauthClient.error, requestContext),
    processEntry: async (refreshToken, index, refreshTokens) => {
      if (!isActiveCreateRequest(requestContext)) {
        return null;
      }
      const tokenInfo = await oauthClient.validateRefreshToken(
        refreshToken,
        form.proxy_id,
        clientId,
      );
      if (!tokenInfo) {
        if (!isActiveCreateRequest(requestContext)) {
          return null;
        }
        return consumeValidationFailureMessage(oauthClient.error);
      }
      if (!isActiveCreateRequest(requestContext)) {
        return null;
      }

      const credentials = oauthClient.buildCredentials(tokenInfo);
      if (clientId) {
        credentials.client_id = clientId;
      }
      const oauthExtra = oauthClient.buildExtraInfo(tokenInfo) as
        | Record<string, unknown>
        | undefined;
      const extra = buildCurrentOpenAIExtra(oauthExtra);

      applyOpenAIModelRestrictionIfNeeded(
        credentials,
        form.platform === "openai" && !isOpenAIModelRestrictionDisabled.value,
      );

      const target = buildCreateOpenAICompatOAuthTarget({
        baseName: form.name,
        credentials,
        extra,
        fallbackBaseName: tokenInfo.email || "OpenAI OAuth Account",
        index,
        platform: "openai",
        total: refreshTokens.length,
      });

      await createOAuthAccount(
        {
          commonPayload,
          ...target,
        },
        requestContext,
      );

      return null;
    },
    resolveUnexpectedError: resolveBatchCreateUnexpectedError,
  });
};

// 手动输入 RT（Codex CLI client_id，默认）
const handleOpenAIValidateRT = (rt: string) => handleOpenAIBatchRT(rt);

// 手动输入 Mobile RT
const handleOpenAIValidateMobileRT = (rt: string) =>
  handleOpenAIBatchRT(rt, OPENAI_MOBILE_RT_CLIENT_ID);

// Antigravity 手动 RT 批量验证和创建
const handleAntigravityValidateRT = async (refreshTokenInput: string) => {
  const commonPayload = buildCurrentCreateSharedPayload();
  const requestContext = beginCreateRequestContext();
  await runBatchCreateFlow({
    rawInput: refreshTokenInput,
    emptyInputMessage: t(
      "admin.accounts.oauth.antigravity.pleaseEnterRefreshToken",
    ),
    loadingRef: antigravityOAuth.loading,
    errorRef: antigravityOAuth.error,
    isActive: () => isActiveCreateRequest(requestContext),
    onComplete: createBatchCompletionHandler(
      antigravityOAuth.error,
      requestContext,
    ),
    processEntry: async (refreshToken, index, refreshTokens) => {
      if (!isActiveCreateRequest(requestContext)) {
        return null;
      }
      const tokenInfo = await antigravityOAuth.validateRefreshToken(
        refreshToken,
        form.proxy_id,
      );
      if (!tokenInfo) {
        if (!isActiveCreateRequest(requestContext)) {
          return null;
        }
        return consumeValidationFailureMessage(antigravityOAuth.error);
      }
      if (!isActiveCreateRequest(requestContext)) {
        return null;
      }

      const credentials = antigravityOAuth.buildCredentials(tokenInfo);
      const createPayload = withAntigravityConfirmFlag(
        buildCreateAccountMutationPayload({
          common: commonPayload,
          name: buildCreateBatchAccountName(
            form.name,
            index,
            refreshTokens.length,
          ),
          platform: "antigravity",
          type: "oauth",
          credentials,
          extra: buildCurrentAntigravityExtra(),
        }),
      );
      if (!isActiveCreateRequest(requestContext)) {
        return null;
      }
      await adminAPI.accounts.create(createPayload);
      return null;
    },
    resolveUnexpectedError: resolveBatchCreateUnexpectedError,
  });
};

// Gemini OAuth 授权码兑换
const handleGeminiExchange = async (authCode: string) => {
  if (!authCode.trim() || !geminiOAuth.sessionId.value) return;
  const requestContext = beginCreateRequestContext();

  await runOAuthExchangeFlow(
    geminiOAuth,
    async () => {
      const stateToUse = resolveCurrentOAuthState(
        geminiOAuth.state.value,
        geminiOAuth.error,
      );
      if (!stateToUse) {
        return;
      }

      const tokenInfo = await geminiOAuth.exchangeAuthCode({
        code: authCode.trim(),
        sessionId: geminiOAuth.sessionId.value,
        state: stateToUse,
        proxyId: form.proxy_id,
        oauthType: geminiOAuthType.value,
        tierId: geminiSelectedTier.value,
      });
      if (!tokenInfo) return;
      if (!isActiveCreateRequest(requestContext)) return;

      const credentials = geminiOAuth.buildCredentials(tokenInfo);
      const extra = geminiOAuth.buildExtraInfo(tokenInfo);
      await createAccountAndFinish(
        "gemini",
        "oauth",
        credentials,
        extra,
        requestContext,
      );
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError,
    {
      isActive: () => isActiveCreateRequest(requestContext),
    },
  );
};

// Antigravity OAuth 授权码兑换
const handleAntigravityExchange = async (authCode: string) => {
  if (!authCode.trim() || !antigravityOAuth.sessionId.value) return;
  const requestContext = beginCreateRequestContext();

  await runOAuthExchangeFlow(
    antigravityOAuth,
    async () => {
      const stateToUse = resolveCurrentOAuthState(
        antigravityOAuth.state.value,
        antigravityOAuth.error,
      );
      if (!stateToUse) {
        return;
      }

      const tokenInfo = await antigravityOAuth.exchangeAuthCode({
        code: authCode.trim(),
        sessionId: antigravityOAuth.sessionId.value,
        state: stateToUse,
        proxyId: form.proxy_id,
      });
      if (!tokenInfo) return;
      if (!isActiveCreateRequest(requestContext)) return;

      const credentials = buildCreateAntigravityOAuthCredentials({
        interceptWarmupRequests: interceptWarmupRequests.value,
        modelMappings: antigravityModelMappings.value,
        tokenInfo: antigravityOAuth.buildCredentials(tokenInfo),
      });
      const extra = buildCurrentAntigravityExtra();
      await createAccountAndFinish(
        "antigravity",
        "oauth",
        credentials,
        extra,
        requestContext,
      );
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError,
    {
      isActive: () => isActiveCreateRequest(requestContext),
    },
  );
};

// Anthropic OAuth 授权码兑换
const handleAnthropicExchange = async (authCode: string) => {
  if (!authCode.trim() || !oauth.sessionId.value) return;
  const requestContext = beginCreateRequestContext();

  await runOAuthExchangeFlow(
    oauth,
    async () => {
      const tokenInfo = await adminAPI.accounts.exchangeCode(
        resolveAnthropicExchangeEndpoint(
          addMethod.value as "oauth" | "setup-token",
          "code",
        ),
        {
          session_id: oauth.sessionId.value,
          code: authCode.trim(),
          ...getCurrentProxyConfig(),
        },
      );

      await doCreateAccount(
        buildCreateAnthropicOAuthAccountPayload({
          common: buildCurrentCreateSharedPayload(),
          name: form.name,
          platform: form.platform,
          type: addMethod.value as AccountType,
          interceptWarmupRequests: interceptWarmupRequests.value,
          tokenInfo,
          extra: buildAnthropicOAuthExtra(tokenInfo),
        }),
        requestContext,
      );
    },
    resolveOAuthAuthErrorMessage,
    appStore.showError,
    {
      isActive: () => isActiveCreateRequest(requestContext),
    },
  );
};

// 主入口：根据平台路由到对应处理函数
const runPlatformOAuthExchange = async (authCode: string) => {
  switch (form.platform) {
    case "openai":
      return handleOpenAIExchange(authCode);
    case "gemini":
      return handleGeminiExchange(authCode);
    case "antigravity":
      return handleAntigravityExchange(authCode);
    default:
      return handleAnthropicExchange(authCode);
  }
};

const handleExchangeCode = async () => {
  await runPlatformOAuthExchange(oauthFlowRef.value?.authCode || "");
};

const handleCookieAuth = async (sessionKey: string) => {
  try {
    const requestContext = beginCreateRequestContext();
    const keys = oauth.parseSessionKeys(sessionKey);

    if (keys.length === 0) {
      oauth.error.value = t("admin.accounts.oauth.pleaseEnterSessionKey");
      return;
    }

    const tempUnschedPayload = buildValidatedTempUnschedPayload();
    if (tempUnschedPayload == null) {
      return;
    }

    const commonPayload = buildCurrentCreateSharedPayload();

    await runOAuthExchangeFlow(
      oauth,
      async () => {
        await runBatchCreateFlow({
          rawInput: keys.join("\n"),
          emptyInputMessage: t("admin.accounts.oauth.pleaseEnterSessionKey"),
          loadingRef: oauth.loading,
          errorRef: oauth.error,
          isActive: () => isActiveCreateRequest(requestContext),
          onComplete: ({ successCount, failedCount, errors }) => {
            if (!isActiveCreateRequest(requestContext)) {
              return;
            }
            if (successCount > 0) {
              appStore.showSuccess(
                t("admin.accounts.oauth.successCreated", {
                  count: successCount,
                }),
              );
              emit("created");
              if (failedCount === 0) {
                handleClose();
              }
            }

            if (failedCount > 0) {
              oauth.error.value = errors.join("\n");
            }
          },
          processEntry: async (key, index, allKeys) => {
            try {
              if (!isActiveCreateRequest(requestContext)) {
                return null;
              }
              const tokenInfo = await adminAPI.accounts.exchangeCode(
                resolveAnthropicExchangeEndpoint(
                  addMethod.value as "oauth" | "setup-token",
                  "cookie",
                ),
                {
                  session_id: "",
                  code: key,
                  ...getCurrentProxyConfig(),
                },
              );
              if (!isActiveCreateRequest(requestContext)) {
                return null;
              }

              await createAnthropicOAuthAccountFromTokenInfo({
                commonPayload,
                index,
                requestContext,
                tempUnschedPayload,
                tokenInfo,
                total: allKeys.length,
              });
              return null;
            } catch (error: any) {
              if (!isActiveCreateRequest(requestContext)) {
                return null;
              }
              return t("admin.accounts.oauth.keyAuthFailed", {
                index: index + 1,
                error: resolveOAuthAuthErrorMessage(error),
              });
            }
          },
        });
      },
      resolveOAuthAuthErrorMessage,
      appStore.showError,
      {
        isActive: () => isActiveCreateRequest(requestContext),
      },
    );
  } catch (error: any) {
    oauth.error.value =
      error.response?.data?.detail ||
      t("admin.accounts.oauth.cookieAuthFailed");
  }
};
</script>

<style scoped>
.create-account-modal__step-node--active {
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.create-account-modal__step-node--idle {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 84%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
}

.create-account-modal__step-label {
  color: var(--theme-page-text);
}

.create-account-modal__step-connector {
  background: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}

.create-account-modal__choice-card {
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.75rem;
  border-color: color-mix(in srgb, var(--theme-card-border) 72%, transparent);
  background: color-mix(
    in srgb,
    var(--theme-surface-soft) 82%,
    var(--theme-surface)
  );
}

.create-account-modal__choice-card:hover {
  border-color: color-mix(
    in srgb,
    var(--theme-page-border) 92%,
    var(--theme-accent)
  );
}

.create-account-modal__choice-card--disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.create-account-modal__choice-card--idle {
  color: var(--theme-page-text);
}

.create-account-modal__choice-icon--idle {
  background: color-mix(
    in srgb,
    var(--theme-page-border) 86%,
    var(--theme-surface)
  );
  color: var(--theme-page-muted);
}

.create-account-modal__choice-icon-control {
  border-radius: calc(var(--theme-button-radius) + 1px);
}

.create-account-modal__choice-title {
  color: var(--theme-page-text);
}

.create-account-modal__choice-description {
  color: var(--theme-page-muted);
}

.create-account-modal__help-button {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: var(--theme-button-radius);
  padding: calc(var(--theme-button-padding-y) * 0.4)
    calc(var(--theme-button-padding-x) * 0.4);
  font-size: 0.75rem;
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 82%,
    var(--theme-page-text)
  );
}

.create-account-modal__help-button:hover {
  background: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 10%,
    var(--theme-surface)
  );
}

.create-account-modal__error-text {
  color: rgb(var(--theme-danger-rgb));
  font-size: 0.75rem;
}

.create-account-modal__inline-toggle {
  color: var(--theme-page-muted);
}

.create-account-modal__inline-toggle:hover {
  color: var(--theme-page-text);
}

.create-account-modal__notice {
  border-radius: var(--theme-auth-feedback-radius);
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.create-account-modal__notice-block {
  padding: var(--theme-auth-callback-feedback-padding);
}

.create-account-modal__notice-tooltip {
  width: 20rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--theme-button-radius);
}

.form-section {
  border-top: 1px solid
    color-mix(in srgb, var(--theme-page-border) 76%, transparent);
  padding-top: 1rem;
}

.create-account-modal__tone-tag {
  display: inline-flex;
  align-items: center;
}

.create-account-modal__tone-tag-control {
  border-radius: var(--theme-button-radius);
  padding: 0.125rem 0.5rem;
}

.create-account-modal__tone-tag-anchor {
  margin-left: auto;
  flex-shrink: 0;
  font-size: 0.75rem;
}

.create-account-modal__choice-card--rose,
.create-account-modal__choice-icon--rose,
.create-account-modal__tone-tag--rose,
.create-account-modal__notice--rose {
  --create-account-tone-rgb: var(--theme-brand-rose-rgb);
}

.create-account-modal__choice-card--orange,
.create-account-modal__choice-icon--orange,
.create-account-modal__tone-tag--orange,
.create-account-modal__notice--orange {
  --create-account-tone-rgb: var(--theme-brand-orange-rgb);
}

.create-account-modal__choice-card--purple,
.create-account-modal__choice-icon--purple,
.create-account-modal__tone-tag--purple,
.create-account-modal__notice--purple {
  --create-account-tone-rgb: var(--theme-brand-purple-rgb);
}

.create-account-modal__choice-card--amber,
.create-account-modal__choice-icon--amber,
.create-account-modal__tone-tag--amber,
.create-account-modal__notice--amber {
  --create-account-tone-rgb: var(--theme-warning-rgb);
}

.create-account-modal__choice-card--green,
.create-account-modal__choice-icon--green,
.create-account-modal__tone-tag--green,
.create-account-modal__notice--green {
  --create-account-tone-rgb: var(--theme-success-rgb);
}

.create-account-modal__choice-card--blue,
.create-account-modal__choice-icon--blue,
.create-account-modal__tone-tag--blue,
.create-account-modal__notice--blue {
  --create-account-tone-rgb: var(--theme-info-rgb);
}

.create-account-modal__choice-card--emerald,
.create-account-modal__choice-icon--emerald,
.create-account-modal__tone-tag--emerald,
.create-account-modal__notice--emerald {
  --create-account-tone-rgb: var(--theme-success-rgb);
}

.create-account-modal__choice-card--rose,
.create-account-modal__choice-card--orange,
.create-account-modal__choice-card--purple,
.create-account-modal__choice-card--amber,
.create-account-modal__choice-card--green,
.create-account-modal__choice-card--blue,
.create-account-modal__choice-card--emerald {
  border-color: rgb(var(--create-account-tone-rgb));
  background: color-mix(
    in srgb,
    rgb(var(--create-account-tone-rgb)) 12%,
    var(--theme-surface)
  );
}

.create-account-modal__choice-icon--rose,
.create-account-modal__choice-icon--orange,
.create-account-modal__choice-icon--purple,
.create-account-modal__choice-icon--amber,
.create-account-modal__choice-icon--green,
.create-account-modal__choice-icon--blue,
.create-account-modal__choice-icon--emerald {
  background: rgb(var(--create-account-tone-rgb));
  color: var(--theme-filled-text);
}

.create-account-modal__tone-tag--rose,
.create-account-modal__tone-tag--orange,
.create-account-modal__tone-tag--purple,
.create-account-modal__tone-tag--amber,
.create-account-modal__tone-tag--green,
.create-account-modal__tone-tag--blue,
.create-account-modal__tone-tag--emerald {
  background: color-mix(
    in srgb,
    rgb(var(--create-account-tone-rgb)) 16%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--create-account-tone-rgb)) 88%,
    var(--theme-page-text)
  );
}

.create-account-modal__notice--rose,
.create-account-modal__notice--orange,
.create-account-modal__notice--purple,
.create-account-modal__notice--amber,
.create-account-modal__notice--green,
.create-account-modal__notice--blue,
.create-account-modal__notice--emerald {
  background: color-mix(
    in srgb,
    rgb(var(--create-account-tone-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--create-account-tone-rgb)) 84%,
    var(--theme-page-text)
  );
}

</style>
