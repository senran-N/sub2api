<template>
  <div
    :class="getRowClasses()"
  >
    <template v-if="mode === 'token'">
      <div class="w-20">
        <label class="interval-row__label">Min</label>
        <input
          :value="interval.min_tokens"
          type="number"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('min_tokens', toInt(($event.target as HTMLInputElement).value))"
        />
      </div>
      <div class="w-20">
        <label class="interval-row__label">Max <span class="interval-row__unit">(含)</span></label>
        <input
          :value="interval.max_tokens ?? ''"
          type="number"
          min="0"
          class="input mt-0.5 text-xs"
          placeholder="∞"
          @input="emitField('max_tokens', toIntOrNull(($event.target as HTMLInputElement).value))"
        />
      </div>
      <div class="flex-1">
        <label class="interval-row__label">
          {{ t('admin.channels.form.inputPrice') }}
          <span v-if="isEmpty" class="interval-row__required">*</span>
          <span class="interval-row__unit">$/M</span>
        </label>
        <input
          :value="interval.input_price"
          type="number"
          step="any"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('input_price', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="flex-1">
        <label class="interval-row__label">
          {{ t('admin.channels.form.outputPrice') }}
          <span v-if="isEmpty" class="interval-row__required">*</span>
          <span class="interval-row__unit">$/M</span>
        </label>
        <input
          :value="interval.output_price"
          type="number"
          step="any"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('output_price', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="flex-1">
        <label class="interval-row__label">
          {{ t('admin.channels.form.cacheWritePrice') }}
          <span class="interval-row__unit">$/M</span>
        </label>
        <input
          :value="interval.cache_write_price"
          type="number"
          step="any"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('cache_write_price', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="flex-1">
        <label class="interval-row__label">
          {{ t('admin.channels.form.cacheReadPrice') }}
          <span class="interval-row__unit">$/M</span>
        </label>
        <input
          :value="interval.cache_read_price"
          type="number"
          step="any"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('cache_read_price', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </template>

    <template v-else>
      <div class="w-24">
        <label class="interval-row__label">
          {{ mode === 'image' ? t('admin.channels.form.resolution') : t('admin.channels.form.tierLabel') }}
        </label>
        <input
          :value="interval.tier_label"
          type="text"
          class="input mt-0.5 text-xs"
          :placeholder="mode === 'image' ? '1K / 2K / 4K' : ''"
          @input="emitField('tier_label', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="w-20">
        <label class="interval-row__label">Min</label>
        <input
          :value="interval.min_tokens"
          type="number"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('min_tokens', toInt(($event.target as HTMLInputElement).value))"
        />
      </div>
      <div class="w-20">
        <label class="interval-row__label">Max <span class="interval-row__unit">(含)</span></label>
        <input
          :value="interval.max_tokens ?? ''"
          type="number"
          min="0"
          class="input mt-0.5 text-xs"
          placeholder="∞"
          @input="emitField('max_tokens', toIntOrNull(($event.target as HTMLInputElement).value))"
        />
      </div>
      <div class="flex-1">
        <label class="interval-row__label">
          {{ t('admin.channels.form.perRequestPrice') }}
          <span v-if="isEmpty" class="interval-row__required">*</span>
          <span class="interval-row__unit">$</span>
        </label>
        <input
          :value="interval.per_request_price"
          type="number"
          step="any"
          min="0"
          class="input mt-0.5 text-xs"
          @input="emitField('per_request_price', ($event.target as HTMLInputElement).value)"
        />
      </div>
    </template>

    <button type="button" :class="getRemoveButtonClasses()" @click="emit('remove')">
      <Icon name="x" size="sm" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { BillingMode } from '@/api/admin/channels'
import Icon from '@/components/icons/Icon.vue'
import type { IntervalFormEntry } from './types'

const { t } = useI18n()

const props = defineProps<{
  interval: IntervalFormEntry
  mode: BillingMode
}>()

const emit = defineEmits<{
  update: [interval: IntervalFormEntry]
  remove: []
}>()

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const isEmpty = computed(() => {
  const iv = props.interval
  return (iv.input_price == null || iv.input_price === '') &&
    (iv.output_price == null || iv.output_price === '') &&
    (iv.cache_write_price == null || iv.cache_write_price === '') &&
    (iv.cache_read_price == null || iv.cache_read_price === '') &&
    (iv.per_request_price == null || iv.per_request_price === '')
})

const getRowClasses = () => {
  return joinClassNames(
    'interval-row',
    isEmpty.value ? 'interval-row--empty' : 'interval-row--filled'
  )
}

const getRemoveButtonClasses = () => {
  return joinClassNames(
    'interval-row__remove-button',
    isEmpty.value && 'interval-row__remove-button--danger'
  )
}

function emitField(field: keyof IntervalFormEntry, value: string | number | null) {
  emit('update', { ...props.interval, [field]: value === '' ? null : value })
}

function toInt(val: string): number {
  const n = parseInt(val, 10)
  return isNaN(n) ? 0 : n
}

function toIntOrNull(val: string): number | null {
  if (val === '') return null
  const n = parseInt(val, 10)
  return isNaN(n) ? null : n
}
</script>

<style scoped>
.interval-row {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-button-radius) + 2px);
  padding: 0.5rem;
}

.interval-row--filled {
  background: var(--theme-surface);
}

.interval-row--empty {
  border-color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 44%, var(--theme-card-border));
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 8%, var(--theme-surface));
}

.interval-row__label {
  color: var(--theme-page-muted);
  font-size: 0.75rem;
}

.interval-row__unit {
  color: color-mix(in srgb, var(--theme-page-muted) 70%, transparent);
}

.interval-row__required {
  color: rgb(var(--theme-danger-rgb));
}

.interval-row__remove-button {
  margin-top: 1rem;
  border-radius: calc(var(--theme-button-radius) - 4px);
  color: var(--theme-page-muted);
  padding: 0.125rem;
  transition: color 0.18s ease, background-color 0.18s ease;
}

.interval-row__remove-button:hover,
.interval-row__remove-button:focus-visible {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, transparent);
  color: rgb(var(--theme-danger-rgb));
  outline: none;
}

.interval-row__remove-button--danger {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 72%, var(--theme-page-muted));
}
</style>
