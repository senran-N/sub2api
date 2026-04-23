<template>
  <div class="card overflow-hidden">
    <div class="settings-wechat-card__header">
      <h2 class="settings-wechat-card__title text-lg font-semibold">
        {{ t('admin.settings.wechatConnect.title') }}
      </h2>
      <p class="settings-wechat-card__description mt-1 text-sm">
        {{ t('admin.settings.wechatConnect.description') }}
      </p>
    </div>

    <div class="settings-wechat-card__body space-y-5">
      <div class="flex items-center justify-between gap-4">
        <div>
          <label class="settings-wechat-card__label font-medium">
            {{ t('admin.settings.wechatConnect.enable') }}
          </label>
          <p class="settings-wechat-card__description text-sm">
            {{ t('admin.settings.wechatConnect.enableHint') }}
          </p>
        </div>
        <Toggle v-model="form.wechat_connect_enabled" :aria-label="t('admin.settings.wechatConnect.enable')" />
      </div>

      <div v-if="form.wechat_connect_enabled" class="settings-wechat-card__section space-y-5 pt-4">
        <div class="grid gap-4 md:grid-cols-3">
          <label class="flex items-start gap-3">
            <input v-model="form.wechat_connect_open_enabled" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">{{ t('admin.settings.wechatConnect.openEnabled') }}</span>
              <span class="settings-wechat-card__description text-xs">{{ t('admin.settings.wechatConnect.openEnabledHint') }}</span>
            </span>
          </label>
          <label class="flex items-start gap-3">
            <input v-model="form.wechat_connect_mp_enabled" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">{{ t('admin.settings.wechatConnect.mpEnabled') }}</span>
              <span class="settings-wechat-card__description text-xs">{{ t('admin.settings.wechatConnect.mpEnabledHint') }}</span>
            </span>
          </label>
          <label class="flex items-start gap-3">
            <input v-model="form.wechat_connect_mobile_enabled" type="checkbox" class="mt-1" />
            <span>
              <span class="block font-medium">{{ t('admin.settings.wechatConnect.mobileEnabled') }}</span>
              <span class="settings-wechat-card__description text-xs">{{ t('admin.settings.wechatConnect.mobileEnabledHint') }}</span>
            </span>
          </label>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.mode') }}</label>
            <select v-model="form.wechat_connect_mode" class="input">
              <option value="open">{{ t('admin.settings.wechatConnect.modeOpen') }}</option>
              <option value="mp">{{ t('admin.settings.wechatConnect.modeMp') }}</option>
              <option value="mobile">{{ t('admin.settings.wechatConnect.modeMobile') }}</option>
            </select>
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.scopes') }}</label>
            <input v-model="form.wechat_connect_scopes" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.wechatConnect.scopesPlaceholder')" />
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.openAppId') }}</label>
            <input v-model="form.wechat_connect_open_app_id" type="text" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.openAppSecret') }}</label>
            <input v-model="form.wechat_connect_open_app_secret" type="password" class="input font-mono text-sm" :placeholder="form.wechat_connect_open_app_secret_configured ? t('admin.settings.wechatConnect.secretConfiguredPlaceholder') : t('admin.settings.wechatConnect.openAppSecretPlaceholder')" />
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.mpAppId') }}</label>
            <input v-model="form.wechat_connect_mp_app_id" type="text" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.mpAppSecret') }}</label>
            <input v-model="form.wechat_connect_mp_app_secret" type="password" class="input font-mono text-sm" :placeholder="form.wechat_connect_mp_app_secret_configured ? t('admin.settings.wechatConnect.secretConfiguredPlaceholder') : t('admin.settings.wechatConnect.mpAppSecretPlaceholder')" />
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.mobileAppId') }}</label>
            <input v-model="form.wechat_connect_mobile_app_id" type="text" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.mobileAppSecret') }}</label>
            <input v-model="form.wechat_connect_mobile_app_secret" type="password" class="input font-mono text-sm" :placeholder="form.wechat_connect_mobile_app_secret_configured ? t('admin.settings.wechatConnect.secretConfiguredPlaceholder') : t('admin.settings.wechatConnect.mobileAppSecretPlaceholder')" />
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-2">
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.redirectUrl') }}</label>
            <input v-model="form.wechat_connect_redirect_url" type="url" class="input font-mono text-sm" :placeholder="t('admin.settings.wechatConnect.redirectUrlPlaceholder')" />
            <div class="mt-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
              <button type="button" class="btn btn-secondary btn-sm w-fit" @click="$emit('quick-set-copy')">
                {{ t('admin.settings.wechatConnect.quickSetCopy') }}
              </button>
              <code v-if="redirectUrlSuggestion" class="settings-wechat-card__suggestion select-all break-all font-mono text-xs">
                {{ redirectUrlSuggestion }}
              </code>
            </div>
          </div>
          <div>
            <label class="settings-wechat-card__field-label mb-2 block text-sm font-medium">{{ t('admin.settings.wechatConnect.frontendRedirectUrl') }}</label>
            <input v-model="form.wechat_connect_frontend_redirect_url" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.wechatConnect.frontendRedirectUrlPlaceholder')" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsWechatFields } from './settingsForm'

defineProps<{
  form: SettingsWechatFields
  redirectUrlSuggestion: string
}>()

defineEmits<{
  'quick-set-copy': []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-wechat-card__header,
.settings-wechat-card__body,
.settings-wechat-card__section {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-wechat-card__header {
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-wechat-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-wechat-card__title,
.settings-wechat-card__label,
.settings-wechat-card__field-label {
  color: var(--theme-page-text);
}

.settings-wechat-card__description {
  color: var(--theme-page-muted);
}

.settings-wechat-card__suggestion {
  border-radius: var(--theme-settings-inline-button-radius);
  padding: var(--theme-settings-code-padding-y) var(--theme-settings-code-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}
</style>
