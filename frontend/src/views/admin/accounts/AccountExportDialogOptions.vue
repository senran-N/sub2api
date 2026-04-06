<template>
  <label class="theme-text-default flex items-center gap-2 text-sm">
    <input
      class="theme-checkbox"
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
