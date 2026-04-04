<template>
  <div
    class="mt-2 flex items-center justify-between rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800 dark:border-amber-700/40 dark:bg-amber-900/20 dark:text-amber-200"
  >
    <span>{{ t('admin.accounts.listPendingSyncHint') }}</span>
    <button
      class="btn btn-secondary px-2 py-1 text-xs"
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
