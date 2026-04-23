<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h3 class="text-lg font-medium text-gray-900 dark:text-white">
        {{ t('admin.settings.notify.title') }}
      </h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.notify.description') }}
      </p>
    </div>

    <div class="space-y-6 px-6 py-6">
      <div class="grid gap-4 md:grid-cols-2">
        <label class="flex items-center justify-between gap-4 rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <div>
            <div class="font-medium">{{ t('admin.settings.notify.balanceEnabled') }}</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.notify.balanceEnabledHint') }}</div>
          </div>
          <input v-model="form.balance_low_notify_enabled" type="checkbox" class="toggle" />
        </label>

        <label class="flex items-center justify-between gap-4 rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <div>
            <div class="font-medium">{{ t('admin.settings.notify.quotaEnabled') }}</div>
            <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.notify.quotaEnabledHint') }}</div>
          </div>
          <input v-model="form.account_quota_notify_enabled" type="checkbox" class="toggle" />
        </label>
      </div>

      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label class="input-label">{{ t('admin.settings.notify.balanceThreshold') }}</label>
          <input v-model.number="form.balance_low_notify_threshold" type="number" min="0" step="0.01" class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('admin.settings.notify.rechargeUrl') }}</label>
          <input v-model.trim="form.balance_low_notify_recharge_url" type="url" class="input" />
        </div>
      </div>

      <div>
        <div class="mb-2 flex items-center justify-between">
          <label class="input-label mb-0">{{ t('admin.settings.notify.quotaEmails') }}</label>
          <button type="button" class="btn btn-secondary btn-sm" @click="addEntry">
            {{ t('common.add') }}
          </button>
        </div>

        <div class="space-y-2">
          <div v-for="(entry, index) in form.account_quota_notify_emails" :key="`${entry.email}-${index}`" class="grid gap-2 rounded-lg border border-gray-200 p-3 dark:border-dark-600 md:grid-cols-[1fr_auto_auto_auto]">
            <input v-model.trim="entry.email" type="email" class="input" :placeholder="t('admin.settings.notify.emailPlaceholder')" />
            <label class="flex items-center gap-2 text-sm"><input v-model="entry.verified" type="checkbox" />{{ t('profile.balanceNotify.verified') }}</label>
            <label class="flex items-center gap-2 text-sm"><input v-model="entry.disabled" type="checkbox" />{{ t('common.disabled') }}</label>
            <button type="button" class="btn btn-danger btn-sm" @click="removeEntry(index)">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { SettingsForm } from './settingsForm'

const props = defineProps<{ form: SettingsForm }>()
const { t } = useI18n()

function ensureEntries() {
  if (!Array.isArray(props.form.account_quota_notify_emails)) {
    props.form.account_quota_notify_emails = []
  }
  return props.form.account_quota_notify_emails
}

function addEntry() {
  ensureEntries().push({ email: '', disabled: false, verified: false })
}

function removeEntry(index: number) {
  ensureEntries().splice(index, 1)
}
</script>
