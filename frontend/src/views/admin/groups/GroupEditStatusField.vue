<template>
  <div>
    <label class="input-label">{{ t('admin.groups.form.status') }}</label>
    <Select v-model="statusModel" :options="statusOptions" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Select, { type SelectOption } from '@/components/common/Select.vue'

const props = defineProps<{
  modelValue: string
  statusOptions: SelectOption[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t } = useI18n()

const statusModel = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})
</script>
