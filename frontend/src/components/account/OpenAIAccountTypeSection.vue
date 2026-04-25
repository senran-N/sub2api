<template>
  <div>
    <label class="input-label">{{ t("admin.accounts.accountType") }}</label>
    <div class="mt-2 grid grid-cols-2 gap-3" data-tour="account-form-type">
      <CreateAccountChoiceCard
        :selected="accountCategory === 'oauth-based'"
        tone="green"
        icon="key"
        title="OAuth"
        :description="t('admin.accounts.types.chatgptOauth')"
        @select="emit('update:accountCategory', 'oauth-based')"
      />
      <CreateAccountChoiceCard
        :selected="accountCategory === 'apikey'"
        tone="purple"
        icon="key"
        title="API Key"
        :description="t('admin.accounts.types.responsesApi')"
        @select="emit('update:accountCategory', 'apikey')"
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
