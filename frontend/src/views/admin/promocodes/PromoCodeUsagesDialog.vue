<template>
  <BaseDialog
    :show="show"
    :title="t('admin.promo.usageRecords')"
    width="wide"
    @close="emit('close')"
  >
    <div v-if="loading" class="flex items-center justify-center py-8">
      <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
    </div>
    <div v-else-if="usages.length === 0" class="py-8 text-center text-gray-500 dark:text-gray-400">
      {{ t('admin.promo.noUsages') }}
    </div>
    <div v-else class="space-y-3">
      <div
        v-for="usage in usages"
        :key="usage.id"
        class="flex items-center justify-between rounded-lg border border-gray-200 p-3 dark:border-dark-600"
      >
        <div class="flex items-center gap-3">
          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
            <Icon name="user" size="sm" class="text-green-600 dark:text-green-400" />
          </div>
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ usage.user?.email || t('admin.promo.userPrefix', { id: usage.user_id }) }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ formatDateTime(usage.used_at) }}
            </p>
          </div>
        </div>
        <div class="text-right">
          <span class="text-sm font-medium text-green-600 dark:text-green-400">
            +${{ usage.bonus_amount.toFixed(2) }}
          </span>
        </div>
      </div>
      <div v-if="total > pageSize" class="mt-4">
        <Pagination
          :page="page"
          :total="total"
          :page-size="pageSize"
          :page-size-options="[10, 20, 50]"
          @update:page="emit('update:page', $event)"
          @update:page-size="emit('update:page-size', $event)"
        />
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/utils/format'
import type { PromoCodeUsage } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

defineProps<{
  show: boolean
  loading: boolean
  usages: PromoCodeUsage[]
  page: number
  pageSize: number
  total: number
}>()

const emit = defineEmits<{
  close: []
  'update:page': [page: number]
  'update:page-size': [pageSize: number]
}>()

const { t } = useI18n()
</script>
