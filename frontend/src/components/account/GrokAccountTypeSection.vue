<template>
  <div>
    <label class="input-label">{{ t("admin.accounts.accountType") }}</label>
    <div
      class="mt-2 grid grid-cols-1 gap-3 xl:grid-cols-3"
      data-tour="account-form-type"
    >
      <CreateAccountChoiceCard
        :selected="accountCategory === 'apikey'"
        tone="purple"
        icon="key"
        title="API Key"
        :description="t('admin.accounts.types.grokApiKey')"
        @select="emit('update:accountCategory', 'apikey')"
      />
      <CreateAccountChoiceCard
        :selected="accountCategory === 'upstream'"
        tone="blue"
        icon="cloud"
        title="Upstream"
        :description="t('admin.accounts.types.grokUpstream')"
        @select="emit('update:accountCategory', 'upstream')"
      />
      <CreateAccountChoiceCard
        :selected="accountCategory === 'session'"
        tone="emerald"
        icon="user"
        title="Session"
        :description="t('admin.accounts.types.grokSession')"
        @select="emit('update:accountCategory', 'session')"
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
