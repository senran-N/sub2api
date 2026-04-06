<template>
  <div>
    <label for="password" class="input-label">
      {{ t('auth.passwordLabel') }}
    </label>
    <div class="relative">
      <div class="register-password-field__affix pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
        <Icon name="lock" size="md" />
      </div>
      <input
        id="password"
        v-model="model"
        :type="showPassword ? 'text' : 'password'"
        required
        autocomplete="new-password"
        :disabled="disabled"
        class="input pl-11 pr-11"
        :class="{ 'input-error': error }"
        :placeholder="t('auth.createPasswordPlaceholder')"
      />
      <button
        type="button"
        class="register-password-field__toggle absolute inset-y-0 right-0 flex items-center pr-3.5"
        @click="showPassword = !showPassword"
      >
        <Icon v-if="showPassword" name="eyeOff" size="md" />
        <Icon v-else name="eye" size="md" />
      </button>
    </div>
    <p v-if="error" class="input-error-text">
      {{ error }}
    </p>
    <p v-else class="input-hint">
      {{ t('auth.passwordHint') }}
    </p>
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

<style scoped>
.register-password-field__affix,
.register-password-field__toggle {
  color: color-mix(in srgb, var(--theme-page-muted) 72%, transparent);
}

.register-password-field__toggle {
  transition: color 0.2s ease;
}

.register-password-field__toggle:hover {
  color: var(--theme-page-text);
}
</style>
