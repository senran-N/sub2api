<template>
  <div class="totp-login-modal fixed inset-0 z-50 overflow-y-auto">
    <div class="totp-login-modal__container flex min-h-full items-center justify-center">
      <div class="totp-login-modal__backdrop fixed inset-0 transition-opacity"></div>

      <div class="totp-login-modal__panel relative w-full transform transition-all">
        <!-- Header -->
        <div class="mb-6 text-center">
          <div class="totp-login-modal__icon-shell mx-auto flex h-12 w-12 items-center justify-center rounded-full">
            <svg class="totp-login-modal__icon h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
            </svg>
          </div>
          <h3 class="totp-login-modal__title mt-4">
            {{ t('profile.totp.loginTitle') }}
          </h3>
          <p class="totp-login-modal__description mt-2 text-sm">
            {{ t('profile.totp.loginHint') }}
          </p>
          <p v-if="userEmailMasked" class="totp-login-modal__email mt-1 text-sm font-medium">
            {{ userEmailMasked }}
          </p>
        </div>

        <!-- Code Input -->
        <div class="mb-6">
          <div class="flex justify-center gap-2">
            <input
              v-for="(_, index) in 6"
              :key="index"
              :ref="(el) => setInputRef(el, index)"
              type="text"
              maxlength="1"
              inputmode="numeric"
              pattern="[0-9]"
              class="totp-login-modal__digit h-12 w-10 text-center text-lg font-semibold"
              :disabled="verifying"
              @input="handleCodeInput($event, index)"
              @keydown="handleKeydown($event, index)"
              @paste="handlePaste"
            />
          </div>
          <!-- Loading indicator -->
          <div v-if="verifying" class="totp-login-modal__description mt-3 flex items-center justify-center gap-2 text-sm">
            <div class="totp-login-modal__spinner h-4 w-4 animate-spin rounded-full border-b-2"></div>
            {{ t('common.verifying') }}
          </div>
        </div>

        <!-- Error -->
        <div v-if="error" class="totp-login-modal__error mb-4 text-sm">
          {{ error }}
        </div>

        <!-- Cancel button only -->
        <button
          type="button"
          class="btn btn-secondary w-full"
          :disabled="verifying"
          @click="$emit('cancel')"
        >
          {{ t('common.cancel') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

defineProps<{
  tempToken: string
  userEmailMasked?: string
}>()

const emit = defineEmits<{
  verify: [code: string]
  cancel: []
}>()

const { t } = useI18n()

const verifying = ref(false)
const error = ref('')
const code = ref<string[]>(['', '', '', '', '', ''])
const inputRefs = ref<(HTMLInputElement | null)[]>([])

// Watch for code changes and auto-submit when 6 digits are entered
watch(
  () => code.value.join(''),
  (newCode) => {
    if (newCode.length === 6 && !verifying.value) {
      emit('verify', newCode)
    }
  }
)

defineExpose({
  setVerifying: (value: boolean) => { verifying.value = value },
  setError: (message: string) => {
    error.value = message
    code.value = ['', '', '', '', '', '']
    // Clear input DOM values
    inputRefs.value.forEach(input => {
      if (input) input.value = ''
    })
    nextTick(() => {
      inputRefs.value[0]?.focus()
    })
  }
})

const setInputRef = (el: any, index: number) => {
  inputRefs.value[index] = el as HTMLInputElement | null
}

const handleCodeInput = (event: Event, index: number) => {
  const input = event.target as HTMLInputElement
  const value = input.value.replace(/[^0-9]/g, '')
  code.value[index] = value

  if (value && index < 5) {
    nextTick(() => {
      inputRefs.value[index + 1]?.focus()
    })
  }
}

const handleKeydown = (event: KeyboardEvent, index: number) => {
  if (event.key === 'Backspace') {
    const input = event.target as HTMLInputElement
    // If current cell is empty and not the first, move to previous cell
    if (!input.value && index > 0) {
      event.preventDefault()
      inputRefs.value[index - 1]?.focus()
    }
    // Otherwise, let the browser handle the backspace naturally
    // The input event will sync code.value via handleCodeInput
  }
}

const handlePaste = (event: ClipboardEvent) => {
  event.preventDefault()
  const pastedData = event.clipboardData?.getData('text') || ''
  const digits = pastedData.replace(/[^0-9]/g, '').slice(0, 6).split('')

  // Update both the ref and the input elements
  digits.forEach((digit, index) => {
    code.value[index] = digit
    if (inputRefs.value[index]) {
      inputRefs.value[index]!.value = digit
    }
  })

  // Clear remaining inputs if pasted less than 6 digits
  for (let i = digits.length; i < 6; i++) {
    code.value[i] = ''
    if (inputRefs.value[i]) {
      inputRefs.value[i]!.value = ''
    }
  }

  const focusIndex = Math.min(digits.length, 5)
  nextTick(() => {
    inputRefs.value[focusIndex]?.focus()
  })
}

onMounted(() => {
  nextTick(() => {
    inputRefs.value[0]?.focus()
  })
})
</script>

<style scoped>
.totp-login-modal__container {
  padding: var(--theme-settings-card-panel-padding);
}

.totp-login-modal__backdrop {
  background: var(--theme-overlay-strong);
}

.totp-login-modal__panel {
  max-width: var(--theme-dialog-width-narrow);
  padding: var(--theme-auth-card-padding);
  border-radius: var(--theme-auth-card-radius);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
}

.totp-login-modal__icon-shell {
  background: color-mix(in srgb, var(--theme-accent-soft) 86%, var(--theme-surface));
}

.totp-login-modal__icon {
  color: var(--theme-accent);
}

.totp-login-modal__title,
.totp-login-modal__email {
  color: var(--theme-page-text);
}

.totp-login-modal__title {
  font-family: var(--theme-auth-section-title-font);
  font-size: var(--theme-auth-section-title-size);
  font-weight: 700;
  letter-spacing: var(--theme-auth-section-title-letter-spacing);
}

.totp-login-modal__description {
  color: var(--theme-page-muted);
}

.totp-login-modal__digit {
  border-radius: var(--theme-button-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: var(--theme-input-bg);
  color: var(--theme-input-text);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.totp-login-modal__digit:focus {
  outline: none;
  border-color: var(--theme-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent-soft) 88%, transparent);
}

.totp-login-modal__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 64%, transparent);
  border-bottom-color: var(--theme-accent);
}

.totp-login-modal__error {
  padding: var(--theme-auth-callback-feedback-padding);
  border-radius: var(--theme-auth-feedback-radius);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}
</style>
