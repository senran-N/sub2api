<template>
  <BaseDialog
    :show="show"
    :title="t('admin.promo.usageRecords')"
    width="wide"
    @close="emit('close')"
  >
    <div
      v-if="loading"
      class="flex items-center justify-center promo-code-usages-dialog__status-state"
    >
      <Icon name="refresh" size="lg" class="promo-code-usages-dialog__spinner animate-spin" />
    </div>
    <div
      v-else-if="usages.length === 0"
      class="promo-code-usages-dialog__muted promo-code-usages-dialog__status-state text-center"
    >
      {{ t('admin.promo.noUsages') }}
    </div>
    <div v-else class="promo-code-usages-dialog__list">
      <div
        v-for="usage in usages"
        :key="usage.id"
        class="promo-code-usages-dialog__item flex items-center justify-between"
      >
        <div class="flex items-center gap-3">
          <div class="promo-code-usages-dialog__icon-shell">
            <Icon name="user" size="sm" class="promo-code-usages-dialog__bonus" />
          </div>
          <div>
            <p class="promo-code-usages-dialog__email text-sm font-medium">
              {{ usage.user?.email || t('admin.promo.userPrefix', { id: usage.user_id }) }}
            </p>
            <p class="promo-code-usages-dialog__muted text-xs">
              {{ formatDateTime(usage.used_at) }}
            </p>
          </div>
        </div>
        <div class="text-right">
          <span class="promo-code-usages-dialog__bonus text-sm font-medium">
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

<style scoped>
.promo-code-usages-dialog__spinner,
.promo-code-usages-dialog__muted {
  color: var(--theme-page-muted);
}

.promo-code-usages-dialog__status-state {
  padding: var(--theme-promo-usages-status-padding-y) 0;
}

.promo-code-usages-dialog__list {
  display: flex;
  flex-direction: column;
  gap: var(--theme-promo-usages-list-gap);
}

.promo-code-usages-dialog__item {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 84%, transparent);
  border-radius: var(--theme-promo-usages-item-radius);
  padding: var(--theme-promo-usages-item-padding);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
}

.promo-code-usages-dialog__icon-shell {
  width: var(--theme-promo-usages-icon-shell-size);
  height: var(--theme-promo-usages-icon-shell-size);
  border-radius: var(--theme-promo-usages-icon-shell-radius);
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 12%, var(--theme-surface));
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.promo-code-usages-dialog__email {
  color: var(--theme-page-text);
}

.promo-code-usages-dialog__bonus {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
