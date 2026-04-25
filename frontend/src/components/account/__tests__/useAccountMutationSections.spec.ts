import { describe, expect, it } from "vitest";
import { ref } from "vue";
import { useAccountMutationSections } from "../useAccountMutationSections";
import type { AccountPlatform, AccountType } from "@/types";

describe("useAccountMutationSections", () => {
  it("resolves compatible OpenAI API-key sections", () => {
    const source = ref<{ platform: AccountPlatform; type: AccountType }>({
      platform: "openai",
      type: "apikey",
    });
    const sections = useAccountMutationSections(() => source.value);

    expect(sections.showCompatibleCredentialsForm.value).toBe(true);
    expect(sections.showQuotaLimitSection.value).toBe(true);
    expect(sections.showOpenAIRuntimeSection.value).toBe(true);
    expect(sections.showAnthropicAPIKeyRuntimeSection.value).toBe(false);
    expect(sections.openAIAccountCategory.value).toBe("apikey");
  });

  it("resolves Anthropic runtime sections", () => {
    const source = ref<{ platform: AccountPlatform; type: AccountType }>({
      platform: "anthropic",
      type: "oauth",
    });
    const sections = useAccountMutationSections(() => source.value);

    expect(sections.showAnthropicQuotaControls.value).toBe(true);
    expect(sections.showWarmupSection.value).toBe(true);
    expect(sections.showCompatibleCredentialsForm.value).toBe(false);

    source.value = { platform: "anthropic", type: "apikey" };
    expect(sections.showAnthropicAPIKeyRuntimeSection.value).toBe(true);
    expect(sections.showAnthropicQuotaControls.value).toBe(false);
  });

  it("returns hidden sections when no source is available", () => {
    const sections = useAccountMutationSections(() => null);

    expect(sections.mutationProfile.value).toBeNull();
    expect(sections.showCompatibleCredentialsForm.value).toBe(false);
    expect(sections.showQuotaLimitSection.value).toBe(false);
    expect(sections.openAIAccountCategory.value).toBe("oauth-based");
  });
});
