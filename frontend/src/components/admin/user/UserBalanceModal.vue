<template>
  <BaseDialog :show="show" :title="operation === 'add' ? t('admin.users.deposit') : t('admin.users.withdraw')" width="narrow" @close="handleClose">
    <form v-if="user" id="balance-form" @submit.prevent="handleBalanceSubmit" class="space-y-5">
      <div class="user-balance-modal__summary flex items-center gap-3">
        <div class="user-balance-modal__avatar flex h-10 w-10 items-center justify-center">
          <span class="user-balance-modal__avatar-text text-lg font-medium">
            {{ user.email.charAt(0).toUpperCase() }}
          </span>
        </div>
        <div class="flex-1">
          <p class="user-balance-modal__email font-medium">{{ user.email }}</p>
          <p class="user-balance-modal__hint text-sm">
            {{ t('admin.users.currentBalance') }}: ${{ formatBalance(user.balance) }}
          </p>
        </div>
      </div>
      <div>
        <label class="input-label">{{ operation === 'add' ? t('admin.users.depositAmount') : t('admin.users.withdrawAmount') }}</label>
        <div class="relative flex gap-2">
          <div class="relative flex-1">
            <div class="user-balance-modal__currency absolute left-3 top-1/2 -translate-y-1/2 font-medium">$</div>
            <input v-model.number="form.amount" type="number" step="any" min="0" required class="input pl-8" />
          </div>
          <button v-if="operation === 'subtract'" type="button" @click="fillAllBalance" class="btn btn-secondary whitespace-nowrap">{{ t('admin.users.withdrawAll') }}</button>
        </div>
      </div>
      <div><label class="input-label">{{ t('admin.users.notes') }}</label><textarea v-model="form.notes" rows="3" class="input"></textarea></div>
      <div v-if="form.amount > 0" class="user-balance-modal__preview border">
        <div class="flex items-center justify-between text-sm">
          <span class="user-balance-modal__preview-label">{{ t('admin.users.newBalance') }}:</span>
          <span class="user-balance-modal__preview-value font-bold">
            ${{ formatBalance(calculateNewBalance()) }}
          </span>
        </div>
      </div>
    </form>
    <template #footer>
      <div class="flex justify-end gap-3">
        <button @click="handleClose" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button
          type="submit"
          form="balance-form"
          :disabled="submitting || !form.amount"
          class="btn"
          :class="operation === 'add' ? 'btn-success' : 'btn-danger'"
        >
          {{ submitting ? t('common.saving') : t('common.confirm') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { resolveErrorMessage } from '@/utils/errorMessage'

const props = defineProps<{
  show: boolean
  user: AdminUser | null
  operation: 'add' | 'subtract'
}>()

const emit = defineEmits(['close', 'success'])
const { t } = useI18n()
const appStore = useAppStore()

const submitting = ref(false)
const form = reactive({ amount: 0, notes: '' })
let balanceRequestSequence = 0

const resetForm = () => {
  form.amount = 0
  form.notes = ''
}

watch(
  () => [props.show, props.user?.id, props.operation] as const,
  ([visible, userId]) => {
    balanceRequestSequence += 1
    submitting.value = false
    if (!visible || userId == null) {
      resetForm()
      return
    }

    resetForm()
  },
  { immediate: true }
)

const formatBalance = (value: number) => {
  if (value === 0) return '0.00'
  const formatted = value.toFixed(8).replace(/\.?0+$/, '')
  const parts = formatted.split('.')
  if (parts.length === 1) return formatted + '.00'
  if (parts[1].length === 1) return formatted + '0'
  return formatted
}

const fillAllBalance = () => {
  if (props.user) {
    form.amount = props.user.balance
  }
}

const calculateNewBalance = () => {
  if (!props.user) return 0
  const result = props.operation === 'add' ? props.user.balance + form.amount : props.user.balance - form.amount
  return Math.abs(result) < 1e-10 ? 0 : result
}

const handleBalanceSubmit = async () => {
  if (!props.user) return
  if (!form.amount || form.amount <= 0) {
    appStore.showError(t('admin.users.amountRequired'))
    return
  }
  if (props.operation === 'subtract' && form.amount > props.user.balance) {
    appStore.showError(t('admin.users.insufficientBalance'))
    return
  }
  const requestSequence = ++balanceRequestSequence
  const userId = props.user.id
  const amount = form.amount
  const notes = form.notes
  const operation = props.operation
  submitting.value = true
  try {
    await adminAPI.users.updateBalance(userId, amount, operation, notes)
    if (requestSequence !== balanceRequestSequence || !props.show || props.user?.id !== userId) {
      return
    }
    appStore.showSuccess(t('common.success'))
    emit('success')
    emit('close')
  } catch (error) {
    if (requestSequence !== balanceRequestSequence || !props.show || props.user?.id !== userId) {
      return
    }
    console.error('Failed to update balance:', error)
    appStore.showError(resolveErrorMessage(error, t('common.error')))
  } finally {
    if (requestSequence === balanceRequestSequence) {
      submitting.value = false
    }
  }
}

const handleClose = () => {
  balanceRequestSequence += 1
  submitting.value = false
  resetForm()
  emit('close')
}
</script>

<style scoped>
.user-balance-modal__summary {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.user-balance-modal__avatar {
  border-radius: var(--theme-version-icon-radius);
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
}

.user-balance-modal__avatar-text {
  color: var(--theme-accent);
}

.user-balance-modal__email,
.user-balance-modal__preview-value {
  color: var(--theme-page-text);
}

.user-balance-modal__hint,
.user-balance-modal__currency,
.user-balance-modal__preview-label {
  color: var(--theme-page-muted);
}

.user-balance-modal__preview {
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-markdown-block-padding);
  border-color: color-mix(in srgb, rgb(var(--theme-info-rgb)) 26%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-info-rgb)) 10%, var(--theme-surface));
}
</style>
