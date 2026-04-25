import { computed } from "vue";
import type { AccountPlatform, AccountType } from "@/types";
import type { CreateAccountCategory } from "@/components/account/createAccountModalHelpers";
import {
  accountMutationProfileHasSection,
  resolveAccountMutationProfile,
} from "@/components/account/accountMutationProfiles";

interface AccountMutationSectionSource {
  platform: AccountPlatform;
  type: AccountType;
}

export function useAccountMutationSections(
  getSource: () => AccountMutationSectionSource | null | undefined,
) {
  const mutationProfile = computed(() => {
    const source = getSource();
    return source
      ? resolveAccountMutationProfile(source.platform, source.type)
      : null;
  });

  const showCompatibleCredentialsForm = computed(() =>
    accountMutationProfileHasSection(
      mutationProfile.value,
      "compatible-credentials",
    ),
  );

  const showQuotaLimitSection = computed(() =>
    accountMutationProfileHasSection(mutationProfile.value, "quota-limits"),
  );

  const showWarmupSection = computed(() =>
    accountMutationProfileHasSection(mutationProfile.value, "warmup"),
  );

  const showOpenAIRuntimeSection = computed(() =>
    accountMutationProfileHasSection(mutationProfile.value, "openai-runtime"),
  );

  const showAnthropicQuotaControls = computed(() =>
    accountMutationProfileHasSection(
      mutationProfile.value,
      "anthropic-runtime",
    ),
  );

  const showAnthropicAPIKeyRuntimeSection = computed(() => {
    const source = getSource();
    return source?.platform === "anthropic" && source.type === "apikey";
  });

  const openAIAccountCategory = computed<CreateAccountCategory>(() =>
    getSource()?.type === "apikey" ? "apikey" : "oauth-based",
  );

  return {
    mutationProfile,
    openAIAccountCategory,
    showAnthropicAPIKeyRuntimeSection,
    showAnthropicQuotaControls,
    showCompatibleCredentialsForm,
    showOpenAIRuntimeSection,
    showQuotaLimitSection,
    showWarmupSection,
  };
}
