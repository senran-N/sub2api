import { describe, expect, it } from "vitest";
import { ref } from "vue";
import { useEditAccountRuntimeOptions } from "../useEditAccountRuntimeOptions";
import type { Account } from "@/types";

function buildAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 1,
    name: "OpenAI",
    notes: "",
    platform: "openai",
    type: "oauth",
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

describe("useEditAccountRuntimeOptions", () => {
  it("hydrates OpenAI OAuth runtime options", () => {
    const currentAccount = ref(buildAccount());
    const runtime = useEditAccountRuntimeOptions(() => currentAccount.value);

    runtime.hydrateRuntimeOptionsFromAccount(
      buildAccount({
        extra: {
          openai_oauth_passthrough: true,
          openai_oauth_responses_websockets_v2_mode: "passthrough",
          openai_apikey_responses_websockets_v2_mode: "ctx_pool",
          codex_cli_only: true,
        },
      }),
    );

    expect(runtime.openaiPassthroughEnabled.value).toBe(true);
    expect(runtime.openaiOAuthResponsesWebSocketV2Mode.value).toBe(
      "passthrough",
    );
    expect(runtime.openaiAPIKeyResponsesWebSocketV2Mode.value).toBe(
      "ctx_pool",
    );
    expect(runtime.openaiResponsesWebSocketV2Mode.value).toBe("passthrough");
    expect(runtime.codexCLIOnlyEnabled.value).toBe(true);
    expect(runtime.isOpenAIModelRestrictionDisabled.value).toBe(true);
    expect(runtime.openAIWSModeConcurrencyHintKey.value).toBe(
      "admin.accounts.openai.wsModePassthroughHint",
    );

    currentAccount.value = buildAccount({ type: "apikey" });
    expect(runtime.openaiResponsesWebSocketV2Mode.value).toBe("ctx_pool");
  });

  it("routes the editable WebSocket mode to the current OpenAI account type", () => {
    const currentAccount = ref(buildAccount());
    const runtime = useEditAccountRuntimeOptions(() => currentAccount.value);

    runtime.openaiResponsesWebSocketV2Mode.value = "passthrough";
    expect(runtime.openaiOAuthResponsesWebSocketV2Mode.value).toBe(
      "passthrough",
    );
    expect(runtime.openaiAPIKeyResponsesWebSocketV2Mode.value).toBe("off");

    currentAccount.value = buildAccount({ type: "apikey" });
    runtime.openaiResponsesWebSocketV2Mode.value = "ctx_pool";

    expect(runtime.openaiOAuthResponsesWebSocketV2Mode.value).toBe(
      "passthrough",
    );
    expect(runtime.openaiAPIKeyResponsesWebSocketV2Mode.value).toBe(
      "ctx_pool",
    );
  });

  it("hydrates Anthropic API-key passthrough and resets unsupported accounts", () => {
    const runtime = useEditAccountRuntimeOptions(() => null);

    runtime.hydrateRuntimeOptionsFromAccount(
      buildAccount({
        platform: "anthropic",
        type: "apikey",
        extra: { anthropic_passthrough: true },
      }),
    );

    expect(runtime.anthropicPassthroughEnabled.value).toBe(true);

    runtime.hydrateRuntimeOptionsFromAccount(
      buildAccount({ platform: "grok", type: "session" }),
    );

    expect(runtime.anthropicPassthroughEnabled.value).toBe(false);
    expect(runtime.openaiPassthroughEnabled.value).toBe(false);
    expect(runtime.openaiResponsesWebSocketV2Mode.value).toBe("off");
    expect(runtime.codexCLIOnlyEnabled.value).toBe(false);
  });
});
