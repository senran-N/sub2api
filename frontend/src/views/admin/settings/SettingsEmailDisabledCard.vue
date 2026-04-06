<template>
  <div class="card">
    <div class="settings-email-disabled-card__body">
      <div class="flex items-start gap-3">
        <Icon name="mail" size="md" class="settings-email-disabled-card__icon mt-0.5 flex-shrink-0" />
        <div>
          <h3 class="theme-text-strong font-medium">
            {{ t('admin.settings.emailTabDisabledTitle') }}
          </h3>
          <p class="theme-text-muted mt-1 text-sm">
            {{ t('admin.settings.emailTabDisabledHint') }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
</script>

<style scoped>
.settings-email-disabled-card__icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.settings-email-disabled-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  align-items: flex-start;
  gap: var(--theme-settings-card-body-padding);
}
</style>
