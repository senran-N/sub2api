<template>
  <div class="card overflow-hidden">
    <div class="settings-oidc-card__header">
      <h2 class="settings-oidc-card__title text-lg font-semibold">
        {{ t('admin.settings.oidc.title') }}
      </h2>
      <p class="settings-oidc-card__description mt-1 text-sm">
        {{ t('admin.settings.oidc.description') }}
      </p>
    </div>

    <div class="settings-oidc-card__body space-y-5">
      <div class="flex items-center justify-between gap-4">
        <div>
          <label class="settings-oidc-card__label font-medium">
            {{ t('admin.settings.oidc.enable') }}
          </label>
          <p class="settings-oidc-card__description text-sm">
            {{ t('admin.settings.oidc.enableHint') }}
          </p>
        </div>
        <Toggle v-model="form.oidc_connect_enabled" :aria-label="t('admin.settings.oidc.enable')" />
      </div>

      <div v-if="form.oidc_connect_enabled" class="settings-oidc-card__section pt-4 space-y-5">
        <div class="grid gap-5 lg:grid-cols-2">
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.providerName') }}
            </label>
            <input v-model="form.oidc_connect_provider_name" type="text" class="input" :placeholder="t('admin.settings.oidc.providerNamePlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.clientId') }}
            </label>
            <input v-model="form.oidc_connect_client_id" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.clientIdPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.clientSecret') }}
            </label>
            <input
              v-model="form.oidc_connect_client_secret"
              type="password"
              class="input font-mono text-sm"
              :placeholder="form.oidc_connect_client_secret_configured ? t('admin.settings.oidc.clientSecretConfiguredPlaceholder') : t('admin.settings.oidc.clientSecretPlaceholder')"
            />
            <p class="settings-oidc-card__description mt-1.5 text-xs">
              {{ form.oidc_connect_client_secret_configured ? t('admin.settings.oidc.clientSecretConfiguredHint') : t('admin.settings.oidc.clientSecretHint') }}
            </p>
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.tokenAuthMethod') }}
            </label>
            <select v-model="form.oidc_connect_token_auth_method" class="input">
              <option value="client_secret_post">client_secret_post</option>
              <option value="client_secret_basic">client_secret_basic</option>
              <option value="none">none</option>
            </select>
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.authorizeUrl') }}
            </label>
            <input v-model="form.oidc_connect_authorize_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.authorizeUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.tokenUrl') }}
            </label>
            <input v-model="form.oidc_connect_token_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.tokenUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.userInfoUrl') }}
            </label>
            <input v-model="form.oidc_connect_userinfo_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.userInfoUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.jwksUrl') }}
            </label>
            <input v-model="form.oidc_connect_jwks_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.jwksUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.issuerUrl') }}
            </label>
            <input v-model="form.oidc_connect_issuer_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.issuerUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.discoveryUrl') }}
            </label>
            <input v-model="form.oidc_connect_discovery_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.discoveryUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.scopes') }}
            </label>
            <input v-model="form.oidc_connect_scopes" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.scopesPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.allowedSigningAlgs') }}
            </label>
            <input v-model="form.oidc_connect_allowed_signing_algs" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.allowedSigningAlgsPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.redirectUrl') }}
            </label>
            <input v-model="form.oidc_connect_redirect_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.redirectUrlPlaceholder')" />
            <div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
              <button type="button" class="btn btn-secondary btn-sm w-fit" @click="$emit('quick-set-copy')">
                {{ t('admin.settings.oidc.quickSetCopy') }}
              </button>
              <code v-if="redirectUrlSuggestion" class="settings-oidc-card__suggestion select-all break-all font-mono text-xs">
                {{ redirectUrlSuggestion }}
              </code>
            </div>
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.frontendRedirectUrl') }}
            </label>
            <input v-model="form.oidc_connect_frontend_redirect_url" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.oidc.frontendRedirectUrlPlaceholder')" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.clockSkewSeconds') }}
            </label>
            <input v-model.number="form.oidc_connect_clock_skew_seconds" type="number" min="0" class="input" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.userInfoEmailPath') }}
            </label>
            <input v-model="form.oidc_connect_userinfo_email_path" type="text" class="input font-mono text-sm" placeholder="email" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.userInfoIdPath') }}
            </label>
            <input v-model="form.oidc_connect_userinfo_id_path" type="text" class="input font-mono text-sm" placeholder="sub" />
          </div>
          <div>
            <label class="settings-oidc-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.oidc.userInfoUsernamePath') }}
            </label>
            <input v-model="form.oidc_connect_userinfo_username_path" type="text" class="input font-mono text-sm" placeholder="preferred_username" />
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-3">
          <label class="flex items-start gap-3">
            <input v-model="form.oidc_connect_use_pkce" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">{{ t('admin.settings.oidc.usePkce') }}</span>
              <span class="settings-oidc-card__description text-xs">{{ t('admin.settings.oidc.usePkceHint') }}</span>
            </span>
          </label>
          <label class="flex items-start gap-3">
            <input v-model="form.oidc_connect_validate_id_token" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">{{ t('admin.settings.oidc.validateIdToken') }}</span>
              <span class="settings-oidc-card__description text-xs">{{ t('admin.settings.oidc.validateIdTokenHint') }}</span>
            </span>
          </label>
          <label class="flex items-start gap-3">
            <input v-model="form.oidc_connect_require_email_verified" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">{{ t('admin.settings.oidc.requireEmailVerified') }}</span>
              <span class="settings-oidc-card__description text-xs">{{ t('admin.settings.oidc.requireEmailVerifiedHint') }}</span>
            </span>
          </label>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsOidcFields } from './settingsForm'

defineProps<{
  form: SettingsOidcFields
  redirectUrlSuggestion: string
}>()

defineEmits<{
  'quick-set-copy': []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-oidc-card__header,
.settings-oidc-card__body,
.settings-oidc-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-oidc-card__header {
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-oidc-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-oidc-card__title,
.settings-oidc-card__label,
.settings-oidc-card__field-label {
  color: var(--theme-page-text);
}

.settings-oidc-card__description {
  color: var(--theme-page-muted);
}

.settings-oidc-card__suggestion {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-code-padding-y) var(--theme-settings-code-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}
</style>
