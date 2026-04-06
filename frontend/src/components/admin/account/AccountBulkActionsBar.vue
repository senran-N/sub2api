<template>
  <div v-if="selectedIds.length > 0" class="account-bulk-actions-bar mb-4 flex items-center justify-between">
    <div class="flex flex-wrap items-center gap-2">
      <span class="account-bulk-actions-bar__label text-sm font-medium">
        {{ t('admin.accounts.bulkActions.selected', { count: selectedIds.length }) }}
      </span>
      <button
        @click="$emit('select-page')"
        class="account-bulk-actions-bar__link text-xs font-medium"
      >
        {{ t('admin.accounts.bulkActions.selectCurrentPage') }}
      </button>
      <span class="account-bulk-actions-bar__separator">•</span>
      <button
        @click="$emit('clear')"
        class="account-bulk-actions-bar__link text-xs font-medium"
      >
        {{ t('admin.accounts.bulkActions.clear') }}
      </button>
    </div>
    <div class="flex gap-2">
      <button @click="$emit('delete')" class="btn btn-danger btn-sm">{{ t('admin.accounts.bulkActions.delete') }}</button>
      <button @click="$emit('reset-status')" class="btn btn-secondary btn-sm">{{ t('admin.accounts.bulkActions.resetStatus') }}</button>
      <button @click="$emit('refresh-token')" class="btn btn-secondary btn-sm">{{ t('admin.accounts.bulkActions.refreshToken') }}</button>
      <button @click="$emit('toggle-schedulable', true)" class="btn btn-success btn-sm">{{ t('admin.accounts.bulkActions.enableScheduling') }}</button>
      <button @click="$emit('toggle-schedulable', false)" class="btn btn-warning btn-sm">{{ t('admin.accounts.bulkActions.disableScheduling') }}</button>
      <button @click="$emit('edit')" class="btn btn-primary btn-sm">{{ t('admin.accounts.bulkActions.edit') }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  selectedIds: number[]
}>()

defineEmits<{
  delete: []
  edit: []
  clear: []
  'select-page': []
  'toggle-schedulable': [enabled: boolean]
  'reset-status': []
  'refresh-token': []
}>()

const { t } = useI18n()
</script>

<style scoped>
.account-bulk-actions-bar {
  padding: calc(var(--theme-markdown-block-padding) * 0.75);
  border: 1px solid color-mix(in srgb, var(--theme-accent) 18%, var(--theme-card-border));
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-accent-soft) 80%, var(--theme-surface));
}

.account-bulk-actions-bar__label {
  color: color-mix(in srgb, var(--theme-accent) 72%, var(--theme-page-text));
}

.account-bulk-actions-bar__link {
  color: color-mix(in srgb, var(--theme-accent) 76%, var(--theme-page-text));
  transition: color 0.2s ease;
}

.account-bulk-actions-bar__link:hover {
  color: color-mix(in srgb, var(--theme-accent) 92%, var(--theme-page-text));
}

.account-bulk-actions-bar__separator {
  color: color-mix(in srgb, var(--theme-page-muted) 58%, transparent);
}
</style>
