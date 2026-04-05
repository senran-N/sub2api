<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ title }}
      </h2>
    </div>
    <div class="p-6">
      <div v-if="loading" class="flex items-center justify-center py-8">
        <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      </div>

      <div v-else-if="history.length > 0" class="space-y-3">
        <div
          v-for="item in history"
          :key="item.id"
          class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
        >
          <div class="flex items-center gap-4">
            <div
              class="flex h-10 w-10 items-center justify-center rounded-xl"
              :class="buildRedeemHistoryPresentation(item).iconBgClass"
            >
              <Icon
                :name="buildRedeemHistoryPresentation(item).iconName"
                size="md"
                :class="buildRedeemHistoryPresentation(item).iconColorClass"
              />
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white">
                {{ resolveRedeemHistoryTitle(item, t) }}
              </p>
              <p class="text-xs text-gray-500 dark:text-dark-400">
                {{ formatDateTime(item.used_at) }}
              </p>
            </div>
          </div>
          <div class="text-right">
            <p
              class="text-sm font-semibold"
              :class="buildRedeemHistoryPresentation(item).valueColorClass"
            >
              {{ formatRedeemHistoryValue(item, t) }}
            </p>
            <p
              v-if="!isAdminAdjustmentRedeemType(item.type)"
              class="font-mono text-xs text-gray-400 dark:text-dark-500"
            >
              {{ item.code.slice(0, 8) }}...
            </p>
            <p v-else class="text-xs text-gray-400 dark:text-dark-500">
              {{ adminAdjustmentLabel }}
            </p>
            <p
              v-if="item.notes"
              class="mt-1 max-w-[200px] truncate text-xs italic text-gray-500 dark:text-dark-400"
              :title="item.notes"
            >
              {{ item.notes }}
            </p>
          </div>
        </div>
      </div>

      <div v-else class="empty-state py-8">
        <div
          class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800"
        >
          <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
        </div>
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ emptyLabel }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { RedeemHistoryItem } from '@/api'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'
import {
  buildRedeemHistoryPresentation,
  formatRedeemHistoryValue,
  isAdminAdjustmentRedeemType,
  resolveRedeemHistoryTitle
} from './redeemView'

defineProps<{
  adminAdjustmentLabel: string
  emptyLabel: string
  history: RedeemHistoryItem[]
  loading: boolean
  title: string
}>()

const { t } = useI18n()
</script>
