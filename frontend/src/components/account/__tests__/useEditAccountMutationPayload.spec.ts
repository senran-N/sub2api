import { describe, expect, it, vi } from "vitest";
import { ref } from "vue";
import type { Account } from "@/types";
import { useEditAccountFormState } from "../useEditAccountFormState";
import { useEditAccountModelRestrictions } from "../useEditAccountModelRestrictions";
import { useEditAccountMutationPayload } from "../useEditAccountMutationPayload";
import { useEditAccountQuotaControls } from "../useEditAccountQuotaControls";
import { useEditAccountQuotaLimits } from "../useEditAccountQuotaLimits";
import { useEditAccountRuntimeOptions } from "../useEditAccountRuntimeOptions";
import { useEditAccountTempUnschedRules } from "../useEditAccountTempUnschedRules";
import { useEditBedrockCredentials } from "../useEditBedrockCredentials";
import { useEditCredentialFields } from "../useEditCredentialFields";
import { useEditCustomErrorCodes } from "../useEditCustomErrorCodes";

function buildAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: "Account",
    notes: null,
    platform: "openai",
    type: "apikey",
    credentials: {
      api_key: "existing-key",
    },
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

function createPayloadBuilder(account: Account) {
  const accountRef = ref(account);
  const bedrockCredentials = useEditBedrockCredentials(() => accountRef.value);
  const credentialFields = useEditCredentialFields();
  const customErrorCodes = useEditCustomErrorCodes({
    confirmSelection: vi.fn(() => true),
    showDuplicate: vi.fn(),
    showInvalid: vi.fn(),
  });
  const defaultBaseUrl = ref("https://api.openai.com");
  const formState = useEditAccountFormState((key) => key);
  const interceptWarmupRequests = ref(false);
  const modelRestrictions = useEditAccountModelRestrictions();
  const quotaControls = useEditAccountQuotaControls();
  const quotaLimits = useEditAccountQuotaLimits();
  const runtimeOptions = useEditAccountRuntimeOptions(() => accountRef.value);
  const tempUnschedRules = useEditAccountTempUnschedRules();
  const payload = useEditAccountMutationPayload({
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
    tempUnschedRules,
  });

  return {
    ...payload,
    accountRef,
    credentialFields,
    customErrorCodes,
    formState,
    interceptWarmupRequests,
    modelRestrictions,
    quotaLimits,
    runtimeOptions,
    tempUnschedRules,
  };
}

describe("useEditAccountMutationPayload", () => {
  it("collects compatible credentials, toggles, and quota limits", () => {
    const account = buildAccount();
    const builder = createPayloadBuilder(account);

    builder.credentialFields.editApiKey.value = "new-key";
    builder.credentialFields.editBaseUrl.value = "https://proxy.example";
    builder.credentialFields.poolModeEnabled.value = true;
    builder.credentialFields.poolModeRetryCount.value = 4;
    builder.customErrorCodes.customErrorCodesEnabled.value = true;
    builder.customErrorCodes.selectedErrorCodes.value = [429, 529];
    builder.interceptWarmupRequests.value = true;
    builder.modelRestrictions.allowedModels.value = ["gpt-5.4"];
    builder.quotaLimits.editQuotaLimit.value = 100;

    const result = builder.buildEditMutationPayload({
      account,
      basePayload: {
        name: "Updated Account",
        status: "active",
      },
    });

    expect(result.error).toBeUndefined();
    expect(result.payload?.credentials).toMatchObject({
      api_key: "new-key",
      base_url: "https://proxy.example",
      custom_error_codes_enabled: true,
      custom_error_codes: [429, 529],
      intercept_warmup_requests: true,
      model_mapping: {
        "gpt-5.4": "gpt-5.4",
      },
      pool_mode: true,
      pool_mode_retry_count: 4,
    });
    expect(result.payload?.extra).toMatchObject({
      quota_limit: 100,
    });
  });

  it("collects Antigravity scheduling flags and model mappings", () => {
    const account = buildAccount({
      platform: "antigravity",
      type: "upstream",
      credentials: {
        api_key: "existing-key",
      },
    });
    const builder = createPayloadBuilder(account);

    builder.credentialFields.editBaseUrl.value = "https://ag.example";
    builder.formState.allowOverages.value = true;
    builder.formState.mixedScheduling.value = true;
    builder.modelRestrictions.antigravityModelMappings.value = [
      { from: "claude-sonnet-4-5", to: "auto" },
    ];

    const result = builder.buildEditMutationPayload({
      account,
      basePayload: {
        name: "Updated Account",
        status: "active",
      },
    });

    expect(result.error).toBeUndefined();
    expect(result.payload?.credentials).toMatchObject({
      base_url: "https://ag.example",
      model_mapping: {
        "claude-sonnet-4-5": "auto",
      },
    });
    expect(result.payload?.extra).toMatchObject({
      allow_overages: true,
      mixed_scheduling: true,
    });
  });
});
