<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-gateway-forwarding-card__title text-lg font-semibold">
        {{ t('admin.settings.gatewayForwarding.title') }}
      </h2>
      <p class="settings-gateway-forwarding-card__description mt-1 text-sm">
        {{ t('admin.settings.gatewayForwarding.description') }}
      </p>
    </div>
    <div class="settings-gateway-forwarding-card__body">
      <div class="flex items-center justify-between">
        <div>
          <label class="settings-gateway-forwarding-card__field-label text-sm font-medium">
            {{ t('admin.settings.gatewayForwarding.fingerprintUnification') }}
          </label>
          <p class="settings-gateway-forwarding-card__description mt-0.5 text-xs">
            {{ t('admin.settings.gatewayForwarding.fingerprintUnificationHint') }}
          </p>
        </div>
        <Toggle v-model="form.enable_fingerprint_unification" />
      </div>

      <div class="flex items-center justify-between">
        <div>
          <label class="settings-gateway-forwarding-card__field-label text-sm font-medium">
            {{ t('admin.settings.gatewayForwarding.metadataPassthrough') }}
          </label>
          <p class="settings-gateway-forwarding-card__description mt-0.5 text-xs">
            {{ t('admin.settings.gatewayForwarding.metadataPassthroughHint') }}
          </p>
        </div>
        <Toggle v-model="form.enable_metadata_passthrough" />
      </div>

      <div class="flex items-center justify-between">
        <div>
          <label class="settings-gateway-forwarding-card__field-label text-sm font-medium">
            {{ t('admin.settings.gatewayForwarding.cchSigning') }}
          </label>
          <p class="settings-gateway-forwarding-card__description mt-0.5 text-xs">
            {{ t('admin.settings.gatewayForwarding.cchSigningHint') }}
          </p>
        </div>
        <Toggle v-model="form.enable_cch_signing" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsGatewayForwardingFields } from './settingsForm'

defineProps<{
  form: SettingsGatewayForwardingFields
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-gateway-forwarding-card__title,
.settings-gateway-forwarding-card__field-label {
  color: var(--theme-page-text);
}

.settings-gateway-forwarding-card__description {
  color: var(--theme-page-muted);
}

.settings-gateway-forwarding-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}
</style>
