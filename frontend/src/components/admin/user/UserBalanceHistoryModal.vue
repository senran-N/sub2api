<template>
  <BaseDialog :show="show" :title="t('admin.users.balanceHistoryTitle')" width="wide" :close-on-click-outside="true" :z-index="40" @close="$emit('close')">
    <div v-if="user" class="space-y-4">
      <!-- User header: two-row layout with full user info -->
      <div class="user-balance-history-modal__hero">
        <!-- Row 1: avatar + email/username/created_at (left) + current balance (right) -->
        <div class="flex items-center gap-3">
          <div class="user-balance-history-modal__avatar user-balance-history-modal__avatar-shape flex flex-shrink-0 items-center justify-center">
            <span class="user-balance-history-modal__avatar-text text-lg font-medium">
              {{ user.email.charAt(0).toUpperCase() }}
            </span>
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <p class="user-balance-history-modal__text-strong truncate font-medium">{{ user.email }}</p>
              <span
                v-if="user.username"
                class="theme-chip theme-chip--compact theme-chip--accent flex-shrink-0"
              >
                {{ user.username }}
              </span>
            </div>
            <p class="user-balance-history-modal__text-soft text-xs">
              {{ t('admin.users.createdAt') }}: {{ formatDateTime(user.created_at) }}
            </p>
          </div>
          <!-- Current balance: prominent display on the right -->
          <div class="flex-shrink-0 text-right">
            <p class="user-balance-history-modal__text-muted text-xs">{{ t('admin.users.currentBalance') }}</p>
            <p class="user-balance-history-modal__text-strong text-xl font-bold">
              ${{ user.balance?.toFixed(2) || '0.00' }}
            </p>
          </div>
        </div>
        <!-- Row 2: notes + total recharged -->
        <div class="user-balance-history-modal__hero-footer user-balance-history-modal__hero-footer-layout flex items-center justify-between">
          <p class="user-balance-history-modal__text-muted min-w-0 flex-1 truncate text-xs" :title="user.notes || ''">
            <template v-if="user.notes">{{ t('admin.users.notes') }}: {{ user.notes }}</template>
            <template v-else>&nbsp;</template>
          </p>
          <p class="user-balance-history-modal__text-muted ml-4 flex-shrink-0 text-xs">
            {{ t('admin.users.totalRecharged') }}: <span class="user-balance-history-modal__text-success font-semibold">${{ totalRecharged.toFixed(2) }}</span>
          </p>
        </div>
      </div>

      <!-- Type filter + Action buttons -->
      <div class="flex items-center gap-3">
        <Select
          v-model="typeFilter"
          :options="typeOptions"
          class="user-balance-history-modal__filter"
          @change="loadHistory(1)"
        />
        <!-- Deposit button - matches menu style -->
        <button
          v-if="!hideActions"
          @click="emit('deposit')"
          class="user-balance-history-modal__action-button user-balance-history-modal__action-button-layout user-balance-history-modal__action-button--deposit flex items-center gap-2 text-sm transition-colors"
        >
          <Icon name="plus" size="sm" class="user-balance-history-modal__action-icon user-balance-history-modal__action-icon--deposit" :stroke-width="2" />
          {{ t('admin.users.deposit') }}
        </button>
        <!-- Withdraw button - matches menu style -->
        <button
          v-if="!hideActions"
          @click="emit('withdraw')"
          class="user-balance-history-modal__action-button user-balance-history-modal__action-button-layout user-balance-history-modal__action-button--withdraw flex items-center gap-2 text-sm transition-colors"
        >
          <svg class="user-balance-history-modal__action-icon user-balance-history-modal__action-icon--withdraw h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
          </svg>
          {{ t('admin.users.withdraw') }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="user-balance-history-modal__state-block flex justify-center">
        <Icon name="refresh" size="lg" class="user-balance-history-modal__loading-icon animate-spin" />
      </div>

      <!-- Empty state -->
      <div v-else-if="history.length === 0" class="user-balance-history-modal__state-block text-center">
        <p class="user-balance-history-modal__text-muted text-sm">{{ t('admin.users.noBalanceHistory') }}</p>
      </div>

      <!-- History list -->
      <div v-else class="user-balance-history-modal__list space-y-3 overflow-y-auto">
        <div
          v-for="item in history"
          :key="item.id"
          class="user-balance-history-modal__history-card"
        >
          <div class="flex items-start justify-between">
            <!-- Left: type icon + description -->
            <div class="flex items-start gap-3">
              <div
                :class="getIconContainerClasses(item)"
              >
                <Icon :name="getIconName(item)" size="sm" :class="getIconClasses(item)" />
              </div>
              <div>
                <p class="user-balance-history-modal__text-strong text-sm font-medium">
                  {{ getItemTitle(item) }}
                </p>
                <!-- Notes (admin adjustment reason) -->
                <p
                  v-if="item.notes"
                  class="user-balance-history-modal__text-muted mt-0.5 text-xs"
                  :title="item.notes"
                >
                  {{ item.notes.length > 60 ? item.notes.substring(0, 55) + '...' : item.notes }}
                </p>
                <p class="user-balance-history-modal__text-soft mt-0.5 text-xs">
                  {{ formatDateTime(item.used_at || item.created_at) }}
                </p>
              </div>
            </div>
            <!-- Right: value -->
            <div class="text-right">
              <p :class="getValueClasses(item)">
                {{ formatValue(item) }}
              </p>
              <p
                v-if="isAdminType(item.type)"
                class="user-balance-history-modal__text-soft text-xs"
              >
                {{ t('redeem.adminAdjustment') }}
              </p>
              <p
                v-else
                class="user-balance-history-modal__text-soft font-mono text-xs"
              >
                {{ item.code.slice(0, 8) }}...
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-center gap-2 pt-2">
        <button
          :disabled="currentPage <= 1"
          class="btn btn-secondary user-balance-history-modal__pagination-button text-sm"
          @click="loadHistory(currentPage - 1)"
        >
          {{ t('pagination.previous') }}
        </button>
        <span class="user-balance-history-modal__text-muted text-sm">
          {{ currentPage }} / {{ totalPages }}
        </span>
        <button
          :disabled="currentPage >= totalPages"
          class="btn btn-secondary user-balance-history-modal__pagination-button text-sm"
          @click="loadHistory(currentPage + 1)"
        >
          {{ t('pagination.next') }}
        </button>
      </div>
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI, type BalanceHistoryItem } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean; user: AdminUser | null; hideActions?: boolean }>()
const emit = defineEmits(['close', 'deposit', 'withdraw'])
const { t } = useI18n()

const history = ref<BalanceHistoryItem[]>([])
const loading = ref(false)
const currentPage = ref(1)
const total = ref(0)
const totalRecharged = ref(0)
const pageSize = 15
const typeFilter = ref('')

const totalPages = computed(() => Math.ceil(total.value / pageSize) || 1)

type HistoryTone = 'success' | 'danger' | 'brand-purple' | 'info' | 'brand-orange'

// Type filter options
const typeOptions = computed(() => [
  { value: '', label: t('admin.users.allTypes') },
  { value: 'balance', label: t('admin.users.typeBalance') },
  { value: 'admin_balance', label: t('admin.users.typeAdminBalance') },
  { value: 'concurrency', label: t('admin.users.typeConcurrency') },
  { value: 'admin_concurrency', label: t('admin.users.typeAdminConcurrency') },
  { value: 'subscription', label: t('admin.users.typeSubscription') }
])

// Watch modal open
watch(() => props.show, (v) => {
  if (v && props.user) {
    typeFilter.value = ''
    loadHistory(1)
  }
}, { immediate: true })

async function loadHistory(page: number) {
  if (!props.user) return
  loading.value = true
  currentPage.value = page
  try {
    const res = await adminAPI.users.getUserBalanceHistory(
      props.user.id,
      page,
      pageSize,
      typeFilter.value || undefined
    )
    history.value = res.items || []
    total.value = res.total || 0
    totalRecharged.value = res.total_recharged || 0
  } catch (error) {
    console.error('Failed to load balance history:', error)
  } finally {
    loading.value = false
  }
}

// Helper: check if admin type
const isAdminType = (type: string) => type === 'admin_balance' || type === 'admin_concurrency'

// Helper: check if balance type (includes admin_balance)
const isBalanceType = (type: string) => type === 'balance' || type === 'admin_balance'

// Helper: check if subscription type
const isSubscriptionType = (type: string) => type === 'subscription'

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const getItemTone = (item: BalanceHistoryItem): HistoryTone => {
  if (isBalanceType(item.type)) {
    return item.value >= 0 ? 'success' : 'danger'
  }
  if (isSubscriptionType(item.type)) {
    return 'brand-purple'
  }
  return item.value >= 0 ? 'info' : 'brand-orange'
}

// Icon name based on type
const getIconName = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) return 'dollar'
  if (isSubscriptionType(item.type)) return 'badge'
  return 'bolt' // concurrency
}

const getIconContainerClasses = (item: BalanceHistoryItem) => {
  const tone = getItemTone(item)
  return joinClassNames(
    'user-balance-history-modal__tone-surface user-balance-history-modal__tone-surface-shape flex flex-shrink-0 items-center justify-center',
    `user-balance-history-modal__tone-surface--${tone}`
  )
}

const getIconClasses = (item: BalanceHistoryItem) => {
  return joinClassNames(
    'user-balance-history-modal__tone-text',
    `user-balance-history-modal__tone-text--${getItemTone(item)}`
  )
}

const getValueClasses = (item: BalanceHistoryItem) => {
  return joinClassNames(
    'user-balance-history-modal__value text-sm font-semibold',
    `user-balance-history-modal__tone-text--${getItemTone(item)}`
  )
}

// Item title
const getItemTitle = (item: BalanceHistoryItem) => {
  switch (item.type) {
    case 'balance':
      return t('redeem.balanceAddedRedeem')
    case 'admin_balance':
      return item.value >= 0 ? t('redeem.balanceAddedAdmin') : t('redeem.balanceDeductedAdmin')
    case 'concurrency':
      return t('redeem.concurrencyAddedRedeem')
    case 'admin_concurrency':
      return item.value >= 0 ? t('redeem.concurrencyAddedAdmin') : t('redeem.concurrencyReducedAdmin')
    case 'subscription':
      return t('redeem.subscriptionAssigned')
    default:
      return t('common.unknown')
  }
}

// Format display value
const formatValue = (item: BalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  }
  if (isSubscriptionType(item.type)) {
    const days = item.validity_days || Math.round(item.value)
    const groupName = item.group?.name || ''
    return groupName ? `${days}d - ${groupName}` : `${days}d`
  }
  // concurrency types
  const sign = item.value >= 0 ? '+' : ''
  return `${sign}${item.value}`
}
</script>

<style scoped>
.user-balance-history-modal__hero,
.user-balance-history-modal__history-card,
.user-balance-history-modal__action-button {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: var(--theme-surface);
}

.user-balance-history-modal__hero {
  border-radius: calc(var(--theme-surface-radius) + 8px);
  background: color-mix(in srgb, var(--theme-surface-soft) 78%, var(--theme-surface));
  padding: var(--theme-balance-history-hero-padding);
}

.user-balance-history-modal__hero-footer {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 62%, transparent);
}

.user-balance-history-modal__hero-footer-layout {
  margin-top: var(--theme-balance-history-hero-footer-margin-top);
  padding-top: var(--theme-balance-history-hero-footer-padding-top);
}

.user-balance-history-modal__filter {
  width: var(--theme-balance-history-filter-width);
}

.user-balance-history-modal__list {
  max-height: var(--theme-balance-history-list-max-height);
}

.user-balance-history-modal__avatar {
  background: color-mix(in srgb, var(--theme-accent-soft) 82%, var(--theme-surface));
}

.user-balance-history-modal__avatar-shape {
  width: var(--theme-balance-history-avatar-size);
  height: var(--theme-balance-history-avatar-size);
  border-radius: 999px;
}

.user-balance-history-modal__avatar-text,
.user-balance-history-modal__loading-icon {
  color: color-mix(in srgb, var(--theme-accent) 84%, var(--theme-page-text));
}

.user-balance-history-modal__text-strong {
  color: var(--theme-page-text);
}

.user-balance-history-modal__text-muted {
  color: var(--theme-page-muted);
}

.user-balance-history-modal__text-soft {
  color: color-mix(in srgb, var(--theme-page-muted) 74%, transparent);
}

.user-balance-history-modal__text-success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__action-button {
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-button-secondary-text);
  box-shadow: var(--theme-card-shadow);
  transition: background-color 0.2s ease, color 0.2s ease, box-shadow 0.2s ease;
}

.user-balance-history-modal__action-button-layout {
  padding: var(--theme-balance-history-action-padding-y) var(--theme-balance-history-action-padding-x);
}

.user-balance-history-modal__action-button:hover {
  background: var(--theme-button-secondary-hover-bg);
  box-shadow: var(--theme-card-shadow-hover);
}

.user-balance-history-modal__action-icon--deposit {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__action-icon--withdraw {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__history-card {
  border-radius: calc(var(--theme-surface-radius) + 8px);
  box-shadow: var(--theme-card-shadow);
  padding: var(--theme-balance-history-card-padding);
}

.user-balance-history-modal__tone-surface {
  --user-balance-history-tone-rgb: var(--theme-info-rgb);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, rgb(var(--user-balance-history-tone-rgb)) 12%, var(--theme-surface));
}

.user-balance-history-modal__tone-surface-shape {
  width: var(--theme-balance-history-tone-surface-size);
  height: var(--theme-balance-history-tone-surface-size);
}

.user-balance-history-modal__state-block {
  padding-block: var(--theme-balance-history-state-padding-y);
}

.user-balance-history-modal__pagination-button {
  padding: var(--theme-balance-history-pagination-padding-y)
    var(--theme-balance-history-pagination-padding-x);
}

.user-balance-history-modal__tone-surface--success {
  --user-balance-history-tone-rgb: var(--theme-success-rgb);
}

.user-balance-history-modal__tone-surface--danger {
  --user-balance-history-tone-rgb: var(--theme-danger-rgb);
}

.user-balance-history-modal__tone-surface--brand-purple {
  --user-balance-history-tone-rgb: var(--theme-brand-purple-rgb);
}

.user-balance-history-modal__tone-surface--info {
  --user-balance-history-tone-rgb: var(--theme-info-rgb);
}

.user-balance-history-modal__tone-surface--brand-orange {
  --user-balance-history-tone-rgb: var(--theme-brand-orange-rgb);
}

.user-balance-history-modal__tone-text--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__tone-text--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__tone-text--brand-purple {
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__tone-text--info {
  color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 84%, var(--theme-page-text));
}

.user-balance-history-modal__tone-text--brand-orange {
  color: color-mix(in srgb, rgb(var(--theme-brand-orange-rgb)) 84%, var(--theme-page-text));
}
</style>
