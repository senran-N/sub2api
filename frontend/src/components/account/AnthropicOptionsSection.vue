<template>
  <div v-if="accountCategory === 'apikey'" class="form-section">
    <div class="flex items-center justify-between gap-4">
      <div>
        <label class="input-label mb-0">{{
          t("admin.accounts.anthropic.apiKeyPassthrough")
        }}</label>
        <p class="anthropic-options-section__description mt-1 text-xs">
          {{ t("admin.accounts.anthropic.apiKeyPassthroughDesc") }}
        </p>
      </div>
      <AccountModalSwitch
        :model-value="apiKeyPassthroughEnabled"
        :aria-label="t('admin.accounts.anthropic.apiKeyPassthrough')"
        @update:model-value="emit('update:apiKeyPassthroughEnabled', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import AccountModalSwitch from "@/components/account/AccountModalSwitch.vue";
import type { CreateAccountCategory } from "@/components/account/createAccountModalHelpers";

defineProps<{
  accountCategory: CreateAccountCategory;
  apiKeyPassthroughEnabled: boolean;
}>();

const emit = defineEmits<{
  "update:apiKeyPassthroughEnabled": [value: boolean];
}>();

const { t } = useI18n();
</script>

<style scoped>
.anthropic-options-section__description {
  color: var(--theme-page-muted);
}
</style>
