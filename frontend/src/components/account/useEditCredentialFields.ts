import { ref } from "vue";
import {
  DEFAULT_POOL_MODE_RETRY_COUNT,
  getDefaultBaseURL,
  normalizePoolModeRetryCount,
} from "@/components/account/credentialsBuilder";

export function useEditCredentialFields() {
  const editBaseUrl = ref(getDefaultBaseURL("anthropic"));
  const editApiKey = ref("");
  const editSessionToken = ref("");
  const poolModeEnabled = ref(false);
  const poolModeRetryCount = ref(DEFAULT_POOL_MODE_RETRY_COUNT);

  const resetCredentialFields = (defaultBaseUrl: string) => {
    editBaseUrl.value = defaultBaseUrl;
    editApiKey.value = "";
    editSessionToken.value = "";
    poolModeEnabled.value = false;
    poolModeRetryCount.value = DEFAULT_POOL_MODE_RETRY_COUNT;
  };

  const hydrateCompatibleCredentialFields = (
    credentials: Record<string, unknown>,
    defaultBaseUrl: string,
  ) => {
    editBaseUrl.value = (credentials.base_url as string) || defaultBaseUrl;
    poolModeEnabled.value = credentials.pool_mode === true;
    poolModeRetryCount.value = normalizePoolModeRetryCount(
      Number(credentials.pool_mode_retry_count ?? DEFAULT_POOL_MODE_RETRY_COUNT),
    );
  };

  const hydrateBedrockPoolMode = (credentials: Record<string, unknown>) => {
    poolModeEnabled.value = credentials.pool_mode === true;
    const retryCount = credentials.pool_mode_retry_count;
    poolModeRetryCount.value =
      typeof retryCount === "number" && retryCount >= 0
        ? retryCount
        : DEFAULT_POOL_MODE_RETRY_COUNT;
  };

  return {
    editApiKey,
    editBaseUrl,
    editSessionToken,
    hydrateBedrockPoolMode,
    hydrateCompatibleCredentialFields,
    poolModeEnabled,
    poolModeRetryCount,
    resetCredentialFields,
  };
}
