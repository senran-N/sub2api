<template>
  <div v-if="options.length > 0">
    <div class="mb-1.5 flex items-center gap-1">
      <label
        :id="selectLabelId"
        :for="selectId"
        class="group-copy-accounts-field__label text-sm font-medium"
      >
        {{ t('admin.groups.copyAccounts.title') }}
      </label>
      <div class="group relative inline-flex">
        <Icon
          name="questionCircle"
          size="sm"
          :stroke-width="2"
          class="group-copy-accounts-field__hint-icon cursor-help transition-colors"
        />
        <div class="pointer-events-none absolute bottom-full left-0 z-50 mb-2 hidden w-72 group-hover:block group-hover:pointer-events-auto">
          <div class="group-copy-accounts-field__tooltip">
            <p class="group-copy-accounts-field__tooltip-text text-xs leading-relaxed">
              {{ tooltipText }}
            </p>
            <div class="group-copy-accounts-field__tooltip-arrow absolute -bottom-1.5 left-3 h-3 w-3 rotate-45"></div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="selectedGroupIds.length > 0" class="mb-2 flex flex-wrap gap-1.5">
      <span
        v-for="groupId in selectedGroupIds"
        :key="groupId"
        class="theme-chip theme-chip--accent theme-chip--compact group-copy-accounts-field__chip"
      >
        {{ getOptionLabel(groupId) }}
        <button
          type="button"
          class="group-copy-accounts-field__chip-remove ml-0.5"
          @click="$emit('remove-group', groupId)"
        >
          <Icon name="x" size="xs" />
        </button>
      </span>
    </div>

    <select :id="selectId" name="copy_accounts_from_group_ids" class="input" @change="handleSelectChange">
      <option value="">{{ t('admin.groups.copyAccounts.selectPlaceholder') }}</option>
      <option
        v-for="option in options"
        :key="option.value"
        :value="option.value"
        :disabled="selectedGroupIds.includes(option.value)"
      >
        {{ option.label }}
      </option>
    </select>
    <p class="input-hint">{{ hintText }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { NumberSelectOption } from './groupsForm'

let copyAccountsFieldIdCounter = 0

const props = defineProps<{
  options: NumberSelectOption[]
  selectedGroupIds: number[]
  tooltipText: string
  hintText: string
}>()

const emit = defineEmits<{
  'add-group': [groupId: number]
  'remove-group': [groupId: number]
}>()

const { t } = useI18n()
const fieldIdPrefix = `group-copy-accounts-field-${++copyAccountsFieldIdCounter}`
const selectId = computed(() => `${fieldIdPrefix}-select`)
const selectLabelId = computed(() => `${fieldIdPrefix}-label`)

function getOptionLabel(groupId: number): string {
  return props.options.find((option) => option.value === groupId)?.label || `#${groupId}`
}

function handleSelectChange(event: Event): void {
  const select = event.target as HTMLSelectElement
  const groupId = Number(select.value)
  if (groupId > 0) {
    emit('add-group', groupId)
  }
  select.value = ''
}
</script>

<style scoped>
.group-copy-accounts-field__label {
  color: var(--theme-page-text);
}

.group-copy-accounts-field__hint-icon {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.group:hover .group-copy-accounts-field__hint-icon {
  color: var(--theme-accent);
}

.group-copy-accounts-field__tooltip {
  border: 1px solid color-mix(in srgb, var(--theme-dropdown-border) 88%, transparent);
  background: var(--theme-dropdown-bg);
  box-shadow: var(--theme-dropdown-shadow);
  border-radius: var(--theme-surface-radius);
  padding: var(--theme-group-selector-padding);
}

.group-copy-accounts-field__tooltip-text {
  color: var(--theme-dropdown-text);
}

.group-copy-accounts-field__tooltip-arrow {
  background: var(--theme-dropdown-bg);
}

.group-copy-accounts-field__chip {
  gap: 0.25rem;
}

.group-copy-accounts-field__chip-remove {
  color: inherit;
  transition: color 0.2s ease;
}

.group-copy-accounts-field__chip-remove:hover {
  color: color-mix(in srgb, var(--theme-accent-strong) 88%, var(--theme-page-text));
}
</style>
