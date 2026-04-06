<template>
  <div class="card">
    <div class="settings-beta-policy-card__header">
      <h2 class="settings-beta-policy-card__title text-lg font-semibold">
        {{ t('admin.settings.betaPolicy.title') }}
      </h2>
      <p class="settings-beta-policy-card__description mt-1 text-sm">
        {{ t('admin.settings.betaPolicy.description') }}
      </p>
    </div>
    <div class="settings-beta-policy-card__content space-y-5">
      <div v-if="loading" class="settings-beta-policy-card__loading flex items-center gap-2">
        <div class="settings-beta-policy-card__spinner h-4 w-4 animate-spin rounded-full border-b-2"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div
          v-for="rule in rules"
          :key="rule.beta_token"
          class="settings-beta-policy-card__rule"
        >
          <div class="mb-3 flex items-center gap-2">
            <span class="settings-beta-policy-card__rule-title text-sm font-medium">
              {{ getDisplayName(rule.beta_token) }}
            </span>
            <span class="settings-beta-policy-card__token text-xs">
              {{ rule.beta_token }}
            </span>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="settings-beta-policy-card__field-label mb-1 block text-xs font-medium">
                {{ t('admin.settings.betaPolicy.action') }}
              </label>
              <Select
                :model-value="rule.action"
                :options="actionOptions"
                @update:model-value="rule.action = $event as BetaPolicyRule['action']"
              />
            </div>

            <div>
              <label class="settings-beta-policy-card__field-label mb-1 block text-xs font-medium">
                {{ t('admin.settings.betaPolicy.scope') }}
              </label>
              <Select
                :model-value="rule.scope"
                :options="scopeOptions"
                @update:model-value="rule.scope = $event as BetaPolicyRule['scope']"
              />
            </div>
          </div>

          <div v-if="rule.action === 'block'" class="mt-3">
            <label class="settings-beta-policy-card__field-label mb-1 block text-xs font-medium">
              {{ t('admin.settings.betaPolicy.errorMessage') }}
            </label>
            <input
              v-model="rule.error_message"
              type="text"
              class="input"
              :placeholder="t('admin.settings.betaPolicy.errorMessagePlaceholder')"
            />
            <p class="settings-beta-policy-card__hint mt-1 text-xs">
              {{ t('admin.settings.betaPolicy.errorMessageHint') }}
            </p>
          </div>
        </div>

        <div class="settings-beta-policy-card__footer flex justify-end pt-4">
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
import type { BetaPolicyRule } from '@/api/admin/settings'
import Select, { type SelectOption } from '@/components/common/Select.vue'

defineProps<{
  loading: boolean
  saving: boolean
  rules: BetaPolicyRule[]
  actionOptions: SelectOption[]
  scopeOptions: SelectOption[]
  getDisplayName: (token: string) => string
}>()

defineEmits<{
  save: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.settings-beta-policy-card__header,
.settings-beta-policy-card__footer {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-beta-policy-card__header {
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-beta-policy-card__title,
.settings-beta-policy-card__rule-title,
.settings-beta-policy-card__field-label {
  color: var(--theme-page-text);
}

.settings-beta-policy-card__header {
  padding: var(--theme-settings-card-header-padding-y) var(--theme-settings-card-header-padding-x);
}

.settings-beta-policy-card__content {
  padding: var(--theme-settings-card-body-padding);
}

.settings-beta-policy-card__description,
.settings-beta-policy-card__loading,
.settings-beta-policy-card__hint {
  color: var(--theme-page-muted);
}

.settings-beta-policy-card__spinner {
  color: var(--theme-accent);
}

.settings-beta-policy-card__rule {
  border-radius: var(--theme-settings-beta-policy-rule-radius);
  padding: var(--theme-settings-beta-policy-rule-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.settings-beta-policy-card__token {
  border-radius: var(--theme-settings-beta-policy-token-radius);
  padding: var(--theme-settings-beta-policy-token-padding-y)
    var(--theme-settings-beta-policy-token-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-muted);
}
</style>
