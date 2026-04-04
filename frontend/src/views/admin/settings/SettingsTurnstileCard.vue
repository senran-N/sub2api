<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.turnstile.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.turnstile.description') }}
      </p>
    </div>
    <div class="space-y-5 p-6">
      <div class="flex items-center justify-between">
        <div>
          <label class="font-medium text-gray-900 dark:text-white">
            {{ t('admin.settings.turnstile.enableTurnstile') }}
          </label>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.settings.turnstile.enableTurnstileHint') }}
          </p>
        </div>
        <Toggle v-model="form.turnstile_enabled" />
      </div>

      <div
        v-if="form.turnstile_enabled"
        class="border-t border-gray-100 pt-4 dark:border-dark-700"
      >
        <div class="grid grid-cols-1 gap-6">
          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.turnstile.siteKey') }}
            </label>
            <input
              v-model="form.turnstile_site_key"
              type="text"
              class="input font-mono text-sm"
              placeholder="0x4AAAAAAA..."
            />
            <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.turnstile.siteKeyHint') }}
              <a
                href="https://dash.cloudflare.com/"
                target="_blank"
                rel="noopener noreferrer"
                class="text-primary-600 hover:text-primary-500"
              >
                {{ t('admin.settings.turnstile.cloudflareDashboard') }}
              </a>
            </p>
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.turnstile.secretKey') }}
            </label>
            <input
              v-model="form.turnstile_secret_key"
              type="password"
              class="input font-mono text-sm"
              placeholder="0x4AAAAAAA..."
            />
            <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
              {{
                form.turnstile_secret_key_configured
                  ? t('admin.settings.turnstile.secretKeyConfiguredHint')
                  : t('admin.settings.turnstile.secretKeyHint')
              }}
            </p>
          </div>
        </div>
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
