<template>
  <div
    class="account-pending-sync-banner mt-2 flex items-center justify-between border text-sm"
  >
    <span>{{ t('admin.accounts.listPendingSyncHint') }}</span>
    <button
      class="account-pending-sync-banner__action btn btn-secondary text-xs"
      @click="emit('sync')"
    >
      {{ t('admin.accounts.listPendingSyncAction') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const emit = defineEmits<{
  sync: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.account-pending-sync-banner {
  padding:
    calc(var(--theme-scheduled-tests-result-card-padding) - 0.25rem)
    var(--theme-scheduled-tests-result-card-padding);
  border-radius: calc(var(--theme-surface-radius) + 2px);
  border-color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 28%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.account-pending-sync-banner__action {
  padding:
    var(--theme-account-usage-action-padding-y)
    var(--theme-settings-code-padding-x);
}
</style>
