<template>
  <BaseDialog
    :show="show"
    :title="t('admin.accounts.tempUnschedulable.statusTitle')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <div v-if="loading" class="temp-unsched-status-modal__loading">
        <svg class="temp-unsched-status-modal__loading-spinner" fill="none" viewBox="0 0 24 24">
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

      <div v-else-if="!isActive" class="temp-unsched-status-modal__empty-state">
        {{ t('admin.accounts.tempUnschedulable.notActive') }}
      </div>

      <div v-else class="space-y-4">
        <div class="temp-unsched-status-modal__hint">
          {{ t('admin.accounts.recoverStateHint') }}
        </div>

        <div class="temp-unsched-status-modal__detail-card temp-unsched-status-modal__detail-card--wide">
          <p class="temp-unsched-status-modal__label">
            {{ t('admin.accounts.tempUnschedulable.accountName') }}
          </p>
          <p class="temp-unsched-status-modal__value mt-1">
            {{ account?.name || '-' }}
          </p>
        </div>

        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div class="temp-unsched-status-modal__detail-card">
            <p class="temp-unsched-status-modal__label">
              {{ t('admin.accounts.tempUnschedulable.triggeredAt') }}
            </p>
            <p class="temp-unsched-status-modal__value mt-1">
              {{ triggeredAtText }}
            </p>
          </div>
          <div class="temp-unsched-status-modal__detail-card">
            <p class="temp-unsched-status-modal__label">
              {{ t('admin.accounts.tempUnschedulable.until') }}
            </p>
            <p class="temp-unsched-status-modal__value mt-1">
              {{ untilText }}
            </p>
          </div>
          <div class="temp-unsched-status-modal__detail-card">
            <p class="temp-unsched-status-modal__label">
              {{ t('admin.accounts.tempUnschedulable.remaining') }}
            </p>
            <p class="temp-unsched-status-modal__value mt-1">
              {{ remainingText }}
            </p>
          </div>
          <div class="temp-unsched-status-modal__detail-card">
            <p class="temp-unsched-status-modal__label">
              {{ t('admin.accounts.tempUnschedulable.errorCode') }}
            </p>
            <p class="temp-unsched-status-modal__value mt-1">
              {{ state?.status_code || '-' }}
            </p>
          </div>
          <div class="temp-unsched-status-modal__detail-card">
            <p class="temp-unsched-status-modal__label">
              {{ t('admin.accounts.tempUnschedulable.matchedKeyword') }}
            </p>
            <p class="temp-unsched-status-modal__value mt-1">
              {{ state?.matched_keyword || '-' }}
            </p>
          </div>
          <div class="temp-unsched-status-modal__detail-card">
            <p class="temp-unsched-status-modal__label">
              {{ t('admin.accounts.tempUnschedulable.ruleOrder') }}
            </p>
            <p class="temp-unsched-status-modal__value mt-1">
              {{ ruleIndexDisplay }}
            </p>
          </div>
        </div>

        <div class="temp-unsched-status-modal__detail-card">
          <p class="temp-unsched-status-modal__label">
            {{ t('admin.accounts.tempUnschedulable.errorMessage') }}
          </p>
          <div class="temp-unsched-status-modal__message-box">
            {{ state?.error_message || '-' }}
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="handleClose">
          {{ t('common.close') }}
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="!isActive || resetting"
          @click="handleReset"
        >
          <svg
            v-if="resetting"
            class="-ml-1 mr-2 h-4 w-4 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
          >
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
          {{ t('admin.accounts.recoverState') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { Account, TempUnschedulableStatus } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{
  show: boolean
  account: Account | null
}>()

const emit = defineEmits<{
  close: []
  reset: [account: Account]
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const resetting = ref(false)
const status = ref<TempUnschedulableStatus | null>(null)

const state = computed(() => status.value?.state || null)

const getErrorMessage = (error: unknown, fallbackMessage: string) => {
  return error instanceof Error && error.message ? error.message : fallbackMessage
}

const isActive = computed(() => {
  if (!status.value?.active || !state.value) return false
  return state.value.until_unix * 1000 > Date.now()
})

const ruleIndexDisplay = computed(() => {
  if (!state.value) return '-'
  return state.value.rule_index + 1
})

const triggeredAtText = computed(() => {
  if (!state.value?.triggered_at_unix) return '-'
  return formatDateTime(new Date(state.value.triggered_at_unix * 1000))
})

const untilText = computed(() => {
  if (!state.value?.until_unix) return '-'
  return formatDateTime(new Date(state.value.until_unix * 1000))
})

const remainingText = computed(() => {
  if (!state.value) return '-'
  const remainingMs = state.value.until_unix * 1000 - Date.now()
  if (remainingMs <= 0) {
    return t('admin.accounts.tempUnschedulable.expired')
  }
  const minutes = Math.ceil(remainingMs / 60000)
  if (minutes < 60) {
    return t('admin.accounts.tempUnschedulable.remainingMinutes', { minutes })
  }
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60
  if (rest === 0) {
    return t('admin.accounts.tempUnschedulable.remainingHours', { hours })
  }
  return t('admin.accounts.tempUnschedulable.remainingHoursMinutes', { hours, minutes: rest })
})

const loadStatus = async () => {
  if (!props.account) return
  loading.value = true
  try {
    status.value = await adminAPI.accounts.getTempUnschedulableStatus(props.account.id)
  } catch (error) {
    appStore.showError(getErrorMessage(error, t('admin.accounts.tempUnschedulable.failedToLoad')))
    status.value = null
  } finally {
    loading.value = false
  }
}

const handleClose = () => {
  emit('close')
}

const handleReset = async () => {
  if (!props.account) return
  resetting.value = true
  try {
    const updated = await adminAPI.accounts.recoverState(props.account.id)
    appStore.showSuccess(t('admin.accounts.recoverStateSuccess'))
    emit('reset', updated)
    handleClose()
  } catch (error) {
    appStore.showError(getErrorMessage(error, t('admin.accounts.recoverStateFailed')))
  } finally {
    resetting.value = false
  }
}

watch(
  () => [props.show, props.account?.id],
  ([visible]) => {
    if (visible && props.account) {
      loadStatus()
      return
    }
    status.value = null
  }
)
</script>

<style scoped>
.temp-unsched-status-modal__loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem 0;
}

.temp-unsched-status-modal__loading-spinner {
  height: 1.5rem;
  width: 1.5rem;
  animation: spin 1s linear infinite;
  color: var(--theme-page-muted);
}

.temp-unsched-status-modal__empty-state,
.temp-unsched-status-modal__detail-card,
.temp-unsched-status-modal__hint {
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-button-radius) + 2px);
}

.temp-unsched-status-modal__empty-state {
  color: var(--theme-page-muted);
  font-size: 0.875rem;
  padding: 1rem;
}

.temp-unsched-status-modal__hint {
  border-color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 32%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 76%, var(--theme-page-text));
  font-size: 0.875rem;
  padding: 0.75rem;
}

.temp-unsched-status-modal__detail-card {
  background: var(--theme-surface);
  padding: 0.75rem;
}

.temp-unsched-status-modal__detail-card--wide {
  padding: 1rem;
}

.temp-unsched-status-modal__label {
  color: var(--theme-page-muted);
  font-size: 0.75rem;
}

.temp-unsched-status-modal__value {
  color: var(--theme-page-text);
  font-size: 0.875rem;
  font-weight: 600;
}

.temp-unsched-status-modal__message-box {
  margin-top: 0.5rem;
  border-radius: calc(var(--theme-button-radius) - 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
  font-size: 0.75rem;
  padding: 0.5rem;
}
</style>
