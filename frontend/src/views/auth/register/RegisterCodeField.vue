<template>
  <div>
    <label :for="id" class="input-label">
      {{ label }}
      <span
        v-if="optionalLabel"
        class="register-code-field__optional ml-1 text-xs font-normal"
      >
        ({{ optionalLabel }})
      </span>
    </label>
    <div class="relative">
      <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
        <Icon
          :name="iconName"
          size="md"
          :class="valid ? 'register-code-field__icon register-code-field__icon--success' : 'register-code-field__icon'"
        />
      </div>
      <input
        :id="id"
        v-model="model"
        type="text"
        :disabled="disabled"
        class="input pl-11 pr-10"
        :class="inputClass"
        :placeholder="placeholder"
        @input="$emit('input', model)"
      />
      <div
        v-if="validating"
        class="absolute inset-y-0 right-0 flex items-center pr-3.5"
      >
        <svg class="register-code-field__spinner h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle
            class="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            stroke-width="4"
          ></circle>
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      </div>
      <div
        v-else-if="valid"
        class="absolute inset-y-0 right-0 flex items-center pr-3.5"
      >
        <Icon :name="successIconName" size="md" class="register-code-field__icon register-code-field__icon--success" />
      </div>
      <div
        v-else-if="hasError"
        class="absolute inset-y-0 right-0 flex items-center pr-3.5"
      >
        <Icon name="exclamationCircle" size="md" class="register-code-field__icon register-code-field__icon--error" />
      </div>
    </div>
    <transition name="fade">
      <div
        v-if="valid && successText"
        class="register-code-field__success mt-2 flex items-center gap-2"
      >
        <Icon :name="successIconName" size="sm" class="register-code-field__icon register-code-field__icon--success" />
        <span class="register-code-field__success-text text-sm">
          {{ successText }}
        </span>
      </div>
      <p v-else-if="errorText" class="input-error-text">
        {{ errorText }}
      </p>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import Icon from '@/components/icons/Icon.vue'

type IconName = InstanceType<typeof Icon>['$props']['name']

const props = withDefaults(
  defineProps<{
    disabled: boolean
    errorText?: string
    iconName: IconName
    id: string
    invalid: boolean
    label: string
    modelValue: string
    optionalLabel?: string
    placeholder: string
    successIconName?: IconName
    successText?: string
    valid: boolean
    validating: boolean
  }>(),
  {
    errorText: '',
    optionalLabel: '',
    successIconName: 'checkCircle',
    successText: ''
  }
)

const emit = defineEmits<{
  input: [value: string]
  'update:modelValue': [value: string]
}>()

const model = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value)
})

const hasError = computed(() => props.invalid || Boolean(props.errorText))

const inputClass = computed(() => ({
  'register-code-field__input--valid': props.valid,
  'register-code-field__input--error': hasError.value
}))
</script>

<style scoped>
.register-code-field__optional,
.register-code-field__icon,
.register-code-field__spinner {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.register-code-field__icon--success {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}

.register-code-field__icon--error {
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.register-code-field__input--valid {
  border-color: rgb(var(--theme-success-rgb));
}

.register-code-field__input--valid:focus {
  border-color: rgb(var(--theme-success-rgb));
  box-shadow: 0 0 0 3px color-mix(in srgb, rgb(var(--theme-success-rgb)) 18%, transparent);
}

.register-code-field__input--error {
  border-color: rgb(var(--theme-danger-rgb));
}

.register-code-field__input--error:focus {
  border-color: rgb(var(--theme-danger-rgb));
  box-shadow: 0 0 0 3px color-mix(in srgb, rgb(var(--theme-danger-rgb)) 18%, transparent);
}

.register-code-field__success {
  padding: var(--theme-register-feedback-padding-y) var(--theme-register-feedback-padding-x);
  border-radius: var(--theme-register-feedback-radius);
  background: color-mix(in srgb, rgb(var(--theme-success-rgb)) 10%, var(--theme-surface));
}

.register-code-field__success-text {
  color: color-mix(in srgb, rgb(var(--theme-success-rgb)) 84%, var(--theme-page-text));
}
</style>
