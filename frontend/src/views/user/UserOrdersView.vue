<template>
  <AppLayout>
    <div class="user-orders-view">
      <div v-if="ordersUnavailable" class="user-orders-view__state">
        <Icon name="document" size="xl" class="user-orders-view__state-icon" />
        <h2 class="user-orders-view__state-title">{{ t('purchase.notEnabledTitle') }}</h2>
        <p class="user-orders-view__state-description">{{ unavailableMessage }}</p>
      </div>
      <template v-else>
      <!-- Filters -->
      <div class="user-orders-view__toolbar">
        <div class="user-orders-view__toolbar-row">
          <Select v-model="currentFilter" :options="statusFilters" class="w-36" @change="fetchOrders" />
          <div class="user-orders-view__toolbar-actions">
            <button @click="fetchOrders" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button class="btn btn-primary" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <OrderTable :orders="orders" :loading="loading">
        <template #actions="{ row }">
          <div class="user-orders-view__row-actions">
            <button v-if="row.status === 'PENDING'" @click="handleCancel(row.id)" class="user-orders-view__row-action user-orders-view__row-action--warning">
              <Icon name="x" size="sm" />
              <span>{{ t('payment.orders.cancel') }}</span>
            </button>
            <button v-if="canRequestRefund(row)" @click="openRefundDialog(row)" class="user-orders-view__row-action user-orders-view__row-action--accent">
              <Icon name="dollar" size="sm" />
              <span>{{ t('payment.orders.requestRefund') }}</span>
            </button>
          </div>
        </template>
      </OrderTable>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
      </template>
    </div>

    <!-- Cancel Confirm Dialog -->
    <BaseDialog :show="!!cancelTargetId" :title="t('payment.orders.cancel')" width="narrow" @close="cancelTargetId = null">
      <p class="user-orders-view__dialog-copy">{{ t('payment.confirmCancel') }}</p>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="cancelTargetId = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">{{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- Refund Dialog -->
    <BaseDialog :show="!!refundTarget" :title="t('payment.orders.requestRefund')" @close="refundTarget = null">
      <div v-if="refundTarget" class="space-y-4">
        <div class="user-orders-view__dialog-summary">
          <div class="user-orders-view__dialog-summary-row">
            <span class="user-orders-view__dialog-summary-label">{{ t('payment.orders.orderId') }}</span>
            <span class="user-orders-view__dialog-summary-value user-orders-view__dialog-summary-value--mono">#{{ refundTarget.id }}</span>
          </div>
          <div class="user-orders-view__dialog-summary-row user-orders-view__dialog-summary-row--spaced">
            <span class="user-orders-view__dialog-summary-label">{{ t('payment.orders.amount') }}</span>
            <span class="user-orders-view__dialog-summary-value">${{ refundTarget.amount.toFixed(2) }}</span>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('payment.refundReason') }}</label>
          <textarea v-model="refundReason" rows="3" class="input mt-1 w-full" :placeholder="t('payment.refundReasonPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button class="btn btn-secondary" @click="refundTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading || !refundReason.trim()" @click="confirmRefund">{{ actionLoading ? t('common.processing') : t('payment.orders.requestRefund') }}</button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import { hasResponseStatus } from '@/utils/requestError'

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const loading = ref(false)
const actionLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const ordersUnavailable = ref(false)
const refundEligibleProviders = ref<Set<string>>(new Set())
const currentFilter = ref('')
const cancelTargetId = ref<number | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])
const unavailableMessage = computed(() => t('purchase.notEnabledDesc'))

async function fetchOrders() {
  loading.value = true
  try {
    const res = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    ordersUnavailable.value = false
    orders.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    if (hasResponseStatus(err, 404)) {
      ordersUnavailable.value = true
      orders.value = []
      pagination.total = 0
      return
    }

    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchOrders() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchOrders() }

function handleCancel(orderId: number) { cancelTargetId.value = orderId }

async function confirmCancel() {
  if (!cancelTargetId.value) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function openRefundDialog(order: PaymentOrder) { refundTarget.value = order; refundReason.value = '' }

async function confirmRefund() {
  if (!refundTarget.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    await paymentAPI.requestRefund(refundTarget.value.id, { reason: refundReason.value.trim() })
    appStore.showSuccess(t('common.success'))
    refundTarget.value = null
    refundReason.value = ''
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function canRequestRefund(order: PaymentOrder): boolean {
  if (order.status !== 'COMPLETED') return false
  if (!order.provider_instance_id) return false
  return refundEligibleProviders.value.has(order.provider_instance_id)
}

async function loadRefundEligibility() {
  try {
    const res = await paymentAPI.getRefundEligibleProviders()
    refundEligibleProviders.value = new Set(res.data.provider_instance_ids || [])
  } catch { /* ignore — default to hiding refund button */ }
}

onMounted(() => { fetchOrders(); loadRefundEligibility() })
</script>

<style scoped>
.user-orders-view {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.user-orders-view__state {
  border: var(--theme-card-border-width) solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 8px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  padding: 2rem 1.5rem;
  text-align: center;
}

.user-orders-view__state-icon {
  margin: 0 auto 0.75rem;
  color: var(--theme-page-muted);
}

.user-orders-view__state-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--theme-page-text);
}

.user-orders-view__state-description {
  margin-top: 0.5rem;
  font-size: 0.875rem;
  color: var(--theme-page-muted);
}

.user-orders-view__toolbar {
  border: var(--theme-card-border-width) solid var(--theme-card-border);
  border-radius: calc(var(--theme-surface-radius) + 8px);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow);
  padding: 1rem;
}

.user-orders-view__toolbar-row,
.user-orders-view__toolbar-actions,
.user-orders-view__row-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.user-orders-view__toolbar-actions {
  flex: 1;
  justify-content: flex-end;
}

.user-orders-view__row-action {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  border: 1px solid transparent;
  border-radius: calc(var(--theme-button-radius) + 4px);
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.user-orders-view__row-action--warning {
  border-color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 8%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}

.user-orders-view__row-action--accent {
  border-color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 24%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 8%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-brand-purple-rgb)) 84%, var(--theme-page-text));
}

.user-orders-view__dialog-copy,
.user-orders-view__dialog-summary-label {
  color: var(--theme-page-muted);
}

.user-orders-view__dialog-copy {
  font-size: 0.875rem;
}

.user-orders-view__dialog-summary {
  border-radius: calc(var(--theme-surface-radius) + 6px);
  background: color-mix(in srgb, var(--theme-surface-soft) 82%, var(--theme-surface));
  padding: 1rem;
}

.user-orders-view__dialog-summary-row {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  font-size: 0.875rem;
}

.user-orders-view__dialog-summary-row--spaced {
  margin-top: 0.5rem;
}

.user-orders-view__dialog-summary-value {
  font-weight: 600;
  color: var(--theme-page-text);
}

.user-orders-view__dialog-summary-value--mono {
  font-family: var(--theme-font-mono);
}
</style>
