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
      <div class="rounded-lg bg-gray-50 p-4 dark:bg-dark-700">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('admin.subscriptions.adjustingFor') }}
          <span class="font-medium text-gray-900 dark:text-white">{{ subscription.user?.email }}</span>
        </p>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          {{ t('admin.subscriptions.currentExpiration') }}:
          <span class="font-medium text-gray-900 dark:text-white">
            {{
              subscription.expires_at
                ? formatDateOnly(subscription.expires_at)
                : t('admin.subscriptions.noExpiration')
            }}
          </span>
        </p>
        <p v-if="subscription.expires_at" class="mt-1 text-sm text-gray-600 dark:text-gray-400">
          {{ t('admin.subscriptions.remainingDays') }}:
          <span class="font-medium text-gray-900 dark:text-white">
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
