<template>
  <div v-if="windows.length > 0" class="min-w-[140px] space-y-1.5">
    <div v-for="window in windows" :key="window.key">
      <div class="flex items-center justify-between text-xs">
        <span class="text-gray-500 dark:text-gray-400">{{ window.label }}</span>
        <span :class="['font-medium tabular-nums', getApiKeyRateLimitTextTone(window.usage, window.limit)]">
          ${{ window.usage?.toFixed(2) || '0.00' }}/${{ window.limit?.toFixed(2) }}
        </span>
      </div>
      <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
        <div
          :class="['h-full rounded-full transition-all', getApiKeyRateLimitBarTone(window.usage, window.limit)]"
          :style="{ width: getApiKeyRateLimitProgressWidth(window.usage, window.limit) }"
        />
      </div>
      <div
        v-if="window.resetAt && formatResetTime(window.resetAt)"
        class="text-[10px] tabular-nums text-gray-400 dark:text-gray-500"
      >
        ⟳ {{ formatResetTime(window.resetAt) }}
      </div>
    </div>

    <button
      v-if="hasApiKeyRateLimitUsage(row)"
      type="button"
      class="mt-0.5 inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
      :title="t('keys.resetRateLimitUsage')"
      @click.stop="emit('reset', row)"
    >
      <Icon name="refresh" size="xs" />
      {{ t('keys.resetUsage') }}
    </button>
  </div>

  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ApiKey } from '@/types'
import {
  getApiKeyRateLimitBarTone,
  getApiKeyRateLimitProgressWidth,
  getApiKeyRateLimitTextTone,
  getApiKeyRateLimitWindows,
  hasApiKeyRateLimitUsage
} from './keysView'

const props = defineProps<{
  row: ApiKey
  formatResetTime: (value: string | null) => string
}>()

const emit = defineEmits<{
  reset: [row: ApiKey]
}>()

const { t } = useI18n()

const windows = computed(() => getApiKeyRateLimitWindows(props.row))
</script>
