<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.betaPolicy.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.betaPolicy.description') }}
      </p>
    </div>
    <div class="space-y-5 p-6">
      <div v-if="loading" class="flex items-center gap-2 text-gray-500">
        <div class="h-4 w-4 animate-spin rounded-full border-b-2 border-primary-600"></div>
        {{ t('common.loading') }}
      </div>

      <template v-else>
        <div
          v-for="rule in rules"
          :key="rule.beta_token"
          class="rounded-lg border border-gray-200 p-4 dark:border-dark-600"
        >
          <div class="mb-3 flex items-center gap-2">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              {{ getDisplayName(rule.beta_token) }}
            </span>
            <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-700 dark:text-gray-400">
              {{ rule.beta_token }}
            </span>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
                {{ t('admin.settings.betaPolicy.action') }}
              </label>
              <Select
                :model-value="rule.action"
                :options="actionOptions"
                @update:model-value="rule.action = $event as BetaPolicyRule['action']"
              />
            </div>

            <div>
              <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
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
            <label class="mb-1 block text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('admin.settings.betaPolicy.errorMessage') }}
            </label>
            <input
              v-model="rule.error_message"
              type="text"
              class="input"
              :placeholder="t('admin.settings.betaPolicy.errorMessagePlaceholder')"
            />
            <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.settings.betaPolicy.errorMessageHint') }}
            </p>
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
