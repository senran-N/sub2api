<template>
  <BaseDialog
    :show="show"
    :title="t('admin.subscriptions.adjustSubscription')"
    width="narrow"
    @close="emit('close')"
  >
    <form
      v-if="subscription"
      id="extend-subscription-form"
      class="space-y-5"
      @submit.prevent="emit('submit')"
    >
      <div class="subscription-extend-dialog__summary">
        <p class="subscription-extend-dialog__description text-sm">
          {{ t('admin.subscriptions.adjustingFor') }}
          <span class="subscription-extend-dialog__value font-medium">{{ subscription.user?.email }}</span>
        </p>
        <p class="subscription-extend-dialog__description mt-1 text-sm">
          {{ t('admin.subscriptions.currentExpiration') }}:
          <span class="subscription-extend-dialog__value font-medium">
            {{
              subscription.expires_at
                ? formatDateOnly(subscription.expires_at)
                : t('admin.subscriptions.noExpiration')
            }}
          </span>
        </p>
        <p v-if="subscription.expires_at" class="subscription-extend-dialog__description mt-1 text-sm">
          {{ t('admin.subscriptions.remainingDays') }}:
          <span class="subscription-extend-dialog__value font-medium">
            {{ getSubscriptionDaysRemaining(subscription.expires_at) ?? 0 }}
          </span>
        </p>
      </div>
      <div>
        <label class="input-label">{{ t('admin.subscriptions.form.adjustDays') }}</label>
        <div class="flex items-center gap-2">
          <input
            v-model.number="form.days"
            type="number"
            required
            class="input text-center"
            :placeholder="t('admin.subscriptions.adjustDaysPlaceholder')"
          />
        </div>
        <p class="input-hint">{{ t('admin.subscriptions.adjustHint') }}</p>
      </div>
    </form>

    <template #footer>
      <div v-if="subscription" class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button
          type="submit"
          form="extend-subscription-form"
          :disabled="submitting"
          class="btn btn-primary"
        >
          {{ submitting ? t('admin.subscriptions.adjusting') : t('admin.subscriptions.adjust') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateOnly } from '@/utils/format'
import BaseDialog from '@/components/common/BaseDialog.vue'
import type { UserSubscription } from '@/types'
import type { ExtendSubscriptionForm } from '../subscriptionForm'
import { getSubscriptionDaysRemaining } from '../subscriptionForm'

defineProps<{
  show: boolean
  subscription: UserSubscription | null
  form: ExtendSubscriptionForm
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.subscription-extend-dialog__summary {
  padding: var(--theme-markdown-block-padding);
  border-radius: var(--theme-subscription-panel-radius);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
}

.subscription-extend-dialog__description {
  color: var(--theme-page-muted);
}

.subscription-extend-dialog__value {
  color: var(--theme-page-text);
}
</style>
