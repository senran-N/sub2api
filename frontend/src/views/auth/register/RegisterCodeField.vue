<template>
  <div>
    <label :for="id" class="input-label">
      {{ label }}
      <span
        v-if="optionalLabel"
        class="ml-1 text-xs font-normal text-gray-400 dark:text-dark-500"
      >
        ({{ optionalLabel }})
      </span>
    </label>
    <div class="relative">
      <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
        <Icon
          :name="iconName"
          size="md"
          :class="valid ? 'text-green-500' : 'text-gray-400 dark:text-dark-500'"
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
        <svg class="h-4 w-4 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24">
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
        <Icon :name="successIconName" size="md" class="text-green-500" />
      </div>
      <div
        v-else-if="hasError"
        class="absolute inset-y-0 right-0 flex items-center pr-3.5"
      >
        <Icon name="exclamationCircle" size="md" class="text-red-500" />
      </div>
    </div>
    <transition name="fade">
      <div
        v-if="valid && successText"
        class="mt-2 flex items-center gap-2 rounded-lg bg-green-50 px-3 py-2 dark:bg-green-900/20"
      >
        <Icon :name="successIconName" size="sm" class="text-green-600 dark:text-green-400" />
        <span class="text-sm text-green-700 dark:text-green-400">
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
  'border-green-500 focus:border-green-500 focus:ring-green-500': props.valid,
  'border-red-500 focus:border-red-500 focus:ring-red-500': hasError.value
}))
</script>
