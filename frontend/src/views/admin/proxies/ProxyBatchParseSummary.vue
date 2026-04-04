<template>
  <div
    v-if="summary.total > 0"
    class="rounded-lg bg-gray-50 p-4 dark:bg-dark-700"
  >
    <div class="flex items-center gap-4 text-sm">
      <div class="flex items-center gap-1.5">
        <Icon name="checkCircle" size="sm" :stroke-width="2" class="text-primary-500" />
        <span class="text-gray-700 dark:text-gray-300">
          {{ t('admin.proxies.parsedCount', { count: summary.valid }) }}
        </span>
      </div>
      <div v-if="summary.invalid > 0" class="flex items-center gap-1.5">
        <Icon
          name="exclamationCircle"
          size="sm"
          :stroke-width="2"
          class="text-amber-500"
        />
        <span class="text-amber-600 dark:text-amber-400">
          {{ t('admin.proxies.invalidCount', { count: summary.invalid }) }}
        </span>
      </div>
      <div v-if="summary.duplicate > 0" class="flex items-center gap-1.5">
        <svg
          class="h-4 w-4 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 01-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 011.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-8.876a9.06 9.06 0 00-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 01-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 00-3.375-3.375h-1.5a1.125 1.125 0 01-1.125-1.125v-1.5a3.375 3.375 0 00-3.375-3.375H9.75"
          />
        </svg>
        <span class="text-gray-500 dark:text-gray-400">
          {{ t('admin.proxies.duplicateCount', { count: summary.duplicate }) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ProxyBatchParseState } from '../proxyForm'

defineProps<{
  summary: ProxyBatchParseState
}>()

const { t } = useI18n()
</script>
