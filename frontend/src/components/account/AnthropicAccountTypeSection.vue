<template>
  <div>
    <label class="input-label">{{ t("admin.accounts.accountType") }}</label>
    <div class="mt-2 grid grid-cols-3 gap-3" data-tour="account-form-type">
      <CreateAccountChoiceCard
        :selected="accountCategory === 'oauth-based'"
        tone="orange"
        icon="sparkles"
        :title="t('admin.accounts.claudeCode')"
        :description="t('admin.accounts.oauthSetupToken')"
        @select="emit('update:accountCategory', 'oauth-based')"
      />
      <CreateAccountChoiceCard
        :selected="accountCategory === 'apikey'"
        tone="purple"
        icon="key"
        :title="t('admin.accounts.claudeConsole')"
        :description="t('admin.accounts.apiKey')"
        @select="emit('update:accountCategory', 'apikey')"
      />
      <CreateAccountChoiceCard
        :selected="accountCategory === 'bedrock'"
        tone="amber"
        icon="cloud"
        :title="t('admin.accounts.bedrockLabel')"
        :description="t('admin.accounts.bedrockDesc')"
        @select="emit('update:accountCategory', 'bedrock')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import CreateAccountChoiceCard from "@/components/account/CreateAccountChoiceCard.vue";
import type { CreateAccountCategory } from "@/components/account/createAccountModalHelpers";

defineProps<{
  accountCategory: CreateAccountCategory;
}>();

const emit = defineEmits<{
  "update:accountCategory": [value: CreateAccountCategory];
}>();

const { t } = useI18n();
</script>
