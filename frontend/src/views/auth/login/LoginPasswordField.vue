<template>
  <div>
    <label for="password" class="input-label">
      {{ t('auth.passwordLabel') }}
    </label>
    <div class="relative">
      <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
        <Icon name="lock" size="md" class="text-gray-400 dark:text-dark-500" />
      </div>
      <input
        id="password"
        v-model="model"
        :type="showPassword ? 'text' : 'password'"
        required
        autocomplete="current-password"
        :disabled="disabled"
        class="input pl-11 pr-11"
        :class="{ 'input-error': error }"
        :placeholder="t('auth.passwordPlaceholder')"
      />
      <button
        type="button"
        class="absolute inset-y-0 right-0 flex items-center pr-3.5 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-dark-300"
        @click="showPassword = !showPassword"
      >
        <Icon v-if="showPassword" name="eyeOff" size="md" />
        <Icon v-else name="eye" size="md" />
      </button>
    </div>
    <div class="mt-1 flex items-center justify-between">
      <p v-if="error" class="input-error-text">
        {{ error }}
      </p>
      <span v-else></span>
      <router-link
        v-if="showForgotPassword"
        to="/forgot-password"
        class="text-sm font-medium text-primary-600 transition-colors hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300"
      >
        {{ t('auth.forgotPassword') }}
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  disabled: boolean
  error: string
  modelValue: string
  showForgotPassword: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t } = useI18n()
const showPassword = ref(false)

const model = computed({
  get: () => props.modelValue,
  set: (value: string) => emit('update:modelValue', value)
})
</script>
