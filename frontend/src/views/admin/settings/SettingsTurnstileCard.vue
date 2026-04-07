<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-turnstile-card__title text-lg font-semibold">
        {{ t('admin.settings.turnstile.title') }}
      </h2>
      <p class="settings-turnstile-card__description mt-1 text-sm">
        {{ t('admin.settings.turnstile.description') }}
      </p>
    </div>
    <div class="settings-turnstile-card__body space-y-5">
      <div class="flex items-center justify-between">
        <div>
          <label class="settings-turnstile-card__title font-medium">
            {{ t('admin.settings.turnstile.enableTurnstile') }}
          </label>
          <p class="settings-turnstile-card__description text-sm">
            {{ t('admin.settings.turnstile.enableTurnstileHint') }}
          </p>
        </div>
        <Toggle v-model="form.turnstile_enabled" />
      </div>

      <div
        v-if="form.turnstile_enabled"
        class="settings-turnstile-card__section border-t pt-4"
      >
        <div class="grid grid-cols-1 gap-6">
          <div>
            <label class="settings-turnstile-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.turnstile.siteKey') }}
            </label>
            <input
              v-model="form.turnstile_site_key"
              type="text"
              class="input font-mono text-sm"
              placeholder="0x4AAAAAAA..."
            />
            <p class="settings-turnstile-card__description mt-1.5 text-xs">
              {{ t('admin.settings.turnstile.siteKeyHint') }}
              <a
                href="https://dash.cloudflare.com/"
                target="_blank"
                rel="noopener noreferrer"
                class="settings-turnstile-card__link"
              >
                {{ t('admin.settings.turnstile.cloudflareDashboard') }}
              </a>
            </p>
          </div>
          <div>
            <label class="settings-turnstile-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.turnstile.secretKey') }}
            </label>
            <input
              v-model="form.turnstile_secret_key"
              type="password"
              class="input font-mono text-sm"
              placeholder="0x4AAAAAAA..."
            />
            <p class="settings-turnstile-card__description mt-1.5 text-xs">
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
import type { SettingsTurnstileFields } from './settingsForm'

defineProps<{
  form: SettingsTurnstileFields
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-turnstile-card__title,
.settings-turnstile-card__field-label {
  color: var(--theme-page-text);
}

.settings-turnstile-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-turnstile-card__description {
  color: var(--theme-page-muted);
}

.settings-turnstile-card__section {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.settings-turnstile-card__link {
  color: var(--theme-accent);
}

.settings-turnstile-card__link:hover {
  color: color-mix(in srgb, var(--theme-accent) 82%, var(--theme-page-text));
}
</style>
