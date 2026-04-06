<template>
  <div>
    <label :for="id" class="input-label">
      {{ label }}
    </label>
    <div class="relative">
      <div class="reset-password-field__affix pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
        <Icon name="lock" size="md" />
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
        class="reset-password-field__toggle absolute inset-y-0 right-0 flex items-center pr-3.5"
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

<style scoped>
.reset-password-field__affix,
.reset-password-field__toggle {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.reset-password-field__toggle {
  transition: color 0.2s ease;
}

.reset-password-field__toggle:hover {
  color: var(--theme-page-text);
}
</style>
