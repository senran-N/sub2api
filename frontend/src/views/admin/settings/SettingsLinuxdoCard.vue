<template>
  <div class="card">
    <div class="settings-linuxdo-card__header">
      <h2 class="settings-linuxdo-card__title text-lg font-semibold">
        {{ t('admin.settings.linuxdo.title') }}
      </h2>
      <p class="settings-linuxdo-card__description mt-1 text-sm">
        {{ t('admin.settings.linuxdo.description') }}
      </p>
    </div>
    <div class="settings-linuxdo-card__body space-y-5">
      <div class="flex items-center justify-between">
        <div>
          <label class="settings-linuxdo-card__label font-medium">
            {{ t('admin.settings.linuxdo.enable') }}
          </label>
          <p class="settings-linuxdo-card__description text-sm">
            {{ t('admin.settings.linuxdo.enableHint') }}
          </p>
        </div>
        <Toggle v-model="form.linuxdo_connect_enabled" />
      </div>

      <div
        v-if="form.linuxdo_connect_enabled"
        class="settings-linuxdo-card__section pt-4"
      >
        <div class="grid grid-cols-1 gap-6">
          <div>
            <label class="settings-linuxdo-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.linuxdo.clientId') }}
            </label>
            <input
              v-model="form.linuxdo_connect_client_id"
              type="text"
              class="input font-mono text-sm"
              :placeholder="t('admin.settings.linuxdo.clientIdPlaceholder')"
            />
            <p class="settings-linuxdo-card__description mt-1.5 text-xs">
              {{ t('admin.settings.linuxdo.clientIdHint') }}
            </p>
          </div>

          <div>
            <label class="settings-linuxdo-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.linuxdo.clientSecret') }}
            </label>
            <input
              v-model="form.linuxdo_connect_client_secret"
              type="password"
              class="input font-mono text-sm"
              :placeholder="
                form.linuxdo_connect_client_secret_configured
                  ? t('admin.settings.linuxdo.clientSecretConfiguredPlaceholder')
                  : t('admin.settings.linuxdo.clientSecretPlaceholder')
              "
            />
            <p class="settings-linuxdo-card__description mt-1.5 text-xs">
              {{
                form.linuxdo_connect_client_secret_configured
                  ? t('admin.settings.linuxdo.clientSecretConfiguredHint')
                  : t('admin.settings.linuxdo.clientSecretHint')
              }}
            </p>
          </div>

          <div>
            <label class="settings-linuxdo-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.linuxdo.redirectUrl') }}
            </label>
            <input
              v-model="form.linuxdo_connect_redirect_url"
              type="url"
              class="input font-mono text-sm"
              :placeholder="t('admin.settings.linuxdo.redirectUrlPlaceholder')"
            />
            <div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
              <button
                type="button"
                class="btn btn-secondary btn-sm w-fit"
                @click="$emit('quick-set-copy')"
              >
                {{ t('admin.settings.linuxdo.quickSetCopy') }}
              </button>
              <code
                v-if="redirectUrlSuggestion"
                class="settings-linuxdo-card__suggestion select-all break-all font-mono text-xs"
              >
                {{ redirectUrlSuggestion }}
              </code>
            </div>
            <p class="settings-linuxdo-card__description mt-1.5 text-xs">
              {{ t('admin.settings.linuxdo.redirectUrlHint') }}
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
import type { SettingsLinuxdoFields } from './settingsForm'

defineProps<{
  form: SettingsLinuxdoFields
  redirectUrlSuggestion: string
}>()

defineEmits<{
  'quick-set-copy': []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-linuxdo-card__header,
.settings-linuxdo-card__body,
.settings-linuxdo-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-linuxdo-card__header {
  padding:
    var(--theme-settings-card-header-padding-y)
    var(--theme-settings-card-header-padding-x);
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-linuxdo-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-linuxdo-card__title,
.settings-linuxdo-card__label,
.settings-linuxdo-card__field-label {
  color: var(--theme-page-text);
}

.settings-linuxdo-card__description {
  color: var(--theme-page-muted);
}

.settings-linuxdo-card__suggestion {
  border-radius: var(--theme-settings-inline-button-radius);
  padding:
    var(--theme-settings-code-padding-y)
    var(--theme-settings-code-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}
</style>
