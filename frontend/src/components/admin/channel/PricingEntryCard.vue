<template>
  <div class="pricing-entry-card">
    <div class="pricing-entry-card__summary" @click="collapsed = !collapsed">
      <Icon
        :name="collapsed ? 'chevronRight' : 'chevronDown'"
        size="sm"
        :stroke-width="2"
        class="pricing-entry-card__summary-icon"
      />

      <div v-if="collapsed" class="flex min-w-0 flex-1 items-center gap-2 overflow-hidden">
        <div class="flex min-w-0 flex-1 flex-wrap items-center gap-1">
          <span
            v-for="(m, i) in entry.models.slice(0, 3)"
            :key="i"
            :class="joinClassNames(getPlatformTagClass(props.platform || ''), 'shrink-0 text-xs')"
          >
            {{ m }}
          </span>
          <span v-if="entry.models.length > 3" class="pricing-entry-card__muted whitespace-nowrap text-xs">
            +{{ entry.models.length - 3 }}
          </span>
          <span v-if="entry.models.length === 0" class="pricing-entry-card__muted text-xs italic">
            {{ t('admin.channels.form.noModels') }}
          </span>
        </div>

        <span :class="getBillingModeChipClasses()">
          {{ billingModeLabel }}
        </span>
      </div>

      <div v-else class="pricing-entry-card__muted flex-1 text-xs font-medium">
        {{ t('admin.channels.form.pricingEntry') }}
      </div>

      <button type="button" class="pricing-entry-card__remove-button" @click.stop="emit('remove')">
        <Icon name="trash" size="sm" />
      </button>
    </div>

    <div class="collapsible-content" :class="{ 'collapsible-content--collapsed': collapsed }">
      <div class="collapsible-inner">
        <div class="mt-3 flex items-start gap-2">
          <div class="flex-1">
            <label class="pricing-entry-card__label">
              {{ t('admin.channels.form.models') }} <span class="pricing-entry-card__required">*</span>
            </label>
            <ModelTagInput
              :models="entry.models"
              :platform="props.platform"
              class="mt-1"
              :placeholder="t('admin.channels.form.modelsPlaceholder')"
              @update:models="onModelsUpdate($event)"
            />
          </div>
          <div class="w-40">
            <label class="pricing-entry-card__label">
              {{ t('admin.channels.form.billingMode') }}
            </label>
            <Select
              :modelValue="entry.billing_mode"
              :options="billingModeOptions"
              class="mt-1"
              @update:modelValue="emit('update', { ...entry, billing_mode: $event as BillingMode, intervals: [] })"
            />
          </div>
        </div>

        <div v-if="entry.billing_mode === 'token'">
          <label class="pricing-entry-card__label mt-3 block">
            {{ t('admin.channels.form.defaultPrices') }}
            <span class="pricing-entry-card__label-unit ml-1 font-normal">$/MTok</span>
          </label>
          <div class="mt-1 grid grid-cols-2 gap-2 sm:grid-cols-5">
            <div>
              <label class="pricing-entry-card__sub-label">{{ t('admin.channels.form.inputPrice') }}</label>
              <input :value="entry.input_price" type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('input_price', ($event.target as HTMLInputElement).value)" />
            </div>
            <div>
              <label class="pricing-entry-card__sub-label">{{ t('admin.channels.form.outputPrice') }}</label>
              <input :value="entry.output_price" type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('output_price', ($event.target as HTMLInputElement).value)" />
            </div>
            <div>
              <label class="pricing-entry-card__sub-label">{{ t('admin.channels.form.cacheWritePrice') }}</label>
              <input :value="entry.cache_write_price" type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('cache_write_price', ($event.target as HTMLInputElement).value)" />
            </div>
            <div>
              <label class="pricing-entry-card__sub-label">{{ t('admin.channels.form.cacheReadPrice') }}</label>
              <input :value="entry.cache_read_price" type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('cache_read_price', ($event.target as HTMLInputElement).value)" />
            </div>
            <div>
              <label class="pricing-entry-card__sub-label">{{ t('admin.channels.form.imageTokenPrice') }}</label>
              <input :value="entry.image_output_price" type="number" step="any" min="0" class="input mt-0.5 text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('image_output_price', ($event.target as HTMLInputElement).value)" />
            </div>
          </div>

          <div class="mt-3">
            <div class="flex items-center justify-between">
              <label class="pricing-entry-card__label">
                {{ t('admin.channels.form.intervals') }}
                <span class="pricing-entry-card__label-unit ml-1 font-normal">(min, max]</span>
              </label>
              <button type="button" class="pricing-entry-card__link-button" @click="addInterval">
                + {{ t('admin.channels.form.addInterval') }}
              </button>
            </div>
            <div v-if="entry.intervals && entry.intervals.length > 0" class="mt-2 space-y-2">
              <IntervalRow
                v-for="(iv, idx) in entry.intervals"
                :key="idx"
                :interval="iv"
                :mode="entry.billing_mode"
                @update="updateInterval(idx, $event)"
                @remove="removeInterval(idx)"
              />
            </div>
          </div>
        </div>

        <div v-else-if="entry.billing_mode === 'per_request'">
          <label class="pricing-entry-card__label mt-3 block">
            {{ t('admin.channels.form.defaultPerRequestPrice') }}
            <span class="pricing-entry-card__label-unit ml-1 font-normal">$</span>
          </label>
          <div class="mt-1 w-48">
            <input :value="entry.per_request_price" type="number" step="any" min="0" class="input text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('per_request_price', ($event.target as HTMLInputElement).value)" />
          </div>

          <div class="mt-3 flex items-center justify-between">
            <label class="pricing-entry-card__label">
              {{ t('admin.channels.form.requestTiers') }}
            </label>
            <button type="button" class="pricing-entry-card__link-button" @click="addInterval">
              + {{ t('admin.channels.form.addTier') }}
            </button>
          </div>
          <div v-if="entry.intervals && entry.intervals.length > 0" class="mt-2 space-y-2">
            <IntervalRow
              v-for="(iv, idx) in entry.intervals"
              :key="idx"
              :interval="iv"
              :mode="entry.billing_mode"
              @update="updateInterval(idx, $event)"
              @remove="removeInterval(idx)"
            />
          </div>
          <div v-else class="pricing-entry-card__empty-state mt-2">
            {{ t('admin.channels.form.noTiersYet') }}
          </div>
        </div>

        <div v-else-if="entry.billing_mode === 'image'">
          <label class="pricing-entry-card__label mt-3 block">
            {{ t('admin.channels.form.defaultImagePrice') }}
            <span class="pricing-entry-card__label-unit ml-1 font-normal">$</span>
          </label>
          <div class="mt-1 w-48">
            <input :value="entry.per_request_price" type="number" step="any" min="0" class="input text-sm" :placeholder="t('admin.channels.form.pricePlaceholder')" @input="emitField('per_request_price', ($event.target as HTMLInputElement).value)" />
          </div>

          <div class="mt-3 flex items-center justify-between">
            <label class="pricing-entry-card__label">
              {{ t('admin.channels.form.imageTiers') }}
            </label>
            <button type="button" class="pricing-entry-card__link-button" @click="addImageTier">
              + {{ t('admin.channels.form.addTier') }}
            </button>
          </div>
          <div v-if="entry.intervals && entry.intervals.length > 0" class="mt-2 space-y-2">
            <IntervalRow
              v-for="(iv, idx) in entry.intervals"
              :key="idx"
              :interval="iv"
              :mode="entry.billing_mode"
              @update="updateInterval(idx, $event)"
              @remove="removeInterval(idx)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { BillingMode } from '@/api/admin/channels'
import channelsAPI from '@/api/admin/channels'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import IntervalRow from './IntervalRow.vue'
import ModelTagInput from './ModelTagInput.vue'
import type { IntervalFormEntry, PricingFormEntry } from './types'
import { getPlatformTagClass, perTokenToMTok } from './types'

const { t } = useI18n()

const props = defineProps<{
  entry: PricingFormEntry
  platform?: string
}>()

const emit = defineEmits<{
  update: [entry: PricingFormEntry]
  remove: []
}>()

const collapsed = ref(props.entry.models.length > 0)

const joinClassNames = (...classNames: Array<string | false | null | undefined>) => {
  return classNames.filter(Boolean).join(' ')
}

const billingModeOptions = computed(() => [
  { value: 'token', label: 'Token' },
  { value: 'per_request', label: t('admin.channels.billingMode.perRequest') },
  { value: 'image', label: t('admin.channels.billingMode.image') }
])

const billingModeLabel = computed(() => {
  const option = billingModeOptions.value.find(item => item.value === props.entry.billing_mode)
  return option ? option.label : props.entry.billing_mode
})

function getBillingModeChipClasses() {
  return joinClassNames(
    'theme-chip pricing-entry-card__billing-mode-chip shrink-0 text-xs font-semibold',
    props.entry.billing_mode === 'token'
      ? 'theme-chip--info'
      : props.entry.billing_mode === 'per_request'
        ? 'theme-chip--accent'
        : 'theme-chip--brand-purple'
  )
}

function emitField(field: keyof PricingFormEntry, value: string) {
  emit('update', { ...props.entry, [field]: value === '' ? null : value })
}

function addInterval() {
  const intervals = [...(props.entry.intervals || [])]
  intervals.push({
    min_tokens: 0,
    max_tokens: null,
    tier_label: '',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    per_request_price: null,
    sort_order: intervals.length
  })
  emit('update', { ...props.entry, intervals })
}

function addImageTier() {
  const intervals = [...(props.entry.intervals || [])]
  const labels = ['1K', '2K', '4K', 'HD']
  intervals.push({
    min_tokens: 0,
    max_tokens: null,
    tier_label: labels[intervals.length] || '',
    input_price: null,
    output_price: null,
    cache_write_price: null,
    cache_read_price: null,
    per_request_price: null,
    sort_order: intervals.length
  })
  emit('update', { ...props.entry, intervals })
}

function updateInterval(idx: number, updated: IntervalFormEntry) {
  const intervals = [...(props.entry.intervals || [])]
  intervals[idx] = updated
  emit('update', { ...props.entry, intervals })
}

function removeInterval(idx: number) {
  const intervals = [...(props.entry.intervals || [])]
  intervals.splice(idx, 1)
  emit('update', { ...props.entry, intervals })
}

async function onModelsUpdate(newModels: string[]) {
  const oldModels = props.entry.models
  emit('update', { ...props.entry, models: newModels })

  const addedModels = newModels.filter(model => !oldModels.includes(model))
  if (addedModels.length === 0) return

  const entry = props.entry
  const hasPrice =
    entry.input_price != null ||
    entry.output_price != null ||
    entry.cache_write_price != null ||
    entry.cache_read_price != null
  if (hasPrice) return

  try {
    const result = await channelsAPI.getModelDefaultPricing(addedModels[0])
    if (!result.found) return

    emit('update', {
      ...props.entry,
      models: newModels,
      input_price: perTokenToMTok(result.input_price ?? null),
      output_price: perTokenToMTok(result.output_price ?? null),
      cache_write_price: perTokenToMTok(result.cache_write_price ?? null),
      cache_read_price: perTokenToMTok(result.cache_read_price ?? null),
      image_output_price: perTokenToMTok(result.image_output_price ?? null)
    })
  } catch {
    // Ignore lookup failures; they shouldn't block editing.
  }
}
</script>

<style scoped>
.pricing-entry-card__billing-mode-chip {
  padding: calc(var(--theme-button-padding-y) * 0.45) calc(var(--theme-button-padding-x) * 0.6);
  border-radius: 9999px;
}
</style>

<style scoped>
.pricing-entry-card {
  border: 1px solid var(--theme-card-border);
  border-radius: calc(var(--theme-button-radius) + 2px);
  background: color-mix(in srgb, var(--theme-surface-soft) 74%, var(--theme-surface));
  padding: 0.75rem;
}

.pricing-entry-card__summary {
  display: flex;
  cursor: pointer;
  user-select: none;
  align-items: center;
  gap: 0.5rem;
}

.pricing-entry-card__summary-icon {
  flex-shrink: 0;
  color: var(--theme-page-muted);
  transition: transform 0.2s ease, color 0.2s ease;
}

.pricing-entry-card__summary:hover .pricing-entry-card__summary-icon {
  color: var(--theme-page-text);
}

.pricing-entry-card__muted,
.pricing-entry-card__sub-label,
.pricing-entry-card__label-unit {
  color: var(--theme-page-muted);
}

.pricing-entry-card__label {
  color: var(--theme-page-text);
  font-size: 0.75rem;
  font-weight: 600;
}

.pricing-entry-card__sub-label {
  font-size: 0.75rem;
}

.pricing-entry-card__required {
  color: rgb(var(--theme-danger-rgb));
}

.pricing-entry-card__remove-button,
.pricing-entry-card__link-button {
  transition: color 0.18s ease, background-color 0.18s ease;
}

.pricing-entry-card__remove-button {
  flex-shrink: 0;
  border-radius: calc(var(--theme-button-radius) - 4px);
  color: var(--theme-page-muted);
  padding: 0.25rem;
}

.pricing-entry-card__remove-button:hover,
.pricing-entry-card__remove-button:focus-visible {
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, transparent);
  color: rgb(var(--theme-danger-rgb));
  outline: none;
}

.pricing-entry-card__link-button {
  color: var(--theme-accent);
  font-size: 0.75rem;
  font-weight: 600;
}

.pricing-entry-card__link-button:hover,
.pricing-entry-card__link-button:focus-visible {
  color: color-mix(in srgb, var(--theme-accent) 74%, var(--theme-accent-strong));
  outline: none;
}

.pricing-entry-card__empty-state {
  border: 1px dashed color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  border-radius: calc(var(--theme-button-radius) + 2px);
  color: var(--theme-page-muted);
  font-size: 0.75rem;
  padding: 0.75rem;
  text-align: center;
}

.collapsible-content {
  display: grid;
  grid-template-rows: 1fr;
  transition: grid-template-rows 0.25s ease;
}

.collapsible-content--collapsed {
  grid-template-rows: 0fr;
}

.collapsible-inner {
  overflow: hidden;
}
</style>
