<template>
  <div class="mb-6 flex border-b border-gray-200 dark:border-dark-600">
    <button
      type="button"
      :class="tabClass('standard')"
      @click="emit('update:modelValue', 'standard')"
    >
      <Icon name="plus" size="sm" class="mr-1.5 inline" />
      {{ t('admin.proxies.standardAdd') }}
    </button>
    <button
      type="button"
      :class="tabClass('batch')"
      @click="emit('update:modelValue', 'batch')"
    >
      <Icon name="menu" size="sm" class="mr-1.5 inline" />
      {{ t('admin.proxies.batchAdd') }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  modelValue: 'standard' | 'batch'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: 'standard' | 'batch']
}>()

const { t } = useI18n()

const tabClass = (mode: 'standard' | 'batch') => [
  '-mb-px border-b-2 px-4 py-2 text-sm font-medium transition-colors',
  props.modelValue === mode
    ? 'border-primary-500 text-primary-600 dark:text-primary-400'
    : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
]
</script>
