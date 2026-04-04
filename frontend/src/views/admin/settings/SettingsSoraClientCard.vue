<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.soraClient.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.soraClient.description') }}
      </p>
    </div>
    <div class="space-y-6 p-6">
      <div class="flex items-center justify-between">
        <div>
          <label class="font-medium text-gray-900 dark:text-white">
            {{ t('admin.settings.soraClient.enabled') }}
          </label>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.settings.soraClient.enabledHint') }}
          </p>
        </div>
        <Toggle v-model="form.sora_client_enabled" />
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
