<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.streamTimeout.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.streamTimeout.description') }}
      </p>
    </div>
    <div class="space-y-5 p-6">
      <div v-if="loading" class="flex items-center gap-2 text-gray-500">
        <div class="h-4 w-4 animate-spin rounded-full border-b-2 border-primary-600"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div class="flex items-center justify-between">
          <div>
            <label class="font-medium text-gray-900 dark:text-white">
              {{ t('admin.settings.streamTimeout.enabled') }}
            </label>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.streamTimeout.enabledHint') }}
            </p>
          </div>
          <Toggle v-model="form.enabled" />
        </div>

        <div
          v-if="form.enabled"
          class="space-y-4 border-t border-gray-100 pt-4 dark:border-dark-700"
        >
          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.streamTimeout.action') }}
            </label>
            <select v-model="form.action" class="input w-64">
              <option value="temp_unsched">
                {{ t('admin.settings.streamTimeout.actionTempUnsched') }}
              </option>
              <option value="error">
                {{ t('admin.settings.streamTimeout.actionError') }}
              </option>
              <option value="none">
                {{ t('admin.settings.streamTimeout.actionNone') }}
              </option>
            </select>
            <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.streamTimeout.actionHint') }}
            </p>
          </div>

          <div v-if="form.action === 'temp_unsched'">
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.streamTimeout.tempUnschedMinutes') }}
            </label>
            <input
              v-model.number="form.temp_unsched_minutes"
              type="number"
              min="1"
              max="60"
              class="input w-32"
            />
            <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.streamTimeout.tempUnschedMinutesHint') }}
            </p>
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.streamTimeout.thresholdCount') }}
            </label>
            <input
              v-model.number="form.threshold_count"
              type="number"
              min="1"
              max="10"
              class="input w-32"
            />
            <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.streamTimeout.thresholdCountHint') }}
            </p>
          </div>

          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.streamTimeout.thresholdWindowMinutes') }}
            </label>
            <input
              v-model.number="form.threshold_window_minutes"
              type="number"
              min="1"
              max="60"
              class="input w-32"
            />
            <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.streamTimeout.thresholdWindowMinutesHint') }}
            </p>
          </div>
        </div>

        <div class="flex justify-end border-t border-gray-100 pt-4 dark:border-dark-700">
          <button
            type="button"
            :disabled="saving"
            class="btn btn-primary btn-sm"
            @click="$emit('save')"
          >
            <svg
              v-if="saving"
              class="mr-1 h-4 w-4 animate-spin"
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
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { StreamTimeoutSettings } from '@/api/admin/settings'
import Toggle from '@/components/common/Toggle.vue'

defineProps<{
  loading: boolean
  saving: boolean
  form: StreamTimeoutSettings
}>()

defineEmits<{
  save: []
}>()

const { t } = useI18n()
</script>
