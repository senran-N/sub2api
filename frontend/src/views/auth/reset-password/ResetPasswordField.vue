<template>
  <div>
    <label :for="id" class="input-label">
      {{ label }}
    </label>
    <div class="relative">
      <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
        <Icon name="lock" size="md" class="text-gray-400 dark:text-dark-500" />
      </div>
      <input
        :id="id"
        v-model="model"
        :type="visible ? 'text' : 'password'"
        required
        autocomplete="new-password"
        :disabled="disabled"
        class="input pl-11 pr-11"
        :class="{ 'input-error': error }"
        :placeholder="placeholder"
      />
      <button
        type="button"
        class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-dark-300"
        @click="visible = !visible"
      >
        <Icon v-if="visible" name="eyeOff" size="md" />
        <Icon v-else name="eye" size="md" />
      </button>
    </div>
    <p v-if="error" class="input-error-text">
      {{ error }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  disabled: boolean
  error: string
  id: string
  label: string
  modelValue: string
  placeholder: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const visible = ref(false)

const model = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value)
})
</script>
