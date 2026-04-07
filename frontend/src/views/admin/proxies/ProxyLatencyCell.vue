<template>
  <div class="flex flex-col gap-1">
    <span
      v-if="proxy.latency_status === 'failed'"
      class="badge badge-danger"
      :title="proxy.latency_message || undefined"
    >
      {{ t('admin.proxies.latencyFailed') }}
    </span>
    <span
      v-else-if="typeof proxy.latency_ms === 'number'"
      :class="latencyClass"
    >
      {{ proxy.latency_ms }}ms
    </span>
    <span v-else class="theme-text-subtle text-sm">-</span>
    <div
      v-if="typeof proxy.quality_checked === 'number'"
      class="theme-text-muted flex items-center gap-1 text-xs"
      :title="proxy.quality_summary || undefined"
    >
      <span>{{ t('admin.proxies.qualityInline', { grade: proxy.quality_grade || '-', score: proxy.quality_score ?? '-' }) }}</span>
      <span class="badge" :class="getQualityOverallClass(proxy.quality_status)">
        {{ getQualityOverallLabel(proxy.quality_status, t) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Proxy } from '@/types'
import { getQualityOverallClass, getQualityOverallLabel } from './proxyPresentation'

const props = defineProps<{
  proxy: Proxy
}>()

const { t } = useI18n()

const latencyClass = computed(() => [
  'badge',
  (props.proxy.latency_ms ?? 0) < 200 ? 'badge-success' : 'badge-warning'
])
</script>
