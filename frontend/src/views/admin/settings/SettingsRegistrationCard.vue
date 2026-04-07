<template>
  <div class="card">
    <div class="settings-registration-card__header">
      <h2 class="settings-registration-card__title text-lg font-semibold">
        {{ t('admin.settings.registration.title') }}
      </h2>
      <p class="settings-registration-card__description mt-1 text-sm">
        {{ t('admin.settings.registration.description') }}
      </p>
    </div>
    <div class="settings-registration-card__content space-y-5">
      <div class="flex items-center justify-between">
        <div>
          <label class="settings-registration-card__label font-medium">
            {{ t('admin.settings.registration.enableRegistration') }}
          </label>
          <p class="settings-registration-card__description text-sm">
            {{ t('admin.settings.registration.enableRegistrationHint') }}
          </p>
        </div>
        <Toggle v-model="form.registration_enabled" />
      </div>

      <div
        class="settings-registration-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-registration-card__label font-medium">
            {{ t('admin.settings.registration.emailVerification') }}
          </label>
          <p class="settings-registration-card__description text-sm">
            {{ t('admin.settings.registration.emailVerificationHint') }}
          </p>
        </div>
        <Toggle v-model="form.email_verify_enabled" />
      </div>

      <div class="settings-registration-card__section pt-4">
        <label class="settings-registration-card__label font-medium">
          {{ t('admin.settings.registration.emailSuffixWhitelist') }}
        </label>
        <p class="settings-registration-card__description mt-1 text-sm">
          {{ t('admin.settings.registration.emailSuffixWhitelistHint') }}
        </p>
        <div
          class="settings-registration-card__tag-editor mt-3"
        >
          <div class="flex flex-wrap items-center gap-2">
            <span
              v-for="suffix in tags"
              :key="suffix"
              class="theme-chip theme-chip--regular theme-chip--neutral inline-flex items-center gap-1 font-mono text-xs"
            >
              <span class="settings-registration-card__tag-prefix">@</span>
              <span>{{ suffix }}</span>
              <button
                type="button"
                class="settings-registration-card__tag-remove rounded-full"
                @click="$emit('remove-tag', suffix)"
              >
                <Icon name="x" size="xs" class="h-3.5 w-3.5" :stroke-width="2" />
              </button>
            </span>

            <div
              class="settings-registration-card__draft-input-wrap flex flex-1 items-center gap-1"
            >
              <span class="settings-registration-card__tag-prefix font-mono text-sm">@</span>
              <input
                :value="draft"
                type="text"
                class="settings-registration-card__draft-input w-full bg-transparent text-sm font-mono outline-none"
                :placeholder="t('admin.settings.registration.emailSuffixWhitelistPlaceholder')"
                @input="handleDraftInput"
                @keydown="$emit('draft-keydown', $event)"
                @blur="$emit('commit-draft')"
                @paste="$emit('draft-paste', $event)"
              />
            </div>
          </div>
        </div>
        <p class="settings-registration-card__description mt-2 text-xs">
          {{ t('admin.settings.registration.emailSuffixWhitelistInputHint') }}
        </p>
      </div>

      <div
        class="settings-registration-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-registration-card__label font-medium">
            {{ t('admin.settings.registration.promoCode') }}
          </label>
          <p class="settings-registration-card__description text-sm">
            {{ t('admin.settings.registration.promoCodeHint') }}
          </p>
        </div>
        <Toggle v-model="form.promo_code_enabled" />
      </div>

      <div
        class="settings-registration-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-registration-card__label font-medium">
            {{ t('admin.settings.registration.invitationCode') }}
          </label>
          <p class="settings-registration-card__description text-sm">
            {{ t('admin.settings.registration.invitationCodeHint') }}
          </p>
        </div>
        <Toggle v-model="form.invitation_code_enabled" />
      </div>

      <div
        v-if="form.email_verify_enabled"
        class="settings-registration-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-registration-card__label font-medium">
            {{ t('admin.settings.registration.passwordReset') }}
          </label>
          <p class="settings-registration-card__description text-sm">
            {{ t('admin.settings.registration.passwordResetHint') }}
          </p>
        </div>
        <Toggle v-model="form.password_reset_enabled" />
      </div>

      <div
        v-if="form.email_verify_enabled && form.password_reset_enabled"
        class="settings-registration-card__section pt-4"
      >
        <label class="settings-registration-card__field-label mb-2 block text-sm font-medium">
          {{ t('admin.settings.registration.frontendUrl') }}
        </label>
        <input
          v-model="form.frontend_url"
          type="url"
          class="input"
          :placeholder="t('admin.settings.registration.frontendUrlPlaceholder')"
        />
        <p class="settings-registration-card__description mt-1.5 text-xs">
          {{ t('admin.settings.registration.frontendUrlHint') }}
        </p>
      </div>

      <div
        class="settings-registration-card__section flex items-center justify-between pt-4"
      >
        <div>
          <label class="settings-registration-card__label font-medium">
            {{ t('admin.settings.registration.totp') }}
          </label>
          <p class="settings-registration-card__description text-sm">
            {{ t('admin.settings.registration.totpHint') }}
          </p>
          <p
            v-if="!form.totp_encryption_key_configured"
            class="settings-registration-card__warning mt-2 text-sm"
          >
            {{ t('admin.settings.registration.totpKeyNotConfigured') }}
          </p>
        </div>
        <Toggle
          v-model="form.totp_enabled"
          :disabled="!form.totp_encryption_key_configured"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import type { SettingsRegistrationFields } from './settingsForm'

defineProps<{
  form: SettingsRegistrationFields
  tags: string[]
  draft: string
}>()

const emit = defineEmits<{
  'remove-tag': [suffix: string]
  'update:draft': [value: string]
  'draft-input': []
  'commit-draft': []
  'draft-keydown': [event: KeyboardEvent]
  'draft-paste': [event: ClipboardEvent]
}>()

const { t } = useI18n()

const handleDraftInput = (event: Event) => {
  emit('update:draft', (event.target as HTMLInputElement).value)
  emit('draft-input')
}
</script>

<style scoped>
.settings-registration-card__header,
.settings-registration-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-registration-card__header {
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-registration-card__title,
.settings-registration-card__label,
.settings-registration-card__field-label {
  color: var(--theme-page-text);
}

.settings-registration-card__header {
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
}

.settings-registration-card__content {
  padding: var(--theme-settings-card-body-padding);
}

.settings-registration-card__description,
.settings-registration-card__tag-prefix {
  color: var(--theme-page-muted);
}

.settings-registration-card__tag-editor {
  border-radius: var(--theme-settings-registration-tag-editor-radius);
  padding: var(--theme-settings-registration-tag-editor-padding);
  border: 1px solid var(--theme-input-border);
  background: color-mix(in srgb, var(--theme-input-bg) 88%, var(--theme-surface-soft));
}

.settings-registration-card__tag-remove {
  color: color-mix(in srgb, var(--theme-page-muted) 78%, transparent);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.settings-registration-card__tag-remove:hover {
  background: var(--theme-button-ghost-hover-bg);
  color: var(--theme-page-text);
}

.settings-registration-card__draft-input-wrap {
  min-width: var(--theme-settings-registration-draft-min-width);
  border-radius: var(--theme-settings-registration-draft-radius);
  padding: var(--theme-settings-registration-draft-padding-y)
    var(--theme-settings-registration-draft-padding-x);
  border: 1px solid transparent;
  transition: border-color 0.2s ease, background-color 0.2s ease;
}

.settings-registration-card__draft-input-wrap:focus-within {
  border-color: color-mix(in srgb, var(--theme-accent) 52%, var(--theme-input-border));
  background: color-mix(in srgb, var(--theme-accent-soft) 48%, var(--theme-input-bg));
}

.settings-registration-card__draft-input {
  color: var(--theme-page-text);
}

.settings-registration-card__draft-input::placeholder {
  color: var(--theme-input-placeholder);
}

.settings-registration-card__warning {
  color: color-mix(in srgb, rgb(var(--theme-warning-rgb)) 84%, var(--theme-page-text));
}
</style>
