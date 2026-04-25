import { computed, ref } from "vue";
import type { Account } from "@/types";
import { deriveOpenAIExtraState } from "@/components/account/editAccountModalHelpers";
import {
  OPENAI_WS_MODE_OFF,
  resolveOpenAIWSModeConcurrencyHintKey,
  type OpenAIWSMode,
} from "@/utils/openaiWsMode";

type AccountGetter = () => Account | null | undefined;

const getAccountExtra = (account: Account): Record<string, unknown> =>
  (account.extra as Record<string, unknown>) || {};

export function useEditAccountRuntimeOptions(getAccount: AccountGetter) {
  const openaiPassthroughEnabled = ref(false);
  const openaiOAuthResponsesWebSocketV2Mode =
    ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
  const openaiAPIKeyResponsesWebSocketV2Mode =
    ref<OpenAIWSMode>(OPENAI_WS_MODE_OFF);
  const codexCLIOnlyEnabled = ref(false);
  const anthropicPassthroughEnabled = ref(false);

  const resetRuntimeOptions = () => {
    openaiPassthroughEnabled.value = false;
    openaiOAuthResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
    openaiAPIKeyResponsesWebSocketV2Mode.value = OPENAI_WS_MODE_OFF;
    codexCLIOnlyEnabled.value = false;
    anthropicPassthroughEnabled.value = false;
  };

  const hydrateRuntimeOptionsFromAccount = (account: Account) => {
    resetRuntimeOptions();

    const extra = getAccountExtra(account);
    if (
      account.platform === "openai" &&
      (account.type === "oauth" || account.type === "apikey")
    ) {
      const nextState = deriveOpenAIExtraState(account.type, extra);
      openaiPassthroughEnabled.value = nextState.openaiPassthroughEnabled;
      openaiOAuthResponsesWebSocketV2Mode.value =
        nextState.openaiOAuthResponsesWebSocketV2Mode;
      openaiAPIKeyResponsesWebSocketV2Mode.value =
        nextState.openaiAPIKeyResponsesWebSocketV2Mode;
      codexCLIOnlyEnabled.value = nextState.codexCLIOnlyEnabled;
    }

    if (account.platform === "anthropic" && account.type === "apikey") {
      anthropicPassthroughEnabled.value =
        extra.anthropic_passthrough === true;
    }
  };

  const openaiResponsesWebSocketV2Mode = computed({
    get: () => {
      if (getAccount()?.type === "apikey") {
        return openaiAPIKeyResponsesWebSocketV2Mode.value;
      }
      return openaiOAuthResponsesWebSocketV2Mode.value;
    },
    set: (mode: OpenAIWSMode) => {
      if (getAccount()?.type === "apikey") {
        openaiAPIKeyResponsesWebSocketV2Mode.value = mode;
        return;
      }
      openaiOAuthResponsesWebSocketV2Mode.value = mode;
    },
  });

  const openAIWSModeConcurrencyHintKey = computed(() =>
    resolveOpenAIWSModeConcurrencyHintKey(
      openaiResponsesWebSocketV2Mode.value,
    ),
  );

  const isOpenAIModelRestrictionDisabled = computed(
    () => getAccount()?.platform === "openai" && openaiPassthroughEnabled.value,
  );

  return {
    anthropicPassthroughEnabled,
    codexCLIOnlyEnabled,
    hydrateRuntimeOptionsFromAccount,
    isOpenAIModelRestrictionDisabled,
    openAIWSModeConcurrencyHintKey,
    openaiAPIKeyResponsesWebSocketV2Mode,
    openaiOAuthResponsesWebSocketV2Mode,
    openaiPassthroughEnabled,
    openaiResponsesWebSocketV2Mode,
    resetRuntimeOptions,
  };
}
