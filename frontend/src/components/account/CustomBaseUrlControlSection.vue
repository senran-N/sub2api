<template>
  <QuotaControlCard
    :enabled="enabled"
    title-key="admin.accounts.quotaControl.customBaseUrl.label"
    hint-key="admin.accounts.quotaControl.customBaseUrl.hint"
    @update:enabled="emit('update:enabled', $event)"
  >
    <div class="mt-3">
      <input
        :value="baseUrl"
        type="text"
        class="input"
        :placeholder="t('admin.accounts.quotaControl.customBaseUrl.urlHint')"
        @input="
          emit('update:baseUrl', ($event.target as HTMLInputElement).value)
        "
      />
    </div>
  </QuotaControlCard>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import QuotaControlCard from "@/components/account/QuotaControlCard.vue";

defineProps<{
  enabled: boolean;
  baseUrl: string;
}>();

const emit = defineEmits<{
  "update:enabled": [value: boolean];
  "update:baseUrl": [value: string];
}>();

const { t } = useI18n();
</script>
