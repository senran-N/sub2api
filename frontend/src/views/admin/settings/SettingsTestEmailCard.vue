<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.testEmail.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.testEmail.description') }}
      </p>
    </div>
    <div class="p-6">
      <div class="flex items-end gap-4">
        <div class="flex-1">
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.settings.testEmail.recipientEmail') }}
          </label>
          <input
            :value="modelValue"
            type="email"
            class="input"
            :placeholder="t('admin.settings.testEmail.recipientEmailPlaceholder')"
            @input="handleInput"
          />
        </div>
        <button
          type="button"
          class="btn btn-secondary"
          :disabled="sending || !modelValue || disabled"
          @click="$emit('send')"
        >
          <svg
            v-if="sending"
            class="h-4 w-4 animate-spin"
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
          {{
            sending
              ? t('admin.settings.testEmail.sending')
              : t('admin.settings.testEmail.sendTestEmail')
          }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

const emit = defineEmits<{
  'update:modelValue': [value: string]
  send: []
}>()

defineProps<{
  modelValue: string
  sending: boolean
  disabled: boolean
}>()

const { t } = useI18n()

function handleInput(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).value)
}
</script>
