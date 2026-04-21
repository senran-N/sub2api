<template>
  <div class="card">
    <div class="settings-rectifier-card__header">
      <h2 class="settings-rectifier-card__title text-lg font-semibold">
        {{ t('admin.settings.rectifier.title') }}
      </h2>
      <p class="settings-rectifier-card__description mt-1 text-sm">
        {{ t('admin.settings.rectifier.description') }}
      </p>
    </div>
    <div class="settings-rectifier-card__body space-y-5">
      <div v-if="loading" class="settings-rectifier-card__loading flex items-center gap-2">
        <div class="settings-rectifier-card__spinner h-4 w-4 animate-spin rounded-full border-b-2"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div class="flex items-center justify-between">
          <div>
            <label class="settings-rectifier-card__label font-medium">
              {{ t('admin.settings.rectifier.enabled') }}
            </label>
            <p class="settings-rectifier-card__description text-sm">
              {{ t('admin.settings.rectifier.enabledHint') }}
            </p>
          </div>
          <Toggle
            v-model="form.enabled"
            :aria-label="t('admin.settings.rectifier.enabled')"
          />
        </div>

        <div
          v-if="form.enabled"
          class="settings-rectifier-card__section space-y-4 pt-4"
        >
          <div class="flex items-center justify-between">
            <div>
              <label class="settings-rectifier-card__field-label text-sm font-medium">
                {{ t('admin.settings.rectifier.thinkingSignature') }}
              </label>
              <p class="settings-rectifier-card__description text-xs">
                {{ t('admin.settings.rectifier.thinkingSignatureHint') }}
              </p>
            </div>
            <Toggle
              v-model="form.thinking_signature_enabled"
              :aria-label="t('admin.settings.rectifier.thinkingSignature')"
            />
          </div>

          <div class="flex items-center justify-between">
            <div>
              <label class="settings-rectifier-card__field-label text-sm font-medium">
                {{ t('admin.settings.rectifier.thinkingBudget') }}
              </label>
              <p class="settings-rectifier-card__description text-xs">
                {{ t('admin.settings.rectifier.thinkingBudgetHint') }}
              </p>
            </div>
            <Toggle
              v-model="form.thinking_budget_enabled"
              :aria-label="t('admin.settings.rectifier.thinkingBudget')"
            />
          </div>

          <div class="flex items-center justify-between">
            <div>
              <label class="settings-rectifier-card__field-label text-sm font-medium">
                {{ t('admin.settings.rectifier.apikeySignature') }}
              </label>
              <p class="settings-rectifier-card__description text-xs">
                {{ t('admin.settings.rectifier.apikeySignatureHint') }}
              </p>
            </div>
            <Toggle
              v-model="form.apikey_signature_enabled"
              :aria-label="t('admin.settings.rectifier.apikeySignature')"
            />
          </div>

          <div
            v-if="form.apikey_signature_enabled"
            class="settings-rectifier-card__patterns ml-4 space-y-3 pl-4"
          >
            <div>
              <label class="settings-rectifier-card__field-label text-sm font-medium">
                {{ t('admin.settings.rectifier.apikeyPatterns') }}
              </label>
              <p class="settings-rectifier-card__description text-xs">
                {{ t('admin.settings.rectifier.apikeyPatternsHint') }}
              </p>
            </div>
            <div
              v-for="(_, index) in form.apikey_signature_patterns"
              :key="index"
              class="flex items-center gap-2"
            >
              <input
                v-model="form.apikey_signature_patterns[index]"
                type="text"
                class="input input-sm flex-1"
                :placeholder="t('admin.settings.rectifier.apikeyPatternPlaceholder')"
              />
              <button
                type="button"
                class="settings-rectifier-card__remove-button btn btn-ghost btn-xs"
                @click="removePattern(index)"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
            <button
              type="button"
              class="settings-rectifier-card__add-button btn btn-ghost btn-xs"
              @click="addPattern"
            >
              + {{ t('admin.settings.rectifier.addPattern') }}
            </button>
          </div>
        </div>

        <div class="settings-rectifier-card__footer flex justify-end pt-4">
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
import type { RectifierSettings } from '@/api/admin/settings'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  loading: boolean
  saving: boolean
  form: RectifierSettings
}>()

defineEmits<{
  save: []
}>()

const { t } = useI18n()

const addPattern = () => {
  props.form.apikey_signature_patterns.push('')
}

const removePattern = (index: number) => {
  props.form.apikey_signature_patterns.splice(index, 1)
}
</script>

<style scoped>
.settings-rectifier-card__header,
.settings-rectifier-card__body,
.settings-rectifier-card__section,
.settings-rectifier-card__footer {
  border-top: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-rectifier-card__header {
  padding:
    var(--theme-settings-card-header-padding-y)
    var(--theme-settings-card-header-padding-x);
  border-top: none;
  border-bottom: 1px solid color-mix(in srgb, var(--theme-card-border) 68%, transparent);
}

.settings-rectifier-card__body {
  padding: var(--theme-settings-card-body-padding);
}

.settings-rectifier-card__title,
.settings-rectifier-card__label,
.settings-rectifier-card__field-label {
  color: var(--theme-page-text);
}

.settings-rectifier-card__description,
.settings-rectifier-card__loading {
  color: var(--theme-page-muted);
}

.settings-rectifier-card__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 70%, transparent);
  border-bottom-color: var(--theme-accent);
}

.settings-rectifier-card__patterns {
  border-left: 2px solid color-mix(in srgb, var(--theme-card-border) 72%, transparent);
}

.settings-rectifier-card__remove-button {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 78%, var(--theme-page-text));
}

.settings-rectifier-card__remove-button:hover {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 92%, var(--theme-page-text));
}

.settings-rectifier-card__add-button {
  color: var(--theme-accent);
}
</style>
