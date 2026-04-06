<template>
  <div class="mx-auto mb-14 max-w-xl">
    <div class="flex gap-3">
      <div class="relative flex-1">
        <div class="key-usage-query-form__input-icon absolute left-4 top-1/2 -translate-y-1/2">
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
          class="key-usage-query-form__input input-ring w-full text-sm transition-all"
          @input="emit('update:apiKey', ($event.target as HTMLInputElement).value)"
          @keydown.enter="emit('query')"
        />
        <button
          type="button"
          class="key-usage-query-form__toggle absolute right-4 top-1/2 -translate-y-1/2 transition-colors"
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
        class="key-usage-query-form__submit flex items-center gap-2 whitespace-nowrap text-sm font-medium transition-all active:scale-[0.97] disabled:opacity-60"
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

    <p class="key-usage-query-form__note mt-3 text-center text-xs">
      {{ privacyNote }}
    </p>

    <div v-if="showDatePicker" class="mt-4">
      <div class="flex flex-wrap items-center justify-center gap-2">
        <span class="key-usage-query-form__range-label text-xs">{{ dateRangeLabel }}</span>
        <button
          v-for="range in dateRanges"
          :key="range.key"
          type="button"
          class="key-usage-query-form__range-chip border text-xs transition-all"
          :class="
            currentRange === range.key
              ? 'key-usage-query-form__range-chip--active'
              : 'key-usage-query-form__range-chip--idle'
          "
          @click="emit('set-range', range.key)"
        >
          {{ range.label }}
        </button>
        <div v-if="currentRange === 'custom'" class="ml-1 flex items-center gap-2">
          <input
            :value="customStartDate"
            type="date"
            class="key-usage-query-form__date-input input-ring text-xs"
            @input="emit('update:customStartDate', ($event.target as HTMLInputElement).value)"
          />
          <span class="key-usage-query-form__separator text-xs">-</span>
          <input
            :value="customEndDate"
            type="date"
            class="key-usage-query-form__date-input input-ring text-xs"
            @input="emit('update:customEndDate', ($event.target as HTMLInputElement).value)"
          />
          <button
            type="button"
            class="key-usage-query-form__apply text-xs"
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

<style scoped>
.key-usage-query-form__input-icon,
.key-usage-query-form__toggle,
.key-usage-query-form__note,
.key-usage-query-form__separator {
  color: var(--theme-page-muted);
}

.key-usage-query-form__toggle:hover {
  color: var(--theme-page-text);
}

.key-usage-query-form__input,
.key-usage-query-form__date-input {
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: var(--theme-surface);
  color: var(--theme-page-text);
}

.key-usage-query-form__input {
  height: var(--theme-key-usage-query-input-height);
  border-radius: var(--theme-key-usage-query-input-radius);
  padding-left: var(--theme-key-usage-query-input-padding-left);
  padding-right: var(--theme-key-usage-query-input-padding-right);
}

.key-usage-query-form__submit {
  height: var(--theme-key-usage-query-input-height);
  border-radius: var(--theme-key-usage-query-submit-radius);
  padding-inline: var(--theme-key-usage-query-submit-padding-x);
}

.key-usage-query-form__range-chip {
  border-radius: var(--theme-key-usage-query-range-chip-radius);
  padding: var(--theme-key-usage-query-range-chip-padding-y)
    var(--theme-key-usage-query-range-chip-padding-x);
}

.key-usage-query-form__date-input {
  border-radius: var(--theme-key-usage-query-date-input-radius);
  padding: var(--theme-key-usage-query-date-input-padding-y)
    var(--theme-key-usage-query-date-input-padding-x);
}

.key-usage-query-form__apply {
  border-radius: var(--theme-key-usage-query-apply-radius);
  padding: var(--theme-key-usage-query-apply-padding-y)
    var(--theme-key-usage-query-apply-padding-x);
}

.key-usage-query-form__input::placeholder {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.key-usage-query-form__submit,
.key-usage-query-form__apply {
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.key-usage-query-form__submit:hover,
.key-usage-query-form__apply:hover {
  background: color-mix(in srgb, var(--theme-accent) 88%, var(--theme-surface-contrast));
}

.key-usage-query-form__range-label {
  color: var(--theme-page-muted);
}

.key-usage-query-form__range-chip--active {
  border-color: var(--theme-accent);
  background: var(--theme-accent);
  color: var(--theme-filled-text);
}

.key-usage-query-form__range-chip--idle {
  border-color: color-mix(in srgb, var(--theme-card-border) 74%, transparent);
  background: var(--theme-surface);
  color: var(--theme-page-text);
}

.key-usage-query-form__range-chip--idle:hover {
  border-color: color-mix(in srgb, var(--theme-accent) 28%, var(--theme-card-border));
}
</style>
