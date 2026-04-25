<template>
  <BulkEditApplySection
    id="bulk-edit-base-url"
    :enabled="enabled"
    label-key="admin.accounts.baseUrl"
    @update:enabled="emit('update:enabled', $event)"
  >
    <input
      id="bulk-edit-base-url"
      :value="baseUrl"
      type="text"
      :disabled="!enabled"
      class="input"
      :class="!enabled && 'cursor-not-allowed opacity-50'"
      :placeholder="t('admin.accounts.bulkEdit.baseUrlPlaceholder')"
      aria-labelledby="bulk-edit-base-url-label"
      @input="
        emit('update:baseUrl', ($event.target as HTMLInputElement).value)
      "
    />
    <p class="input-hint">
      {{ t("admin.accounts.bulkEdit.baseUrlNotice") }}
    </p>
  </BulkEditApplySection>
</template>

<script setup lang="ts">
import { useI18n } from "vue-i18n";
import BulkEditApplySection from "@/components/account/BulkEditApplySection.vue";

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
