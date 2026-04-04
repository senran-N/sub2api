<template>
  <BaseDialog
    :show="show"
    :title="t('admin.proxies.qualityReportTitle')"
    width="normal"
    @close="emit('close')"
  >
    <div v-if="report" class="space-y-4">
      <div class="rounded-lg border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-700">
        <div class="flex items-center justify-between gap-4">
          <div>
            <div class="text-sm text-gray-500 dark:text-gray-400">
              {{ proxyName || '-' }}
            </div>
            <div class="mt-1 text-sm text-gray-700 dark:text-gray-200">
              {{ report.summary }}
            </div>
          </div>
          <div class="text-right">
            <div class="text-2xl font-semibold text-gray-900 dark:text-white">
              {{ report.score }}
            </div>
            <div class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.proxies.qualityGrade', { grade: report.grade }) }}
            </div>
          </div>
        </div>
        <div class="mt-3 grid grid-cols-1 gap-2 text-xs text-gray-600 dark:text-gray-300 sm:grid-cols-2">
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
        class="rounded-lg border border-gray-200 p-4 dark:border-dark-600"
      >
        <div class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">
          {{ t('admin.proxies.qualityCategoryTitle') }}
        </div>
        <div class="space-y-2">
          <div
            v-for="category in categoryScoreEntries"
            :key="category.key"
            class="flex items-center gap-2 text-xs"
          >
            <span class="w-28 shrink-0 text-gray-600 dark:text-gray-300">
              {{ category.label }} ({{ category.weight }}%)
            </span>
            <div class="h-2 flex-1 rounded-full bg-gray-200 dark:bg-dark-600">
              <div
                class="h-2 rounded-full transition-all"
                :class="getProxyScoreBarColor(category.score)"
                :style="{ width: `${category.score}%` }"
              />
            </div>
            <span class="w-8 text-right text-gray-500 dark:text-gray-400">
              {{ category.score }}
            </span>
          </div>
        </div>
      </div>

      <div class="max-h-80 overflow-auto rounded-lg border border-gray-200 dark:border-dark-600">
        <table class="min-w-full divide-y divide-gray-200 text-sm dark:divide-dark-700">
          <thead class="bg-gray-50 text-xs uppercase text-gray-500 dark:bg-dark-800 dark:text-dark-400">
            <tr>
              <th class="px-3 py-2 text-left">{{ t('admin.proxies.qualityTableTarget') }}</th>
              <th class="px-3 py-2 text-left">{{ t('admin.proxies.qualityTableStatus') }}</th>
              <th class="px-3 py-2 text-left">HTTP</th>
              <th class="px-3 py-2 text-left">{{ t('admin.proxies.qualityTableLatency') }}</th>
              <th class="px-3 py-2 text-left">{{ t('admin.proxies.qualityTableMessage') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-900">
            <tr v-for="item in report.items" :key="item.target">
              <td class="px-3 py-2 text-gray-900 dark:text-white">
                {{ getQualityTargetLabel(item.target, t) }}
              </td>
              <td class="px-3 py-2">
                <span class="badge" :class="getQualityStatusClass(item.status)">
                  {{ getQualityStatusLabel(item.status, t) }}
                </span>
              </td>
              <td class="px-3 py-2 text-gray-600 dark:text-gray-300">{{ item.http_status ?? '-' }}</td>
              <td class="px-3 py-2 text-gray-600 dark:text-gray-300">
                {{ typeof item.latency_ms === 'number' ? `${item.latency_ms}ms` : '-' }}
              </td>
              <td class="px-3 py-2 text-gray-600 dark:text-gray-300">
                <span>{{ item.message || '-' }}</span>
                <span v-if="item.cf_ray" class="ml-1 text-xs text-gray-400">
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
