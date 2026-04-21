<template>
  <div class="card">
    <div class="card-header">
      <h2 class="settings-overload-cooldown-card__title text-lg font-semibold">
        {{ t('admin.settings.overloadCooldown.title') }}
      </h2>
      <p class="settings-overload-cooldown-card__description mt-1 text-sm">
        {{ t('admin.settings.overloadCooldown.description') }}
      </p>
    </div>
    <div class="settings-overload-cooldown-card__body">
      <div v-if="loading" class="settings-overload-cooldown-card__description flex items-center gap-2">
        <div class="settings-overload-cooldown-card__spinner h-4 w-4 animate-spin rounded-full border-b-2"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div class="flex items-center justify-between">
          <div>
            <label class="settings-overload-cooldown-card__title font-medium">
              {{ t('admin.settings.overloadCooldown.enabled') }}
            </label>
            <p class="settings-overload-cooldown-card__description text-sm">
              {{ t('admin.settings.overloadCooldown.enabledHint') }}
            </p>
          </div>
          <Toggle
            v-model="form.enabled"
            :aria-label="t('admin.settings.overloadCooldown.enabled')"
          />
        </div>

        <div
          v-if="form.enabled"
          class="settings-overload-cooldown-card__section space-y-4 border-t pt-4"
        >
          <div>
            <label class="settings-overload-cooldown-card__field-label mb-2 block text-sm font-medium">
              {{ t('admin.settings.overloadCooldown.cooldownMinutes') }}
            </label>
            <input
              v-model.number="form.cooldown_minutes"
              type="number"
              min="1"
              max="120"
              class="input w-32"
            />
            <p class="settings-overload-cooldown-card__description mt-1.5 text-xs">
              {{ t('admin.settings.overloadCooldown.cooldownMinutesHint') }}
            </p>
          </div>
        </div>

        <div class="settings-overload-cooldown-card__section flex justify-end border-t pt-4">
          <button
            type="button"
            :disabled="saving"
            class="btn btn-primary btn-sm"
            @click="$emit('save')"
          >
            <svg
              v-if="saving"
              class="mr-1 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
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
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type { OverloadCooldownSettings } from '@/api/admin/settings'
import Toggle from '@/components/common/Toggle.vue'

defineProps<{
  loading: boolean
  saving: boolean
  form: OverloadCooldownSettings
}>()

defineEmits<{
  save: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-overload-cooldown-card__title,
.settings-overload-cooldown-card__field-label {
  color: var(--theme-page-text);
}

.settings-overload-cooldown-card__description {
  color: var(--theme-page-muted);
}

.settings-overload-cooldown-card__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 64%, transparent);
  border-bottom-color: var(--theme-accent);
}

.settings-overload-cooldown-card__section {
  border-color: color-mix(in srgb, var(--theme-card-border) 76%, transparent);
}

.settings-overload-cooldown-card__body {
  padding: var(--theme-settings-card-panel-padding);
  display: flex;
  flex-direction: column;
  gap: var(--theme-settings-card-body-padding);
}
</style>
