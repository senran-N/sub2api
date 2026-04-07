<template>
  <div
    v-if="summary.total > 0"
    class="proxy-batch-parse-summary"
  >
    <div class="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
      <div class="proxy-batch-parse-summary__item proxy-batch-parse-summary__item--success">
        <Icon
          name="checkCircle"
          size="sm"
          :stroke-width="2"
          class="proxy-batch-parse-summary__icon proxy-batch-parse-summary__icon--success"
        />
        <span class="proxy-batch-parse-summary__text">
          {{ t('admin.proxies.parsedCount', { count: summary.valid }) }}
        </span>
      </div>
      <div
        v-if="summary.invalid > 0"
        class="proxy-batch-parse-summary__item proxy-batch-parse-summary__item--warning"
      >
        <Icon
          name="exclamationCircle"
          size="sm"
          :stroke-width="2"
          class="proxy-batch-parse-summary__icon proxy-batch-parse-summary__icon--warning"
        />
        <span class="proxy-batch-parse-summary__text">
          {{ t('admin.proxies.invalidCount', { count: summary.invalid }) }}
        </span>
      </div>
      <div
        v-if="summary.duplicate > 0"
        class="proxy-batch-parse-summary__item proxy-batch-parse-summary__item--neutral"
      >
        <svg
          class="proxy-batch-parse-summary__icon proxy-batch-parse-summary__icon--neutral h-4 w-4"
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
        <span class="proxy-batch-parse-summary__text">
          {{ t('admin.proxies.duplicateCount', { count: summary.duplicate }) }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { ProxyBatchParseState } from './proxyForm'

defineProps<{
  summary: ProxyBatchParseState
}>()

const { t } = useI18n()
</script>

<style scoped>
.proxy-batch-parse-summary {
  padding: var(--theme-markdown-block-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: var(--theme-select-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.proxy-batch-parse-summary__item {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  color: var(--theme-page-muted);
}

.proxy-batch-parse-summary__item--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.proxy-batch-parse-summary__item--warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.proxy-batch-parse-summary__item--neutral {
  color: var(--theme-page-muted);
}

.proxy-batch-parse-summary__icon--success {
  color: rgb(var(--theme-success-rgb));
}

.proxy-batch-parse-summary__icon--warning {
  color: rgb(var(--theme-warning-rgb));
}

.proxy-batch-parse-summary__icon--neutral {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.proxy-batch-parse-summary__text {
  color: inherit;
}
</style>
