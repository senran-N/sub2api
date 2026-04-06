<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-sora-client-card__title text-lg font-semibold">
        {{ t('admin.settings.soraClient.title') }}
      </h2>
      <p class="settings-sora-client-card__description mt-1 text-sm">
        {{ t('admin.settings.soraClient.description') }}
      </p>
    </div>
    <div class="settings-sora-client-card__body">
      <div class="flex items-center justify-between">
        <div>
          <label class="settings-sora-client-card__label font-medium">
            {{ t('admin.settings.soraClient.enabled') }}
          </label>
          <p class="settings-sora-client-card__hint text-sm">
            {{ t('admin.settings.soraClient.enabledHint') }}
          </p>
        </div>
        <Toggle v-model="form.sora_client_enabled" />
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

<style scoped>
.settings-sora-client-card__title,
.settings-sora-client-card__label {
  color: var(--theme-page-text);
}

.settings-sora-client-card__description,
.settings-sora-client-card__hint {
  color: var(--theme-page-muted);
}

.settings-sora-client-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}
</style>
