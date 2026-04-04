<template>
  <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
    <input
      class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
      type="checkbox"
      :checked="modelValue"
      @change="handleChange"
    />
    <span>{{ t('admin.accounts.dataExportIncludeProxies') }}</span>
  </label>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'

defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const { t } = useI18n()

const handleChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.checked)
}
</script>
