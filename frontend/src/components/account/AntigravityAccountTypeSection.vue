<template>
  <div>
    <label class="input-label">{{ t("admin.accounts.accountType") }}</label>
    <div class="mt-2 grid grid-cols-2 gap-3">
      <CreateAccountChoiceCard
        :selected="accountType === 'oauth'"
        tone="purple"
        icon="key"
        title="OAuth"
        :description="t('admin.accounts.types.antigravityOauth')"
        @select="emit('update:accountType', 'oauth')"
      />
      <CreateAccountChoiceCard
        :selected="accountType === 'upstream'"
        tone="purple"
        icon="cloud"
        title="API Key"
        :description="t('admin.accounts.types.antigravityApikey')"
        @select="emit('update:accountType', 'upstream')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import CreateAccountChoiceCard from "@/components/account/CreateAccountChoiceCard.vue";

type AntigravityAccountType = "oauth" | "upstream";

defineProps<{
  accountType: AntigravityAccountType;
}>();

const emit = defineEmits<{
  "update:accountType": [value: AntigravityAccountType];
}>();

const { t } = useI18n();
</script>
