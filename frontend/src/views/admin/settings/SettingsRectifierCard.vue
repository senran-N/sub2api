<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.rectifier.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.rectifier.description') }}
      </p>
    </div>
    <div class="space-y-5 p-6">
      <div v-if="loading" class="flex items-center gap-2 text-gray-500">
        <div class="h-4 w-4 animate-spin rounded-full border-b-2 border-primary-600"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div class="flex items-center justify-between">
          <div>
            <label class="font-medium text-gray-900 dark:text-white">
              {{ t('admin.settings.rectifier.enabled') }}
            </label>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.rectifier.enabledHint') }}
            </p>
          </div>
          <Toggle v-model="form.enabled" />
        </div>

        <div
          v-if="form.enabled"
          class="space-y-4 border-t border-gray-100 pt-4 dark:border-dark-700"
        >
          <div class="flex items-center justify-between">
            <div>
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.settings.rectifier.thinkingSignature') }}
              </label>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.settings.rectifier.thinkingSignatureHint') }}
              </p>
            </div>
            <Toggle v-model="form.thinking_signature_enabled" />
          </div>

          <div class="flex items-center justify-between">
            <div>
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.settings.rectifier.thinkingBudget') }}
              </label>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.settings.rectifier.thinkingBudgetHint') }}
              </p>
            </div>
            <Toggle v-model="form.thinking_budget_enabled" />
          </div>

          <div class="flex items-center justify-between">
            <div>
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.settings.rectifier.apikeySignature') }}
              </label>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.settings.rectifier.apikeySignatureHint') }}
              </p>
            </div>
            <Toggle v-model="form.apikey_signature_enabled" />
          </div>

          <div
            v-if="form.apikey_signature_enabled"
            class="ml-4 space-y-3 border-l-2 border-gray-200 pl-4 dark:border-dark-600"
          >
            <div>
              <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.settings.rectifier.apikeyPatterns') }}
              </label>
              <p class="text-xs text-gray-500 dark:text-gray-400">
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
                class="btn btn-ghost btn-xs text-red-500 hover:text-red-700"
                @click="removePattern(index)"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
            <button
              type="button"
              class="btn btn-ghost btn-xs text-primary-600 dark:text-primary-400"
              @click="addPattern"
            >
              + {{ t('admin.settings.rectifier.addPattern') }}
            </button>
          </div>
        </div>

        <div class="flex justify-end border-t border-gray-100 pt-4 dark:border-dark-700">
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
