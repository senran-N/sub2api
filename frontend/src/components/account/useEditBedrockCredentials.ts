import { computed, ref } from "vue";
import type { Account } from "@/types";
import type { BedrockAuthMode } from "@/components/account/createAccountModalHelpers";

type AccountGetter = () => Account | null | undefined;

const getCredentials = (account: Account): Record<string, unknown> =>
  (account.credentials as Record<string, unknown>) || {};

export function useEditBedrockCredentials(getAccount: AccountGetter) {
  const editBedrockAccessKeyId = ref("");
  const editBedrockSecretAccessKey = ref("");
  const editBedrockSessionToken = ref("");
  const editBedrockRegion = ref("");
  const editBedrockForceGlobal = ref(false);
  const editBedrockApiKeyValue = ref("");

  const isBedrockAPIKeyMode = computed(
    () =>
      getAccount()?.type === "bedrock" &&
      (getAccount()?.credentials as Record<string, unknown>)?.auth_mode ===
        "apikey",
  );

  const editBedrockAuthMode = computed<BedrockAuthMode>(() =>
    isBedrockAPIKeyMode.value ? "apikey" : "sigv4",
  );

  const resetBedrockCredentials = () => {
    editBedrockAccessKeyId.value = "";
    editBedrockSecretAccessKey.value = "";
    editBedrockSessionToken.value = "";
    editBedrockRegion.value = "";
    editBedrockForceGlobal.value = false;
    editBedrockApiKeyValue.value = "";
  };

  const hydrateBedrockCredentialsFromAccount = (account: Account) => {
    resetBedrockCredentials();

    if (account.type !== "bedrock" || !account.credentials) {
      return;
    }

    const credentials = getCredentials(account);
    const authMode = (credentials.auth_mode as string) || "sigv4";
    editBedrockRegion.value = (credentials.aws_region as string) || "";
    editBedrockForceGlobal.value =
      (credentials.aws_force_global as string) === "true";

    if (authMode === "apikey") {
      editBedrockApiKeyValue.value = "";
      return;
    }

    editBedrockAccessKeyId.value =
      (credentials.aws_access_key_id as string) || "";
    editBedrockSecretAccessKey.value = "";
    editBedrockSessionToken.value = "";
  };

  return {
    editBedrockAccessKeyId,
    editBedrockApiKeyValue,
    editBedrockAuthMode,
    editBedrockForceGlobal,
    editBedrockRegion,
    editBedrockSecretAccessKey,
    editBedrockSessionToken,
    hydrateBedrockCredentialsFromAccount,
    isBedrockAPIKeyMode,
    resetBedrockCredentials,
  };
}
