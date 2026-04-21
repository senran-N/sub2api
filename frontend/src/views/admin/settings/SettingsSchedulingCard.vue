<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-scheduling-card__title text-lg font-semibold">
        {{ t('admin.settings.scheduling.title') }}
      </h2>
      <p class="settings-scheduling-card__description mt-1 text-sm">
        {{ t('admin.settings.scheduling.description') }}
      </p>
    </div>
    <div class="settings-scheduling-card__body">
      <div class="flex items-center justify-between">
        <div>
          <label class="settings-scheduling-card__label text-sm font-medium">
            {{ t('admin.settings.scheduling.allowUngroupedKey') }}
          </label>
          <p class="settings-scheduling-card__hint mt-0.5 text-xs">
            {{ t('admin.settings.scheduling.allowUngroupedKeyHint') }}
          </p>
        </div>
        <Toggle
          v-model="form.allow_ungrouped_key_scheduling"
          :aria-label="t('admin.settings.scheduling.allowUngroupedKey')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import type { SettingsSchedulingFields } from './settingsForm'

defineProps<{
  form: SettingsSchedulingFields
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-scheduling-card__title,
.settings-scheduling-card__label {
  color: var(--theme-page-text);
}

.settings-scheduling-card__description,
.settings-scheduling-card__hint {
  color: var(--theme-page-muted);
}

.settings-scheduling-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}
</style>
