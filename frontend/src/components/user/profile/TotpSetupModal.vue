<template>
  <div class="totp-setup-modal fixed inset-0 z-50 overflow-y-auto" @click.self="$emit('close')">
    <div class="totp-setup-modal__viewport flex min-h-full items-center justify-center">
      <div class="totp-setup-modal__backdrop fixed inset-0 transition-opacity" @click="$emit('close')"></div>

      <div class="totp-setup-modal__panel relative w-full transform transition-all">
        <!-- Header -->
        <div class="mb-6 text-center">
          <h3 class="totp-setup-modal__title text-xl font-semibold">
            {{ t('profile.totp.setupTitle') }}
          </h3>
          <p class="totp-setup-modal__description mt-2 text-sm">
            {{ stepDescription }}
          </p>
        </div>

        <!-- Step 0: Identity Verification -->
        <div v-if="step === 0" class="space-y-6">
          <!-- Loading verification method -->
          <div v-if="methodLoading" class="totp-setup-modal__loading flex items-center justify-center">
            <div class="totp-setup-modal__spinner h-8 w-8 animate-spin rounded-full border-b-2"></div>
          </div>

          <template v-else>
            <!-- Email verification -->
            <div v-if="verificationMethod === 'email'" class="space-y-4">
              <div>
                <label class="input-label">{{ t('profile.totp.emailCode') }}</label>
                <div class="flex gap-2">
                  <input
                    v-model="verifyForm.emailCode"
                    type="text"
                    maxlength="6"
                    inputmode="numeric"
                    class="input flex-1"
                    :placeholder="t('profile.totp.enterEmailCode')"
                  />
                  <button
                    type="button"
                    class="btn btn-secondary whitespace-nowrap"
                    :disabled="sendingCode || codeCooldown > 0"
                    @click="handleSendCode"
                  >
                    {{ codeCooldown > 0 ? `${codeCooldown}s` : (sendingCode ? t('common.sending') : t('profile.totp.sendCode')) }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Password verification -->
            <div v-else class="space-y-4">
              <div>
                <label class="input-label">{{ t('profile.currentPassword') }}</label>
                <input
                  v-model="verifyForm.password"
                  type="password"
                  autocomplete="current-password"
                  class="input"
                  :placeholder="t('profile.totp.enterPassword')"
                />
              </div>
            </div>

            <div v-if="verifyError" class="totp-setup-modal__error text-sm">
              {{ verifyError }}
            </div>

            <div class="flex justify-end gap-3 pt-4">
              <button type="button" class="btn btn-secondary" @click="$emit('close')">
                {{ t('common.cancel') }}
              </button>
              <button
                type="button"
                class="btn btn-primary"
                :disabled="!canProceedFromVerify || setupLoading"
                @click="handleVerifyAndSetup"
              >
                {{ setupLoading ? t('common.loading') : t('common.next') }}
              </button>
            </div>
          </template>
        </div>

        <!-- Step 1: Show QR Code -->
        <div v-if="step === 1" class="space-y-6">
          <!-- QR Code and Secret -->
          <template v-if="setupData">
            <div class="flex justify-center">
              <div class="totp-setup-modal__qr-shell">
                <img v-if="qrCodeDataUrl" :src="qrCodeDataUrl" alt="QR Code" class="h-48 w-48" />
                <div
                  v-else
                  class="totp-setup-modal__qr-loading flex h-48 w-48 items-center justify-center"
                >
                  <div class="totp-setup-modal__spinner h-8 w-8 animate-spin rounded-full border-b-2"></div>
                </div>
              </div>
            </div>

            <div class="text-center">
              <p class="totp-setup-modal__description mb-2 text-sm">
                {{ t('profile.totp.manualEntry') }}
              </p>
              <div class="flex items-center justify-center gap-2">
                <code class="totp-setup-modal__secret font-mono text-sm">
                  {{ setupData.secret }}
                </code>
                <button
                  type="button"
                  class="totp-setup-modal__copy-button"
                  @click="copySecret"
                >
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
                  </svg>
                </button>
              </div>
            </div>
          </template>

          <div class="flex justify-end gap-3 pt-4">
            <button type="button" class="btn btn-secondary" @click="$emit('close')">
              {{ t('common.cancel') }}
            </button>
            <button
              type="button"
              class="btn btn-primary"
              :disabled="!setupData"
              @click="step = 2"
            >
              {{ t('common.next') }}
            </button>
          </div>
        </div>

        <!-- Step 2: Verify Code -->
        <div v-if="step === 2" class="space-y-6">
          <form @submit.prevent="handleVerify">
            <div class="mb-6">
              <label class="input-label text-center block mb-3">
                {{ t('profile.totp.enterCode') }}
              </label>
              <div class="flex justify-center gap-2">
                <input
                  v-for="(_, index) in 6"
                  :key="index"
                  :ref="(el) => setInputRef(el, index)"
                  type="text"
                  maxlength="1"
                  inputmode="numeric"
                  pattern="[0-9]"
                  class="totp-setup-modal__digit h-12 w-10 text-center text-lg font-semibold"
                  @input="handleCodeInput($event, index)"
                  @keydown="handleKeydown($event, index)"
                  @paste="handlePaste"
                />
              </div>
            </div>

            <div v-if="error" class="totp-setup-modal__error mb-4 text-sm">
              {{ error }}
            </div>

            <div class="flex justify-end gap-3">
              <button type="button" class="btn btn-secondary" @click="step = 1">
                {{ t('common.back') }}
              </button>
              <button
                type="submit"
                class="btn btn-primary"
                :disabled="verifying || code.join('').length !== 6"
              >
                {{ verifying ? t('common.verifying') : t('profile.totp.verify') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { totpAPI } from '@/api'
import type { TotpSetupResponse } from '@/types'
import { useDocumentThemeVersion } from '@/composables/useDocumentThemeVersion'
import { readThemeCssVariable } from '@/utils/themeStyles'

const emit = defineEmits<{
  close: []
  success: []
}>()

const { t } = useI18n()
const appStore = useAppStore()
const themeVersion = useDocumentThemeVersion()

// Step: 0 = verify identity, 1 = QR code, 2 = verify TOTP code
const step = ref(0)
const methodLoading = ref(true)
const verificationMethod = ref<'email' | 'password'>('password')
const verifyForm = ref({ emailCode: '', password: '' })
const verifyError = ref('')
const sendingCode = ref(false)
const codeCooldown = ref(0)
const cooldownTimer = ref<ReturnType<typeof setInterval> | null>(null)

const setupLoading = ref(false)
const setupData = ref<TotpSetupResponse | null>(null)
const verifying = ref(false)
const error = ref('')
const code = ref<string[]>(['', '', '', '', '', ''])
const inputRefs = ref<(HTMLInputElement | null)[]>([])
const qrCodeDataUrl = ref('')

type QrCodeModule = typeof import('qrcode')

type ErrorMessageLike = {
  response?: {
    data?: {
      message?: string
    }
  }
}

let qrCodeModulePromise: Promise<QrCodeModule> | null = null

const getQrCodeModule = (): Promise<QrCodeModule> => {
  if (!qrCodeModulePromise) {
    qrCodeModulePromise = import('qrcode')
  }

  return qrCodeModulePromise
}

const stepDescription = computed(() => {
  switch (step.value) {
    case 0:
      return verificationMethod.value === 'email'
        ? t('profile.totp.verifyEmailFirst')
        : t('profile.totp.verifyPasswordFirst')
    case 1:
      return t('profile.totp.setupStep1')
    case 2:
      return t('profile.totp.setupStep2')
    default:
      return ''
  }
})

const canProceedFromVerify = computed(() => {
  if (verificationMethod.value === 'email') {
    return verifyForm.value.emailCode.length === 6
  }
  return verifyForm.value.password.length > 0
})

// Generate QR code as base64 when setupData changes
watch(
  [() => setupData.value?.qr_code_url, themeVersion],
  async ([url], _previousUrl, onCleanup) => {
    if (!url) {
      qrCodeDataUrl.value = ''
      return
    }

    let isActive = true
    onCleanup(() => {
      isActive = false
    })

    try {
      qrCodeDataUrl.value = ''
      const QRCode = await getQrCodeModule()
      if (!isActive) {
        return
      }

      const qrDarkColor = readThemeCssVariable('--theme-page-text')
      const qrLightColor = readThemeCssVariable('--theme-surface')
      const nextQrCodeDataUrl = await QRCode.toDataURL(url, {
        width: 200,
        margin: 2,
        color: {
          ['dark']: qrDarkColor,
          ['light']: qrLightColor
        }
      })
      if (!isActive) {
        return
      }

      qrCodeDataUrl.value = nextQrCodeDataUrl
    } catch (err) {
      console.error('Failed to generate QR code:', err)
    }
  },
  { immediate: true }
)

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

const copySecret = async () => {
  if (setupData.value) {
    try {
      await navigator.clipboard.writeText(setupData.value.secret)
      appStore.showSuccess(t('common.copied'))
    } catch {
      appStore.showError(t('common.copyFailed'))
    }
  }
}

const getErrorMessage = (errorLike: unknown, fallbackMessage: string) => {
  const normalizedError = errorLike as ErrorMessageLike | null
  return normalizedError?.response?.data?.message || fallbackMessage
}

const loadVerificationMethod = async () => {
  methodLoading.value = true
  try {
    const method = await totpAPI.getVerificationMethod()
    verificationMethod.value = method.method
  } catch (errorLike) {
    appStore.showError(getErrorMessage(errorLike, t('common.error')))
    emit('close')
  } finally {
    methodLoading.value = false
  }
}

const handleSendCode = async () => {
  sendingCode.value = true
  try {
    await totpAPI.sendVerifyCode()
    appStore.showSuccess(t('profile.totp.codeSent'))
    // Start cooldown
    codeCooldown.value = 60
    if (cooldownTimer.value) {
      clearInterval(cooldownTimer.value)
      cooldownTimer.value = null
    }
    cooldownTimer.value = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) {
        if (cooldownTimer.value) {
          clearInterval(cooldownTimer.value)
          cooldownTimer.value = null
        }
      }
    }, 1000)
  } catch (errorLike) {
    appStore.showError(getErrorMessage(errorLike, t('profile.totp.sendCodeFailed')))
  } finally {
    sendingCode.value = false
  }
}

const handleVerifyAndSetup = async () => {
  setupLoading.value = true
  verifyError.value = ''

  try {
    const request = verificationMethod.value === 'email'
      ? { email_code: verifyForm.value.emailCode }
      : { password: verifyForm.value.password }

    setupData.value = await totpAPI.initiateSetup(request)
    step.value = 1
  } catch (errorLike) {
    verifyError.value = getErrorMessage(errorLike, t('profile.totp.setupFailed'))
  } finally {
    setupLoading.value = false
  }
}

const handleVerify = async () => {
  const totpCode = code.value.join('')
  if (totpCode.length !== 6 || !setupData.value) return

  verifying.value = true
  error.value = ''

  try {
    await totpAPI.enable({
      totp_code: totpCode,
      setup_token: setupData.value.setup_token
    })
    appStore.showSuccess(t('profile.totp.enableSuccess'))
    emit('success')
  } catch (errorLike) {
    error.value = getErrorMessage(errorLike, t('profile.totp.verifyFailed'))
    code.value = ['', '', '', '', '', '']
    nextTick(() => {
      inputRefs.value[0]?.focus()
    })
  } finally {
    verifying.value = false
  }
}

onMounted(() => {
  loadVerificationMethod()
})

onUnmounted(() => {
  if (cooldownTimer.value) {
    clearInterval(cooldownTimer.value)
    cooldownTimer.value = null
  }
})
</script>

<style scoped>
.totp-setup-modal__backdrop {
  background: var(--theme-overlay-strong);
}

.totp-setup-modal__panel {
  max-width: var(--theme-totp-modal-max-width);
  border-radius: var(--theme-totp-modal-radius);
  padding: var(--theme-totp-modal-padding);
  background: var(--theme-surface);
  box-shadow: var(--theme-card-shadow-hover);
}

.totp-setup-modal__viewport {
  padding: var(--theme-totp-modal-viewport-padding);
}

.totp-setup-modal__title {
  color: var(--theme-page-text);
}

.totp-setup-modal__description {
  color: var(--theme-page-muted);
}

.totp-setup-modal__spinner {
  border-color: color-mix(in srgb, var(--theme-card-border) 64%, transparent);
  border-bottom-color: var(--theme-accent);
}

.totp-setup-modal__loading {
  padding-block: var(--theme-totp-modal-loading-padding-y);
}

.totp-setup-modal__qr-loading {
  min-height: 12rem;
}

.totp-setup-modal__error {
  border-radius: var(--theme-totp-modal-error-radius);
  padding: var(--theme-totp-modal-error-padding);
  background: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 10%, var(--theme-surface));
  color: color-mix(in srgb, rgb(var(--theme-danger-rgb)) 84%, var(--theme-page-text));
}

.totp-setup-modal__qr-shell {
  border-radius: var(--theme-totp-modal-qr-radius);
  padding: var(--theme-totp-modal-qr-padding);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: var(--theme-surface);
}

.totp-setup-modal__secret {
  border-radius: var(--theme-totp-modal-secret-radius);
  padding: var(--theme-totp-modal-secret-padding-y) var(--theme-totp-modal-secret-padding-x);
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}

.totp-setup-modal__copy-button {
  border-radius: var(--theme-totp-modal-copy-radius);
  padding: var(--theme-totp-modal-copy-padding);
  color: var(--theme-page-muted);
  transition: background-color 0.2s ease, color 0.2s ease;
}

.totp-setup-modal__copy-button:hover {
  background: color-mix(in srgb, var(--theme-surface-soft) 88%, var(--theme-surface));
  color: var(--theme-page-text);
}

.totp-setup-modal__digit {
  border-radius: var(--theme-totp-modal-digit-radius);
  border: 1px solid color-mix(in srgb, var(--theme-card-border) 88%, transparent);
  background: var(--theme-input-bg);
  color: var(--theme-input-text);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.totp-setup-modal__digit:focus {
  outline: none;
  border-color: var(--theme-accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent-soft) 88%, transparent);
}
</style>
