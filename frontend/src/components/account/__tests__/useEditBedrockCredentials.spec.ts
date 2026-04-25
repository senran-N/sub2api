import { describe, expect, it } from "vitest";
import { ref } from "vue";
import { useEditBedrockCredentials } from "../useEditBedrockCredentials";
import type { Account } from "@/types";

function buildAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: "Bedrock",
    notes: "",
    platform: "anthropic",
    type: "bedrock",
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

describe("useEditBedrockCredentials", () => {
  it("hydrates SigV4 editable fields while keeping secret inputs empty", () => {
    const account = ref(
      buildAccount({
        credentials: {
          auth_mode: "sigv4",
          aws_access_key_id: "AKIA_TEST",
          aws_secret_access_key: "secret",
          aws_session_token: "session",
          aws_region: "us-east-1",
          aws_force_global: "true",
        },
      }),
    );
    const credentials = useEditBedrockCredentials(() => account.value);

    credentials.hydrateBedrockCredentialsFromAccount(account.value);

    expect(credentials.editBedrockAuthMode.value).toBe("sigv4");
    expect(credentials.isBedrockAPIKeyMode.value).toBe(false);
    expect(credentials.editBedrockAccessKeyId.value).toBe("AKIA_TEST");
    expect(credentials.editBedrockSecretAccessKey.value).toBe("");
    expect(credentials.editBedrockSessionToken.value).toBe("");
    expect(credentials.editBedrockRegion.value).toBe("us-east-1");
    expect(credentials.editBedrockForceGlobal.value).toBe(true);
  });

  it("hydrates API-key mode without exposing the saved API key", () => {
    const account = ref(
      buildAccount({
        credentials: {
          auth_mode: "apikey",
          api_key: "saved-key",
          aws_region: "us-west-2",
        },
      }),
    );
    const credentials = useEditBedrockCredentials(() => account.value);

    credentials.hydrateBedrockCredentialsFromAccount(account.value);

    expect(credentials.editBedrockAuthMode.value).toBe("apikey");
    expect(credentials.isBedrockAPIKeyMode.value).toBe(true);
    expect(credentials.editBedrockApiKeyValue.value).toBe("");
    expect(credentials.editBedrockAccessKeyId.value).toBe("");
    expect(credentials.editBedrockRegion.value).toBe("us-west-2");
  });

  it("resets Bedrock inputs for unsupported accounts", () => {
    const account = ref(buildAccount());
    const credentials = useEditBedrockCredentials(() => account.value);

    credentials.hydrateBedrockCredentialsFromAccount(
      buildAccount({
        credentials: {
          aws_access_key_id: "AKIA_TEST",
          aws_region: "us-east-1",
        },
      }),
    );
    account.value = buildAccount({ platform: "openai", type: "apikey" });
    credentials.hydrateBedrockCredentialsFromAccount(account.value);

    expect(credentials.editBedrockAccessKeyId.value).toBe("");
    expect(credentials.editBedrockRegion.value).toBe("");
    expect(credentials.editBedrockForceGlobal.value).toBe(false);
    expect(credentials.editBedrockAuthMode.value).toBe("sigv4");
  });
});
