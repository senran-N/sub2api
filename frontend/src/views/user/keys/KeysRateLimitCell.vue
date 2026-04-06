<template>
  <div v-if="windows.length > 0" class="keys-rate-limit-cell space-y-1.5">
    <div v-for="window in windows" :key="window.key">
      <div class="flex items-center justify-between text-xs">
        <span class="theme-text-muted">{{ window.label }}</span>
        <span :class="['font-medium tabular-nums', getApiKeyRateLimitTextTone(window.usage, window.limit)]">
          ${{ window.usage?.toFixed(2) || '0.00' }}/${{ window.limit?.toFixed(2) }}
        </span>
      </div>
      <div class="theme-progress-track h-1 w-full">
        <div
          :class="['theme-progress-fill', getApiKeyRateLimitBarTone(window.usage, window.limit)]"
          :style="{ width: getApiKeyRateLimitProgressWidth(window.usage, window.limit) }"
        />
      </div>
      <div
        v-if="window.resetAt && formatResetTime(window.resetAt)"
        class="theme-text-subtle text-[10px] tabular-nums"
      >
        ⟳ {{ formatResetTime(window.resetAt) }}
      </div>
    </div>

    <button
      v-if="hasApiKeyRateLimitUsage(row)"
      type="button"
      class="theme-inline-action mt-0.5"
      :title="t('keys.resetRateLimitUsage')"
      @click.stop="emit('reset', row)"
    >
      <Icon name="refresh" size="xs" />
      {{ t('keys.resetUsage') }}
    </button>
  </div>

  <span v-else class="theme-text-subtle text-sm">-</span>
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

<style scoped>
.keys-rate-limit-cell {
  min-width: var(--theme-keys-rate-limit-min-width);
}
</style>
