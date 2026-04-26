import { reactive, ref } from "vue";
import type { AddMethod } from "@/composables/useAccountOAuth";
import type { AccountPlatform } from "@/types";
import {
  createDefaultCreateAccountForm,
  resetCreateAccountForm,
  type CreateAccountForm,
} from "@/components/account/accountModalShared";
import {
  type CreateAccountCategory,
} from "@/components/account/createAccountModalHelpers";
import { getDefaultBaseURL } from "@/components/account/credentialsBuilder";

export function getDefaultCreateAccountCategory(
  platform: AccountPlatform,
): CreateAccountCategory {
  switch (platform) {
    case "grok":
      return "apikey";
    case "anthropic":
    case "openai":
    case "gemini":
    case "antigravity":
    default:
      return "oauth-based";
  }
}

export function useCreateAccountFormState() {
  const form = reactive<CreateAccountForm>(createDefaultCreateAccountForm());
  const step = ref(1);
  const submitting = ref(false);
  const accountCategory = ref<CreateAccountCategory>(
    getDefaultCreateAccountCategory("anthropic"),
  );
  const addMethod = ref<AddMethod>("oauth");
  const apiKeyBaseUrl = ref(getDefaultBaseURL("anthropic"));
  const apiKeyValue = ref("");

  const resetBaseFormState = () => {
    step.value = 1;
    submitting.value = false;
    resetCreateAccountForm(form);
    accountCategory.value = getDefaultCreateAccountCategory(form.platform);
    addMethod.value = "oauth";
    apiKeyBaseUrl.value = getDefaultBaseURL("anthropic");
    apiKeyValue.value = "";
  };

  return {
    accountCategory,
    addMethod,
    apiKeyBaseUrl,
    apiKeyValue,
    form,
    resetBaseFormState,
    step,
    submitting,
  };
}
