import type { Account, UpdateAccountRequest } from "@/types";
import {
  buildEditAccountMutationPayload,
  type AccountMutationPayloadBuildResult,
} from "@/components/account/accountMutationPayload";
import type { useEditAccountFormState } from "@/components/account/useEditAccountFormState";
import type { useEditAccountModelRestrictions } from "@/components/account/useEditAccountModelRestrictions";
import type { useEditAccountQuotaControls } from "@/components/account/useEditAccountQuotaControls";
import type { useEditAccountQuotaLimits } from "@/components/account/useEditAccountQuotaLimits";
import type { useEditAccountRuntimeOptions } from "@/components/account/useEditAccountRuntimeOptions";
import type { useEditAccountTempUnschedRules } from "@/components/account/useEditAccountTempUnschedRules";
import type { useEditBedrockCredentials } from "@/components/account/useEditBedrockCredentials";
import type { useEditCredentialFields } from "@/components/account/useEditCredentialFields";
import type { useEditCustomErrorCodes } from "@/components/account/useEditCustomErrorCodes";

type ValueRef<T> = {
  value: T;
};

type EditAccountFormState = ReturnType<typeof useEditAccountFormState>;
type EditAccountModelRestrictions = ReturnType<
  typeof useEditAccountModelRestrictions
>;
type EditAccountQuotaControls = ReturnType<typeof useEditAccountQuotaControls>;
type EditAccountQuotaLimits = ReturnType<typeof useEditAccountQuotaLimits>;
type EditAccountRuntimeOptions = ReturnType<typeof useEditAccountRuntimeOptions>;
type EditAccountTempUnschedRules = ReturnType<
  typeof useEditAccountTempUnschedRules
>;
type EditBedrockCredentials = ReturnType<typeof useEditBedrockCredentials>;
type EditCredentialFields = ReturnType<typeof useEditCredentialFields>;
type EditCustomErrorCodes = ReturnType<typeof useEditCustomErrorCodes>;

interface UseEditAccountMutationPayloadOptions {
  bedrockCredentials: EditBedrockCredentials;
  credentialFields: EditCredentialFields;
  customErrorCodes: EditCustomErrorCodes;
  defaultBaseUrl: ValueRef<string>;
  formState: Pick<EditAccountFormState, "allowOverages" | "mixedScheduling">;
  interceptWarmupRequests: ValueRef<boolean>;
  modelRestrictions: Pick<
    EditAccountModelRestrictions,
    | "allowedModels"
    | "antigravityModelMappings"
    | "modelMappings"
    | "modelRestrictionMode"
  >;
  quotaControls: EditAccountQuotaControls;
  quotaLimits: EditAccountQuotaLimits;
  runtimeOptions: Pick<
    EditAccountRuntimeOptions,
    | "anthropicPassthroughEnabled"
    | "codexCLIOnlyEnabled"
    | "isOpenAIModelRestrictionDisabled"
    | "openaiAPIKeyResponsesWebSocketV2Mode"
    | "openaiOAuthResponsesWebSocketV2Mode"
    | "openaiPassthroughEnabled"
  >;
  tempUnschedRules: Pick<
    EditAccountTempUnschedRules,
    "tempUnschedEnabled" | "tempUnschedRules"
  >;
}

interface BuildCurrentEditAccountMutationPayloadOptions {
  account: Account;
  basePayload: UpdateAccountRequest;
}

export function useEditAccountMutationPayload(
  options: UseEditAccountMutationPayloadOptions,
) {
  const buildEditMutationPayload = ({
    account,
    basePayload,
  }: BuildCurrentEditAccountMutationPayloadOptions): AccountMutationPayloadBuildResult<UpdateAccountRequest> =>
    buildEditAccountMutationPayload({
      account,
      anthropicAPIKeyExtra: {
        anthropicPassthroughEnabled:
          options.runtimeOptions.anthropicPassthroughEnabled.value,
      },
      anthropicQuotaExtra: {
        baseRpm: options.quotaControls.baseRpm.value,
        cacheTTLOverrideEnabled:
          options.quotaControls.cacheTTLOverrideEnabled.value,
        cacheTTLOverrideTarget:
          options.quotaControls.cacheTTLOverrideTarget.value,
        customBaseUrl: options.quotaControls.customBaseUrl.value,
        customBaseUrlEnabled:
          options.quotaControls.customBaseUrlEnabled.value,
        maxSessions: options.quotaControls.maxSessions.value,
        rpmLimitEnabled: options.quotaControls.rpmLimitEnabled.value,
        rpmStickyBuffer: options.quotaControls.rpmStickyBuffer.value,
        rpmStrategy: options.quotaControls.rpmStrategy.value,
        sessionIdMaskingEnabled:
          options.quotaControls.sessionIdMaskingEnabled.value,
        sessionIdleTimeout: options.quotaControls.sessionIdleTimeout.value,
        sessionLimitEnabled: options.quotaControls.sessionLimitEnabled.value,
        tlsFingerprintEnabled:
          options.quotaControls.tlsFingerprintEnabled.value,
        tlsFingerprintProfileId:
          options.quotaControls.tlsFingerprintProfileId.value,
        userMsgQueueMode: options.quotaControls.userMsgQueueMode.value,
        windowCostEnabled: options.quotaControls.windowCostEnabled.value,
        windowCostLimit: options.quotaControls.windowCostLimit.value,
        windowCostStickyReserve:
          options.quotaControls.windowCostStickyReserve.value,
      },
      antigravity: {
        allowOverages: options.formState.allowOverages.value,
        mixedScheduling: options.formState.mixedScheduling.value,
        modelMappings: options.modelRestrictions.antigravityModelMappings.value,
      },
      basePayload,
      bedrock: {
        accessKeyId: options.bedrockCredentials.editBedrockAccessKeyId.value,
        allowedModels: options.modelRestrictions.allowedModels.value,
        apiKeyInput: options.bedrockCredentials.editBedrockApiKeyValue.value,
        forceGlobal: options.bedrockCredentials.editBedrockForceGlobal.value,
        isApiKeyMode: options.bedrockCredentials.isBedrockAPIKeyMode.value,
        mode: options.modelRestrictions.modelRestrictionMode.value,
        modelMappings: options.modelRestrictions.modelMappings.value,
        poolModeEnabled: options.credentialFields.poolModeEnabled.value,
        poolModeRetryCount: options.credentialFields.poolModeRetryCount.value,
        region: options.bedrockCredentials.editBedrockRegion.value,
        secretAccessKey:
          options.bedrockCredentials.editBedrockSecretAccessKey.value,
        sessionToken: options.bedrockCredentials.editBedrockSessionToken.value,
      },
      compatible: {
        allowedModels: options.modelRestrictions.allowedModels.value,
        apiKeyInput: options.credentialFields.editApiKey.value,
        baseUrlInput: options.credentialFields.editBaseUrl.value,
        customErrorCodesEnabled:
          options.customErrorCodes.customErrorCodesEnabled.value,
        defaultBaseUrl: options.defaultBaseUrl.value,
        isOpenAIModelRestrictionDisabled:
          options.runtimeOptions.isOpenAIModelRestrictionDisabled.value,
        mode: options.modelRestrictions.modelRestrictionMode.value,
        modelMappings: options.modelRestrictions.modelMappings.value,
        poolModeEnabled: options.credentialFields.poolModeEnabled.value,
        poolModeRetryCount: options.credentialFields.poolModeRetryCount.value,
        selectedErrorCodes: options.customErrorCodes.selectedErrorCodes.value,
      },
      openAIExtra: {
        accountType: account.type,
        codexCLIOnlyEnabled: options.runtimeOptions.codexCLIOnlyEnabled.value,
        openaiAPIKeyResponsesWebSocketV2Mode:
          options.runtimeOptions.openaiAPIKeyResponsesWebSocketV2Mode.value,
        openaiOAuthResponsesWebSocketV2Mode:
          options.runtimeOptions.openaiOAuthResponsesWebSocketV2Mode.value,
        openaiPassthroughEnabled:
          options.runtimeOptions.openaiPassthroughEnabled.value,
      },
      quota: {
        dailyResetHour: options.quotaLimits.editDailyResetHour.value,
        dailyResetMode: options.quotaLimits.editDailyResetMode.value,
        quotaDailyLimit: options.quotaLimits.editQuotaDailyLimit.value,
        quotaLimit: options.quotaLimits.editQuotaLimit.value,
        quotaWeeklyLimit: options.quotaLimits.editQuotaWeeklyLimit.value,
        quotaNotifyDailyEnabled:
          options.quotaLimits.editQuotaNotifyDailyEnabled.value,
        quotaNotifyDailyThreshold:
          options.quotaLimits.editQuotaNotifyDailyThreshold.value,
        quotaNotifyDailyThresholdType:
          options.quotaLimits.editQuotaNotifyDailyThresholdType.value,
        quotaNotifyWeeklyEnabled:
          options.quotaLimits.editQuotaNotifyWeeklyEnabled.value,
        quotaNotifyWeeklyThreshold:
          options.quotaLimits.editQuotaNotifyWeeklyThreshold.value,
        quotaNotifyWeeklyThresholdType:
          options.quotaLimits.editQuotaNotifyWeeklyThresholdType.value,
        quotaNotifyTotalEnabled:
          options.quotaLimits.editQuotaNotifyTotalEnabled.value,
        quotaNotifyTotalThreshold:
          options.quotaLimits.editQuotaNotifyTotalThreshold.value,
        quotaNotifyTotalThresholdType:
          options.quotaLimits.editQuotaNotifyTotalThresholdType.value,
        resetTimezone: options.quotaLimits.editResetTimezone.value,
        weeklyResetDay: options.quotaLimits.editWeeklyResetDay.value,
        weeklyResetHour: options.quotaLimits.editWeeklyResetHour.value,
        weeklyResetMode: options.quotaLimits.editWeeklyResetMode.value,
      },
      sessionTokenInput: options.credentialFields.editSessionToken.value,
      sharedCredentials: {
        interceptWarmupRequests: options.interceptWarmupRequests.value,
        tempUnschedEnabled: options.tempUnschedRules.tempUnschedEnabled.value,
        tempUnschedRules: options.tempUnschedRules.tempUnschedRules.value,
      },
    });

  return {
    buildEditMutationPayload,
  };
}
