<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.scheduling.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.scheduling.description') }}
      </p>
    </div>
    <div class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.settings.scheduling.allowUngroupedKey') }}
          </label>
          <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.settings.scheduling.allowUngroupedKeyHint') }}
          </p>
        </div>
        <Toggle v-model="form.allow_ungrouped_key_scheduling" />
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
