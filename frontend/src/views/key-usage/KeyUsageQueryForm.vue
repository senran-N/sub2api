<template>
  <div class="mx-auto mb-14 max-w-xl">
    <div class="flex gap-3">
      <div class="relative flex-1">
        <div class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 dark:text-dark-500">
          <svg
            class="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
        </div>
        <input
          :value="apiKey"
          :type="keyVisible ? 'text' : 'password'"
          :placeholder="placeholder"
          class="input-ring h-12 w-full rounded-xl border border-gray-200 bg-white pl-12 pr-12 text-sm text-gray-900 placeholder:text-gray-400 transition-all dark:border-dark-700 dark:bg-dark-900 dark:text-white dark:placeholder:text-dark-500"
          @input="emit('update:apiKey', ($event.target as HTMLInputElement).value)"
          @keydown.enter="emit('query')"
        />
        <button
          type="button"
          class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 transition-colors hover:text-gray-700 dark:text-dark-500 dark:hover:text-white"
          @click="emit('toggle-visible')"
        >
          <svg
            v-if="!keyVisible"
            class="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"
            />
            <line x1="1" y1="1" x2="23" y2="23" />
          </svg>
          <svg
            v-else
            class="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
            <circle cx="12" cy="12" r="3" />
          </svg>
        </button>
      </div>
      <button
        type="button"
        :disabled="isQuerying"
        class="flex h-12 items-center gap-2 whitespace-nowrap rounded-xl bg-primary-500 px-7 text-sm font-medium text-white transition-all active:scale-[0.97] hover:bg-primary-600 disabled:opacity-60"
        @click="emit('query')"
      >
        <svg v-if="isQuerying" class="h-4 w-4 animate-spin" viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" opacity="0.25" />
          <path d="M12 2a10 10 0 0 1 10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round" />
        </svg>
        <svg
          v-else
          class="h-4 w-4"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        {{ isQuerying ? queryingLabel : queryLabel }}
      </button>
    </div>

    <p class="mt-3 text-center text-xs text-gray-400 dark:text-dark-500">
      {{ privacyNote }}
    </p>

    <div v-if="showDatePicker" class="mt-4">
      <div class="flex flex-wrap items-center justify-center gap-2">
        <span class="text-xs text-gray-500 dark:text-dark-400">{{ dateRangeLabel }}</span>
        <button
          v-for="range in dateRanges"
          :key="range.key"
          type="button"
          class="rounded-lg border px-3 py-1.5 text-xs transition-all"
          :class="
            currentRange === range.key
              ? 'border-primary-500 bg-primary-500 text-white'
              : 'border-gray-200 bg-white text-gray-700 hover:border-primary-300 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-200 dark:hover:border-dark-600'
          "
          @click="emit('set-range', range.key)"
        >
          {{ range.label }}
        </button>
        <div v-if="currentRange === 'custom'" class="ml-1 flex items-center gap-2">
          <input
            :value="customStartDate"
            type="date"
            class="input-ring rounded-lg border border-gray-200 bg-white px-2 py-1.5 text-xs text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
            @input="emit('update:customStartDate', ($event.target as HTMLInputElement).value)"
          />
          <span class="text-xs text-gray-400">-</span>
          <input
            :value="customEndDate"
            type="date"
            class="input-ring rounded-lg border border-gray-200 bg-white px-2 py-1.5 text-xs text-gray-900 dark:border-dark-700 dark:bg-dark-900 dark:text-white"
            @input="emit('update:customEndDate', ($event.target as HTMLInputElement).value)"
          />
          <button
            type="button"
            class="rounded-lg bg-primary-500 px-3 py-1.5 text-xs text-white hover:bg-primary-600"
            @click="emit('query')"
          >
            {{ applyLabel }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type {
  KeyUsageDateRangeKey,
  KeyUsageDateRangeOption
} from './keyUsageView'

defineProps<{
  apiKey: string
  keyVisible: boolean
  isQuerying: boolean
  showDatePicker: boolean
  currentRange: KeyUsageDateRangeKey
  customStartDate: string
  customEndDate: string
  dateRanges: KeyUsageDateRangeOption[]
  placeholder: string
  queryLabel: string
  queryingLabel: string
  privacyNote: string
  dateRangeLabel: string
  applyLabel: string
}>()

const emit = defineEmits<{
  'update:apiKey': [value: string]
  'update:customStartDate': [value: string]
  'update:customEndDate': [value: string]
  'toggle-visible': []
  'set-range': [value: KeyUsageDateRangeKey]
  query: []
}>()
</script>
