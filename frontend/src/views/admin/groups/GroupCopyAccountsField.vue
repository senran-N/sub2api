<template>
  <div v-if="options.length > 0">
    <div class="mb-1.5 flex items-center gap-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('admin.groups.copyAccounts.title') }}
      </label>
      <div class="group relative inline-flex">
        <Icon
          name="questionCircle"
          size="sm"
          :stroke-width="2"
          class="cursor-help text-gray-400 transition-colors hover:text-primary-500 dark:text-gray-500 dark:hover:text-primary-400"
        />
        <div class="pointer-events-none absolute bottom-full left-0 z-50 mb-2 w-72 opacity-0 transition-all duration-200 group-hover:pointer-events-auto group-hover:opacity-100">
          <div class="rounded-lg bg-gray-900 p-3 text-white shadow-lg dark:bg-gray-800">
            <p class="text-xs leading-relaxed text-gray-300">
              {{ tooltipText }}
            </p>
            <div class="absolute -bottom-1.5 left-3 h-3 w-3 rotate-45 bg-gray-900 dark:bg-gray-800"></div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="selectedGroupIds.length > 0" class="mb-2 flex flex-wrap gap-1.5">
      <span
        v-for="groupId in selectedGroupIds"
        :key="groupId"
        class="inline-flex items-center gap-1 rounded-full bg-primary-100 px-2.5 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
      >
        {{ getOptionLabel(groupId) }}
        <button
          type="button"
          class="ml-0.5 text-primary-500 hover:text-primary-700 dark:hover:text-primary-200"
          @click="$emit('remove-group', groupId)"
        >
          <Icon name="x" size="xs" />
        </button>
      </span>
    </div>

    <select class="input" @change="handleSelectChange">
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
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { NumberSelectOption } from '../groupsForm'

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
