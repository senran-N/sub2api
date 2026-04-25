<template>
  <div>
    <div class="flex items-center justify-between gap-3">
      <label class="input-label">{{ t("admin.accounts.accountType") }}</label>
      <button
        type="button"
        class="gemini-account-type-section__help-button"
        @click="emit('openHelp')"
      >
        <Icon name="questionCircle" size="sm" :stroke-width="2" />
        {{ t("admin.accounts.gemini.helpButton") }}
      </button>
    </div>

    <div class="mt-2 grid grid-cols-2 gap-3" data-tour="account-form-type">
      <CreateAccountChoiceCard
        :selected="accountCategory === 'oauth-based'"
        tone="blue"
        icon="key"
        :title="t('admin.accounts.gemini.accountType.oauthTitle')"
        :description="t('admin.accounts.gemini.accountType.oauthDesc')"
        @select="emit('update:accountCategory', 'oauth-based')"
      />
      <CreateAccountChoiceCard
        :selected="accountCategory === 'apikey'"
        tone="purple"
        icon="key"
        :title="t('admin.accounts.gemini.accountType.apiKeyTitle')"
        :description="t('admin.accounts.gemini.accountType.apiKeyDesc')"
        @select="emit('update:accountCategory', 'apikey')"
      />
    </div>

    <div
      v-if="accountCategory === 'apikey'"
      class="gemini-account-type-section__notice mt-3 text-xs"
    >
      <p>{{ t("admin.accounts.gemini.accountType.apiKeyNote") }}</p>
      <div class="mt-2 flex flex-wrap gap-2">
        <a
          :href="apiKeyHelpLink"
          class="gemini-account-type-section__link font-medium"
          target="_blank"
          rel="noreferrer"
        >
          {{ t("admin.accounts.gemini.accountType.apiKeyLink") }}
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import CreateAccountChoiceCard from "@/components/account/CreateAccountChoiceCard.vue";
import Icon from "@/components/icons/Icon.vue";
import type { CreateAccountCategory } from "@/components/account/createAccountModalHelpers";

defineProps<{
  accountCategory: CreateAccountCategory;
  apiKeyHelpLink: string;
}>();

const emit = defineEmits<{
  "update:accountCategory": [value: CreateAccountCategory];
  openHelp: [];
}>();

const { t } = useI18n();
</script>

<style scoped>
.gemini-account-type-section__help-button {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border-radius: var(--theme-button-radius);
  padding: calc(var(--theme-button-padding-y) * 0.4)
    calc(var(--theme-button-padding-x) * 0.4);
  font-size: 0.75rem;
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 82%,
    var(--theme-page-text)
  );
}

.gemini-account-type-section__help-button:hover {
  background: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 10%,
    var(--theme-surface)
  );
}

.gemini-account-type-section__notice {
  border-radius: var(--theme-auth-feedback-radius);
  padding: 0.5rem 0.75rem;
  border-color: color-mix(in srgb, var(--theme-card-border) 68%, transparent);
  background: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 10%,
    var(--theme-surface)
  );
  color: color-mix(
    in srgb,
    rgb(var(--theme-brand-purple-rgb)) 84%,
    var(--theme-page-text)
  );
}

.gemini-account-type-section__link {
  color: color-mix(
    in srgb,
    rgb(var(--theme-info-rgb)) 84%,
    var(--theme-page-text)
  );
}

.gemini-account-type-section__link:hover {
  text-decoration: underline;
}
</style>
