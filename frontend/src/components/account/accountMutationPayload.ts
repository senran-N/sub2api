import type {
  Account,
  AccountPlatform,
  AccountType,
  CreateAccountRequest,
  UpdateAccountRequest,
} from "@/types";
import { buildModelMappingObject } from "@/composables/useModelWhitelist";
import { isOpenAIWSModeEnabled, type OpenAIWSMode } from "@/utils/openaiWsMode";
import { normalizeGrokSessionToken } from "@/utils/grokSessionToken";
import {
  accountMutationProfileHasSection,
  resolveAccountMutationProfile,
} from "./accountMutationProfiles";
import {
  buildAccountQuotaExtra,
  type AccountQuotaExtraOptions,
} from "./accountModalShared";
import {
  applyInterceptWarmup,
  applyTempUnschedConfig,
  type ModelMapping,
  type TempUnschedRuleForm,
} from "./credentialsBuilder";
import { buildCreateAccountRequest } from "./createAccountModalHelpers";
import {
  buildEditableBedrockCredentials,
  buildEditableCompatibleCredentials,
  buildUpdatedAnthropicAPIKeyExtra,
  buildUpdatedAnthropicQuotaControlExtra,
  buildUpdatedAntigravityExtra,
  buildUpdatedOpenAIExtra,
  type AnthropicAPIKeyExtraOptions,
  type AnthropicQuotaControlExtraOptions,
  type AntigravityExtraOptions,
} from "./editAccountModalHelpers";

export type AccountMutationPayloadBuildError =
  | "api_key_required"
  | "grok_session_token_required"
  | "grok_session_token_invalid"
  | "temp_unsched_rules_invalid";

export interface SharedCredentialMutationOptions {
  interceptWarmupRequests: boolean;
  tempUnschedEnabled: boolean;
  tempUnschedRules: TempUnschedRuleForm[];
}

export interface ModelRestrictionMutationOptions {
  allowedModels: string[];
  mode: "whitelist" | "mapping";
  modelMappings: ModelMapping[];
}

export interface CredentialToggleMutationOptions {
  customErrorCodesEnabled: boolean;
  poolModeEnabled: boolean;
  poolModeRetryCount: number;
  selectedErrorCodes: number[];
}

export interface OpenAIAccountRuntimeMutationOptions {
  accountType: AccountType;
  codexCLIOnlyEnabled: boolean;
  openaiAPIKeyResponsesWebSocketV2Mode: OpenAIWSMode;
  openaiOAuthResponsesWebSocketV2Mode: OpenAIWSMode;
  openaiPassthroughEnabled: boolean;
}

export interface BuildCreateAccountMutationPayloadOptions {
  common: Omit<
    CreateAccountRequest,
    "credentials" | "extra" | "name" | "platform" | "type"
  >;
  credentials: Record<string, unknown>;
  extra?: Record<string, unknown>;
  name: string;
  platform: AccountPlatform;
  quota?: AccountQuotaExtraOptions;
  type: AccountType;
}

export interface BuildEditAccountMutationPayloadOptions {
  account: Account;
  anthropicAPIKeyExtra: AnthropicAPIKeyExtraOptions;
  anthropicQuotaExtra: AnthropicQuotaControlExtraOptions;
  antigravity: AntigravityExtraOptions & {
    modelMappings: ModelMapping[];
  };
  basePayload: UpdateAccountRequest;
  bedrock: ModelRestrictionMutationOptions & {
    accessKeyId: string;
    apiKeyInput: string;
    forceGlobal: boolean;
    isApiKeyMode: boolean;
    poolModeEnabled: boolean;
    poolModeRetryCount: number;
    region: string;
    secretAccessKey: string;
    sessionToken: string;
  };
  compatible: ModelRestrictionMutationOptions &
    CredentialToggleMutationOptions & {
      apiKeyInput: string;
      baseUrlInput: string;
      defaultBaseUrl: string;
      isOpenAIModelRestrictionDisabled: boolean;
    };
  currentCredentials?: Record<string, unknown>;
  currentExtra?: Record<string, unknown>;
  openAIExtra: OpenAIAccountRuntimeMutationOptions;
  quota: AccountQuotaExtraOptions;
  sessionTokenInput: string;
  sharedCredentials: SharedCredentialMutationOptions;
}

export interface BuildBulkAccountMutationPayloadOptions {
  baseUrl: {
    enabled: boolean;
    value: string;
  };
  customErrorCodes: {
    enabled: boolean;
    selectedErrorCodes: number[];
  };
  groups: {
    enabled: boolean;
    groupIds: number[];
  };
  interceptWarmup: {
    enabled: boolean;
    value: boolean;
  };
  loadFactor: {
    enabled: boolean;
    value: number | null;
  };
  modelRestriction: ModelRestrictionMutationOptions & {
    disabledByOpenAIPassthrough: boolean;
    enabled: boolean;
  };
  openAI: {
    passthroughEnabled: boolean;
    passthroughValue: boolean;
    wsModeEnabled: boolean;
    wsModeValue: OpenAIWSMode;
  };
  proxy: {
    enabled: boolean;
    proxyId: number | null;
  };
  rpmLimit: {
    baseRpm: number | null;
    enabled: boolean;
    rpmEnabled: boolean;
    stickyBuffer: number | null;
    strategy: "tiered" | "sticky_exempt";
  };
  scalars: {
    concurrency?: number;
    enableConcurrency: boolean;
    enablePriority: boolean;
    enableRateMultiplier: boolean;
    enableStatus: boolean;
    priority?: number;
    rateMultiplier?: number;
    status?: "active" | "inactive";
  };
  userMsgQueueMode: string | null;
}

export interface AccountMutationPayloadBuildResult<TPayload> {
  error?: AccountMutationPayloadBuildError;
  payload?: TPayload;
}

export function resolveAccountMutationPayloadErrorKey(
  error: AccountMutationPayloadBuildError,
) {
  switch (error) {
    case "api_key_required":
      return "admin.accounts.apiKeyIsRequired";
    case "grok_session_token_required":
      return "admin.accounts.grok.sessionTokenRequired";
    case "grok_session_token_invalid":
      return "admin.accounts.grok.sessionTokenInvalidFormat";
    case "temp_unsched_rules_invalid":
      return "admin.accounts.tempUnschedulable.rulesInvalid";
  }
}

export function buildCreateAccountMutationPayload(
  options: BuildCreateAccountMutationPayloadOptions,
): CreateAccountRequest {
  const profile = resolveAccountMutationProfile(options.platform, options.type);
  const extra =
    options.quota && accountMutationProfileHasSection(profile, "quota-limits")
      ? emptyToUndefined(buildAccountQuotaExtra(options.extra, options.quota))
      : options.extra;

  return buildCreateAccountRequest({
    common: options.common,
    name: options.name,
    platform: options.platform,
    type: options.type,
    credentials: options.credentials,
    extra,
  });
}

export function buildEditAccountMutationPayload(
  options: BuildEditAccountMutationPayloadOptions,
): AccountMutationPayloadBuildResult<UpdateAccountRequest> {
  const account = options.account;
  const updatePayload: UpdateAccountRequest = { ...options.basePayload };
  const currentCredentials = options.currentCredentials || getAccountCredentials(account);
  const currentExtra = options.currentExtra || getAccountExtra(account);

  const credentialsResult = buildEditCredentialsMutation({
    account,
    compatible: options.compatible,
    currentCredentials,
    sharedCredentials: options.sharedCredentials,
    bedrock: options.bedrock,
    sessionTokenInput: options.sessionTokenInput,
  });
  if (credentialsResult.error) {
    return { error: credentialsResult.error };
  }
  if (credentialsResult.payload) {
    updatePayload.credentials = credentialsResult.payload;
  }

  applyEditCredentialOverlays(updatePayload, {
    account,
    antigravityModelMappings: options.antigravity.modelMappings,
    compatible: options.compatible,
    currentCredentials,
  });

  applyEditExtraOverlays(updatePayload, {
    account,
    anthropicAPIKeyExtra: options.anthropicAPIKeyExtra,
    anthropicQuotaExtra: options.anthropicQuotaExtra,
    antigravity: options.antigravity,
    currentExtra,
    openAIExtra: options.openAIExtra,
    quota: options.quota,
  });

  return { payload: updatePayload };
}

export function buildBulkAccountMutationPayload(
  options: BuildBulkAccountMutationPayloadOptions,
): Record<string, unknown> | null {
  const updates: Record<string, unknown> = {};
  const credentials: Record<string, unknown> = {};
  let credentialsChanged = false;
  const ensureExtra = (): Record<string, unknown> => {
    if (!updates.extra) {
      updates.extra = {};
    }
    return updates.extra as Record<string, unknown>;
  };

  if (options.proxy.enabled) {
    updates.proxy_id = options.proxy.proxyId === null ? 0 : options.proxy.proxyId;
  }
  if (options.scalars.enableConcurrency) {
    updates.concurrency = options.scalars.concurrency;
  }
  if (options.loadFactor.enabled) {
    const loadFactor = options.loadFactor.value;
    updates.load_factor =
      loadFactor != null && !Number.isNaN(loadFactor) && loadFactor > 0
        ? loadFactor
        : 0;
  }
  if (options.scalars.enablePriority) {
    updates.priority = options.scalars.priority;
  }
  if (options.scalars.enableRateMultiplier) {
    updates.rate_multiplier = options.scalars.rateMultiplier;
  }
  if (options.scalars.enableStatus) {
    updates.status = options.scalars.status;
  }
  if (options.groups.enabled) {
    updates.group_ids = options.groups.groupIds;
  }
  if (options.baseUrl.enabled) {
    const baseUrl = options.baseUrl.value.trim();
    if (baseUrl) {
      credentials.base_url = baseUrl;
      credentialsChanged = true;
    }
  }
  if (
    options.modelRestriction.enabled &&
    !options.modelRestriction.disabledByOpenAIPassthrough
  ) {
    applyBulkModelRestriction(credentials, options.modelRestriction);
    credentialsChanged = true;
  }
  if (options.customErrorCodes.enabled) {
    credentials.custom_error_codes_enabled = true;
    credentials.custom_error_codes = [...options.customErrorCodes.selectedErrorCodes];
    credentialsChanged = true;
  }
  if (options.interceptWarmup.enabled) {
    credentials.intercept_warmup_requests = options.interceptWarmup.value;
    credentialsChanged = true;
  }
  if (credentialsChanged) {
    updates.credentials = credentials;
  }

  if (options.openAI.passthroughEnabled || options.openAI.wsModeEnabled) {
    applyBulkOpenAIExtra(ensureExtra(), options.openAI);
  }
  if (options.rpmLimit.enabled) {
    applyBulkRpmExtra(ensureExtra(), options.rpmLimit);
  }
  if (options.userMsgQueueMode !== null) {
    const extra = ensureExtra();
    extra.user_msg_queue_mode = options.userMsgQueueMode;
    extra.user_msg_queue_enabled = false;
  }

  return Object.keys(updates).length > 0 ? updates : null;
}

function buildEditCredentialsMutation(options: {
  account: Account;
  bedrock: BuildEditAccountMutationPayloadOptions["bedrock"];
  compatible: BuildEditAccountMutationPayloadOptions["compatible"];
  currentCredentials: Record<string, unknown>;
  sessionTokenInput: string;
  sharedCredentials: SharedCredentialMutationOptions;
}): AccountMutationPayloadBuildResult<Record<string, unknown>> {
  let credentials: Record<string, unknown>;

  if (options.account.type === "apikey" || options.account.type === "upstream") {
    const shouldApplyModelMapping =
      options.account.type === "upstream" ||
      !(
        options.account.platform === "openai" &&
        options.compatible.isOpenAIModelRestrictionDisabled
      );
    const result = buildEditableCompatibleCredentials({
      allowedModels: options.compatible.allowedModels,
      apiKeyInput: options.compatible.apiKeyInput,
      baseUrlInput: options.compatible.baseUrlInput,
      currentCredentials: options.currentCredentials,
      customErrorCodesEnabled: options.compatible.customErrorCodesEnabled,
      defaultBaseUrl: options.compatible.defaultBaseUrl,
      mode: options.compatible.mode,
      modelMappings: options.compatible.modelMappings,
      poolModeEnabled: options.compatible.poolModeEnabled,
      poolModeRetryCount: options.compatible.poolModeRetryCount,
      preserveModelMappingWhenDisabled: options.account.type === "apikey",
      selectedErrorCodes: options.compatible.selectedErrorCodes,
      shouldApplyModelMapping,
    });
    if (result.error === "api_key_required" || !result.credentials) {
      return { error: "api_key_required" };
    }
    credentials = result.credentials;
  } else if (options.account.type === "session") {
    const result = buildGrokSessionCredentialsMutation(
      options.currentCredentials,
      options.sessionTokenInput,
    );
    if (result.error || !result.payload) {
      return { error: result.error || "grok_session_token_required" };
    }
    credentials = result.payload;
  } else if (options.account.type === "bedrock") {
    credentials = buildEditableBedrockCredentials({
      accessKeyId: options.bedrock.accessKeyId,
      allowedModels: options.bedrock.allowedModels,
      apiKeyInput: options.bedrock.apiKeyInput,
      currentCredentials: options.currentCredentials,
      forceGlobal: options.bedrock.forceGlobal,
      isApiKeyMode: options.bedrock.isApiKeyMode,
      mode: options.bedrock.mode,
      modelMappings: options.bedrock.modelMappings,
      poolModeEnabled: options.bedrock.poolModeEnabled,
      poolModeRetryCount: options.bedrock.poolModeRetryCount,
      region: options.bedrock.region,
      secretAccessKey: options.bedrock.secretAccessKey,
      sessionToken: options.bedrock.sessionToken,
    });
  } else {
    credentials = { ...options.currentCredentials };
  }

  const sharedError = applySharedCredentialMutation(credentials, {
    ...options.sharedCredentials,
    mode: "edit",
  });
  if (sharedError) {
    return { error: sharedError };
  }

  return { payload: credentials };
}

function buildGrokSessionCredentialsMutation(
  currentCredentials: Record<string, unknown>,
  sessionTokenInput: string,
): AccountMutationPayloadBuildResult<Record<string, unknown>> {
  const credentials: Record<string, unknown> = { ...currentCredentials };
  const sessionToken = sessionTokenInput.trim();
  if (sessionToken) {
    const normalizedSessionToken = normalizeGrokSessionToken(sessionToken);
    if (!normalizedSessionToken) {
      return { error: "grok_session_token_invalid" };
    }
    credentials.session_token = normalizedSessionToken;
    return { payload: credentials };
  }
  if (currentCredentials.session_token) {
    credentials.session_token = currentCredentials.session_token;
    return { payload: credentials };
  }
  return { error: "grok_session_token_required" };
}

function applyEditCredentialOverlays(
  updatePayload: UpdateAccountRequest,
  options: {
    account: Account;
    antigravityModelMappings: ModelMapping[];
    compatible: BuildEditAccountMutationPayloadOptions["compatible"];
    currentCredentials: Record<string, unknown>;
  },
) {
  if (options.account.platform === "openai" && options.account.type === "oauth") {
    const credentials = {
      ...(updatePayload.credentials || options.currentCredentials),
    };
    if (!options.compatible.isOpenAIModelRestrictionDisabled) {
      applyModelMappingReplacement(credentials, options.compatible);
    } else if (options.currentCredentials.model_mapping) {
      credentials.model_mapping = options.currentCredentials.model_mapping;
    }
    updatePayload.credentials = credentials;
  }

  if (options.account.platform === "antigravity") {
    const credentials = {
      ...(updatePayload.credentials || options.currentCredentials),
    };
    applyAntigravityModelMappingReplacement(
      credentials,
      options.antigravityModelMappings,
    );
    updatePayload.credentials = credentials;
  }
}

function applyEditExtraOverlays(
  updatePayload: UpdateAccountRequest,
  options: {
    account: Account;
    anthropicAPIKeyExtra: AnthropicAPIKeyExtraOptions;
    anthropicQuotaExtra: AnthropicQuotaControlExtraOptions;
    antigravity: AntigravityExtraOptions;
    currentExtra: Record<string, unknown>;
    openAIExtra: OpenAIAccountRuntimeMutationOptions;
    quota: AccountQuotaExtraOptions;
  },
) {
  const account = options.account;
  if (account.platform === "antigravity") {
    updatePayload.extra = buildUpdatedAntigravityExtra(
      options.currentExtra,
      options.antigravity,
    );
  }
  if (
    account.platform === "anthropic" &&
    (account.type === "oauth" || account.type === "setup-token")
  ) {
    updatePayload.extra = buildUpdatedAnthropicQuotaControlExtra(
      options.currentExtra,
      options.anthropicQuotaExtra,
    );
  }
  if (account.platform === "anthropic" && account.type === "apikey") {
    updatePayload.extra = buildUpdatedAnthropicAPIKeyExtra(
      options.currentExtra,
      options.anthropicAPIKeyExtra,
    );
  }
  if (
    account.platform === "openai" &&
    (account.type === "oauth" || account.type === "apikey")
  ) {
    updatePayload.extra = buildUpdatedOpenAIExtra(
      options.currentExtra,
      options.openAIExtra,
    );
  }

  const profile = resolveAccountMutationProfile(account.platform, account.type);
  if (accountMutationProfileHasSection(profile, "quota-limits")) {
    updatePayload.extra = buildAccountQuotaExtra(
      (updatePayload.extra as Record<string, unknown>) || options.currentExtra,
      options.quota,
    );
  }
}

function applySharedCredentialMutation(
  credentials: Record<string, unknown>,
  options: SharedCredentialMutationOptions & { mode: "create" | "edit" },
): AccountMutationPayloadBuildError | undefined {
  applyInterceptWarmup(
    credentials,
    options.interceptWarmupRequests,
    options.mode,
  );
  if (
    !applyTempUnschedConfig(
      credentials,
      options.tempUnschedEnabled,
      options.tempUnschedRules,
    )
  ) {
    return "temp_unsched_rules_invalid";
  }
  return undefined;
}

function applyModelMappingReplacement(
  credentials: Record<string, unknown>,
  options: ModelRestrictionMutationOptions,
) {
  const modelMapping = buildModelMappingObject(
    options.mode,
    options.allowedModels,
    options.modelMappings,
  );
  if (modelMapping) {
    credentials.model_mapping = modelMapping;
  } else {
    delete credentials.model_mapping;
  }
}

function applyAntigravityModelMappingReplacement(
  credentials: Record<string, unknown>,
  modelMappings: ModelMapping[],
) {
  delete credentials.model_whitelist;
  applyModelMappingReplacement(credentials, {
    allowedModels: [],
    mode: "mapping",
    modelMappings,
  });
}

function applyBulkModelRestriction(
  credentials: Record<string, unknown>,
  options: ModelRestrictionMutationOptions,
) {
  if (options.mode === "whitelist") {
    const modelMapping: Record<string, string> = {};
    for (const model of options.allowedModels) {
      modelMapping[model] = model;
    }
    credentials.model_mapping = modelMapping;
    return;
  }
  credentials.model_mapping =
    buildModelMappingObject(
      options.mode,
      options.allowedModels,
      options.modelMappings,
    ) || {};
}

function applyBulkOpenAIExtra(
  extra: Record<string, unknown>,
  options: BuildBulkAccountMutationPayloadOptions["openAI"],
) {
  if (options.passthroughEnabled) {
    extra.openai_passthrough = options.passthroughValue;
    if (!options.passthroughValue) {
      extra.openai_oauth_passthrough = false;
    }
  }
  if (options.wsModeEnabled) {
    extra.openai_oauth_responses_websockets_v2_mode = options.wsModeValue;
    extra.openai_oauth_responses_websockets_v2_enabled =
      isOpenAIWSModeEnabled(options.wsModeValue);
    extra.responses_websockets_v2_enabled = false;
    extra.openai_ws_enabled = false;
  }
}

function applyBulkRpmExtra(
  extra: Record<string, unknown>,
  options: BuildBulkAccountMutationPayloadOptions["rpmLimit"],
) {
  if (options.rpmEnabled && options.baseRpm != null && options.baseRpm > 0) {
    extra.base_rpm = options.baseRpm;
    extra.rpm_strategy = options.strategy;
    if (options.stickyBuffer != null && options.stickyBuffer > 0) {
      extra.rpm_sticky_buffer = options.stickyBuffer;
    }
    return;
  }
  extra.base_rpm = 0;
  extra.rpm_strategy = "";
  extra.rpm_sticky_buffer = 0;
}

function getAccountCredentials(account: Account): Record<string, unknown> {
  return (account.credentials as Record<string, unknown>) || {};
}

function getAccountExtra(account: Account): Record<string, unknown> {
  return (account.extra as Record<string, unknown>) || {};
}

function emptyToUndefined<T extends Record<string, unknown>>(
  value: T,
): T | undefined {
  return Object.keys(value).length > 0 ? value : undefined;
}
