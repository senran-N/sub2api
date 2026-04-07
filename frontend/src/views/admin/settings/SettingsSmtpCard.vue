<template>
  <div class="card">
    <div class="settings-smtp-card__header flex items-center justify-between">
      <div>
        <h2 class="settings-smtp-card__title text-lg font-semibold">
          {{ t('admin.settings.smtp.title') }}
        </h2>
        <p class="settings-smtp-card__description mt-1 text-sm">
          {{ t('admin.settings.smtp.description') }}
        </p>
      </div>
      <button
        type="button"
        class="btn btn-secondary btn-sm"
        :disabled="testing || disabled"
        @click="$emit('test-connection')"
      >
        <svg v-if="testing" class="settings-smtp-card__spinner h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
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
        {{ testing ? t('admin.settings.smtp.testing') : t('admin.settings.smtp.testConnection') }}
      </button>
    </div>
    <div class="settings-smtp-card__body space-y-6">
      <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
        <div>
          <label class="settings-smtp-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.smtp.host') }}
          </label>
          <input
            v-model="form.smtp_host"
            type="text"
            class="input"
            :placeholder="t('admin.settings.smtp.hostPlaceholder')"
          />
        </div>
        <div>
          <label class="settings-smtp-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.smtp.port') }}
          </label>
          <input
            v-model.number="form.smtp_port"
            type="number"
            min="1"
            max="65535"
            class="input"
            :placeholder="t('admin.settings.smtp.portPlaceholder')"
          />
        </div>
        <div>
          <label class="settings-smtp-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.smtp.username') }}
          </label>
          <input
            v-model="form.smtp_username"
            type="text"
            class="input"
            :placeholder="t('admin.settings.smtp.usernamePlaceholder')"
          />
        </div>
        <div>
          <label class="settings-smtp-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.smtp.password') }}
          </label>
          <input
            v-model="form.smtp_password"
            type="password"
            class="input"
            autocomplete="new-password"
            autocapitalize="off"
            spellcheck="false"
            :placeholder="
              form.smtp_password_configured
                ? t('admin.settings.smtp.passwordConfiguredPlaceholder')
                : t('admin.settings.smtp.passwordPlaceholder')
            "
            @keydown="$emit('password-interaction')"
            @paste="$emit('password-interaction')"
          />
          <p class="settings-smtp-card__description mt-1.5 text-xs">
            {{
              form.smtp_password_configured
                ? t('admin.settings.smtp.passwordConfiguredHint')
                : t('admin.settings.smtp.passwordHint')
            }}
          </p>
        </div>
        <div>
          <label class="settings-smtp-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.smtp.fromEmail') }}
          </label>
          <input
            v-model="form.smtp_from_email"
            type="email"
            class="input"
            :placeholder="t('admin.settings.smtp.fromEmailPlaceholder')"
          />
        </div>
        <div>
          <label class="settings-smtp-card__field-label mb-2 block text-sm font-medium">
            {{ t('admin.settings.smtp.fromName') }}
          </label>
          <input
            v-model="form.smtp_from_name"
            type="text"
            class="input"
            :placeholder="t('admin.settings.smtp.fromNamePlaceholder')"
          />
        </div>
      </div>

      <div
        class="settings-smtp-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-smtp-card__label font-medium">
            {{ t('admin.settings.smtp.useTls') }}
          </label>
          <p class="settings-smtp-card__description text-sm">
            {{ t('admin.settings.smtp.useTlsHint') }}
          </p>
        </div>
        <Toggle v-model="form.smtp_use_tls" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsSmtpFields } from './settingsForm'

defineProps<{
  form: SettingsSmtpFields
  testing: boolean
  disabled: boolean
}>()

defineEmits<{
  'test-connection': []
  'password-interaction': []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-smtp-card__header,
.settings-smtp-card__body,
.settings-smtp-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-smtp-card__header {
  padding:
    var(--theme-settings-card-header-padding-y)
    var(--theme-settings-card-header-padding-x);
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-smtp-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-smtp-card__title,
.settings-smtp-card__label,
.settings-smtp-card__field-label {
  color: var(--theme-page-text);
}

.settings-smtp-card__description {
  color: var(--theme-page-muted);
}

.settings-smtp-card__spinner {
  color: var(--theme-accent);
}
</style>
