<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.purchase.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.purchase.description') }}
      </p>
    </div>
    <div class="space-y-6 p-6">
      <div class="flex items-center justify-between">
        <div>
          <label class="font-medium text-gray-900 dark:text-white">
            {{ t('admin.settings.purchase.enabled') }}
          </label>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.settings.purchase.enabledHint') }}
          </p>
        </div>
        <Toggle v-model="form.purchase_subscription_enabled" />
      </div>

      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.settings.purchase.url') }}
        </label>
        <input
          v-model="form.purchase_subscription_url"
          type="url"
          class="input font-mono text-sm"
          :placeholder="t('admin.settings.purchase.urlPlaceholder')"
        />
        <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.settings.purchase.urlHint') }}
        </p>
        <p class="mt-2 text-xs text-amber-600 dark:text-amber-400">
          {{ t('admin.settings.purchase.iframeWarning') }}
        </p>
      </div>

      <div class="flex items-center gap-2 text-sm">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          class="h-4 w-4 shrink-0 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <a
          href="https://raw.githubusercontent.com/senran-N/sub2api/main/docs/ADMIN_PAYMENT_INTEGRATION_API.md"
          target="_blank"
          rel="noopener noreferrer"
          class="text-blue-600 hover:underline dark:text-blue-400"
          download="ADMIN_PAYMENT_INTEGRATION_API.md"
        >
          {{ t('admin.settings.purchase.integrationDoc') }}
        </a>
        <span class="text-gray-400 dark:text-gray-500">-</span>
        <span class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.settings.purchase.integrationDocHint') }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsForm } from '../settingsForm'

defineProps<{
  form: SettingsForm
}>()

const { t } = useI18n()
</script>
