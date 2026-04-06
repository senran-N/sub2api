<template>
  <BaseDialog
    :show="show"
    :title="t('admin.proxies.qualityReportTitle')"
    width="normal"
    @close="emit('close')"
  >
    <div v-if="report" class="space-y-4">
      <div class="proxy-quality__panel">
        <div class="flex items-center justify-between gap-4">
          <div>
            <div class="proxy-quality__muted text-sm">
              {{ proxyName || '-' }}
            </div>
            <div class="proxy-quality__text mt-1 text-sm">
              {{ report.summary }}
            </div>
          </div>
          <div class="text-right">
            <div class="proxy-quality__score text-2xl font-semibold">
              {{ report.score }}
            </div>
            <div class="proxy-quality__muted text-xs">
              {{ t('admin.proxies.qualityGrade', { grade: report.grade }) }}
            </div>
          </div>
        </div>
        <div class="proxy-quality__meta mt-3 grid grid-cols-1 gap-2 text-xs sm:grid-cols-2">
          <div>{{ t('admin.proxies.qualityExitIP') }}: {{ report.exit_ip || '-' }}</div>
          <div>{{ t('admin.proxies.qualityCountry') }}: {{ report.country || '-' }}</div>
          <div>
            {{ t('admin.proxies.qualityBaseLatency') }}:
            {{ typeof report.base_latency_ms === 'number' ? `${report.base_latency_ms}ms` : '-' }}
          </div>
          <div>{{ t('admin.proxies.qualityCheckedAt') }}: {{ checkedAtLabel }}</div>
          <div v-if="report.ip_type">
            {{ t('admin.proxies.qualityIPType') }}:
            <span class="badge ml-1" :class="getIpTypeBadgeClass(report.ip_type)">
              {{ getIpTypeLabel(report.ip_type, t) }}
            </span>
          </div>
          <div v-if="report.isp">ISP: {{ report.isp }}</div>
          <div v-if="report.as">AS: {{ report.as }}</div>
          <div v-if="report.dns_leak_risk && report.dns_leak_risk !== 'none'">
            {{ t('admin.proxies.qualityDNSLeak') }}:
            <span class="badge ml-1" :class="getDnsLeakBadgeClass(report.dns_leak_risk)">
              {{ getDnsLeakLabel(report.dns_leak_risk, t) }}
            </span>
          </div>
        </div>
      </div>

      <div
        v-if="report.category_scores"
        class="proxy-quality__panel"
      >
        <div class="proxy-quality__muted mb-2 text-xs font-medium">
          {{ t('admin.proxies.qualityCategoryTitle') }}
        </div>
        <div class="space-y-2">
          <div
            v-for="category in categoryScoreEntries"
            :key="category.key"
            class="flex items-center gap-2 text-xs"
          >
            <span class="proxy-quality__meta w-28 shrink-0">
              {{ category.label }} ({{ category.weight }}%)
            </span>
            <div class="proxy-quality__track h-2 flex-1 rounded-full">
              <div
                class="theme-progress-fill h-2"
                :class="getProxyScoreBarColor(category.score)"
                :style="{ width: `${category.score}%` }"
              />
            </div>
            <span class="proxy-quality__muted w-8 text-right">
              {{ category.score }}
            </span>
          </div>
        </div>
      </div>

      <div class="proxy-quality__table-wrap overflow-auto">
        <table class="table min-w-full text-sm">
          <thead class="text-xs uppercase">
            <tr>
              <th class="proxy-quality__table-head-cell">{{ t('admin.proxies.qualityTableTarget') }}</th>
              <th class="proxy-quality__table-head-cell">{{ t('admin.proxies.qualityTableStatus') }}</th>
              <th class="proxy-quality__table-head-cell">HTTP</th>
              <th class="proxy-quality__table-head-cell">{{ t('admin.proxies.qualityTableLatency') }}</th>
              <th class="proxy-quality__table-head-cell">{{ t('admin.proxies.qualityTableMessage') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in report.items" :key="item.target">
              <td class="proxy-quality__table-cell proxy-quality__text">
                {{ getQualityTargetLabel(item.target, t) }}
              </td>
              <td class="proxy-quality__table-cell">
                <span class="badge" :class="getQualityStatusClass(item.status)">
                  {{ getQualityStatusLabel(item.status, t) }}
                </span>
              </td>
              <td class="proxy-quality__table-cell proxy-quality__meta">{{ item.http_status ?? '-' }}</td>
              <td class="proxy-quality__table-cell proxy-quality__meta">
                {{ typeof item.latency_ms === 'number' ? `${item.latency_ms}ms` : '-' }}
              </td>
              <td class="proxy-quality__table-cell proxy-quality__meta">
                <span>{{ item.message || '-' }}</span>
                <span v-if="item.cf_ray" class="proxy-quality__muted ml-1 text-xs">
                  (cf-ray: {{ item.cf_ray }})
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button @click="emit('close')" class="btn btn-secondary">
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { ProxyQualityCheckResult } from '@/types'
import {
  buildProxyQualityCategoryEntries,
  getDnsLeakBadgeClass,
  getDnsLeakLabel,
  getIpTypeBadgeClass,
  getIpTypeLabel,
  getProxyScoreBarColor,
  getQualityStatusClass,
  getQualityStatusLabel,
  getQualityTargetLabel
} from '../proxyPresentation'

const props = defineProps<{
  show: boolean
  proxyName?: string | null
  report: ProxyQualityCheckResult | null
}>()

const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()

const categoryScoreEntries = computed(() => {
  return buildProxyQualityCategoryEntries(props.report?.category_scores, t)
})

const checkedAtLabel = computed(() => {
  if (!props.report) {
    return '-'
  }

  return new Date(props.report.checked_at * 1000).toLocaleString()
})
</script>

<style scoped>
.proxy-quality__panel,
.proxy-quality__table-wrap {
  border-radius: var(--theme-surface-radius);
  border: 1px solid var(--theme-card-border);
  background: color-mix(in srgb, var(--theme-surface-soft) 86%, var(--theme-surface));
}

.proxy-quality__panel {
  padding: var(--theme-proxy-quality-panel-padding);
}

.proxy-quality__table-wrap {
  max-height: var(--theme-proxy-quality-table-max-height);
  background: var(--theme-surface);
}

.proxy-quality__score,
.proxy-quality__text {
  color: var(--theme-page-text);
}

.proxy-quality__table-head-cell,
.proxy-quality__table-cell {
  padding: var(--theme-table-cell-padding-y) var(--theme-table-cell-padding-x);
  text-align: left;
}

.proxy-quality__table-head-cell {
  font-size: var(--theme-table-head-font-size);
  letter-spacing: var(--theme-table-head-letter-spacing);
  text-transform: var(--theme-table-head-text-transform);
}

.proxy-quality__muted,
.proxy-quality__meta {
  color: var(--theme-page-muted);
}

.proxy-quality__track {
  background: color-mix(in srgb, var(--theme-page-border) 78%, transparent);
}
</style>
